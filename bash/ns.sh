#!/bin/bash
# https://stackoverflow.com/questions/52369247/namespace-stuck-as-terminating-how-i-removed-it

NAMESPACE=db
kubectl proxy &
kubectl get namespace $NAMESPACE -o json |jq '.spec = {"finalizers":[]}' >temp.json
curl -k -H "Content-Type: application/json" -X PUT --data-binary @temp.json 127.0.0.1:8001/api/v1/namespaces/$NAMESPACE/finalize