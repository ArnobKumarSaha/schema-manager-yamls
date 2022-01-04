# This file was written for testing purpose of Schema-manager MongoDB operator 


package framework

import (
	"context"
	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	issuer *certmanager.Issuer
)

func (i *Invocation) GetIssuerSecretSpec() *core.Secret {
	return &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "",
			Namespace: "",
		},
		Data: map[string][]byte{
			"cert": []byte{},
			"key":  []byte{},
		},
		Type: core.SecretTypeTLS,
	}
}

func (i *Invocation) GetIssuerSpec() *certmanager.Issuer {
	return &certmanager.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "",
			Namespace: "",
		},
		Spec: certmanager.IssuerSpec{
			IssuerConfig: certmanager.IssuerConfig{
				CA: &certmanager.CAIssuer{
					SecretName: "",
				},
			},
		},
	}
}

func (i *TestOptions) CreateIssuer() error {
	secret = i.GetIssuerSecretSpec()
	err := i.myClient.Create(context.TODO(), secret)
	if err != nil {
		return err
	}

	issuer = i.GetIssuerSpec()
	err = i.myClient.Create(context.TODO(), issuer)
	if err != nil {
		return err
	}
	return nil
}

func (i *TestOptions) DeleteIssuer() error {
	err := i.myClient.Delete(context.TODO(), secret)
	if err != nil {
		return err
	}
	err = i.myClient.Delete(context.TODO(), issuer)
	if err != nil {
		return err
	}
	return nil
}
