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
	"unicode"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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

			var fileOpen multipart.File
			fileOpen, err = file.Open()
			if err != nil {
				c.JSON(200, gin.H{"status": "error", "message": "Error opening file"})
				return
			}
			defer fileOpen.Close()

			var hasher hash.Hash = sha256.New()
			if _, err = io.Copy(hasher, fileOpen); err != nil {
				c.JSON(200, gin.H{"status": "error", "message": "Error hashing file"})
				return
			}
			var hashBytes []byte = hasher.Sum(nil)
			var hashString string = hex.EncodeToString(hashBytes)

			var db *gorm.DB
			var dbErr error
			db, dbErr = connectDB()
			if dbErr != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Error connecting to the database"})
				return
			}

			var existingFileName string
			var existingFilePath string
			var fileHashScan *sql.Row = db.Raw(
				"SELECT fileName, filePath FROM imageuploads where fileHash = ?",
				hashString,
			).Row()

			var scanErr error
			scanErr = fileHashScan.Scan(&existingFileName, &existingFilePath)
			if scanErr == nil {
				var fileID string = hashString[:8]
				c.JSON(200, gin.H{
					"status":  "success",
					"message": "File already exists",
					"path":    existingFilePath,
					"fileID":  fileID,
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

			var ext = filepath.Ext(file.Filename)
			file.Filename = fmt.Sprintf("%s%s", randomiseName(), ext)

			var moveError string
			var moved bool
			moved, moveError = storeImageFromHeader(file)
			if !moved {
				c.JSON(200, gin.H{"status": "error", "message": moveError})
				return
			}

			var sessionID string = ""
			sessionID, _ = c.Cookie("session_id")
			log.Printf("Session ID: %s", sessionID)

			var userID interface{}
			if sessionID != "" {

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

			result := db.Exec(
				"INSERT INTO imageuploads (userID, fileName, fileSize, mimeType, filePath, isPublic, fileHash) VALUES (?, ?, ?, ?, ?, ?, ?)",
				userID,
				file.Filename,
				file.Size,
				mime,
				fmt.Sprintf("/images/%s", file.Filename),
				1,
				hashString,
			)

			var fileID string = hashString[:8]
			if result.Error != nil {
				fileID = ""
			}

			c.JSON(200, gin.H{
				"status":  "success",
				"message": "Image uploaded successfully!",
				"path":    fmt.Sprintf("/images/%s", file.Filename),
				"fileID":  fileID,
			})
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
				return
			}
			var validated bool
			var errorMsg string
			validated, errorMsg = validateFile(file)
			if !validated {
				c.JSON(200, gin.H{"status": "error", "message": errorMsg})
				return
			}

			var fileOpen multipart.File
			fileOpen, check = file.Open()
			if check != nil {
				c.JSON(200, gin.H{"status": "error", "message": "Error opening file"})
				return
			}
			defer fileOpen.Close()

			var hasher hash.Hash = sha256.New()
			if _, check = io.Copy(hasher, fileOpen); check != nil {
				c.JSON(200, gin.H{"status": "error", "message": "Error hashing file"})
				return
			}
			var hashBytes []byte = hasher.Sum(nil)
			var hashString string = hex.EncodeToString(hashBytes)

			var db *gorm.DB
			var dbErr error
			db, dbErr = connectDB()
			if dbErr != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Error connecting to the database"})
				return
			}

			var existingFileName string
			var existingFilePath string
			var fileHashScan *sql.Row = db.Raw(
				"SELECT fileName, filePath FROM fileuploads where fileHash = ?",
				hashString,
			).Row()

			var scanErr error
			scanErr = fileHashScan.Scan(&existingFileName, &existingFilePath)
			if scanErr == nil {
				c.JSON(200, gin.H{
					"status":  "success",
					"message": "File already exists",
					"path":    existingFilePath,
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

			var ext = filepath.Ext(file.Filename)
			file.Filename = fmt.Sprintf("%s%s", randomiseName(), ext)

			var moveError string
			var moved bool
			moved, moveError = storeFileFromHeader(file)
			if !moved {
				c.JSON(200, gin.H{"status": "error", "message": moveError})
				return
			}

			var sessionID string = ""
			sessionID, _ = c.Cookie("session_id")
			log.Printf("Session ID: %s", sessionID)

			var userID interface{}
			if sessionID != "" {

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

		api.POST("/createAccount", func(c *gin.Context) {
			type bodyType struct {
				Username  string `json:"username"`
				Password  string `json:"password"`
				Email     string `json:"email"`
				Turnstile string `json:"turnstile"`
			}

			var body bodyType
			var check error
			if check = c.ShouldBindBodyWithJSON(&body); check != nil {
				c.JSON(200, gin.H{"status": "error", "message": "invalid form"})
				return
			}

			var errMsg string
			var err bool
			err, errMsg = checkTurnstile(body.Turnstile)
			if !err {
				c.JSON(200, gin.H{"status": "error", "message": "failed bot verifcation", "errormessage": errMsg})
				return
			}

			var db *gorm.DB
			var dbErr error
			db, dbErr = connectDB()
			if dbErr != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Error connecting to the database"})
				return
			}

			//Checking if the user already exist

			var userCheck *sql.Row = db.Raw(
				"SELECT username FROM users WHERE username = ?",
				body.Username,
			).Row()

			var existingUsername string

			check = userCheck.Scan(&existingUsername)
			if check == nil {
				c.JSON(200, gin.H{"status": "error", "message": "account with username already exists."})
				return
			}
			if check != sql.ErrNoRows {
				c.JSON(500, gin.H{"status": "error", "message": check.Error()})
				return
			}

			//Check if password is ok
			var password string = body.Password
			if len(password) < 8 {
				c.JSON(200, gin.H{"status": "error", "message": "password needs to be atleast 8 chars."})
				return
			}

			var hasDigit bool = false
			var r rune

			for _, r = range password {
				if unicode.IsDigit(r) {
					hasDigit = true
					break
				}
			}

			if !hasDigit {
				c.JSON(200, gin.H{"status": "error", "message": "Password must contain atleast 1 number"})
				return
			}

			//hawk tuah hash that thang... idk
			var bytes []byte
			bytes, check = bcrypt.GenerateFromPassword([]byte(password), 14)
			var hash string = string(bytes)

			execute := db.Exec(
				"INSERT INTO users (username, email, password, hasAgreedToTOS, isActive) VALUES (?, ?, ?, ?, ?)",
				body.Username,
				body.Email,
				hash,
				1,
				1,
			)

			if execute.Error != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Database execution failed."})
			}

			c.JSON(200, gin.H{"status": "success", "message": "account created successfully!"})

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

		api.GET("/checkSession", func(c *gin.Context) {
			sessionID, err := c.Cookie("session_id")
			if err != nil || sessionID == "" {
				c.JSON(200, gin.H{"loggedIn": false})
				return
			}

			db, dbErr := connectDB()
			if dbErr != nil {
				c.JSON(200, gin.H{"loggedIn": false})
				return
			}

			var username string
			row := db.Raw("SELECT username FROM sessions WHERE sessionID = ?", sessionID).Row()
			if scanErr := row.Scan(&username); scanErr != nil {
				c.JSON(200, gin.H{"loggedIn": false})
				return
			}

			c.JSON(200, gin.H{"loggedIn": true, "username": username})
		})

		api.POST("/validateKey", validateKeyHandler)
		api.POST("/deleteImage", deleteImageHandler)
		api.GET("/proxy", proxyHandler)
	}
}
