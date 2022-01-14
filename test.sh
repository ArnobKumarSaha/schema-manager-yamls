#!/bin/bash

# out=$(mongo bla bla)
# echo $?
# echo $out

echo $MONGODB_DATABASE_NAME
mongo --host=mongodb.db.svc.cluster.local  --authenticationDatabase=$MONGODB_DATABASE_NAME --username=$MONGODB_USERNAME --password=$MONGODB_PASSWORD < /tmp/ini.js;
echo $? print