package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"geo-tracker/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

var (
	upgrader      = websocket.Upgrader{}
	connections   = []*websocket.Conn{}
	locations     = []Location{} // Initialize locations here
	totalDistance float64
	mu            sync.Mutex
)

func ServeHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	mu.Lock()
	connections = append(connections, conn)
	mu.Unlock()

	var previousLocation *Location

	for {
		var message map[string]interface{}
		err := conn.ReadJSON(&message)
		if err != nil {
			mu.Lock()
			removeConnection(conn)
			mu.Unlock()
			break
		}

		lat, ok := message["lat"].(float64)
		if !ok {
			log.Printf("Invalid latitude format")
			continue
		}
		lng, ok := message["lng"].(float64)
		if !ok {
			log.Printf("Invalid longitude format")
			continue
		}

		loc := Location{Lat: lat, Lng: lng}

		if previousLocation != nil {
			distance := calculateDistance(previousLocation.Lat, previousLocation.Lng, loc.Lat, loc.Lng)
			totalDistance += distance
			log.Printf("New Location: %v, Total Distance: %v km\n", loc, totalDistance)
		}

		previousLocation = &loc

		mu.Lock()
		locations = append(locations, loc)
		mu.Unlock()

		log.Printf("Broadcasting Location: %v", loc) // Debug logging
		broadcastLocation(loc)
	}
}

func HandleClear(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// Save individual marking records
	for _, loc := range locations {
		record := models.MarkingRecord{
			Header:       "header_007", // Example data; replace with actual
			Imei:         "678901234567890",
			PacketType:   "update",
			Time:         "2024-08-04T18:00:00Z",
			Lat:          loc.Lat,
			DirectionLat: "N",
			Lng:          loc.Lng,
			DirectionLng: "E",
			Date:         "2024-08-04",
			Checksum:     "checksum_value_6",
		}

		if err := models.DB.Create(&record).Error; err != nil {
			http.Error(w, "Error saving marking record to database", http.StatusInternalServerError)
			log.Printf("Error saving marking record to database: %v", err)
			return
		}
	}

	// Save total distance
	totalDistanceRecord := models.TotalDistance{
		Imei:          "678901234567890", // Example data; replace with actual
		TotalDistance: totalDistance,
	}

	if err := models.DB.Create(&totalDistanceRecord).Error; err != nil {
		http.Error(w, "Error saving total distance to database", http.StatusInternalServerError)
		log.Printf("Error saving total distance to database: %v", err)
		return
	}

	// Reset data
	totalDistance = 0
	locations = nil

	// Notify all WebSocket clients to clear their data
	for _, conn := range connections {
		err := conn.WriteJSON(map[string]interface{}{
			"type": "clear",
		})
		if err != nil {
			log.Printf("Error sending clear message: %v", err)
		}
	}

	fmt.Fprintln(w, "Distance data saved and tracking cleared")
}

func HandleSave(w http.ResponseWriter, r *http.Request) {
	var loc Location
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Save location to database
	record := models.MarkingRecord{
		Header:       "",
		Imei:         "",
		PacketType:   "",
		Time:         "",
		Lat:          loc.Lat,
		DirectionLat: "",
		Lng:          loc.Lng,
		DirectionLng: "",
		Date:         "",
		Checksum:     "",
	}

	if err := models.DB.Create(&record).Error; err != nil {
		http.Error(w, "Error saving data to database", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Data saved successfully")
}

func removeConnection(conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()
	for i, c := range connections {
		if c == conn {
			connections = append(connections[:i], connections[i+1:]...)
			break
		}
	}
}

func broadcastLocation(loc Location) {
	mu.Lock()
	defer mu.Unlock()
	for _, conn := range connections {
		err := conn.WriteJSON(loc)
		if err != nil {
			removeConnection(conn)
		}
	}
}

func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371 // Radius of the Earth in km
	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLng := (lng2 - lng1) * (math.Pi / 180)
	a := (math.Sin(dLat/2) * math.Sin(dLat/2)) + (math.Cos(lat1*(math.Pi/180)) * math.Cos(lat2*(math.Pi/180)) * math.Sin(dLng/2) * math.Sin(dLng/2))
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// HandleDashboardByIMEI handles the GET /dashboard/{imei} request
func HandleDashboardByIMEI(w http.ResponseWriter, r *http.Request) {
	imei := mux.Vars(r)["imei"]

	// Fetch total distance for the specified IMEI
	var totalDistanceForIMEI float64
	if err := models.DB.Model(&models.TotalDistance{}).Where("imei = ?", imei).Select("SUM(total_distance)").Row().Scan(&totalDistanceForIMEI); err != nil {
		http.Error(w, "IMEI not found", http.StatusNotFound)
		return
	}

	// Calculate the total distance for the past day for the specified IMEI
	var totalDistancePastDay float64
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	if err := models.DB.Model(&models.TotalDistance{}).
		Where("imei = ? AND created_at > ?", imei, oneDayAgo).
		Select("SUM(total_distance)").
		Row().
		Scan(&totalDistancePastDay); err != nil {
		http.Error(w, "Error fetching recent data", http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"total_distance_for_imei": map[string]interface{}{
			"imei":           imei,
			"total_distance": totalDistanceForIMEI,
		},
		"total_distance_past_day": totalDistancePastDay,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleDashboard handles the GET /dashboard request
func HandleDashboard(w http.ResponseWriter, r *http.Request) {
	var allDistances []models.TotalDistance
	if err := models.DB.Find(&allDistances).Error; err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}

	// Calculate the most traveled device overall
	deviceTotals := make(map[string]float64)
	for _, distance := range allDistances {
		deviceTotals[distance.Imei] += distance.TotalDistance
	}

	var mostTraveledDevice string
	var highestDistance float64
	for device, totalDistance := range deviceTotals {
		if totalDistance > highestDistance {
			mostTraveledDevice = device
			highestDistance = totalDistance
		}
	}

	// Calculate the most traveled device in the past day
	var recentDistances []models.TotalDistance
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	if err := models.DB.Where("created_at > ?", oneDayAgo).Find(&recentDistances).Error; err != nil {
		http.Error(w, "Error fetching recent data", http.StatusInternalServerError)
		return
	}

	recentDeviceTotals := make(map[string]float64)
	for _, distance := range recentDistances {
		recentDeviceTotals[distance.Imei] += distance.TotalDistance
	}

	var mostTraveledDevicePastDay string
	var highestDistancePastDay float64
	for device, totalDistance := range recentDeviceTotals {
		if totalDistance > highestDistancePastDay {
			mostTraveledDevicePastDay = device
			highestDistancePastDay = totalDistance
		}
	}

	// Prepare the response
	response := map[string]interface{}{
		"total_distances": deviceTotals,
		"most_traveled_device": map[string]interface{}{
			"imei":     mostTraveledDevice,
			"distance": highestDistance,
		},
		"most_traveled_device_past_day": map[string]interface{}{
			"imei":     mostTraveledDevicePastDay,
			"distance": highestDistancePastDay,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
