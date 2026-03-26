// backend/main.go
package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kerke/cronjob-panel/auth"
	"github.com/kerke/cronjob-panel/db"
	"github.com/kerke/cronjob-panel/router"
)

func main() {
	database, err := db.InitDB("data/cronjob-panel.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	discordCfg := auth.DiscordConfig{
		ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("DISCORD_REDIRECT_URI"),
	}

	r := gin.Default()
	router.Setup(r, database, discordCfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
