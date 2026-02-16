package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSClient represents a connected WebSocket client
type WSClient struct {
	hub  *WSHub
	conn *websocket.Conn
	send chan []byte
}

// WSHub manages WebSocket connections and broadcasts messages
type WSHub struct {
	clients    map[*WSClient]bool
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
	mu         sync.RWMutex

	// Crafting state tracked by the hub
	state          string           // "idle", "running", "paused", "stopped"
	activeSession  *CraftingSession // pointer to active session (nil when idle)
	activeConfig   *Config          // pointer to active config (nil when idle)
	currentItem    int              // current item number in batch
	currentAttempt int              // current attempt within item
	maxAttempts    int              // max attempts per item
	lastOCRText    string           // last OCR text captured
}

// NewWSHub creates a new WebSocket hub
func NewWSHub() *WSHub {
	return &WSHub{
		clients:    make(map[*WSClient]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
		state:      "idle",
	}
}

// Run starts the hub's event loop
func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			fmt.Printf("[WSHub] Client connected (%d total)\n", len(h.clients))

			// Send current state to newly connected client
			h.sendCurrentState(client)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			fmt.Printf("[WSHub] Client disconnected (%d total)\n", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client buffer full, disconnect
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast implements EventBroadcaster interface
func (h *WSHub) Broadcast(msgType string, data interface{}) {
	// Update internal state based on message type
	h.updateState(msgType, data)

	msg, err := marshalWSMessage(msgType, data)
	if err != nil {
		fmt.Printf("[WSHub] Error marshaling message: %v\n", err)
		return
	}

	select {
	case h.broadcast <- msg:
	default:
		fmt.Println("[WSHub] Broadcast channel full, dropping message")
	}
}

// updateState updates the hub's tracked state based on events
func (h *WSHub) updateState(msgType string, data interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch msgType {
	case "state_change":
		if d, ok := data.(StateChangeData); ok {
			h.state = d.State
		}
	case "roll_attempted":
		if d, ok := data.(RollAttemptedData); ok {
			h.currentAttempt = d.AttemptNum
			h.maxAttempts = d.MaxAttempts
		}
	case "item_started":
		if d, ok := data.(ItemStartedData); ok {
			h.currentItem = d.ItemNumber
		}
	case "mods_tracked":
		if d, ok := data.(ModsTrackedData); ok {
			h.lastOCRText = d.OCRText
		}
	}
}

// sendCurrentState sends the current state snapshot to a newly connected client
func (h *WSHub) sendCurrentState(client *WSClient) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Send state
	stateMsg, _ := marshalWSMessage("state_change", StateChangeData{State: h.state})
	select {
	case client.send <- stateMsg:
	default:
	}

	// If running, send current progress
	if h.state == "running" && h.activeSession != nil {
		duration := time.Since(h.activeSession.StartTime)
		rollsPerMin := 0.0
		if duration.Minutes() > 0 {
			rollsPerMin = float64(h.activeSession.TotalRolls) / duration.Minutes()
		}

		rollMsg, _ := marshalWSMessage("roll_attempted", RollAttemptedData{
			AttemptNum:  h.currentAttempt,
			MaxAttempts: h.maxAttempts,
			TotalRolls:  h.activeSession.TotalRolls,
			RollsPerMin: rollsPerMin,
		})
		select {
		case client.send <- rollMsg:
		default:
		}

		// Send current mod stats
		if len(h.activeSession.ModStats) > 0 {
			modsMsg, _ := marshalWSMessage("mods_tracked", ModsTrackedData{
				OCRText:    h.lastOCRText,
				ModStats:   h.activeSession.ModStats,
				TotalRolls: h.activeSession.TotalRolls,
			})
			select {
			case client.send <- modsMsg:
			default:
			}
		}
	}
}

// GetStatusJSON returns current crafting status as JSON
func (h *WSHub) GetStatusJSON() ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	status := map[string]interface{}{
		"state":       h.state,
		"currentItem": h.currentItem,
		"attempt":     h.currentAttempt,
		"maxAttempts": h.maxAttempts,
	}

	if h.activeSession != nil {
		duration := time.Since(h.activeSession.StartTime)
		rollsPerMin := 0.0
		if duration.Minutes() > 0 {
			rollsPerMin = float64(h.activeSession.TotalRolls) / duration.Minutes()
		}
		status["totalRolls"] = h.activeSession.TotalRolls
		status["rollsPerMin"] = rollsPerMin
		status["duration"] = duration.Round(time.Second).String()
		status["targetModHit"] = h.activeSession.TargetModHit
		if h.activeSession.TargetModHit {
			status["targetModName"] = h.activeSession.TargetModName
			status["targetValue"] = h.activeSession.TargetValue
		}
	}

	return json.Marshal(status)
}

// writePump pumps messages from the hub to the websocket connection
func (c *WSClient) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *WSClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// Client messages are currently ignored (control is via REST API)
	}
}
