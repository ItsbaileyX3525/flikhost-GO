package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"html"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
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

			var imageData []byte
			imageData, err = io.ReadAll(fileOpen)
			if err != nil {
				c.JSON(200, gin.H{"status": "error", "message": "Error reading image data"})
				return
			}

			var hasher hash.Hash = sha256.New()
			hasher.Write(imageData)
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

			var userID interface{}
			if sessionID != "" {

				var row *sql.Row = db.Raw(
					"SELECT userID FROM sessions WHERE ID = ?",
					sessionID,
				).Row()

				if scanErr = row.Scan(&userID); scanErr != nil {

					//c.JSON(200, gin.H{"status": "error", "message": "Session not found, login"})
					//return
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
					"SELECT userID FROM sessions WHERE ID = ?",
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
				"INSERT INTO fileuploads (userID, fileName, fileSize, filePath, isPublic, fileHash) VALUES (?, ?, ?, ?, ?, ?)",
				userID,
				file.Filename,
				file.Size,
				fmt.Sprintf("/files/%s", file.Filename),
				1,
				hashString,
			)

			c.JSON(200, gin.H{
				"status":  "success",
				"message": "File stored successfully",
				"path":    fmt.Sprintf("/files/%s", file.Filename),
			})
		})

		api.POST("/apiUpload", func(c *gin.Context) {
			type SubmitBody struct {
				WebsiteName string `form:"websiteName"`
				ApiKey      string `form:"apiKey"`
			}

			var body SubmitBody
			var err error

			if err = c.ShouldBind(&body); err != nil {
				c.JSON(400, gin.H{"success": false, "message": "Invalid form fields"})
				return
			}

			// Validate API key exists and get userID
			if body.ApiKey == "" {
				c.JSON(200, gin.H{"success": false, "message": "API key is required."})
				return
			}

			// Connect to database early to validate API key
			var db *gorm.DB
			var dbErr error
			db, dbErr = connectDB()
			if dbErr != nil {
				c.JSON(500, gin.H{"success": false, "message": "Error connecting to the database"})
				return
			}

			// Validate API key and get user info
			var userID int
			var username string
			var apiKeyRow *sql.Row = db.Raw(
				"SELECT userID, username FROM users WHERE apiKey = ? AND isActive = 1",
				body.ApiKey,
			).Row()

			var scanErr error
			scanErr = apiKeyRow.Scan(&userID, &username)
			if scanErr != nil {
				c.JSON(200, gin.H{"success": false, "message": "API key is invalid or has expired."})
				return
			}

			const maxUploadSize int64 = 81 << 20 // 81 MB

			if c.Request.ContentLength > maxUploadSize {
				c.JSON(200, gin.H{"success": false, "message": "File too big (max 80MB)"})
				return
			}

			if err = c.Request.ParseMultipartForm(maxUploadSize); err != nil {
				c.JSON(200, gin.H{"success": false, "message": "Invalid multipart form"})
				return
			}

			// Try to get the file (support both "imageUpload" and "file" field names)
			var file *multipart.FileHeader
			file, err = c.FormFile("imageUpload")
			if err != nil {
				file, err = c.FormFile("file")
				if err != nil {
					c.JSON(200, gin.H{"success": false, "message": "No file uploaded."})
					return
				}
			}

			// Basic file size validation
			if file.Size > 80<<20 { // 80 MB
				c.JSON(200, gin.H{"success": false, "message": "File too big (max 80MB)"})
				return
			}

			// Open file to read content
			var fileOpen multipart.File
			fileOpen, err = file.Open()
			if err != nil {
				c.JSON(200, gin.H{"success": false, "message": "Error opening file"})
				return
			}
			defer fileOpen.Close()

			// Read file data for mime detection and hashing
			var fileData []byte
			fileData, err = io.ReadAll(fileOpen)
			if err != nil {
				c.JSON(200, gin.H{"success": false, "message": "Error reading file data"})
				return
			}

			// Detect mime type
			var mime string = http.DetectContentType(fileData)

			// Determine if it's an image or file
			var fileType string = "file"
			for _, validMime := range validImageMimeTypes {
				if mime == validMime {
					fileType = "image"
					break
				}
			}

			// Hash the file
			var hasher hash.Hash = sha256.New()
			hasher.Write(fileData)
			var hashBytes []byte = hasher.Sum(nil)
			var hashString string = hex.EncodeToString(hashBytes)

			// Check for duplicate uploads in apiuploads table
			var existingFilePath string
			var fileHashScan *sql.Row = db.Raw(
				"SELECT filePath FROM apiuploads WHERE fileHash = ?",
				hashString,
			).Row()

			scanErr = fileHashScan.Scan(&existingFilePath)
			if scanErr == nil {
				// File already exists, return existing link
				var link string = fmt.Sprintf("https://%s%s", c.Request.Host, existingFilePath)
				c.JSON(200, gin.H{
					"success": true,
					"message": "Uploaded successfully",
					"link":    link,
				})
				return
			}

			if scanErr != sql.ErrNoRows {
				c.JSON(200, gin.H{
					"success": false,
					"message": scanErr.Error(),
				})
				return
			}

			// Generate unique filename
			var ext = filepath.Ext(file.Filename)
			var uniqueFilename = fmt.Sprintf("%s%s", randomiseName(), ext)

			// Determine storage directory based on file type
			var directory string
			if fileType == "image" {
				directory = "images"
			} else {
				directory = "files"
			}

			// Save file to disk
			var filePath string = fmt.Sprintf("./%s/%s", directory, uniqueFilename)
			var dst *os.File
			dst, err = os.Create(filePath)
			if err != nil {
				c.JSON(200, gin.H{"success": false, "message": "Error creating file"})
				return
			}
			defer dst.Close()

			if _, err = dst.Write(fileData); err != nil {
				c.JSON(200, gin.H{"success": false, "message": "Error writing file"})
				return
			}

			// Insert into apiuploads table
			var uploadIP string = c.ClientIP()
			var dbFilePath string = fmt.Sprintf("/%s/%s", directory, uniqueFilename)

			var result = db.Exec(
				"INSERT INTO apiuploads (userID, websiteName, fileName, fileSize, mimeType, fileType, filePath, uploadIP, fileHash) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
				userID,
				body.WebsiteName,
				file.Filename,
				file.Size,
				mime,
				fileType,
				dbFilePath,
				uploadIP,
				hashString,
			)

			if result.Error != nil {
				c.JSON(200, gin.H{"success": false, "message": "Database insert failed: " + result.Error.Error()})
				return
			}

			log.Printf("API upload from user %s (ID: %d) via %s: %s (%s, %s) - IP: %s", username, userID, body.WebsiteName, uniqueFilename, fileType, mime, uploadIP)

			var link string = fmt.Sprintf("https://%s/%s/%s", c.Request.Host, directory, uniqueFilename)
			c.JSON(200, gin.H{
				"success": true,
				"message": "Uploaded successfully",
				"link":    link,
			})
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

			var turnstile = body.Turnstile

			var errMsg string
			var err bool
			err, errMsg = checkTurnstile(turnstile)
			if !err {
				c.JSON(200, gin.H{"status": "error", "message": "failed bot verifcation", "errormessage": errMsg})
				return
			}

			var username = body.Username
			var password = body.Password
			var email = body.Email

			username = html.EscapeString(username)

			var db *gorm.DB
			var dbErr error
			db, dbErr = connectDB()
			if dbErr != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Error connecting to the database"})
				return
			}

			//Checking if the user already exist

			var userCheck *sql.Row = db.Raw(
				"SELECT username FROM users WHERE username = ? or email = ?",
				username,
				email,
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
				username,
				email,
				hash,
				1,
				1,
			)

			if execute.Error != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Database execution failed."})
				return
			}

			var userID string

			var userIDFetch *sql.Row = db.Raw(
				"SELECT userID FROM users WHERE username = ? LIMIT 1",
				username,
			).Row()

			check = userIDFetch.Scan(&userID)
			if check != nil {
				c.JSON(200, gin.H{"status": "error", "message": "Error fetching userID... Server error?"})
				return
			}

			var sessionID string
			var sessionError error
			sessionID, sessionError = generateRandomToken()

			if sessionError != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Session generation failed."})
				return
			}

			var result *gorm.DB = db.Exec(
				"INSERT INTO sessions (ID, username, userID, token, expiresAt) VALUES (?, ?, ?, ?, ?)",
				sessionID,
				username,
				userID,
				sessionID,
				time.Now().Add(time.Hour*24*30),
			)

			if result.Error != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Failed to store session in database"})
				return
			}

			c.SetCookie(
				"session_id",
				sessionID,
				60*60*24*30,
				"/",
				"",
				cookieSeure,
				true,
			)

			c.JSON(200, gin.H{
				"status":   "success",
				"message":  "account created successfully!",
				"username": username,
			})

		})

		// Login endpoint
		api.POST("/login", func(c *gin.Context) {
			type bodyType struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}

			var body bodyType
			var check error
			if check = c.ShouldBindBodyWithJSON(&body); check != nil {
				c.JSON(200, gin.H{"status": "error", "message": "Invalid form"})
				return
			}

			var username = body.Username
			var password = body.Password

			if username == "" || password == "" {
				c.JSON(200, gin.H{"status": "error", "message": "Username and password are required"})
				return
			}

			var db *gorm.DB
			var dbErr error
			db, dbErr = connectDB()
			if dbErr != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Error connecting to the database"})
				return
			}

			var userID int
			var storedHash string
			var row *sql.Row = db.Raw(
				"SELECT userID, password FROM users WHERE username = ?",
				username,
			).Row()

			check = row.Scan(&userID, &storedHash)
			if check == sql.ErrNoRows {
				c.JSON(200, gin.H{"status": "error", "message": "Invalid username or password"})
				return
			}
			if check != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Database error"})
				return
			}

			// Verify password
			check = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
			if check != nil {
				c.JSON(200, gin.H{"status": "error", "message": "Invalid username or password"})
				return
			}

			// Generate session
			var sessionID string
			var sessionError error
			sessionID, sessionError = generateRandomToken()

			if sessionError != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Session generation failed"})
				return
			}

			var result *gorm.DB = db.Exec(
				"INSERT INTO sessions (ID, username, userID, token, expiresAt) VALUES (?, ?, ?, ?, ?)",
				sessionID,
				username,
				userID,
				sessionID,
				time.Now().Add(time.Hour*24*30),
			)

			if result.Error != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Failed to create session"})
				return
			}

			c.SetCookie(
				"session_id",
				sessionID,
				60*60*24*30,
				"/",
				"",
				cookieSeure,
				true,
			)

			c.JSON(200, gin.H{
				"status":   "success",
				"message":  "Logged in successfully!",
				"username": username,
			})
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
			row := db.Raw("SELECT username FROM sessions WHERE ID = ?", sessionID).Row()
			if scanErr := row.Scan(&username); scanErr != nil {
				c.JSON(200, gin.H{"loggedIn": false})
				return
			}

			c.JSON(200, gin.H{"loggedIn": true, "username": username})
		})

		api.POST("/validateKey", validateKeyHandler)
		api.POST("/deleteImage", deleteImageHandler)
		api.GET("/proxy", proxyHandler)

		// Get user account info
		api.GET("/getUserInfo", func(c *gin.Context) {
			sessionID, err := c.Cookie("session_id")
			if err != nil || sessionID == "" {
				c.JSON(200, gin.H{"status": "error", "message": "Not logged in"})
				return
			}

			db, dbErr := connectDB()
			if dbErr != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Database connection error"})
				return
			}

			var userID int
			var username string
			row := db.Raw("SELECT userID, username FROM sessions WHERE ID = ?", sessionID).Row()
			if scanErr := row.Scan(&userID, &username); scanErr != nil {
				c.JSON(200, gin.H{"status": "error", "message": "Session not found"})
				return
			}

			var email string
			var createdAt time.Time
			var apiKey string
			userRow := db.Raw("SELECT email, createdAt, apiKey FROM users WHERE userID = ?", userID).Row()
			if scanErr := userRow.Scan(&email, &createdAt, &apiKey); scanErr != nil {
				c.JSON(200, gin.H{"status": "error", "message": "User not found"})
				return
			}

			c.JSON(200, gin.H{
				"status":    "success",
				"username":  username,
				"email":     email,
				"apiKey":    apiKey,
				"createdAt": createdAt.Format("2006-01-02 15:04:05"),
			})
		})

		// Get user's uploaded images
		api.GET("/getUserImages", func(c *gin.Context) {
			sessionID, err := c.Cookie("session_id")
			if err != nil || sessionID == "" {
				c.JSON(200, gin.H{"status": "error", "message": "Not logged in"})
				return
			}

			db, dbErr := connectDB()
			if dbErr != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Database connection error"})
				return
			}

			var userID int
			var username string
			row := db.Raw("SELECT userID, username FROM sessions WHERE ID = ?", sessionID).Row()
			if scanErr := row.Scan(&userID, &username); scanErr != nil {
				c.JSON(200, gin.H{"status": "error", "message": "Session not found"})
				return
			}

			rows, queryErr := db.Raw(
				"SELECT uploadID, fileName, filePath, uploadedAt FROM imageuploads WHERE userID = ? ORDER BY uploadedAt DESC",
				userID,
			).Rows()
			if queryErr != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Error fetching images"})
				return
			}
			defer rows.Close()

			var images []gin.H
			for rows.Next() {
				var uploadID int
				var fileName string
				var filePath string
				var uploadedAt time.Time
				if scanErr := rows.Scan(&uploadID, &fileName, &filePath, &uploadedAt); scanErr != nil {
					continue
				}
				images = append(images, gin.H{
					"id":         uploadID,
					"name":       fileName,
					"path":       filePath,
					"uploadDate": uploadedAt.Format("2006-01-02 15:04:05"),
				})
			}

			c.JSON(200, gin.H{
				"status":   "success",
				"username": username,
				"images":   images,
			})
		})

		// Logout endpoint
		api.POST("/logout", func(c *gin.Context) {
			sessionID, err := c.Cookie("session_id")
			if err != nil || sessionID == "" {
				c.JSON(200, gin.H{"status": "error", "message": "Not logged in"})
				return
			}

			db, dbErr := connectDB()
			if dbErr != nil {
				c.JSON(500, gin.H{"status": "error", "message": "Database connection error"})
				return
			}

			db.Exec("DELETE FROM sessions WHERE ID = ?", sessionID)

			c.SetCookie(
				"session_id",
				"",
				-1,
				"/",
				"",
				cookieSeure,
				true,
			)

			c.JSON(200, gin.H{"status": "success", "message": "Logged out successfully"})
		})
	}
}
