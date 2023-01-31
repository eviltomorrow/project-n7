#!/bin/bash

# create dir
mkdir -p $(pwd)/data/mongo/{db,conf,logs,init}
chmod 777 $(pwd)/data/mongo/db $(pwd)/data/mongo/logs

mkdir -p $(pwd)/data/mysql/{db,conf,logs,init}
chmod 777 $(pwd)/data/mysql/db $(pwd)/data/mysql/logs

mkdir -p $(pwd)/data/etcd
chmod 777 $(pwd)/data/etcd

mkdir -p $(pwd)/logs/{n7-collector,n7-email,n7-repository,n7-finder}

# mongodb.conf
cat > $(pwd)/data/mongo/conf/mongod.conf <<EOF
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
cat > $(pwd)/data/mongo/init/init_mongo.js <<EOF
db = db.getSiblingDB('n7');
db.createUser({"user":"admin","pwd":"admin123","roles":[{"role":"dbOwner","db":"n7"}]});
db.createCollection('metadata');
db.metadata.createIndex({date: 1, code: 1},{background: true});
EOF
chmod a+x $(pwd)/data/mongo/init/init_mongo.js

# my.cnf 
cat > $(pwd)/data/mysql/conf/my.cnf <<EOF
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
socket = /run/mysqld/mysqld.sock
sql_mode = 'STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION,PIPES_AS_CONCAT,ANSI_QUOTES'
EOF

# init.sql
cat > $(pwd)/data/mysql/init/init_mysql.sql <<EOF
CREATE USER 'admin'@'%' IDENTIFIED BY 'admin123';
CREATE DATABASE \`n7_repository\` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
GRANT ALL ON n7_repository.* TO 'admin'@'%';

-- create table quote_day
drop table if exists \`n7_repository\`.\`quote_day\`;
create table \`n7_repository\`.\`quote_day\` (
    \`id\` BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    \`code\` CHAR(8) NOT NULL COMMENT '股票代码',
    \`open\` DECIMAL(10,2) NOT NULL COMMENT '开盘价',
    \`close\` DECIMAL(10,2) NOT NULL COMMENT '收盘价',
    \`high\` DECIMAL(10,2) NOT NULL COMMENT '最高价',
    \`low\` DECIMAL(10,2) NOT NULL COMMENT '最低价',
    \`yesterday_closed\` DECIMAL(10,2) NOT NULL COMMENT '昨日收盘价',
    \`volume\` BIGINT NOT NULL COMMENT '交易量',
    \`account\` DECIMAL(18,2) NOT NULL COMMENT '金额',
    \`date\` TIMESTAMP NOT NULL COMMENT '日期',
    \`num_of_year\` INT NOT NULL COMMENT '天数',
    \`xd\` DOUBLE NOT NULL COMMENT '前复权比例',
    \`create_timestamp\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    \`modify_timestamp\` TIMESTAMP COMMENT '修改时间'
);
create index idx_code_date on \`n7_repository\`.\`quote_day\`(\`code\`,\`date\`);

drop table if exists \`n7_repository\`.\`quote_week\`;
create table \`n7_repository\`.\`quote_week\` (
    \`id\` BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    \`code\` CHAR(8) NOT NULL COMMENT '股票代码',
    \`open\` DECIMAL(10,2) NOT NULL COMMENT '开盘价',
    \`close\` DECIMAL(10,2) NOT NULL COMMENT '收盘价',
    \`high\` DECIMAL(10,2) NOT NULL COMMENT '最高价',
    \`low\` DECIMAL(10,2) NOT NULL COMMENT '最低价',
    \`yesterday_closed\` DECIMAL(10,2) NOT NULL COMMENT '昨日收盘价',
    \`volume\` BIGINT NOT NULL COMMENT '交易量',
    \`account\` DECIMAL(18,2) NOT NULL COMMENT '金额',
    \`date\` TIMESTAMP NOT NULL COMMENT '开始时期',
    \`num_of_year\` INT NOT NULL COMMENT '周数',
    \`xd\` DOUBLE NOT NULL COMMENT '前复权比例',
    \`create_timestamp\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    \`modify_timestamp\` TIMESTAMP COMMENT '修改时间'
);
create index idx_code_date_end on \`n7_repository\`.\`quote_week\`(\`code\`,\`date\`);

-- create table taskrecord
drop table if exists \`n7_repository\`.\`stock\`;
create table \`n7_repository\`.\`stock\` (
    \`code\` CHAR(8) NOT NULL COMMENT '股票代码',
    \`name\` VARCHAR(32) NOT NULL COMMENT '名称',
    \`suspend\` VARCHAR(32) NOT NULL COMMENT '停牌状态',
    \`create_timestamp\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    \`modify_timestamp\` TIMESTAMP COMMENT '修改时间',
     PRIMARY KEY(\`code\`)
);
EOF