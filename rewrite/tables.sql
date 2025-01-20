

CREATE TABLE `users` (
	`userid` BIGINT NOT NULL AUTO_INCREMENT,
	`email` VARCHAR(320) NOT NULL COLLATE 'latin1_swedish_ci',
	`created_at` TIMESTAMP NOT NULL DEFAULT current_timestamp(),
	`buddy_id` INT(11) NULL DEFAULT NULL,
	PRIMARY KEY (`userid`) USING BTREE,
	UNIQUE INDEX `uq_buddy_id` (`buddy_id`) USING BTREE
)
COLLATE='latin1_swedish_ci'
ENGINE=InnoDB
;

ALTER TABLE `users`
	ADD UNIQUE INDEX `uq_email` (`email`);

CREATE TABLE `records` (
	`id_record` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
	`record` VARCHAR(255) NOT NULL COLLATE 'latin1_swedish_ci',
	PRIMARY KEY (`id_record`) USING BTREE,
	UNIQUE INDEX `idx_records_unique` (`record`) USING BTREE
)
COLLATE='latin1_swedish_ci'
ENGINE=InnoDB
;



CREATE TABLE `authattempt` (
	`key` VARCHAR(36) NOT NULL COLLATE 'latin1_swedish_ci',
	`code` VARCHAR(9) NOT NULL DEFAULT '' COLLATE 'latin1_swedish_ci',
	`created_at` TIMESTAMP NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
	`email` VARCHAR(320) NOT NULL,
	PRIMARY KEY (`key`, `code`)
)
COLLATE='latin1_swedish_ci'
ENGINE=InnoDB
;
ALTER TABLE `authattempt`
	COMMENT='Contains a publicly known "token" and a secret pin code used to authenticate for a given user. \r\n\r\nUpon presenting both the token and the code to an API, an access token is inserted into "authentry" and returned to the user.';


CREATE TABLE `authentry` (
	`userid` BIGINT NOT NULL,
	`token` VARCHAR(64) NOT NULL DEFAULT '',
	`email` VARCHAR(320) NOT NULL COLLATE 'latin1_swedish_ci',
	`ts` TIMESTAMP NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
	PRIMARY KEY (`userid`, `token`)
)
COMMENT='GDIA: Auth tokens for the backup API'
COLLATE='latin1_swedish_ci'
ENGINE=InnoDB
;


CREATE TABLE `deleteditem` (
	`userid` BIGINT NOT NULL,
	`id` VARCHAR(36) NOT NULL COMMENT 'Item ID',
	`ts` BIGINT NOT NULL,
	PRIMARY KEY (`userid`, `id`)
)
COMMENT='GDIA: Items which have been deleted. ID is stored here so that other clients can sync down and delete the item.'
COLLATE='latin1_swedish_ci'
;


CREATE TABLE `throttleentry` (
	`id` INT NOT NULL AUTO_INCREMENT,
	`userid` VARCHAR(320) NULL,
	`ip` VARCHAR(512) NULL DEFAULT NULL,
	`created_at` TIMESTAMP NOT NULL DEFAULT now(),
	PRIMARY KEY (`id`)
)
COLLATE='latin1_swedish_ci'
;



CREATE TABLE `item` (
	`id` VARCHAR(36) NOT NULL COLLATE 'latin1_swedish_ci',
	`userid` BIGINT NOT NULL COLLATE 'latin1_swedish_ci',
	`id_baserecord` BIGINT(20) UNSIGNED NOT NULL DEFAULT '0',
	`id_prefixrecord` BIGINT(20) UNSIGNED NULL DEFAULT '0',
	`id_suffixrecord` BIGINT(20) UNSIGNED NULL DEFAULT '0',
	`id_modifierrecord` BIGINT(20) UNSIGNED NULL DEFAULT '0',
	`id_transmuterecord` BIGINT(20) UNSIGNED NULL DEFAULT '0',
	`seed` BIGINT(20) NOT NULL,
	`id_reliccompletionbonusrecord` BIGINT(20) UNSIGNED NULL DEFAULT NULL,
	`id_enchantmentrecord` BIGINT(20) UNSIGNED NULL DEFAULT NULL,
	`prefixrarity` INT(11) NULL DEFAULT NULL,
	`unknown` INT(11) NULL DEFAULT NULL,
	`enchantmentseed` BIGINT(20) NULL DEFAULT NULL,
	`materiacombines` INT(11) NULL DEFAULT NULL,
	`stackcount` INT(11) NOT NULL,
	`name` VARCHAR(255) NULL DEFAULT '' COLLATE 'latin1_swedish_ci',
	`namelowercase` VARCHAR(255) NULL DEFAULT '' COLLATE 'latin1_swedish_ci',
	`rarity` VARCHAR(255) NULL DEFAULT '' COLLATE 'latin1_swedish_ci',
	`mod` VARCHAR(255) NULL DEFAULT '' COLLATE 'latin1_swedish_ci',
	`levelrequirement` DOUBLE NULL DEFAULT '0',
	`ishardcore` TINYINT(1) NULL DEFAULT '0',
	`created_at` BIGINT(20) NOT NULL DEFAULT '0' COMMENT 'Determined by IA',
	`ts` BIGINT(20) NOT NULL DEFAULT '0' COMMENT 'Time of creation online',
	`relicseed` BIGINT(20) NULL DEFAULT NULL,
	`id_materiarecord` BIGINT(20) UNSIGNED NULL DEFAULT NULL,
	PRIMARY KEY (`id`, `userid`) USING BTREE,
	INDEX `FK_item_records` (`id_baserecord`) USING BTREE,
	INDEX `FK_item_id_enchantmentrecord` (`id_enchantmentrecord`) USING BTREE,
	INDEX `FK_item_id_reliccompletionbonusrecord` (`id_reliccompletionbonusrecord`) USING BTREE,
	INDEX `FK_item_id_transmuterecord` (`id_transmuterecord`) USING BTREE,
	INDEX `FK_item_id_modifierrecord` (`id_modifierrecord`) USING BTREE,
	INDEX `FK_item_id_suffixrecord` (`id_suffixrecord`) USING BTREE,
	INDEX `FK_item_id_prefixrecord` (`id_prefixrecord`) USING BTREE,
	INDEX `FK_item_id_materiarecord` (`id_materiarecord`) USING BTREE,
	INDEX `FK_item_userid` (`userid`) USING BTREE,
	CONSTRAINT `FK_item_id_enchantmentrecord` FOREIGN KEY (`id_enchantmentrecord`) REFERENCES `records` (`id_record`) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT `FK_item_id_materiarecord` FOREIGN KEY (`id_materiarecord`) REFERENCES `records` (`id_record`) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT `FK_item_id_modifierrecord` FOREIGN KEY (`id_modifierrecord`) REFERENCES `records` (`id_record`) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT `FK_item_id_prefixrecord` FOREIGN KEY (`id_prefixrecord`) REFERENCES `records` (`id_record`) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT `FK_item_id_reliccompletionbonusrecord` FOREIGN KEY (`id_reliccompletionbonusrecord`) REFERENCES `records` (`id_record`) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT `FK_item_id_suffixrecord` FOREIGN KEY (`id_suffixrecord`) REFERENCES `records` (`id_record`) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT `FK_item_id_transmuterecord` FOREIGN KEY (`id_transmuterecord`) REFERENCES `records` (`id_record`) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT `FK_item_records` FOREIGN KEY (`id_baserecord`) REFERENCES `records` (`id_record`) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT `FK_item_userid` FOREIGN KEY (`userid`) REFERENCES `users` (`userid`) ON UPDATE RESTRICT ON DELETE RESTRICT
)
COLLATE='latin1_swedish_ci'
ENGINE=InnoDB
;

ALTER TABLE `item`
	CHANGE COLUMN `name` `name` VARCHAR(255) NULL DEFAULT '' COLLATE 'utf8_general_ci' AFTER `stackcount`,
	CHANGE COLUMN `namelowercase` `namelowercase` VARCHAR(255) NULL DEFAULT '' COLLATE 'utf8_general_ci' AFTER `name`;



ALTER TABLE `records` ROW_FORMAT=COMPRESSED;
ALTER TABLE `item` ROW_FORMAT=COMPRESSED;
ALTER TABLE `deleteditem` ROW_FORMAT=COMPRESSED;


CREATE TABLE `characters` (
	`userid` BIGINT(20) NOT NULL,
	`name` VARCHAR(50) NOT NULL,
	`filename` VARCHAR(400) NOT NULL,
	`created_at` TIMESTAMP NOT NULL DEFAULT current_timestamp(),
	`updated_at` TIMESTAMP NOT NULL DEFAULT current_TIMESTAMP(),
	PRIMARY KEY (`userid`, `name`),
	CONSTRAINT `FK_characters_users` FOREIGN KEY (`userid`) REFERENCES `users` (`userid`)
)
COMMENT='Stores filename mappings for characters backups'
COLLATE='latin1_swedish_ci'
ENGINE=InnoDB
;

ALTER TABLE `deleteditem` ADD CONSTRAINT `FK_deleteditem_users` FOREIGN KEY (`userid`) REFERENCES `users` (`userid`);
ALTER TABLE `authentry` ADD CONSTRAINT `FK_authentry_users` FOREIGN KEY (`userid`) REFERENCES `users` (`userid`);
ALTER TABLE `characters` CHANGE COLUMN `name` `name` VARCHAR(50) NOT NULL COLLATE 'utf8_general_ci' AFTER `userid`;
ALTER TABLE `characters`	ROW_FORMAT=COMPRESSED;
ALTER TABLE `users`
	COMMENT='List of users in the backup system.\r\nHelps keep track of new users and returning users (check if they have items in the old solution, notify them that they may have entered the wrong email etc)\';\r\n',
	CHANGE COLUMN `buddy_id` `buddy_id` INT(11) NULL DEFAULT NULL COMMENT 'Activated/created when the user enabled buddy sharing in settings in IA' AFTER `created_at`;

ALTER TABLE `throttleentry`
	COMMENT='GDIA: Throttle entries to prevent brute force attempts / email spam';

ALTER TABLE `item`
	COMMENT='GDIA: Items for the backup system',
	CHANGE COLUMN `id` `id` VARCHAR(36) NOT NULL COMMENT 'GUID provided by client' COLLATE 'latin1_swedish_ci' FIRST;
ALTER TABLE `authattempt`
	COMMENT='Contains a publicly known "token" and a secret pin code used to authenticate for a given user. \r\n\r\nUpon presenting both the token and the code to an API, an access token is inserted into "authentry" and returned to the user.\r\nIf the authenticating user does not exist, he will be created upon verification.';
