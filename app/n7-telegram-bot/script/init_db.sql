-- create database
CREATE DATABASE `n7_telegrambot` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;

-- create table quote_day
DROP TABLE IF EXISTS `n7_telegrambot`.`subscribe`;
CREATE TABLE `n7_telegrambot`.`subscribe` (
    `id` BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    `username` VARCHAR(32) NOT NULL COMMENT '用户名',
    `chart_id` INT NOT NULL COMMENT 'chat id',
    `status` INT NOT NULL COMMENT '状态',
    `create_timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `modify_timestamp` TIMESTAMP COMMENT '修改时间'
);
ALTER TABLE `n7_telegrambot`.`subscribe` ADD UNIQUE(`username`);

