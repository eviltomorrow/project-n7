# 使用说明

## 第一步

    执行 setup.sh

## 第二步

    docker-compose up

## 第三步

    - 配置 mongodb

    use n7
    db.createUser({user:"admin", pwd:"admin123", roles:[{role: "dbAdmin", db: "n7"}]})

    - 配置 mysql

    CREATE USER 'admin'@'%' IDENTIFIED BY 'admin123';
    CREATE DATABASE `n7_repository` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
    GRANT ALL ON n7_repository.* TO 'admin'@'%';
