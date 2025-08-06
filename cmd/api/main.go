package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/samrityy/go-oauth/internal/app"
	"github.com/samrityy/go-oauth/internal/db"
	"golang.org/x/oauth2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() {
	godotenv.Load()

	database, err := db.DB()
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		os.Exit(1)
	}
	defer database.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	tmpl := template.Must(template.New("").Funcs(
		template.FuncMap{
			"join":  strings.Join,
			"title": cases.Title(language.English).String,
		},
	).ParseFiles("internal/templates/signin.html", "internal/templates/index.html"))

	appInstance := &app.App{
		Logger:   logger,
		Template: tmpl,
		DB:       database,
		OAuthConfigs: map[string]*oauth2.Config{
			"github": {
				ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
				ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
				RedirectURL:  "https://go-oauth-production.up.railway.app/oauth2/callback/github",
				Scopes:       []string{"read:user", "user:email"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://github.com/login/oauth/authorize",
					TokenURL: "https://github.com/login/oauth/access_token",
				},
			},
			"facebook": {
				ClientID:     os.Getenv("FACEBOOK_CLIENT_ID"),
				ClientSecret: os.Getenv("FACEBOOK_CLIENT_SECRET"),
				RedirectURL:  "http://localhost:3000/oauth2/callback/facebook",
				Scopes:       []string{"email", "public_profile"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://www.facebook.com/v10.0/dialog/oauth",
					TokenURL: "https://graph.facebook.com/v10.0/oauth/access_token",
				},
			},
			"google": {
				ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
				RedirectURL:  "http://localhost:3000/oauth2/callback/google",
				Scopes:       []string{"openid", "email", "profile"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://accounts.google.com/o/oauth2/auth",
					TokenURL: "https://oauth2.googleapis.com/token",
				},
			},
			"instagram": {
				ClientID:     os.Getenv("INSTAGRAM_CLIENT_ID"),
				ClientSecret: os.Getenv("INSTAGRAM_CLIENT_SECRET"),
				RedirectURL:  "http://localhost:3000/oauth2/callback/instagram",
				Scopes:       []string{"user_profile", "user_media"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://api.instagram.com/oauth/authorize",
					TokenURL: "https://api.instagram.com/oauth/access_token",
				},
			},
		},
	}

	handler := appInstance.SetupRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // fallback for local development
	}

	server := http.Server{
		Addr:    ":" + port,
		Handler:handler, // replace with your actual handler
	}

	logger.Info("starting server", "addr", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}
