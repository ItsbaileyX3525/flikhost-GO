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
var validationKey string
var websiteURL string
var cookieSeure bool
var secretTurnstileToken string

func checkSSL() (string, string) {
	certPath := filepath.Join("ssl", "cert.pem")
	keyPath := filepath.Join("ssl", "key.pem")

	if _, err := os.Stat(certPath); err != nil {
		return "", ""
	}
	if _, err := os.Stat(keyPath); err != nil {
		return "", ""
	}

	return certPath, keyPath
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
	validationKey = os.Getenv("VALIDATIONKEY")
	secretTurnstileToken = os.Getenv("turnstileKey")
	var PORT string = os.Getenv("PORT")
	var cookieSeureString string = os.Getenv("SECURE")
	var boolParseError error
	cookieSeure, boolParseError = strconv.ParseBool(cookieSeureString)
	if boolParseError != nil {
		log.Fatal("Secure method not found")
	}

	// Initialize visitor tracking database
	visitorDB := initVisitorDB()

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
	serveFiles(router)
	serveImages(router)

	registerAPIRoutes(router, visitorDB)

	router.Static("/assets", "./assets")

	serveHTML(router)

	certPath, keyPath := checkSSL()
	if certPath != "" && keyPath != "" {
		log.Print("SSL certificates found, starting on port 443")
		router.RunTLS(":443", certPath, keyPath)
	} else {
		log.Printf("No SSL certificates, starting on port %s", PORT)
		router.Run(fmt.Sprintf(":%s", PORT))
	}
}
