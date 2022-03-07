#!/bin/bash


cd /home/arnob/go/src/kubedb.dev/mongodb && make push-to-kind

prov=$(kubectl get pods --selector=app.kubernetes.io/name=kubedb-provisioner -n kubedb --output=jsonpath={.items..metadata.name})
hook=$(kubectl get pods --selector=app.kubernetes.io/name=kubedb-webhook-server -n kubedb --output=jsonpath={.items..metadata.name})

kubectl delete pod/${prov} -n kubedb
kubectl delete pod/${hook} -n kubedb


