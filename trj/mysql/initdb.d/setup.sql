CREATE DATABASE trojan IF NOT EXISTS;

GRANT ALL PRIVILEGES ON trojan.* to trojan;

CREATE TABLE users (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    username VARCHAR(64) NOT NULL,
    password VARCHAR(128) NOT NULL DEFAULT "",
    rawpassword VARCHAR(128) NOT NULL,
    quota BIGINT NOT NULL DEFAULT 0,
    download BIGINT UNSIGNED NOT NULL DEFAULT 0,
    upload BIGINT UNSIGNED NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    INDEX (password)
);
