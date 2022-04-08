#!/bin/bash
# `./init-rebuild init` to just edit the mongodbversion 4.4.6

if [[ "$1" == "edit" ]];then
    INITDB=arnobkumarsaha/mongodb-init:hello1
    kubectl get mongodbversion 4.4.6 -oyaml | sed "s~kubedb/mongodb-init:4.2-v3~"$INITDB"~g" | kubectl replace -f -
    exit 0
fi 

docker rmi $(docker images -f dangling=true -q)

images=(arnobkumarsaha/mongodb-init:hello1) # alpine:latest debian:stretch
for image in "${images[@]}"; do
    echo hi $image
    if [[ "$(docker images -q "$image" 2> /dev/null)" != "" ]]; then
        im=$(docker images -q "$image")
        docker rmi $im
    fi
done

# rebuild the init-docker-container & replace it in mongoversion yaml
cd /home/arnob/go/src/kubedb.dev/mongodb-init-docker && make push-to-kind
INITDB=arnobkumarsaha/mongodb-init:hello1
kubectl get mongodbversion 4.4.6 -oyaml | sed "s~kubedb/mongodb-init:4.2-v3~"$INITDB"~g" | kubectl replace -f -
