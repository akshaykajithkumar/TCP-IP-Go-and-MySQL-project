package models

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MarkingRecord struct {
	gorm.Model
	Header       string  `json:"header"`
	Imei         string  `json:"imei"`
	PacketType   string  `json:"packet_type"`
	Time         string  `json:"time"`
	Lat          float64 `json:"lat"`
	DirectionLat string  `json:"direction_lat"`
	Lng          float64 `json:"lng"`
	DirectionLng string  `json:"direction_lng"`
	Date         string  `json:"date"`
	Checksum     string  `json:"checksum"`
}

type TotalDistance struct {
	gorm.Model
	Imei          string  `json:"imei"`
	TotalDistance float64 `json:"total_distance"`
}

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := "akshay:Admin123!@tcp(localhost:3306)/geo_tracking?charset=utf8mb4&parseTime=True&loc=Local"

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := DB.AutoMigrate(&MarkingRecord{}, &TotalDistance{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
}
