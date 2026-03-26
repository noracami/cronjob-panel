// backend/main.go
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kerke/cronjob-panel/auth"
	"github.com/kerke/cronjob-panel/db"
	"github.com/kerke/cronjob-panel/router"
)

//go:embed static/*
var staticFS embed.FS

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

	// Serve frontend static files
	frontendFS, _ := fs.Sub(staticFS, "static")
	r.NoRoute(gin.WrapH(http.FileServer(http.FS(frontendFS))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
