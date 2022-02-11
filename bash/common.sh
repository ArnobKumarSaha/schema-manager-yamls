#!/bin/bash

retry() {
    local delay=1
    while true; do
        str_command="$*"
        echo "Running command $str_command . . ."
        out=$("$@")
        echo "$out"
        return 0

        # if [ "$(echo $out | jq -r '.ok')" == "1" ]; then
        #     return 0
        # elif echo $out | grep "failed: Name or service not known"; then
        #     sleep $delay
        # else
        #     return 0
        # fi
    done
}