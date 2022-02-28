#!/bin/bash

function get_pods {
    local HOSTS=$(echo $1 | tr "/" "\n")
    # convert to an array
    local pods=($HOSTS)

    echo "1 ->" "${pods[@]}"

    # checking if it really becomes an array
    if [[ "$(declare -p pods)" =~ "declare -a" ]]; then
        echo "yaah ! this is an array"
    fi

    unset pods[0]  # first index containes the string "replicaset". removing it

    # pods are comma separated. make it an array.
    HOSTS=$(echo "${pods[@]}" | tr "," "\n")
    peers=($HOSTS)

    echo "2 ->" "${peers[@]}"

    for pod in "${peers[@]}"; do
        echo "$pod"
    done
}
HOSTS=replicaset/mongodb-0.mongodb-pods.db.svc,mongodb-1.mongodb-pods.db.svc,mongodb-2.mongodb-pods.db.svc
get_pods $HOSTS

echo $HOSTS
echo $pods
echo ${peers[@]}