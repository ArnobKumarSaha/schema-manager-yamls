#!/bin/bash

# AUTHOR : Arnob kumar saha
# Date created : 06-01-2022
# Lase modified : 06-01-2022

# USAGE : ./delete.sh DBTYPE
# where DBTYPE = alone / replica / shard

#demo
kubectl delete -f /home/arnob/files/webinar/vault.yaml
#kubectl delete secret vault-keys -n demo


# db
kubectl delete secret -n db minio-secret

if [ "$1" = alone ]
then
    kubectl delete -f /home/arnob/files/webinar/db/standalone.yaml
elif [ "$1" = replica ]
then 
    kubectl delete -f /home/arnob/files/webinar/db/replica.yaml
else 
    kubectl delete -f /home/arnob/files/webinar/db/shard.yaml
fi

kubectl delete -f /home/arnob/files/webinar/stash/repository.yaml


# dev
kubectl delete -f /home/arnob/files/webinar/init/configmap.yaml



# namespaces
sleep 15
kubectl delete ns dev
kubectl delete ns db
kubectl delete ns demo