#!/bin/bash

# AUTHOR : Arnob kumar saha
# Date created : 06-01-2022
# Lase modified : 06-01-2022

# USAGE : ./delete.sh DBTYPE
# where DBTYPE = alone / replica / shard

#demo
kubectl delete -f /home/arnob/files/workspace/vaultserver.yaml
kubectl delete secret vault-keys -n demo


# db
kubectl delete secret -n db minio-secret

if [ "$1" = alone ]
then
    kubectl delete -f /home/arnob/files/stash/alone-mongo.yaml
elif [ "$1" = replica ]
then 
    kubectl delete -f /home/arnob/files/stash/replica-mongo.yaml
else 
    kubectl delete -f /home/arnob/files/stash/shard-mongo.yaml
fi

kubectl delete -f /home/arnob/files/stash/minio/repository.yaml


# dev
kubectl delete -f /home/arnob/files/workspace/configmap.yaml
