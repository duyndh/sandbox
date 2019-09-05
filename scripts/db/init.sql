DROP TABLE IF EXISTS `todo`;
CREATE TABLE `todo` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `title` VARCHAR(200) DEFAULT NULL,
    `description` VARCHAR(1024) DEFAULT NULL,
    `reminder` TIMESTAMP NULL DEFAULT NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT NOW(),
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),
    PRIMARY KEY (`id`),
    UNIQUE KEY `id_unique` (`id`)
) ENGINE = InnoDB;