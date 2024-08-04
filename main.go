package main

import (
	"geo-tracker/handlers"
	"geo-tracker/models"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Database connection setup
	dsn := "akshay:Admin123!@tcp(localhost:3306)/geo_tracking?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Set up the database connection for the models package
	models.DB = db

	// Auto-migrate models
	err = db.AutoMigrate(&models.MarkingRecord{}, &models.TotalDistance{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Create a new router
	r := mux.NewRouter()

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	// Routes
	r.HandleFunc("/", handlers.ServeHTML)
	r.HandleFunc("/ws", handlers.HandleWebSocket)
	r.HandleFunc("/clear", handlers.HandleClear).Methods("POST")

	r.HandleFunc("/total_distance/{imei}", handlers.HandleDashboardByIMEI).Methods("GET")
	r.HandleFunc("/dashboard", handlers.HandleDashboard).Methods("GET")

	// Start server
	serverAddr := ":8282"
	log.Printf("Server listening on %s\n", serverAddr)
	if err := http.ListenAndServe(serverAddr, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
