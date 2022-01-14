#!/bin/bash

set -x

helm repo update

helm install kubedb appscode/kubedb \
    --version v2021.12.21 \
    --namespace kubedb --create-namespace \
    --set kubedb-enterprise.enabled=true \
    --set kubedb-autoscaler.enabled=true \
    --set-file global.license=/home/arnob/files/license/kubedb.txt 

helm install kubevault appscode/kubevault \
    --version v2022.01.11 \
    --namespace kubevault --create-namespace \
    --set-file global.license=/home/arnob/files/license/kubevault.txt
    
helm install stash appscode/stash             \
  --version v2021.11.24                  \
  --namespace kube-system                       \
  --set features.enterprise=true                \
  --set-file global.license=/home/arnob/files/license/stash.txt

helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.6.1 \
  --set installCRDs=true

helm install minio minio/minio \
  --namespace minio --create-namespace 
  --values=/home/arnob/files/stash/minio/value.yaml

