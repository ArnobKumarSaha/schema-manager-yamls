#!/bin/bash
# new changes in `mongodb` repo


cd /home/arnob/go/src/kubedb.dev/enterprise && make push-to-kind

prov=$(kubectl get pods --selector=app.kubernetes.io/name=kubedb-ops-manager -n kubedb --output=jsonpath={.items..metadata.name})

kubectl delete pod/${prov} -n kubedb
