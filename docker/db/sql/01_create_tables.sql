-- Table for tasks
DROP TABLE IF EXISTS `tasks`;
DROP TABLE IF EXISTS `users`;

CREATE TABLE `users` (
    `id`         bigint(20) NOT NULL AUTO_INCREMENT,
    `name`       varchar(50) NOT NULL UNIQUE,
    `password`   binary(32) NOT NULL,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) DEFAULT CHARSET=utf8mb4;

CREATE TABLE `tasks` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `title` varchar(50) NOT NULL,
    `is_done` boolean NOT NULL DEFAULT b'0',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `comment` varchar(256) NOT NULL DEFAULT '未記入',
    PRIMARY KEY (`id`)
) DEFAULT CHARSET=utf8mb4;
