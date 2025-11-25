package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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

	router.Static("/assets", "./assets")

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

	router.Run(fmt.Sprintf(":%s", PORT))
}
