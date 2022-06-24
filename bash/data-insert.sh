#!/bin/bash

for i in {1..100}
do
   mongo mydb --host localhost --quiet < insert.js
done

# mongo mydb --host localhost --quiet --eval "rs.status()"

# mongo mydb --host localhost --quiet --eval "db.coll.insert({"hello": "world"})"

