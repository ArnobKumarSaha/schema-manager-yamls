#!/bin/bash

# find . -type f ! -iname "*jpg" ! -iname "*.jpeg" ! -iname "*png" -delete

retry() {
    local delay=1
    local numberOfTry=3
    local tryNo=1
    while [[ $tryNo -le $numberOfTry ]] ; do
        echo "try= $tryNo"
        str_command="$*"
        echo "Running command $str_command . . ."
        out=$("$@")
        echo "$out"

        if [ "$(echo $out | jq -r '.ok')" == "1" ]; then
            return 0
        elif echo $out | jq -r '.errmsg' | grep "HostUnreachable"; then
            sleep $delay
        elif echo $out | jq -r '.errmsg' | grep "Host not found"; then
            sleep $delay
        elif echo $out | grep "SocketException"; then
            tryNo=$((tryNo + 1))
            sleep $delay
            echo "try inside = $tryNo $numberOfTry"
        elif [ "$(echo $out | jq -r '.ok')" == "0" ]; then
            exit 1  # kill the container
        else
            return 0
        fi
    done
}

retry echo SocketException