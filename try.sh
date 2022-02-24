#!/bin/bash

ls -la docker-entrypoint-initdb.d
for f in docker-entrypoint-initdb.d/*; do
    echo bla
    echo $f
done

for f in webinar/db/*; do
    echo webinar
    echo $f
done