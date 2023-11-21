-- Table for tasks
DROP TABLE IF EXISTS `tasks`;
DROP TABLE IF EXISTS `users`;
DROP TABLE IF EXISTS `ownership`;
 
CREATE TABLE `ownership` (
    `user_id` bigint(20) NOT NULL,
    `task_id` bigint(20) NOT NULL,
    PRIMARY KEY (`user_id`, `task_id`)
) DEFAULT CHARSET=utf8mb4;

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
    `deadline` datetime NOT NULL DEFAULT '2100-01-01 00:00:00',
    `comment` varchar(256) NOT NULL DEFAULT '未記入',
    PRIMARY KEY (`id`)
) DEFAULT CHARSET=utf8mb4;

ALTER TABLE ownership
ADD CONSTRAINT fk_ownership_user
FOREIGN KEY (user_id)
REFERENCES users(id)
ON DELETE CASCADE;

ALTER TABLE ownership
ADD CONSTRAINT fk_ownership_task
FOREIGN KEY (task_id)
REFERENCES tasks(id)
ON DELETE CASCADE;

-- CREATE TABLE `tags` (
--     `id` bigint(20) NOT NULL AUTO_INCREMENT,
--     `name` varchar(50) NOT NULL,
--     `user_id` bigint(20) NOT NULL,
--     PRIMARY KEY (`id`),
--     FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
-- ) DEFAULT CHARSET=utf8mb4;

-- CREATE TABLE `task_tags` (
--     `task_id` bigint(20) NOT NULL,
--     `tag_id` bigint(20) NOT NULL,
--     PRIMARY KEY (`task_id`, `tag_id`),
--     FOREIGN KEY (`task_id`) REFERENCES `tasks`(`id`) ON DELETE CASCADE,
--     FOREIGN KEY (`tag_id`) REFERENCES `tags`(`id`) ON DELETE CASCADE
-- ) DEFAULT CHARSET=utf8mb4;
