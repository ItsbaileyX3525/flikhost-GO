-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Generation Time: Nov 27, 2025 at 11:21 AM
-- Server version: 10.4.32-MariaDB
-- PHP Version: 8.2.12

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `flikhost`
--

-- --------------------------------------------------------

--
-- Table structure for table `sessions`
--

CREATE TABLE IF NOT EXISTS `sessions` (
  `ID` varchar(255) NOT NULL DEFAULT (UUID()),
  `username` varchar(255) NOT NULL,
  `userID` int(11) NOT NULL,
  `token` varchar(255) NOT NULL UNIQUE,
  `expiresAt` timestamp NOT NULL,
  `creation` timestamp NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`ID`),
  UNIQUE KEY `unique_token` (`token`),
  KEY `userID` (`userID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- --------------------------------------------------------

--
-- Table structure for table `fileuploads`
--

CREATE TABLE IF NOT EXISTS `fileuploads` (
  `uploadID` int(11) NOT NULL AUTO_INCREMENT,
  `userID` int(11) DEFAULT NULL,
  `fileName` varchar(255) NOT NULL,
  `fileSize` int(11) NOT NULL,
  `filePath` varchar(500) NOT NULL,
  `uploadedAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `expiresAt` timestamp NULL DEFAULT NULL,
  `isPublic` tinyint(1) NOT NULL DEFAULT 0,
  `downloadCount` int(11) NOT NULL DEFAULT 0,
  `fileHash` varchar(64) DEFAULT NULL,
  PRIMARY KEY (`uploadID`),
  UNIQUE KEY `filePath` (`filePath`),
  KEY `idx_fileUploads_userID` (`userID`),
  KEY `idx_fileUploads_uploadedAt` (`uploadedAt`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- --------------------------------------------------------

--
-- Table structure for table `imageuploads`
--

CREATE TABLE IF NOT EXISTS `imageuploads` (
  `uploadID` int(11) NOT NULL AUTO_INCREMENT,
  `userID` int(11) DEFAULT NULL,
  `fileName` varchar(255) NOT NULL,
  `fileSize` int(11) NOT NULL,
  `mimeType` varchar(100) NOT NULL,
  `filePath` varchar(500) NOT NULL,
  `uploadedAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `expiresAt` timestamp NULL DEFAULT NULL,
  `isPublic` tinyint(1) NOT NULL DEFAULT 0,
  `downloadCount` int(11) NOT NULL DEFAULT 0,
  `fileHash` varchar(64) DEFAULT NULL,
  PRIMARY KEY (`uploadID`),
  UNIQUE KEY `filePath` (`filePath`),
  KEY `idx_imageUploads_userID` (`userID`),
  KEY `idx_imageUploads_uploadedAt` (`uploadedAt`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- --------------------------------------------------------

--
-- Table structure for table `apiuploads`
--

CREATE TABLE IF NOT EXISTS `apiuploads` (
  `uploadID` int(11) NOT NULL AUTO_INCREMENT,
  `userID` int(11) DEFAULT NULL,
  `websiteName` varchar(255) DEFAULT NULL,
  `fileName` varchar(255) NOT NULL,
  `fileSize` bigint(20) NOT NULL,
  `mimeType` varchar(100) NOT NULL,
  `fileType` enum('image','file') NOT NULL,
  `filePath` varchar(500) NOT NULL,
  `uploadedAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `uploadIP` varchar(45) DEFAULT NULL,
  `fileHash` varchar(64) DEFAULT NULL,
  `downloadCount` int(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`uploadID`),
  UNIQUE KEY `filePath` (`filePath`),
  KEY `idx_apiUploads_userID` (`userID`),
  KEY `idx_apiUploads_websiteName` (`websiteName`),
  KEY `idx_apiUploads_uploadedAt` (`uploadedAt`),
  KEY `idx_apiUploads_fileHash` (`fileHash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- --------------------------------------------------------

--
-- Table structure for table `users`
--

CREATE TABLE IF NOT EXISTS `users` (
  `userID` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `apiKey` varchar(36) DEFAULT (UUID()),
  `hasAgreedToTOS` tinyint(1) NOT NULL DEFAULT 0,
  `createdAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `isActive` tinyint(1) NOT NULL DEFAULT 1,
  PRIMARY KEY (`userID`),
  UNIQUE KEY `email` (`email`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `apiKey` (`apiKey`),
  KEY `idx_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
