package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Reads from the env file
var dbName string
var dbUser string
var dbPass string
var websiteURL string
var cookieSeure bool
var secretTurnstileToken string

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var validImageMimeTypes = []string{
	"image/png",
	"image/gif",
	"image/jpeg",
	"image/tiff",
}

var disallowedExtensions = []string{
	".html", ".htm", ".svg", ".js", ".mjs",
	".json", ".xml", ".zip", ".rar", ".tar",
	".7z", ".gz", ".tpl", ".tmpl",
}

func validateImage(image *multipart.FileHeader) (bool, string) {
	if image.Size > 50000000 { //50MB
		return false, "file too big"
	}

	var file multipart.File
	var fileError error
	file, fileError = image.Open()
	if fileError != nil {
		return false, "Error opening image data"
	}
	defer file.Close()

	var buf []byte = make([]byte, 512)
	var n int
	var readError error
	n, readError = file.Read(buf)
	if readError != nil {
		return false, "Error reading image data"
	}

	var mime string = http.DetectContentType(buf[:n])

	if !slices.Contains(validImageMimeTypes, mime) {
		return false, "Incorrect mime type"
	}

	return true, ""
}

func validateFile(file *multipart.FileHeader) (bool, string) {
	if file.Size > 80000000 {
		return false, "File too big"
	} //80MB

	var fileOpen multipart.File
	var fileError error
	fileOpen, fileError = file.Open()
	if fileError != nil {
		return false, "Error opening file data"
	}
	defer fileOpen.Close()

	var buf []byte = make([]byte, 512)
	var n int
	var readError error
	n, readError = fileOpen.Read(buf)
	if readError != nil {
		return false, "Error reading file data"
	}

	var extension = filepath.Ext(file.Filename)

	if slices.Contains(disallowedExtensions, extension) {
		return false, "File type not allowed."
	}

	var mime string = http.DetectContentType(buf[:n])

	if slices.Contains(validImageMimeTypes, mime) {
		return false, "File is an image, use image upload."
	}

	return true, ""
}

func base62Encode(n uint32) string {
	if n == 0 {
		return string(base62Alphabet[0])
	}

	out := make([]byte, 0, 6)
	for n > 0 {
		r := n % 62
		out = append([]byte{base62Alphabet[r]}, out...)
		n /= 62
	}
	return string(out)
}

func randomiseName() string {
	var b [4]byte
	var err error
	_, err = rand.Read(b[:])
	if err != nil {
		panic(err)
	}

	var randomInt uint32 = binary.BigEndian.Uint32(b[:])

	var uniqueID string = base62Encode(randomInt)

	return uniqueID

}

func storeImage(image *multipart.FileHeader) (bool, string) {
	var src multipart.File
	var err error
	src, err = image.Open()
	if err != nil {
		return false, "Error opening image."
	}
	defer src.Close()

	var dst *os.File
	dst, err = os.Create(fmt.Sprintf("./images/%s", image.Filename))
	if err != nil {
		return false, "Unable to create source file."
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return false, "Failed to copy image contents."
	}

	return true, ""
}

func storeFile(file *multipart.FileHeader) (bool, string) {
	var src multipart.File
	var err error
	src, err = file.Open()
	if err != nil {
		return false, "Error opening file contents"
	}
	defer src.Close()

	var dst *os.File
	dst, err = os.Create(fmt.Sprintf("./files/%s", file.Filename))
	if err != nil {
		return false, "Unable to allocate the file space."
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return false, "Failed to copy file."
	}

	return true, ""
}

func checkTurnstile(token string) (bool, string) {
	var payload map[string]string = map[string]string{
		"secret":   secretTurnstileToken,
		"response": token,
	}
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post("https://challenges.cloudflare.com/turnstile/v0/siteverify", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()

	return true, ""
}

func createEndpoints(router *gin.Engine) {
	var api *gin.RouterGroup = router.Group("/api")
	{
		api.POST("/uploadImage", func(c *gin.Context) {
			type SubmitBody struct {
				Token string `form:"token"`
				Image *multipart.FileHeader
			}

			var body SubmitBody

			if err := c.ShouldBind(&body); err != nil {
				c.JSON(400, gin.H{"status": "error", "message": "Invalid form fields"})
				return
			}

			const maxUploadSize int64 = 51 << 20 //51 MB (1mb for turnstile token)

			if c.Request.ContentLength > maxUploadSize {
				c.JSON(200, gin.H{"status": "error", "message": "Image too big"})
				return
			}

			if err := c.Request.ParseMultipartForm(maxUploadSize); err != nil {
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

			file, fileError := c.FormFile("image")
			if fileError != nil {
				log.Print("Image ain't an image tbh")
				c.JSON(200, gin.H{"status": "error", "message": "Not an image"})
				return
			}
			var errorMsg string
			var validated bool
			validated, errorMsg = validateImage(file)
			if !validated {
				c.JSON(200, gin.H{"status": "error", "message": errorMsg})
				return
			}

			var ext = filepath.Ext(file.Filename)
			file.Filename = fmt.Sprintf("%s%s", randomiseName(), ext)

			var moveError string
			var moved bool
			moved, moveError = storeImage(file)

			if !moved {
				c.JSON(200, gin.H{"status": "error", "message": moveError})
				return
			}

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

			var moveError string
			var moved bool

			moved, moveError = storeFile(file)

			if !moved {
				c.JSON(200, gin.H{"status": "error", "message": moveError})
				return
			}

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
	}
}

func serveHTML(router *gin.Engine) {
	router.NoRoute(func(c *gin.Context) {
		if c.Request.Method == "GET" {
			var path string = filepath.Join("./public", c.Request.URL.Path)
			var info os.FileInfo
			var pathError error
			if info, pathError = os.Stat(path); pathError == nil && !info.IsDir() {
				c.File(path)
				return
			}

			if filepath.Ext(path) == "" {
				var htmlPath string = path + ".html"
				var htmlError error
				if _, htmlError = os.Stat(htmlPath); htmlError == nil {
					c.File(htmlPath)
					return
				}
			}

			if info, err := os.Stat(path); err == nil && info.IsDir() {
				var indexPath string = filepath.Join(path, "index.html")
				var indexPathError error
				if _, indexPathError = os.Stat(indexPath); indexPathError == nil {
					c.File(indexPath)
					return
				}
			}

			c.File("./public/404.html")
		}
	})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Print(".ENV file not found")
	}

	dbName = os.Getenv("DBNAME")
	dbUser = os.Getenv("DBUSER")
	dbPass = os.Getenv("DBPASS")
	websiteURL = os.Getenv("WEBSITEURL")
	secretTurnstileToken = os.Getenv("turnstileKey")
	var PORT string = os.Getenv("PORT")
	var cookieSeureString string = os.Getenv("SECURE")
	var boolParseError error
	cookieSeure, boolParseError = strconv.ParseBool(cookieSeureString)
	if boolParseError != nil {
		log.Fatal("Secure method not found")
	}

	//gin.SetMode(gin.ReleaseMode) //Uncomment in prod
	var router *gin.Engine = gin.Default()

	router.Use(func(c *gin.Context) {
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"frame-src https://challenges.cloudflare.com https://*.challenges.cloudflare.com; "+
				"connect-src * https://cloudflare.com https://*.cloudflare.com https://challenges.cloudflare.com https://*.challenges.cloudflare.com; "+
				"font-src * https://cloudflare.com https://*.cloudflare.com https://challenges.cloudflare.com https://*.challenges.cloudflare.com; "+
				"script-src * 'unsafe-inline' 'unsafe-eval' https://cloudflare.com https://*.cloudflare.com https://challenges.cloudflare.com https://*.challenges.cloudflare.com; "+
				"script-src-elem * 'unsafe-inline' https://cloudflare.com https://*.cloudflare.com https://challenges.cloudflare.com https://*.challenges.cloudflare.com; "+
				"img-src * data: https://cloudflare.com https://*.cloudflare.com https://challenges.cloudflare.com https://*.challenges.cloudflare.com; "+
				"style-src * 'unsafe-inline' https://cloudflare.com https://*.cloudflare.com https://challenges.cloudflare.com https://*.challenges.cloudflare.com;",
		)

	})

	createEndpoints(router)
	serveHTML(router)

	router.Static("/assets", "./assets")

	router.Run(fmt.Sprintf(":%s", PORT))
}
