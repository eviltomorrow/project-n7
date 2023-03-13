#!/bin/bash
#存放目录
backup_dir=/home/shepard/db
#数据库库名
db_name=n7_repository
#日期命名
date_tag=`date +%Y%m%d`
#sql脚本名字
sqlfile=$db_name'_'$date_tag'.'sql
#压缩文件名字
tarfile=$sqlfile'.'tar'.'gz
#备份
mysqldump -h localhost -uroot -proot --databases $db_name > $backup_dir/$sqlfile 
#进行压缩并删除原文件
cd $backup_dir
tar -czf  $tarfile $sqlfile
rm -rf $sqlfile
#定时清除文件，以访长期堆积占用磁盘空间(删除5天以前带有tar.gz文件)
find $backup_dir -mtime +5 -name '*.tar.gz' -exec rm -rf {} \;