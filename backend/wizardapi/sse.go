package wizardapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/andriyg76/bgl/services"
	log "github.com/andriyg76/glog"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// subscribeToEvents handles SSE subscription for real-time game updates
func (h *Handler) subscribeToEvents(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		http.Error(w, "Game code is required", http.StatusBadRequest)
		return
	}

	// Verify game exists
	_, err := h.wizardRepo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Check if the ResponseWriter supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Generate unique client ID
	clientID := uuid.New().String()

	// Subscribe to game events
	client := h.eventHub.Subscribe(code, clientID)
	defer h.eventHub.Unsubscribe(client)

	log.Info("SSE: Client %s connected to game %s", clientID, code)

	// Send initial connection event
	initialEvent := &services.GameEvent{
		Type:      "connected",
		GameCode:  code,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"client_id":   clientID,
			"subscribers": h.eventHub.GetSubscriberCount(code),
		},
	}
	if !sendSSEEvent(w, flusher, initialEvent) {
		log.Info("SSE: Client %s failed to connect to game %s", clientID, code)
		return
	}

	// Start heartbeat ticker
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	// Listen for events
	for {
		select {
		case <-r.Context().Done():
			log.Info("SSE: Client %s disconnected from game %s (context done)", clientID, code)
			return
		case <-client.Done:
			log.Info("SSE: Client %s unsubscribed from game %s", clientID, code)
			return
		case event := <-client.Channel:
			if !sendSSEEvent(w, flusher, event) {
				log.Info("SSE: Client %s disconnected from game %s (write failed)", clientID, code)
				return
			}
		case <-heartbeat.C:
			// Send heartbeat to keep connection alive and detect dead connections
			heartbeatEvent := &services.GameEvent{
				Type:      "heartbeat",
				GameCode:  code,
				Timestamp: time.Now(),
			}
			if !sendSSEEvent(w, flusher, heartbeatEvent) {
				log.Info("SSE: Client %s disconnected from game %s (heartbeat failed)", clientID, code)
				return
			}
		}
	}
}

// sendSSEEvent sends an event in SSE format
// Returns false if writing failed (client disconnected)
func sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, event *services.GameEvent) bool {
	data, err := services.FormatSSEEvent(event)
	if err != nil {
		log.Warn("SSE: Failed to format event: %v", err)
		return false
	}

	// Write SSE format: event: <type>\ndata: <json>\n\n
	_, err = fmt.Fprintf(w, "event: %s\n", event.Type)
	if err != nil {
		log.Info("SSE: Write failed (client disconnected): %v", err)
		return false
	}
	_, err = fmt.Fprintf(w, "data: %s\n\n", data)
	if err != nil {
		log.Info("SSE: Write failed (client disconnected): %v", err)
		return false
	}
	flusher.Flush()
	return true
}
