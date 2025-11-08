package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const port = "8092"

var (
	logger   *log.Logger
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for demo
		},
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.RWMutex
)

type InventoryUpdate struct {
	ProductID   string    `json:"product_id"`
	Warehouse   string    `json:"warehouse"`
	Quantity    int       `json:"quantity"`
	Change      int       `json:"change"`
	Timestamp   time.Time `json:"timestamp"`
	UpdateType  string    `json:"update_type"` // sale, restock, adjustment
}

type Product struct {
	ProductID    string    `json:"product_id"`
	Name         string    `json:"name"`
	TotalStock   int       `json:"total_stock"`
	ReservedStock int      `json:"reserved_stock"`
	AvailableStock int    `json:"available_stock"`
	Warehouses   map[string]int `json:"warehouses"`
	LastUpdated  time.Time `json:"last_updated"`
}

var (
	inventory   = make(map[string]*Product)
	inventoryMu sync.RWMutex
)

func init() {
	logger = log.New(os.Stdout, "[INVENTORY] ", log.LstdFlags)

	// Initialize sample inventory
	initializeSampleInventory()
}

func main() {
	logger.Printf("Starting Real-time Inventory Service on port %s", port)

	router := mux.NewRouter()

	// REST endpoints
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")
	router.HandleFunc("/", rootHandler).Methods("GET")
	router.HandleFunc("/inventory", getAllInventoryHandler).Methods("GET")
	router.HandleFunc("/inventory/{product_id}", getProductInventoryHandler).Methods("GET")
	router.HandleFunc("/inventory/{product_id}/update", updateInventoryHandler).Methods("POST")
	router.HandleFunc("/inventory/{product_id}/reserve", reserveInventoryHandler).Methods("POST")

	// WebSocket endpoint
	router.HandleFunc("/ws", websocketHandler)

	logger.Fatal(http.ListenAndServe(":"+port, router))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "inventory-service",
		"timestamp": time.Now().UTC(),
		"clients":   len(clients),
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"service": "Real-time Inventory Service",
		"version": "1.0.0",
		"features": []string{
			"Real-time stock updates",
			"WebSocket notifications",
			"Multi-warehouse support",
			"Stock reservation",
		},
	})
}

func getAllInventoryHandler(w http.ResponseWriter, r *http.Request) {
	inventoryMu.RLock()
	defer inventoryMu.RUnlock()

	products := make([]*Product, 0, len(inventory))
	for _, product := range inventory {
		products = append(products, product)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"products": products,
		"total":    len(products),
	})
}

func getProductInventoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["product_id"]

	inventoryMu.RLock()
	product, exists := inventory[productID]
	inventoryMu.RUnlock()

	if !exists {
		respondError(w, http.StatusNotFound, "Product not found")
		return
	}

	respondJSON(w, http.StatusOK, product)
}

func updateInventoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["product_id"]

	var req struct {
		Warehouse  string `json:"warehouse"`
		Change     int    `json:"change"`
		UpdateType string `json:"update_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	inventoryMu.Lock()
	product, exists := inventory[productID]
	if !exists {
		inventoryMu.Unlock()
		respondError(w, http.StatusNotFound, "Product not found")
		return
	}

	// Update warehouse stock
	if product.Warehouses == nil {
		product.Warehouses = make(map[string]int)
	}

	oldQuantity := product.Warehouses[req.Warehouse]
	newQuantity := oldQuantity + req.Change
	if newQuantity < 0 {
		newQuantity = 0
	}

	product.Warehouses[req.Warehouse] = newQuantity
	product.TotalStock = calculateTotalStock(product.Warehouses)
	product.AvailableStock = product.TotalStock - product.ReservedStock
	product.LastUpdated = time.Now()

	inventoryMu.Unlock()

	// Broadcast update to WebSocket clients
	update := InventoryUpdate{
		ProductID:  productID,
		Warehouse:  req.Warehouse,
		Quantity:   newQuantity,
		Change:     req.Change,
		Timestamp:  time.Now(),
		UpdateType: req.UpdateType,
	}

	broadcastUpdate(update)

	respondJSON(w, http.StatusOK, product)
}

func reserveInventoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["product_id"]

	var req struct {
		Quantity int `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	inventoryMu.Lock()
	defer inventoryMu.Unlock()

	product, exists := inventory[productID]
	if !exists {
		respondError(w, http.StatusNotFound, "Product not found")
		return
	}

	if product.AvailableStock < req.Quantity {
		respondError(w, http.StatusConflict, "Insufficient stock")
		return
	}

	product.ReservedStock += req.Quantity
	product.AvailableStock = product.TotalStock - product.ReservedStock

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"reserved":       req.Quantity,
		"available_stock": product.AvailableStock,
	})
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Printf("WebSocket upgrade error: %v", err)
		return
	}

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	logger.Printf("New WebSocket client connected. Total clients: %d", len(clients))

	// Send initial inventory snapshot
	inventoryMu.RLock()
	snapshot := make([]*Product, 0, len(inventory))
	for _, product := range inventory {
		snapshot = append(snapshot, product)
	}
	inventoryMu.RUnlock()

	if err := conn.WriteJSON(map[string]interface{}{
		"type": "snapshot",
		"data": snapshot,
	}); err != nil {
		logger.Printf("Error sending snapshot: %v", err)
	}

	// Keep connection alive and handle disconnect
	go func() {
		defer func() {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			conn.Close()
			logger.Printf("Client disconnected. Total clients: %d", len(clients))
		}()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

func broadcastUpdate(update InventoryUpdate) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	message := map[string]interface{}{
		"type": "update",
		"data": update,
	}

	for client := range clients {
		if err := client.WriteJSON(message); err != nil {
			logger.Printf("Error broadcasting to client: %v", err)
		}
	}

	logger.Printf("Broadcast update for %s: %+d (%s)", update.ProductID, update.Change, update.UpdateType)
}

func initializeSampleInventory() {
	products := []struct {
		ID   string
		Name string
		Warehouses map[string]int
	}{
		{"OLJCESPC7Z", "Sunglasses", map[string]int{"US-WEST": 150, "US-EAST": 200}},
		{"66VCHSJNUP", "Tank Top", map[string]int{"US-WEST": 300, "EU-CENTRAL": 250}},
		{"1YMWWN1N4O", "Watch", map[string]int{"US-EAST": 75, "ASIA-PACIFIC": 100}},
		{"L9ECAV7KIM", "Loafers", map[string]int{"US-WEST": 120, "US-EAST": 180}},
		{"2ZYFJ3GM2N", "Hairdryer", map[string]int{"EU-CENTRAL": 200, "ASIA-PACIFIC": 150}},
	}

	for _, p := range products {
		totalStock := calculateTotalStock(p.Warehouses)
		inventory[p.ID] = &Product{
			ProductID:      p.ID,
			Name:           p.Name,
			TotalStock:     totalStock,
			ReservedStock:  0,
			AvailableStock: totalStock,
			Warehouses:     p.Warehouses,
			LastUpdated:    time.Now(),
		}
	}

	logger.Printf("Initialized %d products in inventory", len(inventory))
}

func calculateTotalStock(warehouses map[string]int) int {
	total := 0
	for _, qty := range warehouses {
		total += qty
	}
	return total
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}
