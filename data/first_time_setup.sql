/**
 * Firsttime only setup relational database
 * 
 * Change values for cnotest_test and password before executing
 */

USE mysql;
CREATE DATABASE IF NOT EXISTS cnotest_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci;
CREATE user IF NOT EXISTS 'app_user' IDENTIFIED BY '***';
GRANT SELECT ON cnotest_test.* TO 'app_user'@'%';
