package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func validateKeyHandler(c *gin.Context) {
	type AuthRequest struct {
		Password string `form:"gimmieServerKey"`
	}

	var req AuthRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(200, gin.H{"success": false, "message": "Invalid form data"})
		return
	}

	currentIP := c.ClientIP()
	rateLimitFile := "./assets/ratelimits/admin.txt"

	os.MkdirAll("./assets/ratelimits", 0755)

	var rateLimits map[string]int64 = make(map[string]int64)
	if data, err := os.ReadFile(rateLimitFile); err == nil {
		json.Unmarshal(data, &rateLimits)
	}

	currentTime := time.Now().UnixMilli()
	if lastRequest, exists := rateLimits[currentIP]; exists {
		if (currentTime - lastRequest) < 10000 {
			c.JSON(200, gin.H{"success": false, "message": "Sorry, you have been rate-limited. Please try again in 10 seconds."})
			return
		}
	}

	rateLimits[currentIP] = currentTime
	if data, err := json.MarshalIndent(rateLimits, "", "  "); err == nil {
		os.WriteFile(rateLimitFile, data, 0644)
	}

	var key string = ""
	switch req.Password {
	case dbPass:
		key = validationKey
	case "RewriteRule ^assets/ - [L]":
		key = "LOLYouActuallyThoughtThatWasGoingToWorkLMAOAnywaysJustEnjoyTheTerminalBecauseThatsWhatItIsHereFor"
	case "43_u9IO'IeV*;.f":
		key = "LelandComeOnBroLikeIdActuallyGiveYouThePasswordToFuckWithTheServer"
	case "YouWouldntThinkThisIsTheRealPasswordButItIsBecauseItIsEasyToRemember":
		key = "BroLelandHowDidYouFallForItAgainThoLOL"
	default:
		key = "FakeVaildationKeyLOLYouGotThePasswordWrongSoYouDontGetTheProperKeyAnywaysFeelFreeToPlayAroundWithTheTerminalAnyways"
	}

	c.JSON(200, gin.H{"success": true, "key": key})
}

func deleteImageHandler(c *gin.Context) {
	type DeleteRequest struct {
		ImageID string `form:"image_id"`
		Server  string `form:"server"`
	}

	var req DeleteRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(200, gin.H{"success": false, "message": "Invalid form data"})
		return
	}

	var userID interface{}

	if req.Server != "" {
		if req.Server != validationKey {
			c.JSON(200, gin.H{"success": false, "message": "Bypass token invalid, you cannot perform this action!"})
			return
		}
		userID = nil
	} else {
		sessionID, _ := c.Cookie("session_id")
		if sessionID == "" {
			c.JSON(200, gin.H{"success": false, "message": "No session found"})
			return
		}

		db, dbErr := connectDB()
		if dbErr != nil {
			c.JSON(500, gin.H{"success": false, "message": "Database connection error"})
			return
		}

		var scanErr error
		row := db.Raw("SELECT userID FROM sessions WHERE sessionID = ?", sessionID).Row()
		if scanErr = row.Scan(&userID); scanErr != nil {
			c.JSON(200, gin.H{"success": false, "message": "Session not found"})
			return
		}
	}

	db, dbErr := connectDB()
	if dbErr != nil {
		c.JSON(500, gin.H{"success": false, "message": "Database connection error"})
		return
	}

	var imagePath string
	if userID == nil {
		if err := db.Raw("SELECT filePath FROM imageuploads WHERE fileID = ?", req.ImageID).Row().Scan(&imagePath); err != nil {
			c.JSON(200, gin.H{"success": false, "message": "Error fetching image"})
			return
		}
	} else {
		if err := db.Raw("SELECT filePath FROM imageuploads WHERE fileID = ? AND userID = ?", req.ImageID, userID).Row().Scan(&imagePath); err != nil {
			c.JSON(200, gin.H{"success": false, "message": "Error fetching image"})
			return
		}
	}

	if userID == nil {
		db.Exec("DELETE FROM imageuploads WHERE fileID = ?", req.ImageID)
	} else {
		db.Exec("DELETE FROM imageuploads WHERE fileID = ? AND userID = ?", req.ImageID, userID)
	}

	filePath := filepath.Join("./images", filepath.Base(imagePath))
	if err := os.Remove(filePath); err != nil {
		c.JSON(200, gin.H{"success": false, "message": "Error deleting image file"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Image deleted successfully"})
}

func proxyHandler(c *gin.Context) {
	url := c.Query("url")
	message := c.Query("message")

	if url == "" || message == "" {
		c.JSON(400, gin.H{"error": "URL and message parameters are required"})
		return
	}

	apiURL := fmt.Sprintf("%s?message=%s", url, message)

	for key, values := range c.Request.URL.Query() {
		if key != "url" && key != "message" {
			for _, value := range values {
				apiURL += fmt.Sprintf("&%s=%s", key, value)
			}
		}
	}

	resp, err := http.Get(apiURL)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch data", "details": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(500, gin.H{"error": "Failed to fetch data", "status": resp.StatusCode})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read response"})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}
