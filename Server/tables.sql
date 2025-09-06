-- --------------------------------------------------------
-- Host:                         127.0.0.1
-- Server version:               8.4.3 - MySQL Community Server - GPL
-- Server OS:                    Linux
-- HeidiSQL Version:             12.6.0.6765
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

-- Dumping structure for table iagd_backup.authattempt
CREATE TABLE IF NOT EXISTS `authattempt` (
  `key` varchar(36) NOT NULL,
  `code` varchar(9) NOT NULL DEFAULT '',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `email` varchar(320) NOT NULL,
  `status` enum('CREATED', 'COMPLETED') NOT NULL,
  PRIMARY KEY (`key`,`code`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='Contains a publicly known "token" and a secret pin code used to authenticate for a given user. \r\n\r\nUpon presenting both the token and the code to an API, an access token is inserted into "authentry" and returned to the user.\r\nIf the authenticating user does not exist, he will be created upon verification.';

-- Data exporting was unselected.

-- Dumping structure for table iagd_backup.authentry
CREATE TABLE IF NOT EXISTS `authentry` (
  `userid` bigint NOT NULL,
  `token` varchar(64) NOT NULL DEFAULT '',
  `email` varchar(320) NOT NULL,
  `ts` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`userid`,`token`),
  CONSTRAINT `FK_authentry_users` FOREIGN KEY (`userid`) REFERENCES `users` (`userid`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='GDIA: Auth tokens for the backup API';

-- Data exporting was unselected.

-- Dumping structure for table iagd_backup.characters
CREATE TABLE IF NOT EXISTS `characters` (
  `userid` bigint NOT NULL,
  `name` varchar(50) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NOT NULL,
  `filename` varchar(400) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`userid`,`name`),
  CONSTRAINT `FK_character_users` FOREIGN KEY (`userid`) REFERENCES `users` (`userid`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 ROW_FORMAT=COMPRESSED COMMENT='Stores filename mappings for character backups';

-- Data exporting was unselected.

-- Dumping structure for table iagd_backup.deleteditem
CREATE TABLE IF NOT EXISTS `deleteditem` (
  `userid` bigint NOT NULL,
  `id` varchar(36) NOT NULL COMMENT 'Item ID',
  `ts` bigint NOT NULL,
  PRIMARY KEY (`userid`,`id`),
  KEY `idx_userid_ts` (`userid`,`ts`),
  CONSTRAINT `FK_deleteditem_users` FOREIGN KEY (`userid`) REFERENCES `users` (`userid`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 ROW_FORMAT=COMPRESSED COMMENT='GDIA: Items which have been deleted. ID is stored here so that other clients can sync down and delete the item.';

-- Data exporting was unselected.

-- Dumping structure for table iagd_backup.item
CREATE TABLE IF NOT EXISTS `item` (
  `id` varchar(36) NOT NULL COMMENT 'GUID provided by client',
  `userid` bigint NOT NULL,
  `id_baserecord` bigint unsigned NOT NULL DEFAULT '0',
  `id_prefixrecord` bigint unsigned DEFAULT '0',
  `id_suffixrecord` bigint unsigned DEFAULT '0',
  `id_modifierrecord` bigint unsigned DEFAULT '0',
  `id_transmuterecord` bigint unsigned DEFAULT '0',
  `seed` bigint NOT NULL,
  `id_reliccompletionbonusrecord` bigint unsigned DEFAULT NULL,
  `id_enchantmentrecord` bigint unsigned DEFAULT NULL,
  `prefixrarity` int DEFAULT NULL,
  `unknown` int DEFAULT NULL,
  `enchantmentseed` bigint DEFAULT NULL,
  `materiacombines` int DEFAULT NULL,
  `stackcount` int NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT '',
  `namelowercase` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT '',
  `rarity` varchar(255) DEFAULT '',
  `mod` varchar(255) DEFAULT '',
  `levelrequirement` double DEFAULT '0',
  `ishardcore` tinyint(1) DEFAULT '0',
  `created_at` bigint NOT NULL DEFAULT '0' COMMENT 'Determined by IA',
  `ts` bigint NOT NULL DEFAULT '0' COMMENT 'Time of creation online',
  `relicseed` bigint DEFAULT NULL,
  `id_materiarecord` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`,`userid`) USING BTREE,
  KEY `FK_item_records` (`id_baserecord`) USING BTREE,
  KEY `FK_item_id_enchantmentrecord` (`id_enchantmentrecord`) USING BTREE,
  KEY `FK_item_id_reliccompletionbonusrecord` (`id_reliccompletionbonusrecord`) USING BTREE,
  KEY `FK_item_id_transmuterecord` (`id_transmuterecord`) USING BTREE,
  KEY `FK_item_id_modifierrecord` (`id_modifierrecord`) USING BTREE,
  KEY `FK_item_id_suffixrecord` (`id_suffixrecord`) USING BTREE,
  KEY `FK_item_id_prefixrecord` (`id_prefixrecord`) USING BTREE,
  KEY `FK_item_id_materiarecord` (`id_materiarecord`) USING BTREE,
  KEY `FK_item_userid` (`userid`) USING BTREE,
  KEY `idx_userid_ts` (`userid`,`ts`),
  CONSTRAINT `FK_item_id_enchantmentrecord` FOREIGN KEY (`id_enchantmentrecord`) REFERENCES `records` (`id_record`),
  CONSTRAINT `FK_item_id_materiarecord` FOREIGN KEY (`id_materiarecord`) REFERENCES `records` (`id_record`),
  CONSTRAINT `FK_item_id_modifierrecord` FOREIGN KEY (`id_modifierrecord`) REFERENCES `records` (`id_record`),
  CONSTRAINT `FK_item_id_prefixrecord` FOREIGN KEY (`id_prefixrecord`) REFERENCES `records` (`id_record`),
  CONSTRAINT `FK_item_id_reliccompletionbonusrecord` FOREIGN KEY (`id_reliccompletionbonusrecord`) REFERENCES `records` (`id_record`),
  CONSTRAINT `FK_item_id_suffixrecord` FOREIGN KEY (`id_suffixrecord`) REFERENCES `records` (`id_record`),
  CONSTRAINT `FK_item_id_transmuterecord` FOREIGN KEY (`id_transmuterecord`) REFERENCES `records` (`id_record`),
  CONSTRAINT `FK_item_records` FOREIGN KEY (`id_baserecord`) REFERENCES `records` (`id_record`),
  CONSTRAINT `FK_item_userid` FOREIGN KEY (`userid`) REFERENCES `users` (`userid`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 ROW_FORMAT=COMPRESSED COMMENT='GDIA: Items for the backup system';

-- Data exporting was unselected.

-- Dumping structure for table iagd_backup.records
CREATE TABLE IF NOT EXISTS `records` (
  `id_record` bigint unsigned NOT NULL AUTO_INCREMENT,
  `record` varchar(255) NOT NULL,
  PRIMARY KEY (`id_record`) USING BTREE,
  UNIQUE KEY `idx_records_unique` (`record`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=147493359 DEFAULT CHARSET=latin1 ROW_FORMAT=COMPRESSED;

-- Data exporting was unselected.

-- Dumping structure for table iagd_backup.throttleentry
CREATE TABLE IF NOT EXISTS `throttleentry` (
  `id` int NOT NULL AUTO_INCREMENT,
  `userid` varchar(320) DEFAULT NULL,
  `ip` varchar(512) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=46648 DEFAULT CHARSET=latin1 COMMENT='GDIA: Throttle entries to prevent brute force attempts / email spam';

-- Data exporting was unselected.

-- Dumping structure for table iagd_backup.users
CREATE TABLE IF NOT EXISTS `users` (
  `userid` bigint NOT NULL AUTO_INCREMENT,
  `email` varchar(320) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `buddy_id` int DEFAULT NULL COMMENT 'Activated/created when the user enabled buddy sharing in settings in IA',
  PRIMARY KEY (`userid`) USING BTREE,
  UNIQUE KEY `uq_email` (`email`),
  UNIQUE KEY `uq_buddy_id` (`buddy_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=20133 DEFAULT CHARSET=latin1 COMMENT='List of users in the backup system.\r\nHelps keep track of new users and returning users (check if they have items in the old solution, notify them that they may have entered the wrong email etc)'';\r\n';


ALTER TABLE `item`
    ADD COLUMN `id_ascendantaffixname` BIGINT(20) UNSIGNED NULL DEFAULT NULL AFTER `id_enchantmentrecord`,
    ADD COLUMN `id_ascendantaffix2hname` BIGINT(20) UNSIGNED NULL DEFAULT NULL AFTER `id_ascendantaffixname`,
    ADD COLUMN `rerollsused` INT(11) NULL AFTER `stackcount`;

-- Data exporting was unselected.

/*!40103 SET TIME_ZONE=IFNULL(@OLD_TIME_ZONE, 'system') */;
/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IFNULL(@OLD_FOREIGN_KEY_CHECKS, 1) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40111 SET SQL_NOTES=IFNULL(@OLD_SQL_NOTES, 1) */;

