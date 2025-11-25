package main

import (
	"log"
	"net/http"
	"os"

	"example.com/m/apis/findplaces"
	"example.com/m/auth"
	"example.com/m/auth/login"
	"example.com/m/auth/signup"
	"example.com/m/db"
	"example.com/m/middleware"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file not found or could not be loaded")
	}

	// Check Geoapify API key
	if os.Getenv("GEOAPIFY_API_KEY") == "" {
		log.Fatal("❌ GEOAPIFY_API_KEY missing in environment")
	}

	// ✅ Initialize PostgreSQL
	db.InitDB()

	// ✅ Initialize Redis
	// findplaces.InitRedis()

	mux := http.NewServeMux()

	// Authentication routes
	mux.HandleFunc("/login", login.Handler)
	mux.HandleFunc("/signup", signup.Handler)
	mux.HandleFunc("/refresh", auth.RefreshHandler)

	// ✅ Find Places API route (protected by middleware)
	mux.Handle("/findplaces", middleware.Auth(http.HandlerFunc(findplaces.FindPlacesHandler)))

	log.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
