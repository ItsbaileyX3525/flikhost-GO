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

CREATE TABLE `sessions` (
  `ID` varchar(255) NOT NULL DEFAULT (UUID()),
  `username` varchar(255) NOT NULL,
  `userID` int(11) NOT NULL,
  `token` varchar(255) NOT NULL UNIQUE,
  `expiresAt` timestamp NOT NULL,
  `creation` timestamp NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- --------------------------------------------------------

--
-- Table structure for table `fileuploads`
--

CREATE TABLE `fileuploads` (
  `uploadID` int(11) NOT NULL,
  `userID` int(11) DEFAULT NULL,
  `fileName` varchar(255) NOT NULL,
  `fileSize` int(11) NOT NULL,
  `filePath` varchar(500) NOT NULL,
  `uploadedAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `expiresAt` timestamp NULL DEFAULT NULL,
  `isPublic` tinyint(1) NOT NULL DEFAULT 0,
  `downloadCount` int(11) NOT NULL DEFAULT 0,
  `fileHash` varchar(64) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- --------------------------------------------------------

--
-- Table structure for table `imageuploads`
--

CREATE TABLE `imageuploads` (
  `uploadID` int(11) NOT NULL,
  `userID` int(11) DEFAULT NULL,
  `fileName` varchar(255) NOT NULL,
  `fileSize` int(11) NOT NULL,
  `mimeType` varchar(100) NOT NULL,
  `filePath` varchar(500) NOT NULL,
  `uploadedAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `expiresAt` timestamp NULL DEFAULT NULL,
  `isPublic` tinyint(1) NOT NULL DEFAULT 0,
  `downloadCount` int(11) NOT NULL DEFAULT 0,
  `fileHash` varchar(64) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- --------------------------------------------------------

--
-- Table structure for table `users`
--

CREATE TABLE `users` (
  `userID` int(11) NOT NULL,
  `username` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `hasAgreedToTOS` tinyint(1) NOT NULL DEFAULT 0,
  `createdAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `isActive` tinyint(1) NOT NULL DEFAULT 1
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

--
-- Indexes for dumped tables
--

--
-- Indexes for table `sessions`
--
ALTER TABLE `sessions`
  ADD UNIQUE KEY `unique_token` (`token`),
  ADD KEY `userID` (`userID`);

--
-- Indexes for table `fileuploads`
--
ALTER TABLE `fileuploads`
  ADD PRIMARY KEY (`uploadID`),
  ADD UNIQUE KEY `filePath` (`filePath`),
  ADD KEY `idx_fileUploads_userID` (`userID`),
  ADD KEY `idx_fileUploads_uploadedAt` (`uploadedAt`);

--
-- Indexes for table `imageuploads`
--
ALTER TABLE `imageuploads`
  ADD PRIMARY KEY (`uploadID`),
  ADD UNIQUE KEY `filePath` (`filePath`),
  ADD KEY `idx_imageUploads_userID` (`userID`),
  ADD KEY `idx_imageUploads_uploadedAt` (`uploadedAt`);

--
-- Indexes for table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`userID`),
  ADD UNIQUE KEY `email` (`email`),
  ADD UNIQUE KEY `username` (`username`),
  ADD KEY `idx_users_email` (`email`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `fileuploads`
--
ALTER TABLE `fileuploads`
  MODIFY `uploadID` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `imageuploads`
--
ALTER TABLE `imageuploads`
  MODIFY `uploadID` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `users`
--
ALTER TABLE `users`
  MODIFY `userID` int(11) NOT NULL AUTO_INCREMENT;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `sessions`
--
ALTER TABLE `sessions`
  ADD CONSTRAINT `sessions_ibfk_1` FOREIGN KEY (`userID`) REFERENCES `users` (`userID`),
  ADD CONSTRAINT `sessions_ibfk_2` FOREIGN KEY (`username`) REFERENCES `users` (`username`);

--
-- Constraints for table `fileuploads`
--
ALTER TABLE `fileuploads`
  ADD CONSTRAINT `fileuploads_ibfk_1` FOREIGN KEY (`userID`) REFERENCES `users` (`userID`) ON DELETE SET NULL;

--
-- Constraints for table `imageuploads`
--
ALTER TABLE `imageuploads`
  ADD CONSTRAINT `imageuploads_ibfk_1` FOREIGN KEY (`userID`) REFERENCES `users` (`userID`) ON DELETE SET NULL;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
