#!/bin/bash

# AUTHOR : Arnob kumar saha
# Date created : 06-01-2022
# Lase modified : 06-01-2022

# USAGE : ./create.sh DBTYPE NS_TYPE
# where DBTYPE = alone / replica / shard
# NS_TYPE = yes / no

x=$2

#demo
if [ "$x" = yes ]
then
    kubectl create ns demo
fi

kubectl apply -f /home/arnob/files/webinar/vault.yaml


# db
if [ "$x" = yes ]
then
    kubectl create ns db
fi
kubectl create secret generic -n db minio-secret \
    --from-file=/home/arnob/files/stash/minio/RESTIC_PASSWORD \
    --from-file=/home/arnob/files/stash/minio/AWS_ACCESS_KEY_ID \
    --from-file=/home/arnob/files/stash/minio/AWS_SECRET_ACCESS_KEY

kubectl apply -f /home/arnob/files/webinar/stash/repository.yaml

if [ "$1" = alone ]
then
    kubectl apply -f /home/arnob/files/webinar/db/standalone.yaml
    kubectl patch repository -n db minio-repo --type="merge" --patch='{"spec": {"backend": {"s3": {"prefix": "standalone"}}}}'

elif [ "$1" = replica ]
then 
    kubectl apply -f /home/arnob/files/webinar/db/replica.yaml
    kubectl patch repository -n db minio-repo --type="merge" --patch='{"spec": {"backend": {"s3": {"prefix": "replica"}}}}'

else 
    kubectl apply -f /home/arnob/files/webinar/db/shard.yaml
    kubectl patch repository -n db minio-repo --type="merge" --patch='{"spec": {"backend": {"s3": {"prefix": "shard"}}}}'

fi


# dev
if [ "$x" = yes ]
then
    kubectl create ns dev
fi
kubectl apply -f /home/arnob/files/webinar/init/configmap.yaml
