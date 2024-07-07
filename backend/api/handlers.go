package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/suhas-24/apica-fullstack-assignment/backend/cache"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Be cautious with this in production
	},
}

type Handler struct {
	cache            *cache.LRUCache
	subscribers      map[*websocket.Conn]bool
	subscribersMutex sync.RWMutex
}

func NewHandler(cache *cache.LRUCache) *Handler {
	h := &Handler{
		cache:       cache,
		subscribers: make(map[*websocket.Conn]bool),
	}
	go h.broadcastLoop()
	return h
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, ok := h.cache.Get(key)
	if !ok {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"value": value})
}

func (h *Handler) SetHandler(w http.ResponseWriter, r *http.Request) {
    var data struct {
        Key        string `json:"key"`
        Value      string `json:"value"`
        Expiration int    `json:"expiration"` // expiration in seconds
    }

    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    h.cache.Set(data.Key, data.Value, time.Duration(data.Expiration)*time.Second)
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    response := map[string]string{
        "message": "Key set successfully",
        "key":     data.Key,
        "value":   data.Value,
    }
    json.NewEncoder(w).Encode(response)
    
    h.notifySubscribers()
}

func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	h.cache.Delete(key)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Key deleted successfully"})
	h.notifySubscribers()
}

func (h *Handler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	items := h.cache.GetAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	h.subscribersMutex.Lock()
	h.subscribers[conn] = true
	h.subscribersMutex.Unlock()

	defer func() {
		h.subscribersMutex.Lock()
		delete(h.subscribers, conn)
		h.subscribersMutex.Unlock()
	}()

	// Send initial cache state
	h.sendCacheState(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}
	}
}

func (h *Handler) broadcastLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.broadcastCacheState()
	}
}

func (h *Handler) broadcastCacheState() {
	items := h.cache.GetAll()
	h.subscribersMutex.RLock()
	defer h.subscribersMutex.RUnlock()

	for conn := range h.subscribers {
		err := conn.WriteJSON(items)
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			conn.Close()
			delete(h.subscribers, conn)
		}
	}
}

func (h *Handler) sendCacheState(conn *websocket.Conn) {
	items := h.cache.GetAll()
	err := conn.WriteJSON(items)
	if err != nil {
		log.Printf("WebSocket write error: %v", err)
	}
}

func (h *Handler) notifySubscribers() {
	h.broadcastCacheState()
}