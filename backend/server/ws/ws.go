package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow all origins for dev
}

var Clients = make(map[*websocket.Conn]bool)

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	log.Println("Client connected:", conn.RemoteAddr())
	Clients[conn] = true

	// Optional: read loop to handle incoming messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket disconnected:", err)
			delete(Clients, conn)
			conn.Close()
			break
		}
	}
}

// Send message to all connected clients
func BroadcastJSON(data interface{}) {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Println("Failed to marshal message:", err)
		return
	}
	for client := range Clients {
		err := client.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("Write error:", err)
			client.Close()
			delete(Clients, client)
		}
	}
}
