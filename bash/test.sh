#!/bin/bash


# Remove everything before the last '/' of $0
# script_name=${0##*/}
# echo "${script_name}"

# Remove the prefix "te" from $script_name. If no 'te' prefix found, return the value of $script_name.
# script_name=${script_name#"te"}
# echo "${script_name}"



# command | tee file1.out file2.out file3.out
# command | tee -a file.out
# echo "newline" | sudo tee -a /etc/file.conf

# Running the command given after the script name
# out=$("$@")
# if [ "$(echo $out | jq -r '.demo')" == "1" ]; then
#     echo hi
#     return 0
# fi

source ./common.sh

retry ls

retry echo "hello"