# Technical Challenge in TCP/IP, Go, and MySQL

Welcome to the Technical Challenge project, where we dive into the exciting world of TCP/IP, Go, and MySQL. This project demonstrates the integration of various technologies to create a robust data collection and visualization system.

## Technical Requirements

- **TCP/IP Client:** Develop a TCP/IP client to collect data in a specified format.
- **Go Server:** Implement a Go server to receive data and store it in a MySQL database.
- **Go Client:** Create a Go client to fetch data from the server and plot it on a live and playback map.
- **Dashboard:** Build a dashboard using Go and MySQL to display relevant information.

## Technologies Used

- **WebSocket:** For real-time communication.
- **TCP/IP:** For data collection.
- **Google Maps API:** For map visualization.
- **MySQL:** For data storage.
- **Postman:** For testing endpoints.

## Setup Instructions

### Start the Project

Run the following command to start the Go server:

    ```bash
    go run main.go
    ```

### Load the Application

Open your browser and navigate to:

    ```text
    http://localhost:8282/
    ```

### Connect WebSocket Client

Use Postman to connect to the WebSocket server with the following URL:

    ```text
    ws://localhost:8282/ws
    ```

### Send Data

In Postman, send a WebSocket message with the following format:

    ```json
    {
      "header": "header_007",
      "imei": "678901234567890",
      "packet_type": "update",
      "time": "2024-08-04T18:00:00Z",
      "lat": 28.6139,
      "direction_lat": "N",
      "lng": 77.2090,
      "direction_lng": "E",
      "date": "2024-08-04",
      "checksum": "checksum_value_6"
    }
    ```

### Visualize Data

- **Marked Locations:** View the marked locations in your browser.
- **Add More Points:** Send additional data through WebSocket to mark new locations.
- **Clear Points:** Use the "Clear" button in the browser to reset and start marking new points.

### Check Dashboard

- **Specific Device Details:** Get detailed information for a specific device using its IMEI number:

    ```text
    http://localhost:8282/dashboard/{imei}
    ```

- **Main Dashboard:** View overall dashboard details:

    ```text
    http://localhost:8282/dashboard
    ```

## Conclusion

This project illustrates a seamless integration of various technologies to handle real-time data, provide insightful visualizations, and offer comprehensive data analysis. Enjoy exploring the capabilities and features of this technical challenge!
