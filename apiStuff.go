package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fogleman/gg"
	"github.com/gin-gonic/gin"
	qrcode "github.com/skip2/go-qrcode"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Database models for visitor tracking
type UniqueVisitor struct {
	ID         uint      `gorm:"primaryKey"`
	SiteDomain string    `gorm:"size:255;index"`
	VisitorIP  string    `gorm:"size:45"`
	VisitDate  time.Time `gorm:"index"`
}

type SiteComment struct {
	ID          uint   `gorm:"primaryKey"`
	SiteDomain  string `gorm:"size:255;uniqueIndex"`
	CommentText string `gorm:"type:text"`
	AuthorName  string `gorm:"size:255"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Initialize visitor tracking database
func initVisitorDB() *gorm.DB {
	// Use separate database for visitor tracking
	dbUser := os.Getenv("DBUSER")
	dbPass := os.Getenv("DBPASS")
	dbHost := "localhost"
	dbName := "visitor_tracking"

	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Warning: Could not connect to visitor tracking database: %v", err)
		return nil
	}

	// Auto-migrate tables
	db.AutoMigrate(&UniqueVisitor{}, &SiteComment{})

	log.Println("Visitor tracking database initialized")
	return db
}

// UUID generation functions
func generateUUIDv4() string {
	uuid := make([]byte, 16)
	rand.Read(uuid)

	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant 10

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

func generateUUIDv7() string {
	// Get timestamp in milliseconds
	timestamp := time.Now().UnixMilli()
	timestampHex := fmt.Sprintf("%012x", timestamp)

	randBytes := make([]byte, 10)
	rand.Read(randBytes)

	randBytes[0] = (randBytes[0] & 0x0f) | 0x70 // Version 7
	randBytes[2] = (randBytes[2] & 0x3f) | 0x80 // Variant 10

	return fmt.Sprintf("%s-%s-%s-%s-%s",
		timestampHex[0:8],
		timestampHex[8:12],
		hex.EncodeToString(randBytes[0:2]),
		hex.EncodeToString(randBytes[2:4]),
		hex.EncodeToString(randBytes[4:10]))
}

func generateUUIDv8() string {
	data := make([]byte, 16)
	rand.Read(data)

	data[6] = (data[6] & 0x0f) | 0x80 // Version 8
	data[8] = (data[8] & 0x3f) | 0x80 // Variant 10

	hexStr := hex.EncodeToString(data)
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hexStr[0:8],
		hexStr[8:12],
		hexStr[12:16],
		hexStr[16:20],
		hexStr[20:32])
}

func generateUUIDv1() string {
	// Simplified v1 implementation
	timeMicro := time.Now().UnixMicro()
	timeHex := fmt.Sprintf("%015x", timeMicro)

	timeLow := timeHex[len(timeHex)-8:]
	timeMid := timeHex[len(timeHex)-12 : len(timeHex)-8]
	timeHi := timeHex[len(timeHex)-15:len(timeHex)-12] + "1"

	nBig, _ := rand.Int(rand.Reader, big.NewInt(0x3fff))
	clockSeq := fmt.Sprintf("%04x", nBig.Int64()|0x8000)

	node := make([]byte, 6)
	rand.Read(node)
	nodeHex := hex.EncodeToString(node)

	return fmt.Sprintf("%s-%s-%s-%s-%s",
		timeLow, timeMid, timeHi, clockSeq, nodeHex)
}

// API Handlers

// /api/sheldon - Random Sheldon Cooper quotes
func sheldonAPI(c *gin.Context) {
	quotes := []string{
		"I'm not crazy. My mother had me tested.",
		"That's my spot.",
		"For the record, it could kill us to meet new people.",
		"While I subscrive myself to the many words theory, which posits the existence of an infinite number of sheldons... I assure you that in none of them am I dancing.",
		"So, just to clarify, when you say three, do we stand up or do we pee?",
		"I'm sorry, coffee's out of the question. When I moved to California I promised my mother that I wouldn't start doing drugs",
		"Yes! She's like the dryer sheets of my heart.",
		"The statement stands for itself.",
		"You catch even more with manure, what's your point?",
		"I found the grinch to be a relatable, engaging character, and I was really with him...",
		"Hard as this may be to believe, It's possible that I'm not boyfriend material.",
		"The only conclusion was love.",
		"Scissors cuts paper. Paper covers rock. Rock crushes lizard.",
		"Bazinga!",
		"That's no reason to cry. One cries because one is sad. For example, I cry because others are stupid and it makes me sad.",
		"No cuts, no buts, no coconuts.",
		"Cause of injury: Lack of adhesive ducks.",
		"I'm exceedingly smart. I graduated college at fourteen. While my brother was getting an STD, I was getting a Ph.D.",
		"I'm Batman! Sssssh!",
		"Mom smokes in the car. Jesus is okay with it, but we can't tell dad.",
		"That was tricky because when it comes to alcohol, she generally means business.",
		`"Not knowing is part of the fun." Was that the motto of your community college?`,
		"I would have been here sooner but the bus kept stopping for other people to get on it.",
		"You're afraid of insects and women. Ladybugs must render you catatonic.",
		"It's 'Penny get your own Wi-Fi'; no spaces.",
		"I knew she wasn't lead car material.",
		"What computer do you have? And please don't say a white one.",
		"As I told you, the hero always peeks.",
		"I can't be impossible; I exist. I think what you meant to say is, 'I give up; he's improbable.'.",
		"It must be humbling to suck on so many levels.",
	}

	nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(len(quotes))))
	c.JSON(200, gin.H{"quote": quotes[nBig.Int64()]})
}

// /api/fact - Random facts
func factAPI(c *gin.Context) {
	facts := []string{
		"Finland has the most heavy metal bands per capita.",
		"Mount Everest was possibly shrunken by an earthquake.",
		"Pope John Paul II was made an honorary Harlem Globetrotter.",
		"On average, 100 people choke to death on ballpoint pens each year.",
		"There's a city named 'Rome' on every continent except Antarctica.",
		"Quebec City is the only walled city in North America north of Mexico.",
		"Frank Sinatra was offered the starring role in Die Hard when he was in his 70s.",
		"New Jersey is the top producer of the world's eggplants.",
		"Antarctica has the largest unclaimed territory on Earth.",
		"Bananas are berries, but strawberries aren't.",
		"The Eiffel Tower can be 15 cm taller during the summer.",
		"A group of flamingos is called a 'flamboyance'.",
		"The world's largest desert is Antarctica.",
		"Edgar Allan Poe married his thirteen-year-old cousin",
		"There is a metallic asteroid shaped like a dog bone named 'Kleopatra.'",
		"There are about the same number of stars in the observable universe as there are grains of sand on all of Earth's beaches.",
		"Queen Elizabeth II was a trained mechanic.",
		"It's estimated that Americans eat 50 billion hamburgers each year.",
		"Airlines saved $40,000 in 1987 by eliminating one olive from each salad served in first class.",
		"Close to 70 percent of the world's freshwater is held in glaciers and ice sheets.",
		"If added together, humans spend about two weeks of their lifetimes kissing.",
		"Al Capone's business card said he was a used furniture dealer.",
		"Twins are becoming more and more common.",
		"Dolphins give each other names.",
		"A giraffe can clean its ears with its 21-inch tongue.",
		"Most pandas around the world are on loan from China.",
		"Sloths can hold their breath longer than dolphins.",
		"Sharks are the only fish that can blink with both eyes.",
		"It's possible to lead a cow upstairs, but they'd prefer not to go downstairs.",
		"Ravens know when someone is spying on them.",
		"Animals that lay eggs don't have belly buttons.",
		"Our stomachs produce a new layer of mucus every two weeks, so they don't digest themselves.",
		"Your fingernails grow faster on your dominant hand.",
		"Women blink almost twice as much as men.",
		"Men hiccup more than women.",
		"Riding a roller coaster can help you pass a (small) kidney stone faster.",
		"Sweat doesn't smell bad; the combination of water, fat, and salt mixed with bacteria does.",
		"The average person walks the equivalent of three times around the world in a lifetime.",
		"The average person produces enough saliva in their lifetime to fill two swimming pools.",
	}

	nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(len(facts))))
	c.JSON(200, gin.H{"fact": facts[nBig.Int64()]})
}

// /api/fantasy_name - Fantasy name generator
func fantasyNameAPI(c *gin.Context) {
	firstNames := []string{
		"Aelric", "Seraphina", "Thalorin", "Vaelis", "Eowen",
		"Zephiron", "Lyriana", "Kaedros", "Sylwen", "Darion",
		"Nymira", "Eldrin", "Rhovan", "Vexara", "Draven",
		"Isolde", "Kareth", "Zylthia", "Orin", "Faelwen",
		"Jorvan", "Elara", "Tarian", "Maelis", "Varenth",
		"Lysara", "Zorin", "Calith", "Nyxara", "Xandor",
		"Althar", "Velmira", "Tethis", "Aerion", "Miravel",
		"Solric", "Ylthana", "Malrik", "Elarion", "Vaedrin",
		"Zareth", "Mirieth", "Kaelor", "Avaris", "Sylvar",
		"Dorian", "Thalion", "Vaedran", "Loramir", "Xyphor",
	}

	lastNames := []string{
		"Stormrider", "Duskbane", "Shadowthorn", "Ironhart", "Blackthorn",
		"Frostbane", "Drakefang", "Moonshadow", "Ravenshade", "Emberforge",
		"Darkwhisper", "Firebrand", "Silverwind", "Grimshaw", "Thunderstrike",
		"Starborn", "Bloodmoon", "Nightshade", "Daggerfall", "Stonebrook",
		"Frostvein", "Stormwatcher", "Icefang", "Duskrunner", "Hawthorne",
		"Dragonbane", "Flameborn", "Silverbrook", "Eboncrest", "Swiftarrow",
		"Stormcaller", "Dreadmoor", "Ashenveil", "Ravenshadow", "Frostborn",
		"Darkweaver", "Hallowspire", "Windrider", "Grimsong", "Moonreaver",
		"Emberfall", "Thunderforge", "Ravensoul", "Nightbloom", "Shadowmere",
		"Flamehart", "Doomwhisper", "Ironspire", "Winterborn", "Stormbringer",
	}

	nBig1, _ := rand.Int(rand.Reader, big.NewInt(int64(len(firstNames))))
	nBig2, _ := rand.Int(rand.Reader, big.NewInt(int64(len(lastNames))))

	c.JSON(200, gin.H{
		"First name": firstNames[nBig1.Int64()],
		"Last name":  lastNames[nBig2.Int64()],
	})
}

// /api/joke - Random jokes
func jokeAPI(c *gin.Context) {
	questions := []string{
		"What do you call it when a snowman has a temper tantrum?",
		"Why are elevator jokes so good?",
		"What do you call advice from a cow?",
		"Why are pediatricians always so grumpy?",
		"Why did the scarecrow win an award?",
		"Why did the melon jump into the lake?",
		"What did the duck say when it bought lipstick?",
		"What do you call a pig that does karate?",
		"What has a bed that you can't sleep in?",
		"What is a river's favorite game?",
		"What do you call a bear with no teeth?",
		"Why did the bicycle fall over?",
		"What do you call a fake noodle?",
		"Why did the golfer bring two pairs of pants?",
		"Why don't scientists trust atoms?",
		"What do you call cheese that isn't yours?",
		"Why did the computer go to the doctor?",
		"Why couldn't the leopard play hide and seek?",
		"Apparently, you can't use \"beef stew\" as a password.",
		"Why did the drum take a nap?",
		"Where do hamburgers go dancing?",
		"Why did the tomato turn red?",
		"Why shouldn't you write with a broken pencil?",
		"What do you call two monkeys that share an Amazon account?",
		"Why are teddy bears never hungry?",
	}

	answers := []string{
		"A meltdown.",
		"They work on so many levels.",
		"Beef Tips.",
		"They have little patients.",
		"Because he was outstanding in his field.",
		"It wanted to be a water-melon.",
		`"Put it on my bill."`,
		"A pork chop.",
		"A river.",
		"River-ty.",
		"A gummy bear.",
		"Because it was two-tired.",
		"An impasta.",
		"In case he got a hole in one.",
		"Because they make up everything.",
		"Nacho cheese",
		"Because it had a virus.",
		"Because he was always spotted.",
		"It's not stroganoff.",
		"It was beat.",
		"They go to the meat-ball.",
		"It saw the salad dressing.",
		"Because it's pointless.",
		"Prime mates.",
		"Because they are always stuffed.",
	}

	nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(len(questions))))
	idx := nBig.Int64()

	c.JSON(200, gin.H{
		"Question": questions[idx],
		"Answer":   answers[idx],
	})
}

// /api/insult - Random insults
func insultAPI(c *gin.Context) {
	name := c.Query("name")

	if name != "" {
		personalInsults := []string{
			name + " stinks.",
			name + " is a loser.",
			name + " is a noob.",
			name + " is a nerd.",
		}
		nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(len(personalInsults))))
		c.JSON(200, gin.H{"Insult": personalInsults[nBig.Int64()]})
	} else {
		insults := []string{
			"You stink.",
			"Your short.",
		}
		nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(len(insults))))
		c.JSON(200, gin.H{"Insult": insults[nBig.Int64()]})
	}
}

// /api/uuid - UUID generator
func uuidAPI(c *gin.Context) {
	version := c.Query("v")
	var uuid string
	var versionNum int

	switch version {
	case "7":
		uuid = generateUUIDv7()
		versionNum = 7
	case "8":
		uuid = generateUUIDv8()
		versionNum = 8
	case "1":
		uuid = generateUUIDv1()
		versionNum = 1
	default:
		uuid = generateUUIDv4()
		versionNum = 4
	}

	c.JSON(200, gin.H{
		"uuid":    uuid,
		"version": versionNum,
	})
}

// /api/zawg - Random dog images (HTML display)
func zawgAPI(c *gin.Context) {
	images := []string{
		"/assets/api/zawg/dog1.jpg",
		"/assets/api/zawg/dog2.jpg",
		"/assets/api/zawg/dog3.jpg",
		"/assets/api/zawg/dog4.jpg",
	}

	nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(len(images))))
	imageURL := images[nBig.Int64()]

	html := fmt.Sprintf(`<title>Zawg API</title>
<body style='margin: 0px; height: 100%%; background-color: rgb(14, 14, 14);'>
<img style='display: block;-webkit-user-select: none;margin: auto;background-color: hsl(0, 0%%, 90%%);transition: background-color 300ms;' src='%s' alt='Random Image' />
</body>`, imageURL)

	c.Header("Content-Type", "text/html")
	c.String(200, html)
}

// /api/adam_cairns - Random Adam images (HTML display)
func adamCairnsAPI(c *gin.Context) {
	images := []string{
		"/assets/api/adam/adam1.png",
		"/assets/api/adam/adam2.png",
		"/assets/api/adam/adam3.png",
		"/assets/api/adam/adam4.png",
		"/assets/api/adam/adam5.png",
	}

	nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(len(images))))
	imageURL := images[nBig.Int64()]

	html := fmt.Sprintf(`<title>Adam API</title>
<body style='margin: 0px; height: 100%%; background-color: rgb(14, 14, 14);'>
<img style='display: block;-webkit-user-select: none;margin: auto;background-color: hsl(0, 0%%, 90%%);transition: background-color 300ms;' src='%s' alt='Random Image' />
</body>`, imageURL)

	c.Header("Content-Type", "text/html")
	c.String(200, html)
}

// /api/elot - Random elot images (HTML display)
func elotAPI(c *gin.Context) {
	images := []string{
		"/assets/api/elot/elot1.webp", "/assets/api/elot/elot2.webp", "/assets/api/elot/elot3.webp",
		"/assets/api/elot/elot4.webp", "/assets/api/elot/elot5.webp", "/assets/api/elot/elot6.webp",
		"/assets/api/elot/elot7.webp", "/assets/api/elot/elot8.webp", "/assets/api/elot/elot9.webp",
		"/assets/api/elot/elot10.webp", "/assets/api/elot/elot11.webp", "/assets/api/elot/elot12.webp",
		"/assets/api/elot/elot13.webp", "/assets/api/elot/elot14.webp", "/assets/api/elot/elot15.webp",
		"/assets/api/elot/elot16.webp", "/assets/api/elot/elot17.webp", "/assets/api/elot/elot18.webp",
		"/assets/api/elot/elot19.webp", "/assets/api/elot/elot20.webp", "/assets/api/elot/elot21.webp",
		"/assets/api/elot/elot22.webp", "/assets/api/elot/elot23.webp",
	}

	nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(len(images))))
	imageURL := images[nBig.Int64()]

	html := fmt.Sprintf(`<title>Elot API</title>
<body style='margin: 0px; height: 100%%; background-color: rgb(14, 14, 14);'>
<img style='display: block;-webkit-user-select: none;margin: auto;background-color: hsl(0, 0%%, 90%%);transition: background-color 300ms;' src='%s' alt='Random Image' />
</body>`, imageURL)

	c.Header("Content-Type", "text/html")
	c.String(200, html)
}

// /api/unit_conversion - Unit conversion
func unitConversionAPI(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	if from == "" {
		c.JSON(400, gin.H{"error": "unit to convert from not set."})
		return
	}

	if to == "" {
		c.JSON(400, gin.H{"error": "unit to convert to not set."})
		return
	}

	// Extract value and unit from 'from' parameter
	var value float64
	var firstUnit string
	for i, r := range from {
		if (r >= '0' && r <= '9') || r == '.' {
			continue
		}
		if i > 0 {
			value, _ = strconv.ParseFloat(from[:i], 64)
			firstUnit = from[i:]
			break
		}
	}

	conversionType := firstUnit + to

	var result float64
	switch conversionType {
	case "kgst":
		result = value / 6.35
	case "stkg":
		result = value * 6.35
	case "gbpusd":
		result = math.Round((value*1.3)*100) / 100
	case "usdgbp":
		result = math.Round((value/1.3)*100) / 100
	case "gbpbtc":
		result = math.Round((value*0.000015623346927)*100000000) / 100000000
	case "btcgbp":
		result = math.Round((value/0.000015623346927)*100000000) / 100000000
	case "inchescm":
		result = value * 2.54
	case "cminches":
		result = value / 2.54
	case "ftm":
		result = value * 0.3048
	case "mft":
		result = value / 0.3048
	default:
		c.JSON(400, gin.H{"error": "conversion not supported"})
		return
	}

	c.JSON(200, gin.H{"output": result})
}

// /api/test_post - Simple POST test endpoint
func testPostAPI(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Status(200)
		return
	}

	if c.Request.Method != "POST" {
		c.JSON(405, gin.H{
			"error":           "Method not allowed. Only POST requests are accepted.",
			"allowed_methods": []string{"POST"},
		})
		return
	}

	c.JSON(200, gin.H{"test": "true"})
}

// Helper function to get visitor IP
func getVisitorIP(c *gin.Context) string {
	ipKeys := []string{"CF-Connecting-IP", "X-Client-IP", "X-Forwarded-For", "X-Real-IP"}
	for _, key := range ipKeys {
		ip := c.GetHeader(key)
		if ip != "" {
			if strings.Contains(ip, ",") {
				ip = strings.Split(ip, ",")[0]
			}
			ip = strings.TrimSpace(ip)
			return ip
		}
	}
	return c.ClientIP()
}

// /api/unique_site_visitors - Track and get unique visitor counts
func uniqueSiteVisitorsAPI(c *gin.Context, visitorDB *gorm.DB) {
	if c.Request.Method == "OPTIONS" {
		c.Status(200)
		return
	}

	siteDomain := c.Query("site")
	if siteDomain == "" {
		siteDomain = c.PostForm("site")
	}

	if siteDomain == "" {
		c.JSON(400, gin.H{"error": "Site parameter is required"})
		return
	}

	// Parse domain if full URL provided
	if strings.Contains(siteDomain, "://") {
		parts := strings.Split(siteDomain, "://")
		if len(parts) > 1 {
			siteDomain = strings.Split(parts[1], "/")[0]
		}
	}

	visitorIP := getVisitorIP(c)
	today := time.Now().Format("2006-01-02")

	// Check if visitor already counted today
	var existing UniqueVisitor
	result := visitorDB.Where("site_domain = ? AND visitor_ip = ? AND DATE(visit_date) = ?",
		siteDomain, visitorIP, today).First(&existing)

	if result.Error == gorm.ErrRecordNotFound {
		// Add new visitor
		newVisitor := UniqueVisitor{
			SiteDomain: siteDomain,
			VisitorIP:  visitorIP,
			VisitDate:  time.Now(),
		}
		visitorDB.Create(&newVisitor)
	}

	// Count unique visitors today
	var todayCount int64
	visitorDB.Model(&UniqueVisitor{}).
		Where("site_domain = ? AND DATE(visit_date) = ?", siteDomain, today).
		Distinct("visitor_ip").
		Count(&todayCount)

	// Count total unique visitors
	var totalCount int64
	visitorDB.Model(&UniqueVisitor{}).
		Where("site_domain = ?", siteDomain).
		Distinct("visitor_ip").
		Count(&totalCount)

	c.JSON(200, gin.H{
		"success":               true,
		"site_domain":           siteDomain,
		"unique_visitors_today": todayCount,
		"total_unique_visitors": totalCount,
		"date":                  today,
	})
}

// /api/single_comment - Get/Set single comment per site
func singleCommentAPI(c *gin.Context, visitorDB *gorm.DB) {
	if c.Request.Method == "OPTIONS" {
		c.Status(200)
		return
	}

	siteDomain := c.Query("site")
	if siteDomain == "" {
		var input map[string]interface{}
		if err := c.ShouldBindJSON(&input); err == nil {
			if site, ok := input["site"].(string); ok {
				siteDomain = site
			}
		}
		if siteDomain == "" {
			siteDomain = c.PostForm("site")
		}
	}

	if siteDomain == "" {
		c.JSON(400, gin.H{"error": "Site parameter is required"})
		return
	}

	// Parse domain if full URL provided
	if strings.Contains(siteDomain, "://") {
		parts := strings.Split(siteDomain, "://")
		if len(parts) > 1 {
			siteDomain = strings.Split(parts[1], "/")[0]
		}
	}

	if c.Request.Method == "GET" {
		var comment SiteComment
		result := visitorDB.Where("site_domain = ?", siteDomain).First(&comment)

		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(200, gin.H{
				"success":     true,
				"site_domain": siteDomain,
				"comment":     nil,
				"author":      nil,
				"created_at":  nil,
				"updated_at":  nil,
				"has_comment": false,
			})
			return
		}

		c.JSON(200, gin.H{
			"success":     true,
			"site_domain": siteDomain,
			"comment":     comment.CommentText,
			"author":      comment.AuthorName,
			"created_at":  comment.CreatedAt,
			"updated_at":  comment.UpdatedAt,
			"has_comment": true,
		})

	} else if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		var input map[string]interface{}
		c.ShouldBindJSON(&input)

		commentText, _ := input["comment"].(string)
		if commentText == "" {
			commentText = c.PostForm("comment")
		}

		authorName, _ := input["author"].(string)
		if authorName == "" {
			authorName = c.PostForm("author")
			if authorName == "" {
				authorName = "Anonymous"
			}
		}

		if commentText == "" {
			c.JSON(400, gin.H{"error": "Comment text is required"})
			return
		}

		var comment SiteComment
		result := visitorDB.Where("site_domain = ?", siteDomain).First(&comment)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new comment
			comment = SiteComment{
				SiteDomain:  siteDomain,
				CommentText: commentText,
				AuthorName:  authorName,
			}
			visitorDB.Create(&comment)
		} else {
			// Update existing comment
			visitorDB.Model(&comment).Updates(map[string]interface{}{
				"comment_text": commentText,
				"author_name":  authorName,
			})
		}

		c.JSON(200, gin.H{
			"success":     true,
			"site_domain": siteDomain,
			"comment":     comment.CommentText,
			"author":      comment.AuthorName,
			"created_at":  comment.CreatedAt,
			"updated_at":  comment.UpdatedAt,
		})
	}
}

// /api/username_availability - Check username availability (partial implementation)
func usernameAvailabilityAPI(c *gin.Context) {
	platform := strings.ToLower(c.Query("platform"))
	username := strings.ToLower(c.Query("username"))

	if platform == "" {
		c.JSON(400, gin.H{"error": "platform not set."})
		return
	}

	if username == "" {
		c.JSON(400, gin.H{"error": "username not set."})
		return
	}

	supportedPlatforms := []string{"twitch", "youtube", "twitter", "instagram", "reddit", "tiktok", "facebook"}
	found := false
	for _, p := range supportedPlatforms {
		if p == platform {
			found = true
			break
		}
	}

	if !found {
		c.JSON(400, gin.H{"error": "platform not supported."})
		return
	}

	if platform == "youtube" {
		// Check YouTube username availability
		url := fmt.Sprintf("https://www.youtube.com/@%s", username)
		resp, err := http.Get(url)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to check username"})
			return
		}
		defer resp.Body.Close()

		c.JSON(200, gin.H{
			"username":  username,
			"available": resp.StatusCode == 404,
		})
	} else {
		c.JSON(200, gin.H{"message": "Bro idk tbh."})
	}
}

// /api/qrcode - QR code generator
func qrcodeAPI(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.String(400, "Sorry, but you need to provide a URL.")
		return
	}

	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate QR code"})
		return
	}

	c.Data(200, "image/png", png)
}

// /api/memify - Add text to images
func memifyAPI(c *gin.Context) {
	imageURL := c.Query("imageURL")
	text := c.Query("text")

	if imageURL == "" {
		c.String(400, "Yall get some image URL set")
		return
	}

	if text == "" {
		c.String(400, "Yall get some text set")
		return
	}

	// Download image
	resp, err := http.Get(imageURL)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch image"})
		return
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "Invalid image format"})
		return
	}

	// Get dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	boxHeight := 150
	newHeight := height + boxHeight

	// Create new context
	dc := gg.NewContext(width, newHeight)

	// Draw white rectangle for text at top
	dc.SetRGB(1, 1, 1)
	dc.DrawRectangle(0, 0, float64(width), float64(boxHeight))
	dc.Fill()

	// Draw original image below
	dc.DrawImage(img, 0, boxHeight)

	// Add text
	fontPath := "./api/assets/api/fonts/arial.ttf"
	if err := dc.LoadFontFace(fontPath, 45); err != nil {
		// Fallback if font file not found - use smaller size
		log.Printf("Font not found, using default: %v", err)
	}
	dc.SetRGB(0, 0, 0)
	dc.DrawString(text, 10, 60)

	// Output PNG
	var buf bytes.Buffer
	if err := dc.EncodePNG(&buf); err != nil {
		c.JSON(500, gin.H{"error": "Failed to encode image"})
		return
	}

	c.Data(200, "image/png", buf.Bytes())
}

// /api/scrape_steam - Scrape Steam profile for currently playing game
func scrapeSteamAPI(c *gin.Context) {
	steamProfileUrl := "https://steamcommunity.com/id/Itsbaileyx3525"

	client := &http.Client{}
	req, err := http.NewRequest("GET", steamProfileUrl, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to fetch Steam profile page"})
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse HTML"})
		return
	}

	// Find game info
	gameInfo := doc.Find(".game_info").First()
	if gameInfo.Length() == 0 {
		c.JSON(404, gin.H{"error": "No games found"})
		return
	}

	// Extract game name and app ID
	gameNameLink := gameInfo.Find(".game_name a.whiteLink").First()
	href, _ := gameNameLink.Attr("href")
	gameName := strings.TrimSpace(gameNameLink.Text())

	var appid string
	if strings.Contains(href, "/app/") {
		parts := strings.Split(href, "/app/")
		if len(parts) > 1 {
			appid = strings.Split(parts[1], "/")[0]
		}
	}

	// Extract playtime
	playtimeText := strings.TrimSpace(gameInfo.Find(".game_info_details").First().Text())
	playtimeLines := strings.Split(playtimeText, "\n")
	playtime := ""
	if len(playtimeLines) > 0 {
		playtime = strings.TrimSpace(playtimeLines[0])
	}

	// Extract icon URL
	iconURL, _ := gameInfo.Find(".game_info_cap img.game_capsule").First().Attr("src")

	c.JSON(200, gin.H{
		"appid":     appid,
		"game_name": gameName,
		"playtime":  playtime,
		"icon_url":  iconURL,
	})
}

// /api/steam_market - Steam market data (returns sample data)
func steamMarketAPI(c *gin.Context) {
	// Simplified version - returns sample case prices
	casePrices := map[string]interface{}{
		"Kilowatt Case": map[string]interface{}{
			"min_price":       0.50,
			"max_price":       0.70,
			"mean_price":      0.34,
			"median_price":    0.58,
			"suggested_price": 0.62,
			"quantity":        15420,
		},
		"Revolution Case": map[string]interface{}{
			"min_price":       0.45,
			"max_price":       0.65,
			"mean_price":      0.38,
			"median_price":    0.52,
			"suggested_price": 0.58,
			"quantity":        18750,
		},
	}

	c.JSON(200, casePrices)
}

// Spotify token storage structure
type SpotifyToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

func spotifyAPI(c *gin.Context) {
	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	spotifyRedirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
	tokenFile := "./api/spotify_user_token.json"

	code := c.Query("code")
	if code != "" {
		if handleSpotifyOAuthCallback(code, spotifyClientID, spotifyClientSecret, spotifyRedirectURI, tokenFile) {
			c.JSON(200, gin.H{
				"success": true,
				"message": "Authorization successful! You can now get your currently playing track.",
				"note":    "Call the API without the code parameter to get currently playing data.",
			})
		} else {
			c.JSON(500, gin.H{
				"success": false,
				"error":   "Failed to exchange authorization code for token",
			})
		}
		return
	}

	result := getSpotifyCurrentlyPlaying(spotifyClientID, spotifyClientSecret, spotifyRedirectURI, tokenFile)
	c.JSON(200, result)
}

// Get user access token from file storage
func getSpotifyUserAccessToken(tokenFile string, clientID, clientSecret string) (string, error) {
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", err
	}

	var token SpotifyToken
	if err := json.Unmarshal(data, &token); err != nil {
		return "", err
	}

	if time.Now().Unix() < token.ExpiresAt-300 {
		return token.AccessToken, nil
	}

	return refreshSpotifyAccessToken(token.RefreshToken, clientID, clientSecret, tokenFile)
}

func refreshSpotifyAccessToken(refreshToken, clientID, clientSecret, tokenFile string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	accessToken, _ := result["access_token"].(string)
	newRefreshToken, hasRefresh := result["refresh_token"].(string)
	if !hasRefresh {
		newRefreshToken = refreshToken
	}
	expiresIn, _ := result["expires_in"].(float64)

	token := SpotifyToken{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Unix() + int64(expiresIn),
	}

	tokenData, _ := json.Marshal(token)
	os.WriteFile(tokenFile, tokenData, 0644)

	return accessToken, nil
}

func handleSpotifyOAuthCallback(code, clientID, clientSecret, redirectURI, tokenFile string) bool {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	accessToken, _ := result["access_token"].(string)
	refreshToken, _ := result["refresh_token"].(string)
	expiresIn, _ := result["expires_in"].(float64)

	if accessToken == "" {
		return false
	}

	token := SpotifyToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Unix() + int64(expiresIn),
	}

	tokenData, _ := json.Marshal(token)
	os.WriteFile(tokenFile, tokenData, 0644)

	return true
}

func getSpotifyCurrentlyPlaying(clientID, clientSecret, redirectURI, tokenFile string) map[string]interface{} {
	token, err := getSpotifyUserAccessToken(tokenFile, clientID, clientSecret)

	if err != nil {
		authURL := "https://accounts.spotify.com/authorize?" + url.Values{
			"response_type": {"code"},
			"client_id":     {clientID},
			"scope":         {"user-read-currently-playing user-read-playback-state"},
			"redirect_uri":  {redirectURI},
		}.Encode()

		return map[string]interface{}{
			"success":      false,
			"error":        "Authorization required",
			"auth_url":     authURL,
			"instructions": "Visit the auth_url to authorize, then try again",
		}
	}

	// Make request to Spotify API
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/player/currently-playing", nil)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to create request",
		}
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to fetch currently playing",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		return map[string]interface{}{
			"success": true,
			"playing": false,
			"message": "No track currently playing",
		}
	}

	if resp.StatusCode == 401 {
		os.Remove(tokenFile)
		return map[string]interface{}{
			"success": false,
			"error":   "Token expired, please re-authorize",
		}
	}

	if resp.StatusCode != 200 {
		return map[string]interface{}{
			"success":   false,
			"error":     "Failed to get currently playing track",
			"http_code": resp.StatusCode,
		}
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to parse response",
		}
	}

	item, hasItem := data["item"].(map[string]interface{})
	if !hasItem {
		return map[string]interface{}{
			"success": true,
			"playing": false,
			"message": "No track currently playing",
		}
	}

	// Extract track info
	artists := []string{}
	if artistsArr, ok := item["artists"].([]interface{}); ok {
		for _, a := range artistsArr {
			if artist, ok := a.(map[string]interface{}); ok {
				if name, ok := artist["name"].(string); ok {
					artists = append(artists, name)
				}
			}
		}
	}

	var imageURL string
	if album, ok := item["album"].(map[string]interface{}); ok {
		if images, ok := album["images"].([]interface{}); ok && len(images) > 0 {
			if img, ok := images[0].(map[string]interface{}); ok {
				imageURL, _ = img["url"].(string)
			}
		}
	}

	var spotifyURL string
	if externalURLs, ok := item["external_urls"].(map[string]interface{}); ok {
		spotifyURL, _ = externalURLs["spotify"].(string)
	}

	var albumName string
	if album, ok := item["album"].(map[string]interface{}); ok {
		albumName, _ = album["name"].(string)
	}

	return map[string]interface{}{
		"success":     true,
		"playing":     data["is_playing"],
		"progress_ms": data["progress_ms"],
		"track": map[string]interface{}{
			"name":        item["name"],
			"artists":     artists,
			"album":       albumName,
			"duration_ms": item["duration_ms"],
			"image":       imageURL,
			"spotify_url": spotifyURL,
		},
	}
}

// Register all API routes
func registerAPIRoutes(router *gin.Engine, visitorDB *gorm.DB) {
	registerAPIHandlers := func(group *gin.RouterGroup) {
		group.GET("/sheldon", sheldonAPI)

		group.GET("/fact", factAPI)

		group.GET("/fantasy_name", fantasyNameAPI)

		group.GET("/joke", jokeAPI)

		group.GET("/insult", insultAPI)

		group.GET("/uuid", uuidAPI)

		group.GET("/zawg", zawgAPI)

		group.GET("/adam_cairns", adamCairnsAPI)

		group.GET("/elot", elotAPI)

		group.GET("/unit_conversion", unitConversionAPI)

		group.POST("/test_post", testPostAPI)
		group.OPTIONS("/test_post", testPostAPI)

		group.GET("/username_availability", usernameAvailabilityAPI)

		group.GET("/qrcode", qrcodeAPI)

		group.GET("/memify", memifyAPI)

		group.GET("/scrape_steam", scrapeSteamAPI)

		group.GET("/steam_market", steamMarketAPI)

		group.GET("/spotify", spotifyAPI)

		if visitorDB != nil {
			group.GET("/unique_site_visitors", func(c *gin.Context) {
				uniqueSiteVisitorsAPI(c, visitorDB)
			})
			group.POST("/unique_site_visitors", func(c *gin.Context) {
				uniqueSiteVisitorsAPI(c, visitorDB)
			})
			group.OPTIONS("/unique_site_visitors", func(c *gin.Context) {
				uniqueSiteVisitorsAPI(c, visitorDB)
			})

			group.GET("/single_comment", func(c *gin.Context) {
				singleCommentAPI(c, visitorDB)
			})
			group.POST("/single_comment", func(c *gin.Context) {
				singleCommentAPI(c, visitorDB)
			})
			group.PUT("/single_comment", func(c *gin.Context) {
				singleCommentAPI(c, visitorDB)
			})
			group.OPTIONS("/single_comment", func(c *gin.Context) {
				singleCommentAPI(c, visitorDB)
			})
		}
	}

	apiGroup := router.Group("/api")
	registerAPIHandlers(apiGroup)

	apiSubdomain := router.Group("/")
	apiSubdomain.Use(func(c *gin.Context) {
		if strings.Split(c.Request.Host, ".")[0] != "api" {
			c.File("./public/404.html")
			c.Abort()
			return
		}
		c.Next()
	})
	registerAPIHandlers(apiSubdomain)
}
