-- feed 项目数据库初始化
-- 用法: mysql -u root < mysql-setup-feed.sql
CREATE DATABASE IF NOT EXISTS feed CHARACTER SET utf8mb4;
CREATE USER IF NOT EXISTS 'goapp'@'localhost' IDENTIFIED BY 'goapp123';
GRANT ALL PRIVILEGES ON feed.* TO 'goapp'@'localhost';
FLUSH PRIVILEGES;
