#!/bin/bash

echo "Checking for KubeVault"
KUBEVAULT_POD_COUNT=1
while [[ true ]]; do
    out=$(kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubevault" -o=jsonpath='{.items[?(@.status.phase=="Running")].metadata.name}')
    counter=0
    for i in $out
        do
        :
        counter=$( expr $counter + 1 )
    done
    #echo kubevault pod count $counter
    if [[ $counter = $KUBEVAULT_POD_COUNT ]]; then
        break
    fi
done

echo "Checking for Stash"
STASH_POD_COUNT=1
while [[ true ]]; do
    out=$(kubectl get pods --all-namespaces -l "app.kubernetes.io/name=stash-enterprise" -o=jsonpath='{.items[?(@.status.phase=="Running")].metadata.name}')
    counter=0
    for i in $out
        do
        :
        counter=$( expr $counter + 1 )
    done
    if [[ $counter = $STASH_POD_COUNT ]]; then
        break
    fi
done

# echo "Checking for KubeDB"
# KUBEDB_POD_COUNT=3
# while [[ true ]]; do
#     out=$(kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb" -o=jsonpath='{.items[?(@.status.phase=="Running")].metadata.name}')
#     counter=0
#     for i in $out
#         do
#         :
#         counter=$( expr $counter + 1 )
#     done
#     if [[ $counter = $KUBEDB_POD_COUNT ]]; then
#         break
#     fi
# done

# This should be used if install kubedb manually (from mongo repo)
echo "Checking for KubeDB"
KUBEDB_POD_COUNT=1
while [[ true ]]; do
    out=$(kubectl get pods --all-namespaces -l "app.kubernetes.io/name=kubedb-community" -o=jsonpath='{.items[?(@.status.phase=="Running")].metadata.name}')
    counter=0
    for i in $out
        do
        :
        counter=$( expr $counter + 1 )
    done
    if [[ $counter = $KUBEDB_POD_COUNT ]]; then
        break
    fi
done

echo "Checking for cert-manager"
CERT_MANAGER_POD_COUNT=3
while [[ true ]]; do
    out=$(kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=cert-manager" -o=jsonpath='{.items[?(@.status.phase=="Running")].metadata.name}')
    counter=0
    for i in $out
        do
        :
        counter=$( expr $counter + 1 )
    done
    if [[ $counter = $CERT_MANAGER_POD_COUNT ]]; then
        break
    fi
done

echo "Checking for minio server"
MINIO_POD_COUNT=1
while [[ true ]]; do
    out=$(kubectl get pods --all-namespaces -l "app=minio" -o=jsonpath='{.items[?(@.status.phase=="Running")].metadata.name}')
    counter=0
    for i in $out
        do
        :
        counter=$( expr $counter + 1 )
    done
    if [[ $counter = $MINIO_POD_COUNT ]]; then
        break
    fi
done