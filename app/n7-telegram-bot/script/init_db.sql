-- create database
CREATE DATABASE `n7_telegrambot` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;

-- create table quote_day
drop table if exists `n7_telegrambot`.`channel`;
create table `n7_telegrambot`.`channel` (
    `id` BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    `chrt_id` INT NOT NULL COMMENT 'chat id',
    `name` VARCHAR(32) NOT NULL COMMENT '频道名称',
    `create_timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `modify_timestamp` TIMESTAMP COMMENT '修改时间'
);
create index idx_code_date on `n7_telegrambot`.`channel`(`code`,`date`);

// chat_id
// name
// 