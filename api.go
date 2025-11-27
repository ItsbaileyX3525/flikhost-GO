package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func createEndpoints(router *gin.Engine) {
	var api *gin.RouterGroup = router.Group("/api")
	{
		api.POST("/uploadImage", func(c *gin.Context) {
			type SubmitBody struct {
				Token string `form:"token"`
				Image *multipart.FileHeader
			}

			var body SubmitBody
			var err error

			if err = c.ShouldBind(&body); err != nil {
				c.JSON(400, gin.H{"status": "error", "message": "Invalid form fields"})
				return
			}

			const maxUploadSize int64 = 51 << 20 //51 MB (1mb for turnstile token)

			if c.Request.ContentLength > maxUploadSize {
				c.JSON(200, gin.H{"status": "error", "message": "Image too big"})
				return
			}

			if err = c.Request.ParseMultipartForm(maxUploadSize); err != nil {
				c.JSON(200, gin.H{"status": "error", "message": "Invalid multipart form"})
				return
			}

			var token string = body.Token

			var passedTurnstile bool = false
			var turnstileError string = ""
			passedTurnstile, turnstileError = checkTurnstile(token)
			if !passedTurnstile {
				c.JSON(200, gin.H{"status": "error", "message": "Invalid turnstile."})
				return
			}

			if turnstileError != "" {
				c.JSON(200, gin.H{"status": "error", "message": turnstileError})
				return
			}

			var file *multipart.FileHeader
			file, err = c.FormFile("image")
			if err != nil {
				log.Print("Image ain't an image tbh")
				c.JSON(200, gin.H{"status": "error", "message": "Not an image"})
				return
			}
			var errorMsg string
			var validated bool
			var mime string
			validated, errorMsg, mime = validateImage(file)
			if !validated {
				c.JSON(200, gin.H{"status": "error", "message": errorMsg})
				return
			}

			var ext = filepath.Ext(file.Filename)
			file.Filename = fmt.Sprintf("%s%s", randomiseName(), ext)

			var moveError string
			var moved bool
			var fileOpen multipart.File
			moved, moveError, fileOpen = storeImage(file)

			if !moved {
				c.JSON(200, gin.H{"status": "error", "message": moveError})
				return
			}

			var hasher hash.Hash = sha256.New()

			if _, err = io.Copy(hasher, fileOpen); err != nil {
				return
			}

			var hashBytes []byte = hasher.Sum(nil)

			var hashString string = hex.EncodeToString(hashBytes)

			var sessionID string = ""
			sessionID, _ = c.Cookie("session_id")

			log.Printf("Session ID: %s", sessionID)

			var db *gorm.DB
			var dbErr error
			db, dbErr = connectDB()
			if dbErr != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Error connecting to the database"})
				return
			}
			var scanErr error
			var userID interface{}

			var fileName string
			var filePath string
			var fileHashScan *sql.Row = db.Raw(
				"SELECT fileName, filePath FROM imageuploads where fileHash = ?",
				hashString,
			).Row()

			scanErr = fileHashScan.Scan(&fileName, &filePath)
			if scanErr == nil {
				c.JSON(200, gin.H{
					"status":  "success",
					"message": "File already exists",
					"path":    filePath,
				})
				return
			}

			if scanErr != sql.ErrNoRows {
				c.JSON(200, gin.H{
					"status":  "error",
					"message": scanErr.Error(),
				})
				return
			}

			if sessionID != "" { //Has an account

				var row *sql.Row = db.Raw(
					"SELECT userID FROM sessions WHERE sessionID = ?",
					sessionID,
				).Row()

				if scanErr = row.Scan(&userID); scanErr != nil {
					c.JSON(200, gin.H{"status": "error", "message": "Session not found, login"})
					return
				}

			} else {
				userID = nil
			}

			db.Exec(
				"INSERT INTO imageuploads (userID, fileName, fileSize, mimeType, filePath, isPublic, fileHash) VALUES (?, ?, ?, ?, ?, ?, ?)",
				userID,
				file.Filename,
				file.Size,
				mime,
				fmt.Sprintf("/files/%s", file.Filename),
				1,
				hashString,
			)

			c.JSON(200, gin.H{"status": "success", "message": "Image uploaded successfully!"})
		})

		api.POST("/uploadFile", func(c *gin.Context) {
			type SubmitBody struct {
				Token string `form:"token"`
				File  *multipart.FileHeader
			}

			var body SubmitBody
			var check error

			if check = c.ShouldBind(&body); check != nil {
				c.JSON(200, gin.H{"status": "error", "message": "Invalid form, little cheat"})
				return
			}

			const maxUploadSize int64 = 81 << 20 //81 MB (1mb for turnstile token)

			if c.Request.ContentLength > maxUploadSize {
				c.JSON(200, gin.H{"status": "error", "message": "File too big!"})
			}

			if check = c.Request.ParseMultipartForm(maxUploadSize); check != nil {
				c.JSON(200, gin.H{"status": "error", "message": "Invalid mutlipart form"})
				return
			}

			var token string = body.Token

			var passedTurnstile bool = false
			var turnstileError string = ""
			passedTurnstile, turnstileError = checkTurnstile(token)
			if !passedTurnstile {
				c.JSON(200, gin.H{"status": "error", "message": "invalid turnstile"})
				return
			}

			if turnstileError != "" {
				c.JSON(200, gin.H{"status": "error", "message": turnstileError})
			}

			var file *multipart.FileHeader
			file, check = c.FormFile("file")
			if check != nil {
				c.JSON(200, gin.H{"status": "error", "message": "File aint a file tbh."})
			}
			var validated bool
			var errorMsg string
			validated, errorMsg = validateFile(file)
			if !validated {
				c.JSON(200, gin.H{"status": "error", "message": errorMsg})
				return
			}

			var ext = filepath.Ext(file.Filename)
			file.Filename = fmt.Sprintf("%s%s", randomiseName(), ext)

			var moveError string
			var moved bool
			var fileOpen multipart.File

			moved, moveError, fileOpen = storeFile(file)

			if !moved {
				c.JSON(200, gin.H{"status": "error", "message": moveError})
				return
			}

			var hasher hash.Hash = sha256.New()

			if _, check = io.Copy(hasher, fileOpen); check != nil {
				c.JSON(200, gin.H{"status": "error", "message": "error reading file hash"})
				return
			}

			var hashBytes []byte = hasher.Sum(nil)

			var hashString string = hex.EncodeToString(hashBytes)

			//Store all the shit in the database now

			var sessionID string = ""
			sessionID, _ = c.Cookie("session_id")

			log.Printf("Session ID: %s", sessionID)

			var db *gorm.DB
			var dbErr error
			db, dbErr = connectDB()
			if dbErr != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Error connecting to the database"})
				return
			}
			var scanErr error
			var userID interface{}

			var fileName string
			var filePath string
			var fileHashScan *sql.Row = db.Raw(
				"SELECT fileName, filePath FROM fileuploads where fileHash = ?",
				hashString,
			).Row()

			scanErr = fileHashScan.Scan(&fileName, &filePath)
			if scanErr == nil {
				c.JSON(200, gin.H{
					"status":  "success",
					"message": "File already exists",
					"path":    filePath,
				})
				return
			}

			if scanErr != sql.ErrNoRows {
				c.JSON(200, gin.H{
					"status":  "error",
					"message": scanErr.Error(),
				})
				return
			}

			if sessionID != "" { //Has an account

				var row *sql.Row = db.Raw(
					"SELECT userID FROM sessions WHERE sessionID = ?",
					sessionID,
				).Row()

				if scanErr = row.Scan(&userID); scanErr != nil {
					c.JSON(200, gin.H{"status": "error", "message": "Session not found, login"})
					return
				}

			} else {
				userID = nil
			}

			db.Exec(
				"INSERT INTO fileuploads (userID, fileName, fileSize, mimeType, filePath, isPublic, fileHash) VALUES (?, ?, ?, ?, ?, ?, ?)",
				userID,
				file.Filename,
				file.Size,
				file.Header.Get("Content-Type"),
				fmt.Sprintf("/files/%s", file.Filename),
				1,
				hashString,
			)

			c.JSON(200, gin.H{"status": "success", "message": "File stored successfully"})
		})

		//Get stuff for dirty shnitzels trynna see my endpoints

		api.GET("/uploadImage", func(c *gin.Context) {
			c.File("./dev/youlittleshnitzel.html")
		})
		api.GET("/uploadFile", func(c *gin.Context) {
			c.File("./dev/youlittleshnitzel.html")
		})
		api.GET("/createAccount", func(c *gin.Context) {
			c.File("./dev/youlittleshnitzel.html")
		})

		api.GET("/info", func(c *gin.Context) {
			ipAddress := c.GetHeader("CF-Connecting-IP")
			if ipAddress == "" {
				ipAddress = c.Query("testIP")
				if ipAddress == "" {
					ipAddress = c.ClientIP()
				}
			}

			userAgent := c.GetHeader("User-Agent")

			if ipAddress == "::1" || ipAddress == "127.0.0.1" {
				c.JSON(200, gin.H{
					"address": ipAddress,
					"country": "Local",
					"city":    "Local",
					"region":  "Local",
					"isp":     "Local",
					"loc":     "0,0",
					"post":    "Local",
					"ua":      userAgent,
				})
				return
			}

			accessToken := "b017e2178303a1"
			resp, err := http.Get(fmt.Sprintf("https://ipinfo.io/%s?token=%s", ipAddress, accessToken))
			if err != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Failed to fetch IP info"})
				return
			}
			defer resp.Body.Close()

			var ipinfoData map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&ipinfoData); err != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Failed to parse IP info"})
				return
			}

			info := gin.H{
				"address": ipAddress,
				"country": ipinfoData["country"],
				"city":    ipinfoData["city"],
				"region":  ipinfoData["region"],
				"isp":     ipinfoData["org"],
				"loc":     ipinfoData["loc"],
				"post":    ipinfoData["postal"],
				"ua":      userAgent,
			}

			c.JSON(200, info)
		})
	}
}
