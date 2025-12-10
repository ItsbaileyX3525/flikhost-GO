package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

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

func validateImage(image *multipart.FileHeader) (bool, string, string) {
	var mime string
	if image.Size > 50000000 { //50MB
		return false, "file too big", mime
	}

	var file multipart.File
	var fileError error
	file, fileError = image.Open()
	if fileError != nil {
		return false, "Error opening image data", mime
	}
	defer file.Close()

	var buf []byte = make([]byte, 512)
	var n int
	var readError error
	n, readError = file.Read(buf)
	if readError != nil {
		return false, "Error reading image data", mime
	}

	mime = http.DetectContentType(buf[:n])

	if !slices.Contains(validImageMimeTypes, mime) {
		return false, "Incorrect mime type", mime
	}

	return true, "", mime
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

func storeImageFromHeader(image *multipart.FileHeader) (bool, string) {
	src, err := image.Open()
	if err != nil {
		return false, "Error opening image."
	}
	defer src.Close()

	dst, err := os.Create(fmt.Sprintf("./images/%s", image.Filename))
	if err != nil {
		return false, "Unable to create destination file."
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return false, "Failed to copy image contents."
	}

	return true, ""
}

func storeFileFromHeader(file *multipart.FileHeader) (bool, string) {
	src, err := file.Open()
	if err != nil {
		return false, "Error opening file contents"
	}
	defer src.Close()

	dst, err := os.Create(fmt.Sprintf("./files/%s", file.Filename))
	if err != nil {
		return false, "Unable to allocate the file space."
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return false, "Failed to copy file."
	}

	return true, ""
}

func connectDB() (*gorm.DB, error) {
	var dsn string = fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s?charset=utf8mb4&parseTime=True&loc=UTC", dbUser, dbPass, dbName)
	var db *gorm.DB
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
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

func serveHTML(router *gin.Engine) {
	router.NoRoute(func(c *gin.Context) {
		if c.Request.Method == "GET" {
			var baseDir string = "./public"

			log.Print(strings.Split(c.Request.Host, ".")[0])

			if (strings.Split(c.Request.Host, ".")[0]) == "api" {
				baseDir = "./api"
			}

			var path string = filepath.Join(baseDir, c.Request.URL.Path)
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

func serveFiles(router *gin.Engine) {
	var files *gin.RouterGroup = router.Group("/files")
	{
		files.GET("/:id", func(c *gin.Context) {
			var id string = c.Param("id")
			c.File(fmt.Sprintf("./files/%s", id))
			log.Printf("ID: %s", id)
		})
	}
}

func serveImages(router *gin.Engine) {
	var images *gin.RouterGroup = router.Group("/images")
	{
		images.GET("/:id", func(c *gin.Context) {
			var id string = c.Param("id")
			c.File(fmt.Sprintf("./images/%s", id))
			log.Printf("ID: %s", id)
		})
	}
}

// Stack overflow
func generateRandomToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}
