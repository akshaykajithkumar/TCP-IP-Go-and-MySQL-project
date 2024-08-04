package tcp_client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

const (
	serverAddr = "localhost:8282"
	tcpAddr    = "localhost:12347" // Change to your TCP server address
)

func Start() error {
	// Connect to TCP server
	conn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to TCP server: %v", err)
	}
	defer conn.Close()

	// Connect to WebSocket server
	ws, _, err := websocket.DefaultDialer.Dial("ws://localhost:8282/ws", nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket server: %v", err)
	}
	defer ws.Close()

	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read from TCP server: %v", err)
		}

		// Process message
		parsedMessage, err := parseMessage(message)
		if err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Send parsed message to WebSocket server
		if err := ws.WriteJSON(parsedMessage); err != nil {
			log.Printf("Error sending message to WebSocket server: %v", err)
		}
	}
}

func parseMessage(message string) (map[string]interface{}, error) {
	// Remove trailing '#'
	if strings.HasSuffix(message, "#") {
		message = strings.TrimSuffix(message, "#")
	}

	// Split message by commas
	parts := strings.Split(message, ",")
	if len(parts) < 10 {
		return nil, fmt.Errorf("invalid message format")
	}

	// Convert latitude and longitude to float64
	lat, err := strconv.ParseFloat(parts[4], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude format: %v", err)
	}
	lng, err := strconv.ParseFloat(parts[6], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude format: %v", err)
	}

	return map[string]interface{}{
		"header":        parts[0],
		"imei":          parts[1],
		"packet_type":   parts[2],
		"time":          parts[3],
		"lat":           lat,
		"direction_lat": parts[5],
		"lng":           lng,
		"direction_lng": parts[7],
		"date":          parts[8],
		"checksum":      parts[9],
	}, nil
}
