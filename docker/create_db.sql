CREATE USER 'statusengine'@'localhost' IDENTIFIED BY 'password';
CREATE DATABASE IF NOT EXISTS `statusengine` DEFAULT CHARACTER SET utf8 DEFAULT COLLATE utf8_general_ci;
GRANT ALL PRIVILEGES ON `statusengine`.* TO 'statusengine'@'localhost';
