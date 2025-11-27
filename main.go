package main

import (
	"fmt"
	"log"
	"os"
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
