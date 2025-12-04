package main

import (
	"log"
	"net/http"
	"os"

	"example.com/m/apis/findplaces"
	"example.com/m/apis/users"
	"example.com/m/auth"
	"example.com/m/auth/login"
	"example.com/m/auth/signup"
	"example.com/m/db"
	_ "example.com/m/docs" // Required for swagger
	"example.com/m/middleware"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/joho/godotenv"
)

// @title           Backend API
// @version         1.0
// @description     This is the backend API for the project.
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file not found or could not be loaded")
	}

	if err := middleware.InitCloudinary(); err != nil {
		log.Fatalf("Failed to init Cloudinary: %v", err)
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

	// Swagger
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Authentication routes
	mux.HandleFunc("/login", login.Handler)
	mux.HandleFunc("/signup", signup.Handler)
	mux.HandleFunc("/refresh", auth.RefreshHandler)

	// ✅ Find Places API route (protected by middleware)
	mux.Handle("/findplaces", middleware.Auth(http.HandlerFunc(findplaces.Handler)))

	mux.Handle("/api/user/", http.StripPrefix("/api/user", users.Routes()))

	log.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
