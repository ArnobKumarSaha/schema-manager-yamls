#!/bin/bash

if [[ "$1" == "arbiter" ]]; then 
    echo "using my changed code"
    kubectl set image -n kubedb deploy kubedb-kubedb-webhook-server operator=kubedb/mg-operator:arb-shard_linux_amd64
    kubectl set image -n kubedb deploy kubedb-kubedb-provisioner operator=kubedb/mg-operator:arb-shard_linux_amd64
elif [[ "$1" == "master" ]]; then
    echo "using master code"
    kubectl set image -n kubedb deploy kubedb-kubedb-webhook-server operator=kubedb/mg-operator:v0.18.0-1-gb4922308_linux_amd64
    kubectl set image -n kubedb deploy kubedb-kubedb-provisioner operator=kubedb/mg-operator:v0.18.0-1-gb4922308_linux_amd64
fi
