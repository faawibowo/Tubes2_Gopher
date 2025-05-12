package ws

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/faawibowo/Tubes2_Gopher/internal/algorithm/bfs"
	"github.com/faawibowo/Tubes2_Gopher/internal/algorithm/bidirectional"
	"github.com/faawibowo/Tubes2_Gopher/internal/algorithm/dfs"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/gorilla/websocket"
)

type Request struct {
	Target   string `json:"target"`
	MaxPaths int    `json:"maxPaths"`
	DelayMs  int    `json:"delay"`
}

var (
	graphMap map[string]*Graph.ElementGraph
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "http://localhost:3000"
	}}
)

func InjectGraph(m map[string]*Graph.ElementGraph) { graphMap = m }

// helper to stream TreeUpdates
func streamUpdates(conn *websocket.Conn, updates <-chan *bidirectional.TreeUpdate, done <-chan struct{}) {
	for {
		select {
		case update := <-updates:
			if err := conn.WriteJSON(update); err != nil {
				log.Println("WebSocket write error:", err)
				return
			}
		case <-done:
			return
		}
	}
}

func decodeRequest(conn *websocket.Conn) (Request, *Graph.ElementGraph, bool) {
	var req Request
	if err := conn.ReadJSON(&req); err != nil {
		log.Println("WebSocket read error:", err)
		return req, nil, false
	}

	target := graphMap[normalizeName(req.Target)]
	if target == nil {
		log.Println("Sending JSON error: Target not found")

		_ = conn.WriteJSON(map[string]string{
			"error": "Target not found",
		})

		time.Sleep(100 * time.Millisecond)

		conn.Close()

		return req, nil, false
	}

	return req, target, true
}

// ---------------------- BFS BUILD TREE -----------------------
func HandleBFS(w http.ResponseWriter, r *http.Request) {
	conn, ok := upgradeAndCheck(w, r)
	if !ok {
		return
	}
	defer conn.Close()

	req, target, ok := decodeRequest(conn)
	if !ok {
		return
	}

	updates := make(chan bfs.BFSResult, 1)
	done := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case update := <-updates:
				if err := conn.WriteJSON(update); err != nil {
					log.Println("WebSocket write error:", err)
					return
				}
			case <-done:
				return
			}
		}
	}()

	// 	func FindMultiplePath(
	// 	target *Graph.ElementGraph,
	// 	maxPaths int,
	// 	delay time.Duration,
	// 	updates chan<- *BFSResult,
	// 	_ map[string]*Graph.ElementGraph,
	// )

	result := bfs.FindMultiplePath(
		target,
		req.MaxPaths,
		time.Duration(req.DelayMs)*time.Millisecond,
		updates,
		graphMap,
	)

	if updates != nil {
		log.Println("Multithread BFS trying for the last update: ", normalizeName(req.Target))
		time.Sleep(time.Duration(req.DelayMs) * time.Millisecond)
		updates <- result
	}
	close(done)
	wg.Wait()
	conn.Close()
}

// ---------------------- DFS BUILD TREE -----------------------
func HandleDFS(w http.ResponseWriter, r *http.Request) {
	conn, ok := upgradeAndCheck(w, r)
	if !ok {
		return
	}
	defer conn.Close()

	req, target, ok := decodeRequest(conn)
	if !ok {
		return
	}

	updates := make(chan *dfs.DFSResult, 1)
	done := make(chan struct{})
	var wg sync.WaitGroup

	// stream updates
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case update := <-updates:
				if err := conn.WriteJSON(update); err != nil {
					log.Println("WebSocket write error:", err)
					return
				}
			case <-done:
				return
			}
		}
	}()

	// streaming build
	result := dfs.FindMultiplePathDree(
		target,
		req.MaxPaths,
		graphMap,
		time.Duration(req.DelayMs)*time.Millisecond,
		// updates,
	)

	if updates != nil {
		log.Println("Multithread DFS trying for the last update: ", normalizeName(req.Target))
		time.Sleep(time.Duration(req.DelayMs) * time.Millisecond)
		updates <- result
	}
	close(done)
	wg.Wait()
	conn.Close()
}

// ---------------------- BFS SHORTEST PATH -----------------------
func HandleShortestBFS(w http.ResponseWriter, r *http.Request) {
	conn, ok := upgradeAndCheck(w, r)
	if !ok {
		return
	}
	defer conn.Close()

	req, target, ok := decodeRequest(conn)
	if !ok {
		return
	}

	updates := make(chan bfs.BFSResult, 1)
	done := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case update := <-updates:
				if err := conn.WriteJSON(update); err != nil {
					log.Println("WebSocket write error:", err)
					return
				}
			case <-done:
				return
			}
		}
	}()

	result := bfs.FindFirstPath(
		target,
		time.Duration(req.DelayMs)*time.Millisecond,
		updates,
	)

	if updates != nil {
		log.Println("Shortest BFS trying for the last update: ", normalizeName(req.Target))
		time.Sleep(time.Duration(req.DelayMs) * time.Millisecond)
		updates <- result
	}
	close(done)
	wg.Wait()
	conn.Close()
}

// ---------------------- DFS SHORTEST PATH -----------------------
func HandleShortestDFS(w http.ResponseWriter, r *http.Request) {
	conn, ok := upgradeAndCheck(w, r)
	if !ok {
		return
	}
	defer conn.Close()

	req, target, ok := decodeRequest(conn)
	if !ok {
		return
	}

	updates := make(chan dfs.DFSResult, 1)
	done := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case update := <-updates:
				if err := conn.WriteJSON(update); err != nil {
					log.Println("WebSocket write error:", err)
					return
				}
			case <-done:
				return
			}
		}
	}()

	// Run algorithm
	result := dfs.FindShortestPath(
		target,
		graphMap,
		time.Duration(req.DelayMs)*time.Millisecond,
		updates,
	)

	if updates != nil {
		log.Println("Shortest DFS trying for the last update: ", normalizeName(req.Target))
		time.Sleep(time.Duration(req.DelayMs) * time.Millisecond)
		updates <- result
	}
	close(done)
	wg.Wait()
	conn.Close()
}

// ---------------------- COMMON UPGRADE -----------------------
func upgradeAndCheck(w http.ResponseWriter, r *http.Request) (*websocket.Conn, bool) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return nil, false
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return nil, false
	}
	return conn, true
}

func normalizeName(s string) string {
	if len(s) == 0 {
		return s
	}
	s = strings.ToLower(s)
	return strings.ToUpper(s[:1]) + s[1:]
}
