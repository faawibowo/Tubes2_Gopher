// ──────────────────────────────────────────────────────────────
// cmd/server/main.go
// ──────────────────────────────────────────────────────────────
package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/faawibowo/Tubes2_Gopher/configs"
	"github.com/faawibowo/Tubes2_Gopher/internal/algorithm/bfs"
	"github.com/faawibowo/Tubes2_Gopher/internal/algorithm/dfs"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/faawibowo/Tubes2_Gopher/server/ws"

	"github.com/gorilla/handlers"
)

const serverPort = ":8080"

var graphMap map[string]*Graph.ElementGraph

// ──────────────────────────────────────────────────────────────
// shared request struct
// ──────────────────────────────────────────────────────────────
type request struct {
	Target          string `json:"target"`
	MaxPaths        int    `json:"maxPaths"`
	DelayMs         int    `json:"delay"`       // optional for REST; WS always sends it
	ElementJSONPath string `json:"elementPath"` // only /api/config
}

func main() {
	// ── REST routes
	http.HandleFunc("/api/config", withCORS(configHandler))
	http.HandleFunc("/api/bfs", withCORS(bfsHandler))
	http.HandleFunc("/api/dfs", withCORS(dfsHandler))
	http.HandleFunc("/api/shortest/bfs", withCORS(shortestBFSHandler))
	http.HandleFunc("/api/shortest/dfs", withCORS(shortestDFSHandler))
	// ── WS route
	http.HandleFunc("/ws/bfs", ws.HandleBFS)
	http.HandleFunc("/ws/dfs", ws.HandleDFS)
	http.HandleFunc("/ws/shortest/bfs", ws.HandleShortestBFS)
	http.HandleFunc("/ws/shortest/dfs", ws.HandleShortestDFS)

	// ── CORS middleware (allows your Next/React dev-server on :3000)
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:3000"}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
		handlers.AllowCredentials(),
	)

	log.Printf("▶ server running on http://localhost%s\n", serverPort)
	log.Fatal(http.ListenAndServe(serverPort, cors(http.DefaultServeMux)))
}

// ──────────────────────────────────────────────────────────────
// /api/config ‒ POST {elementPath: ".../elements.json"}
// ──────────────────────────────────────────────────────────────
func configHandler(w http.ResponseWriter, r *http.Request) {
	var req request
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	elements, err := configs.LoadElementsJSON(req.ElementJSONPath)
	if err != nil {
		http.Error(w, "failed to load elements: "+err.Error(), http.StatusInternalServerError)
		return
	}

	scraped := configs.ToScrapedMapWithTiers(elements)
	tierMap := make(map[string]int)
	ingredientMap := make(map[string][]string)
	for name, data := range scraped {
		tierMap[name] = data.Tier
		ingredientMap[name] = data.Ingredients
	}

	graphMap = Graph.CreateElementGraphMap(ingredientMap, tierMap)
	ws.InjectGraph(graphMap) // hand a copy to the WS package

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("graph loaded successfully"))
}

// ──────────────────────────────────────────────────────────────
// /api/bfs – POST {target, maxPaths, delay}
// ──────────────────────────────────────────────────────────────
func bfsHandler(w http.ResponseWriter, r *http.Request) {
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	target := graphMap[req.Target]
	if target == nil {
		http.Error(w, "target not found / graph not loaded", http.StatusBadRequest)
		return
	}

	result := bfs.BuildRecipeTree(
		target,
		graphMap,
		req.MaxPaths,
		time.Duration(req.DelayMs)*time.Millisecond,
		nil, // no WebSocket streaming
	)
	respondJSON(w, result)
}

// ──────────────────────────────────────────────────────────────
func dfsHandler(w http.ResponseWriter, r *http.Request) {
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	target := graphMap[req.Target]
	if target == nil {
		http.Error(w, "target not found / graph not loaded", http.StatusBadRequest)
		return
	}

	result := dfs.BuildRecipeTree(
		target,
		graphMap,
		req.MaxPaths,
		time.Duration(req.DelayMs)*time.Millisecond,
		nil,
	)

	respondJSON(w, result)
}

// ──────────────────────────────────────────────────────────────
func shortestBFSHandler(w http.ResponseWriter, r *http.Request) {
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	target := graphMap[req.Target]
	if target == nil {
		http.Error(w, "target not found / graph not loaded", http.StatusBadRequest)
		return
	}

	result := bfs.FindShortestPath(
		target,
		graphMap,
		time.Duration(req.DelayMs)*time.Millisecond,
		nil, // no WebSocket streaming
	)
	respondJSON(w, result)
}

func shortestDFSHandler(w http.ResponseWriter, r *http.Request) {
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	target := graphMap[req.Target]
	if target == nil {
		http.Error(w, "target not found / graph not loaded", http.StatusBadRequest)
		return
	}

	result := dfs.FindShortestPath(
		target,
		graphMap,
		time.Duration(req.DelayMs)*time.Millisecond,
		nil, // 🔸 No WebSocket updates
	)

	respondJSON(w, result)
}

// ──────────────────────────────────────────────────────────────
func respondJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("JSON encode error:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// tiny helper
func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}
