#!/bin/bash

# create dir
mkdir -p $(pwd)/mongo/{db,conf,logs,init}
chmod 777 $(pwd)/mongo/db $(pwd)/mongo/logs

mkdir -p $(pwd)/mysql/{db,conf,logs,init}
chmod 777 $(pwd)/mysql/db $(pwd)/mysql/logs

mkdir -p $(pwd)/app/{n7-collector,n7-email,n7-repository,n7-finder}

# mongodb.conf
cat > $(pwd)/mongo/conf/mongod.conf <<EOF
# mongod.conf

# for documentation of all options, see:
#   http://docs.mongodb.org/manual/reference/configuration-options/

# where to write logging data.
systemLog:
  destination: file
  logAppend: true
  path: /var/log/mongo/mongod.log

# Where and how to store data.
storage:
  dbPath: /var/lib/mongo
  journal:
    enabled: true
#  engine:
#  wiredTiger:

# how the process runs
processManagement:
  # fork: true  # fork and run in background
  # pidFilePath: /var/run/mongo/mongod.pid  # location of pidfile
  timeZoneInfo: /usr/share/zoneinfo

# network interfaces
net:
  port: 27017
  bindIp: 0.0.0.0  # Enter 0.0.0.0,:: to bind to all IPv4 and IPv6 addresses or, alternatively, use the net.bindIpAll setting.


#security:

#operationProfiling:

#replication:

#sharding:

## Enterprise-Only Options

#auditLog:

#snmp:
EOF

# init_mongo.js
cat > $(pwd)/mongo/init/init_mongo.js <<EOF
db = db.getSiblingDB('n7');
db.createUser({"user":"admin","pwd":"admin123","roles":[{"role":"dbOwner","db":"n7"}]});
db.createCollection('metadata');
db.metadata.createIndex({date: 1, code: 1},{background: true});
EOF
chmod a+x $(pwd)/mongo/init/init_mongo.js

# my.cnf 
cat > $(pwd)/mysql/conf/my.cnf <<EOF
[client]
port = 3306
default-character-set = utf8

[mysqld]
user = mysql
server-id = 1
port = 3306
character-set-server = utf8mb4
authentication_policy = mysql_native_password
secure_file_priv = /var/lib/mysql
expire_logs_days = 7
max_connections = 1000
log_error = /var/log/mysql/error.log
socket = /var/run/mysqld/mysql.sock
sql_mode = 'STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION,PIPES_AS_CONCAT,ANSI_QUOTES'
EOF

# init.sql
cat > $(pwd)/mysql/init/init_mysql.sql <<EOF
CREATE USER 'admin'@'%' IDENTIFIED BY 'admin123';
CREATE DATABASE n7_repository DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
GRANT ALL ON n7_repository.* TO 'admin'@'%';
EOF