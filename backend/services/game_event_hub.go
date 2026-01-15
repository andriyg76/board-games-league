package services

import (
	"encoding/json"
	"sync"
	"time"

	log "github.com/andriyg76/glog"
)

// GameEvent represents an event to be broadcast to connected clients
type GameEvent struct {
	Type      string      `json:"type"`
	GameCode  string      `json:"game_code"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// GameEventClient represents a connected SSE client
type GameEventClient struct {
	ID       string
	GameCode string
	Channel  chan *GameEvent
	Done     chan struct{}
}

// GameEventHub manages SSE connections for game updates
type GameEventHub interface {
	// Subscribe adds a client to receive updates for a specific game
	Subscribe(gameCode string, clientID string) *GameEventClient
	// Unsubscribe removes a client from receiving updates
	Unsubscribe(client *GameEventClient)
	// Broadcast sends an event to all clients subscribed to a game
	Broadcast(gameCode string, eventType string, data interface{})
	// GetSubscriberCount returns the number of subscribers for a game
	GetSubscriberCount(gameCode string) int
}

type gameEventHub struct {
	mu      sync.RWMutex
	clients map[string]map[string]*GameEventClient // gameCode -> clientID -> client
}

// NewGameEventHub creates a new game event hub
func NewGameEventHub() GameEventHub {
	return &gameEventHub{
		clients: make(map[string]map[string]*GameEventClient),
	}
}

// Subscribe adds a client to receive updates for a specific game
func (h *gameEventHub) Subscribe(gameCode string, clientID string) *GameEventClient {
	h.mu.Lock()
	defer h.mu.Unlock()

	client := &GameEventClient{
		ID:       clientID,
		GameCode: gameCode,
		Channel:  make(chan *GameEvent, 10), // buffered channel to prevent blocking
		Done:     make(chan struct{}),
	}

	if h.clients[gameCode] == nil {
		h.clients[gameCode] = make(map[string]*GameEventClient)
	}

	h.clients[gameCode][clientID] = client
	log.Info("SSE: Client %s subscribed to game %s (total: %d)", clientID, gameCode, len(h.clients[gameCode]))

	return client
}

// Unsubscribe removes a client from receiving updates
func (h *gameEventHub) Unsubscribe(client *GameEventClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if gameClients, ok := h.clients[client.GameCode]; ok {
		if _, exists := gameClients[client.ID]; exists {
			close(client.Done)
			close(client.Channel)
			delete(gameClients, client.ID)
			log.Info("SSE: Client %s unsubscribed from game %s (remaining: %d)", client.ID, client.GameCode, len(gameClients))

			// Clean up game entry if no more clients
			if len(gameClients) == 0 {
				delete(h.clients, client.GameCode)
			}
		}
	}
}

// Broadcast sends an event to all clients subscribed to a game
func (h *gameEventHub) Broadcast(gameCode string, eventType string, data interface{}) {
	h.mu.RLock()
	gameClients, ok := h.clients[gameCode]
	if !ok || len(gameClients) == 0 {
		h.mu.RUnlock()
		return
	}

	// Copy clients to avoid holding lock while sending
	clients := make([]*GameEventClient, 0, len(gameClients))
	for _, client := range gameClients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	event := &GameEvent{
		Type:      eventType,
		GameCode:  gameCode,
		Timestamp: time.Now(),
		Data:      data,
	}

	log.Info("SSE: Broadcasting %s event to %d clients for game %s", eventType, len(clients), gameCode)

	for _, client := range clients {
		select {
		case client.Channel <- event:
			// Event sent successfully
		default:
			// Channel full, skip (client is slow or disconnected)
			log.Warn("SSE: Channel full for client %s, skipping event", client.ID)
		}
	}
}

// GetSubscriberCount returns the number of subscribers for a game
func (h *gameEventHub) GetSubscriberCount(gameCode string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if gameClients, ok := h.clients[gameCode]; ok {
		return len(gameClients)
	}
	return 0
}

// FormatSSEEvent formats a GameEvent as SSE data
func FormatSSEEvent(event *GameEvent) ([]byte, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	return data, nil
}
