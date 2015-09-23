CREATE DATABASE  IF NOT EXISTS `coreos_artifactory_monitor` /*!40100 DEFAULT CHARACTER SET utf8 */;
USE `coreos_artifactory_monitor`;
-- MySQL dump 10.13  Distrib 5.6.24, for osx10.8 (x86_64)

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `artifactory_auth_tokens`
--

DROP TABLE IF EXISTS `artifactory_auth_tokens`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `artifactory_auth_tokens` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'The primary key for each entry in the table.',
  `token` varchar(255) NOT NULL COMMENT 'The oauth or bearer token assigned to this user/service.',
  `created_at` datetime NOT NULL COMMENT 'The create date for this row.',
  `updated_at` datetime NOT NULL COMMENT 'The last update date for this row.',
  `name` varchar(255) DEFAULT NULL COMMENT 'The name of the user or service that has been granted authority.',
  `notes` text COMMENT 'General comments.',
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  UNIQUE KEY `token_UNIQUE` (`token`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `artifactory_deploys`
--

DROP TABLE IF EXISTS `artifactory_deploys`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
/* note status: Started = 1, Success = 2, Failed = 3 */;
CREATE TABLE `artifactory_deploys` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'The unique identifier for each row.',
  `deploy_id` varchar(255) NOT NULL COMMENT 'The last UUID assigned to this deployment.',
  `domain` varchar(255) NOT NULL COMMENT 'The domain that is being covered by this deploy. ',
  `environment` varchar(255) NOT NULL COMMENT 'The environment, for example: development, staging, QA, production, that this deploy is being performed.',
  `service_name` varchar(255) NOT NULL COMMENT 'The service name being deployed, for example acme-video-mobile',
  `version` varchar(255) NOT NULL COMMENT 'The version of the service being deployed e.g. 1.0.2',
  `status` int(11) NOT NULL DEFAULT '1' COMMENT 'The current status of the deploy: Started, Failed, Success.',
  `updated_at` datetime NOT NULL COMMENT 'The update date and time of the deploy.',
  `created_at` datetime NOT NULL COMMENT 'The create date and time of the deploy.',
  PRIMARY KEY (`id`),
  UNIQUE KEY `key_UNIQUE` (`domain`, `environment`, `service_name`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
