/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mongodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	dbaapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"
	dbutil "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	dbautil "kubedb.dev/apimachinery/client/clientset/versioned/typed/ops/v1alpha1/util"
	"kubedb.dev/enterprise/pkg/controller/lib"

	"github.com/google/go-cmp/cmp"
	cm_api "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	cm_util "kmodules.xyz/cert-manager-util/certmanager/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	v1 "kmodules.xyz/offshoot-api/api/v1"
)

type checkCertRevision struct {
	*mongoOpsReqController
	revisionMap              map[string]int
	updateCertificateRetries map[string]*Retries
}

func (c *mongoOpsReqController) newCheckCertVersion(revisionMap map[string]int) func() (bool, error) {
	opts := &checkCertRevision{
		mongoOpsReqController:    c,
		revisionMap:              revisionMap,
		updateCertificateRetries: make(map[string]*Retries),
	}

	for certName := range opts.revisionMap {
		opts.updateCertificateRetries[certName] = c.newRetries()
	}

	return opts.run
}

func (opts *checkCertRevision) run() (bool, error) {
	for certName, revision := range opts.revisionMap {
		cert, err := opts.CertManagerClient.CertmanagerV1().Certificates(opts.db.Namespace).Get(context.TODO(), certName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if *cert.Status.Revision == revision {
			return opts.updateCertificateRetries[certName].Wait(), fmt.Errorf("new certificate has not issued yet for certificate: %s", certName)
		}
		opts.updateCertificateRetries[certName].Initialize()

	}

	return false, nil
}

type updateTLS struct {
	*mongoOpsReqController
	dbCopy            *api.MongoDB
	stsNames          []string
	tlsUpdatedRetries map[string]*Retries
}

func (c *mongoOpsReqController) newUpdateTLS(dbCopy *api.MongoDB, stsNames []string) func() (bool, error) {
	opts := &updateTLS{
		mongoOpsReqController: c,
		dbCopy:                dbCopy,
		stsNames:              stsNames,
		tlsUpdatedRetries:     make(map[string]*Retries),
	}

	for _, stsName := range opts.stsNames {
		opts.tlsUpdatedRetries[stsName] = c.newRetries()
	}

	return opts.run
}

func (opts *updateTLS) run() (bool, error) {
	r, err := opts.newReconciler()
	if err != nil {
		return true, err
	}
	err = opts.manageTLS(opts.db)
	if err != nil {
		return true, err
	}
	err = r.Reconcile(opts.db, true)
	if err != nil {
		return true, err
	}

	stsList, err := opts.Client.AppsV1().StatefulSets(opts.db.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labels.Set(opts.db.OffshootSelectors()).String(),
	})
	if err != nil {
		return false, err
	}
	sslMode := string(api.SSLModeRequireSSL)
	if opts.req.Spec.TLS.Remove {
		sslMode = string(api.SSLModeDisabled)
	}

	opts.log.Info("/////////////// run starts //////////////////")
	for _, item := range stsList.Items {
		opts.log.Info("----", item.Name, item.Spec.Template.Spec.InitContainers)
	}

	for _, sts := range stsList.Items {
		for _, initContainer := range sts.Spec.Template.Spec.InitContainers {
			if initContainer.Name == api.MongoDBInitInstallContainerName {
				for _, env := range initContainer.Env {
					if env.Name == "SSL_MODE" {
						if env.Value != sslMode {
							return opts.tlsUpdatedRetries[sts.Name].Wait(), fmt.Errorf("TLS is not updated yet for statefulSet: %s", sts.Name)
						}
						opts.tlsUpdatedRetries[sts.Name].Initialize()
					}
				}
			}
		}
	}

	opts.log.Info("///////////////// in run(),sslMode is requireSSL now in sts ///////////////////")

	_, _, err = dbutil.CreateOrPatchMongoDB(context.TODO(), opts.DBClient.KubedbV1alpha2(), opts.db.ObjectMeta, func(db *api.MongoDB) *api.MongoDB {
		db.Spec.SSLMode = opts.dbCopy.Spec.SSLMode
		db.Spec.ClusterAuthMode = opts.dbCopy.Spec.ClusterAuthMode
		db.Spec.PodTemplate = opts.dbCopy.Spec.PodTemplate
		db.Spec.TLS = opts.dbCopy.Spec.TLS
		db.Spec.KeyFileSecret = opts.dbCopy.Spec.KeyFileSecret
		return db
	}, metav1.PatchOptions{})
	if err != nil {
		opts.log.Info("failed to patch mongoDB", "error", err)
		return false, err
	}

	opts.log.Info("///////////////// run() ends ///////////////////")
	return false, nil
}

func (c *mongoOpsReqController) ReconfigureTLS() error {
	log := c.log
	if c.req.Status.Phase == dbaapi.OpsRequestPhasePending {
		newOpsReq, err := dbautil.UpdateMongoDBOpsRequestStatus(
			context.TODO(),
			c.DBClient.OpsV1alpha1(),
			c.req.ObjectMeta,
			func(status *dbaapi.MongoDBOpsRequestStatus) (types.UID, *dbaapi.MongoDBOpsRequestStatus) {
				status.Phase = dbaapi.Progressing
				status.ObservedGeneration = c.req.Generation
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.NewCondition(string(dbaapi.OpsRequestTypeReconfigureTLSs), "MongoDB ops request is reconfiguring TLS", c.req.Generation))
				return c.req.UID, status
			}, metav1.UpdateOptions{})
		if err != nil {
			log.Error(err, "failed to update status")
			return err
		}
		c.req.Status = newOpsReq.Status
	}

	err := c.pauseMongoDB()
	if err != nil {
		c.log.Error(err, "failed to pause mongodb")
		return err
	}

	updateDB := false
	tlsConfig := c.req.Spec.TLS
	dbCopy := c.db.DeepCopy()
	if c.req.Spec.TLS.IssuerRef != nil {
		if dbCopy.Spec.TLS == nil {
			dbCopy.Spec.TLS = &kmapi.TLSConfig{
				IssuerRef: c.req.Spec.TLS.IssuerRef,
			}
		} else {
			dbCopy.Spec.TLS.IssuerRef = c.req.Spec.TLS.IssuerRef
		}
	}

	allSelector := labels.Set(c.db.OffshootSelectors()).String()
	stsList, err := c.Client.AppsV1().StatefulSets(c.db.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: allSelector,
	})
	if err != nil {
		log.Error(err, "failed to list Statefulsets", "LabelSelector", allSelector)
		return err
	}

	stsNames := make([]string, len(stsList.Items))
	for i, sts := range stsList.Items {
		stsNames[i] = sts.Name
	}

	var aliases []string

	if tlsConfig.Remove {
		newDBCopy := c.db.DeepCopy()

		if !kmapi.IsConditionTrue(c.req.Status.Conditions, dbaapi.TLSRemoved) {
			db, _, err := dbutil.CreateOrPatchMongoDB(context.TODO(), c.DBClient.KubedbV1alpha2(), c.db.ObjectMeta, func(db *api.MongoDB) *api.MongoDB {
				db.Spec.SSLMode = api.SSLModeDisabled
				db.Spec.ClusterAuthMode = api.ClusterAuthModeKeyFile
				db.Spec.PodTemplate = nil
				db.Spec.TLS = nil

				if db.Spec.ShardTopology != nil {
					db.Spec.ShardTopology.Mongos.PodTemplate = v1.PodTemplateSpec{}
					db.Spec.ShardTopology.Shard.PodTemplate = v1.PodTemplateSpec{}
					db.Spec.ShardTopology.ConfigServer.PodTemplate = v1.PodTemplateSpec{}
				}

				return db
			}, metav1.PatchOptions{})
			if err != nil {
				log.Error(err, "failed to patch mongoDB")
				return err
			}

			r, err := c.newReconciler()
			if err != nil {
				return err
			}
			err = r.Reconcile(db)
			if err != nil {
				return err
			}

			c.RunParallel(
				dbaapi.TLSRemoved,
				"Successfully Updated StatefulSets",
				c.newUpdateTLS(newDBCopy, stsNames))
			return nil
		}

		ok := c.restartPods()
		if !ok {
			return nil
		}

		_, _, err := dbutil.CreateOrPatchMongoDB(context.TODO(), c.DBClient.KubedbV1alpha2(), c.db.ObjectMeta, func(db *api.MongoDB) *api.MongoDB {
			db.Spec.SSLMode = api.SSLModeDisabled
			db.Spec.ClusterAuthMode = ""
			db.Spec.PodTemplate = nil
			db.Spec.KeyFileSecret = nil
			db.Spec.TLS = nil

			if db.Spec.ShardTopology != nil {
				db.Spec.ShardTopology.Mongos.PodTemplate = v1.PodTemplateSpec{}
				db.Spec.ShardTopology.Shard.PodTemplate = v1.PodTemplateSpec{}
				db.Spec.ShardTopology.ConfigServer.PodTemplate = v1.PodTemplateSpec{}
			}

			return db
		}, metav1.PatchOptions{})
		if err != nil {
			log.Error(err, "failed to patch mongoDB")
			return err
		}

		certList, err := c.CertManagerClient.CertmanagerV1().Certificates(c.db.Namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: labels.Set(c.db.OffshootSelectors()).String(),
		})
		if err != nil {
			return err
		}

		for _, cert := range certList.Items {
			err = c.Client.CoreV1().Secrets(c.db.Namespace).Delete(context.TODO(), cert.Spec.SecretName, metav1.DeleteOptions{})
			if err != nil {
				return err
			}

			err = c.CertManagerClient.CertmanagerV1().Certificates(c.db.Namespace).Delete(context.TODO(), cert.Name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	} else if tlsConfig.RotateCertificates || (!tlsConfig.RotateCertificates && c.db.Spec.TLS != nil && !kmapi.HasCondition(c.req.Status.Conditions, dbaapi.TLSAdded)) {
		for _, currentCert := range tlsConfig.Certificates {
			for j, prevCert := range dbCopy.Spec.TLS.Certificates {
				if currentCert.Alias == prevCert.Alias {
					dbCopy.Spec.TLS.Certificates[j] = currentCert
				}
			}
		}

		dbCopy.SetTLSDefaults()

		for _, prevCert := range c.db.Spec.TLS.Certificates {
			for _, currentCert := range dbCopy.Spec.TLS.Certificates {
				if currentCert.Alias == prevCert.Alias && !cmp.Equal(currentCert, prevCert) {
					aliases = append(aliases, currentCert.Alias)
				}
			}
		}

		if tlsConfig.IssuerRef != nil || !cmp.Equal(dbCopy.Spec.TLS, c.db.Spec.TLS) {
			updateDB = true
		}

		revisionMap := make(map[string]int)
		certList, err := c.CertManagerClient.CertmanagerV1().Certificates(c.db.Namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: labels.Set(c.db.OffshootSelectors()).String(),
		})
		if err != nil {
			return err
		}

		for _, cert := range certList.Items {
			if c.req.Spec.TLS.RotateCertificates {
				revisionMap[cert.Name] = *cert.Status.Revision
			} else {
				for _, alias := range aliases {
					if strings.HasSuffix(cert.Name, fmt.Sprintf("%s-cert", alias)) {
						revisionMap[cert.Name] = *cert.Status.Revision
					}
				}
			}
		}

		if tlsConfig.RotateCertificates {
			if !kmapi.IsConditionTrue(c.req.Status.Conditions, dbaapi.IssuingConditionUpdated) {
				for name := range revisionMap {
					exist := false
					for _, alias := range aliases {
						if strings.HasSuffix(name, fmt.Sprintf("%s-cert", alias)) {
							exist = true
							break
						}
					}
					if exist {
						continue
					}
					_, err := cm_util.UpdateCertificateStatus(
						context.TODO(),
						c.CertManagerClient.CertmanagerV1(),
						metav1.ObjectMeta{
							Name:      name,
							Namespace: c.db.Namespace,
						},
						func(status *cm_api.CertificateStatus) *cm_api.CertificateStatus {
							status.Conditions = getCertificateCondition(status.Conditions,
								cm_api.CertificateConditionIssuing,
								cmmeta.ConditionTrue,
								"RotateCertificate",
								"Rotating Certificate for KubeDB")

							return status
						}, metav1.UpdateOptions{},
					)
					if err != nil {
						return err
					}
				}

				err = c.UpdateMongoOpsReqConditions(dbaapi.IssuingConditionUpdated, "Successfully Added Issuing Condition in Certificates")
				if err != nil {
					log.Error(err, "failed to update condition")
					return err
				}
			}
		}

		if updateDB {
			_, err = c.certs.Get(c.db.Name + "-old-cert")
			if err != nil {
				clientCertSecret, err := c.Client.CoreV1().Secrets(c.db.Namespace).Get(context.TODO(), c.db.GetCertSecretName(api.MongoDBClientCert, ""), metav1.GetOptions{})
				if err != nil {
					return err
				}

				oldClientCertSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      c.db.Name + "-old-cert",
						Namespace: c.db.Namespace,
					},
					Data: clientCertSecret.Data,
				}
				core_util.EnsureOwnerReference(oldClientCertSecret, metav1.NewControllerRef(c.req, api.SchemeGroupVersion.WithKind(dbaapi.ResourceKindMongoDBOpsRequest)))

				_, err = c.Client.CoreV1().Secrets(c.db.Namespace).Create(context.TODO(), oldClientCertSecret, metav1.CreateOptions{})
				if err != nil && !errors.IsAlreadyExists(err) {
					return err
				}
			}
			db, _, err := dbutil.CreateOrPatchMongoDB(context.TODO(), c.DBClient.KubedbV1alpha2(), c.db.ObjectMeta, func(db *api.MongoDB) *api.MongoDB {
				db.Spec.TLS = dbCopy.Spec.TLS
				return db
			}, metav1.PatchOptions{})
			if err != nil {
				log.Error(err, "failed to patch mongoDB")
				return err
			}
			c.db = db

			r, err := c.newReconciler()
			if err != nil {
				return err
			}
			err = c.manageTLS(db)
			if err != nil {
				return err
			}
			log.Info("///////////*/*************/*********************///")
			err = r.Reconcile(db)
			if err != nil {
				return err
			}
		}

		if !kmapi.IsConditionTrue(c.req.Status.Conditions, dbaapi.CertificateIssuingSuccessful) {
			c.RunParallel(
				dbaapi.CertificateIssuingSuccessful,
				"Successfully Issued New Certificates",
				c.newCheckCertVersion(revisionMap))

			return nil
		}

		ok := c.restartPods()
		if !ok {
			return nil
		}

		err = c.Client.CoreV1().Secrets(c.db.Namespace).Delete(context.TODO(), c.db.Name+"-old-cert", metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
	} else if c.db.Spec.TLS == nil || kmapi.IsConditionTrue(c.req.Status.Conditions, dbaapi.TLSAdded) {
		// add as false TLSADDED
		newOpsReq, err := dbautil.UpdateMongoDBOpsRequestStatus(
			context.TODO(),
			c.DBClient.OpsV1alpha1(),
			c.req.ObjectMeta,
			func(status *dbaapi.MongoDBOpsRequestStatus) (types.UID, *dbaapi.MongoDBOpsRequestStatus) {
				status.Phase = dbaapi.Progressing
				status.ObservedGeneration = c.req.Generation
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.NewCondition(dbaapi.TLSAdded, "TLSAdded is being set to be false", c.req.Generation, false))
				return c.req.UID, status
			}, metav1.UpdateOptions{})
		if err != nil {
			log.Error(err, "failed to update status")
			return err
		}
		c.req.Status = newOpsReq.Status

		log.Info("++++++++++++++++++++++++++ Start of else if  +++++++++++++++++++++++++++")
		if c.req.Spec.TLS.Certificates != nil {
			if dbCopy.Spec.TLS == nil {
				dbCopy.Spec.TLS = &kmapi.TLSConfig{
					Certificates: c.req.Spec.TLS.Certificates,
				}
			} else {
				dbCopy.Spec.TLS.Certificates = c.req.Spec.TLS.Certificates
			}
		}
		newDBCopy := c.db.DeepCopy()
		if !kmapi.IsConditionTrue(c.req.Status.Conditions, dbaapi.TLSAdded) {
			db, _, err := dbutil.CreateOrPatchMongoDB(context.TODO(), c.DBClient.KubedbV1alpha2(), c.db.ObjectMeta, func(db *api.MongoDB) *api.MongoDB {
				db.Spec.SSLMode = api.SSLModeRequireSSL
				db.Spec.ClusterAuthMode = api.ClusterAuthModeX509
				db.Spec.PodTemplate = nil
				db.Spec.KeyFileSecret = nil
				db.Spec.TLS = dbCopy.Spec.TLS

				if db.Spec.ShardTopology != nil {
					db.Spec.ShardTopology.Mongos.PodTemplate = v1.PodTemplateSpec{}
					db.Spec.ShardTopology.Shard.PodTemplate = v1.PodTemplateSpec{}
					db.Spec.ShardTopology.ConfigServer.PodTemplate = v1.PodTemplateSpec{}
				}

				return db
			}, metav1.PatchOptions{})
			if err != nil {
				log.Error(err, "failed to patch mongoDB")
				return err
			}

			c.RunParallel(
				dbaapi.TLSAdded,
				"Successfully Updated StatefulSets",
				c.newUpdateTLS(newDBCopy, stsNames))

			log.Info("cccccccccccccccccccccccccccccccccccccccccccc")
			return nil
		}

		log.Info("++++++++++++++++++++++++++ End of if  +++++++++++++++++++++++++++")

		ok := c.restartPods()
		if !ok {
			return nil
		}

		ddd, _, err := dbutil.CreateOrPatchMongoDB(context.TODO(), c.DBClient.KubedbV1alpha2(), c.db.ObjectMeta, func(db *api.MongoDB) *api.MongoDB {
			db.Spec.SSLMode = api.SSLModeRequireSSL
			db.Spec.ClusterAuthMode = api.ClusterAuthModeX509
			db.Spec.TLS = dbCopy.Spec.TLS
			db.Spec.PodTemplate = nil
			db.Spec.KeyFileSecret = nil

			if db.Spec.ShardTopology != nil {
				db.Spec.ShardTopology.Mongos.PodTemplate = v1.PodTemplateSpec{}
				db.Spec.ShardTopology.Shard.PodTemplate = v1.PodTemplateSpec{}
				db.Spec.ShardTopology.ConfigServer.PodTemplate = v1.PodTemplateSpec{}
			}

			return db
		}, metav1.PatchOptions{})
		if err != nil {
			log.Error(err, "failed to patch mongoDB")
			return err
		}

		log.Info("+++++++++++++++++++++ end of else if +++++++++++++++++++++", "sts", ddd.Spec)
	}

	log.Info("1111111111111111111111111111111111111111")

	err = c.resumeMongoDB()
	if err != nil {
		log.Error(err, "failed to resume mongodb")
		return err
	}

	log.Info("2222222222222222222222222222222222222222222")

	conditions, err := lib.ResumeBackupConfiguration(c.Initializers.Stash.StashClient.StashV1beta1(), c.db.ObjectMeta, c.req.Status.Conditions, c.req.Generation)
	if err != nil {
		return err
	}
	if conditions != nil {
		newOpsReq, err := dbautil.UpdateMongoDBOpsRequestStatus(
			context.TODO(),
			c.DBClient.OpsV1alpha1(),
			c.req.ObjectMeta,
			func(in *dbaapi.MongoDBOpsRequestStatus) (types.UID, *dbaapi.MongoDBOpsRequestStatus) {
				in.Conditions = conditions
				return c.req.UID, in
			}, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

		c.req.Status = newOpsReq.Status
	}

	log.Info("333333333333333333333333333333333333333333")

	c.Recorder.Event(
		c.req,
		corev1.EventTypeNormal,
		dbaapi.Successful,
		"Successfully Reconfigured TLS",
	)
	log.V(2).Info("Successfully Reconfigured TLS")

	newOpsReq, err := dbautil.UpdateMongoDBOpsRequestStatus(
		context.TODO(),
		c.DBClient.OpsV1alpha1(),
		c.req.ObjectMeta,
		func(status *dbaapi.MongoDBOpsRequestStatus) (types.UID, *dbaapi.MongoDBOpsRequestStatus) {
			status.Phase = dbaapi.OpsRequestPhaseSuccessful
			status.ObservedGeneration = c.req.Generation
			status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.NewCondition(dbaapi.Successful, "Successfully Reconfigured TLS", c.req.Generation))
			return c.req.UID, status
		}, metav1.UpdateOptions{})
	if err != nil {
		log.Error(err, "failed to update status")
		return err
	}
	log.Info("4444444444444444444444444444444444444444444444")
	c.req.Status = newOpsReq.Status

	return nil
}

func getCertificateCondition(conditions []cm_api.CertificateCondition, conditionType cm_api.CertificateConditionType, status cmmeta.ConditionStatus, reason, message string) []cm_api.CertificateCondition {
	newCondition := cm_api.CertificateCondition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: &metav1.Time{Time: time.Now()},
	}

	for idx, cond := range conditions {
		if cond.Type != conditionType {
			continue
		}

		if cond.Status == status {
			newCondition.LastTransitionTime = cond.LastTransitionTime
		}
		conditions[idx] = newCondition
		return conditions
	}

	conditions = append(conditions, newCondition)

	return conditions
}

func (c *mongoOpsReqController) restartPods() bool {
	// TODO: move this to restart and share with restart func

	if c.db.Spec.ShardTopology == nil && c.db.Spec.ReplicaSet == nil {
		if !kmapi.IsConditionTrue(c.req.Status.Conditions, dbaapi.RestartStandalone) {
			c.RunParallel(dbaapi.RestartStandalone,
				"Successfully Restarted standalone node",
				c.newRestartFunc([][]string{c.podNames()},
					nil,
					nil))
		}

		return kmapi.IsConditionTrue(c.req.Status.Conditions, dbaapi.RestartStandalone)
	} else if c.db.Spec.ShardTopology == nil && c.db.Spec.ReplicaSet != nil {
		if !kmapi.IsConditionTrue(c.req.Status.Conditions, dbaapi.RestartReplicaSet) {
			c.RunParallel(dbaapi.RestartReplicaSet,
				"Successfully Restarted ReplicaSet nodes",
				c.newRestartFunc([][]string{c.podNames()},
					nil,
					nil,
					true))
		}

		return kmapi.IsConditionTrue(c.req.Status.Conditions, dbaapi.RestartReplicaSet)
	} else if c.db.Spec.ShardTopology != nil {
		if !kmapi.IsConditionTrue(c.req.Status.Conditions, dbaapi.RestartConfigServer) {
			c.RunParallel(dbaapi.RestartConfigServer,
				"Successfully Restarted ConfigServer nodes",
				c.newRestartFunc([][]string{c.configServerPodNames()},
					nil,
					nil,
					true))
		}

		if !kmapi.IsConditionTrue(c.req.Status.Conditions, dbaapi.RestartShard) {
			c.RunParallel(dbaapi.RestartShard,
				"Successfully Restarted Shard nodes",
				c.newRestartFunc(c.shardPodNames(),
					nil,
					nil,
					true))
		}

		if !kmapi.IsConditionTrue(c.req.Status.Conditions, dbaapi.RestartMongos) {
			c.RunParallel(dbaapi.RestartMongos,
				"Successfully Restarted Mongos nodes",
				c.newRestartFunc([][]string{c.mongosPodNames()},
					nil,
					nil))
		}
		return kmapi.IsConditionTrue(c.req.Status.Conditions, dbaapi.RestartConfigServer) &&
			kmapi.IsConditionTrue(c.req.Status.Conditions, dbaapi.RestartShard) &&
			kmapi.IsConditionTrue(c.req.Status.Conditions, dbaapi.RestartMongos)
	}

	return false
}
