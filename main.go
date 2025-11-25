package main

import (
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

var validImageMimeTypes = []string{
	"image/png",
	"image/gif",
	"image/jpeg",
	"image/tiff",
}

func validateImage(image *multipart.FileHeader) (bool, string) {
	if image.Size > 80000000 { //80MB
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

func storeImage(image *multipart.FileHeader) (bool, string) {
	var src multipart.File
	var readError error
	src, readError = image.Open()
	if readError != nil {
		return false, "Error opening image."
	}
	defer src.Close()

	dst, err := os.Create(fmt.Sprintf("./images/%s", image.Filename))
	if err != nil {
		return false, "Unable to create source file."
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return false, "Failed to copy image contents."
	}

	return true, ""
}

func createEndpoints(router *gin.Engine) {
	var api *gin.RouterGroup = router.Group("/api")
	{
		api.POST("/uploadImage", func(c *gin.Context) {
			const maxUploadSize int64 = 80 << 20 //80 MB

			if c.Request.ContentLength > maxUploadSize {
				c.JSON(200, gin.H{"status": "error", "message": "File too big"})
				return
			}

			if err := c.Request.ParseMultipartForm(maxUploadSize); err != nil {
				c.JSON(200, gin.H{"status": "error", "message": "Invalid multipart form"})
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

			var moveError string
			var moved bool
			moved, moveError = storeImage(file)

			if !moved {
				c.JSON(200, gin.H{"status": "error", "message": moveError})
				return
			}

			c.JSON(200, gin.H{"status": "success", "message": "File uploaded successfully!"})
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
	var PORT string = os.Getenv("PORT")
	var cookieSeureString string = os.Getenv("SECURE")
	var boolParseError error
	cookieSeure, boolParseError = strconv.ParseBool(cookieSeureString)
	if boolParseError != nil {
		log.Fatal("Secure method not found")
	}

	//gin.SetMode(gin.ReleaseMode) //Uncomment in prod
	var router *gin.Engine = gin.Default()

	createEndpoints(router)
	serveHTML(router)

	router.Static("/assets", "./assets")

	router.Run(fmt.Sprintf(":%s", PORT))
}
