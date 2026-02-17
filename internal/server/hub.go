package server

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"poe2-chaos-crafter/internal/config"
	"poe2-chaos-crafter/internal/engine"

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
	state          string                  // "idle", "running", "paused", "stopped"
	activeSession  *engine.CraftingSession // pointer to active session (nil when idle)
	activeConfig   *config.Config          // pointer to active config (nil when idle)
	currentItem    int
	currentAttempt int
	maxAttempts    int
	lastOCRText    string
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
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast implements engine.EventBroadcaster interface
func (h *WSHub) Broadcast(msgType string, data interface{}) {
	h.updateState(msgType, data)

	msg, err := engine.MarshalWSMessage(msgType, data)
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

// OnSessionStart implements engine.SessionManager
func (h *WSHub) OnSessionStart(session *engine.CraftingSession, cfg *config.Config) {
	h.mu.Lock()
	h.activeSession = session
	h.activeConfig = cfg
	h.state = "running"
	h.mu.Unlock()
}

// OnSessionEnd implements engine.SessionManager
func (h *WSHub) OnSessionEnd() {
	h.mu.Lock()
	h.activeSession = nil
	h.activeConfig = nil
	h.state = "idle"
	h.mu.Unlock()
}

// GetState implements engine.SessionManager
func (h *WSHub) GetState() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.state
}

// SetState implements engine.SessionManager
func (h *WSHub) SetState(state string) {
	h.mu.Lock()
	h.state = state
	h.mu.Unlock()
}

// updateState updates the hub's tracked state based on events
func (h *WSHub) updateState(msgType string, data interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch msgType {
	case "state_change":
		if d, ok := data.(engine.StateChangeData); ok {
			h.state = d.State
		}
	case "roll_attempted":
		if d, ok := data.(engine.RollAttemptedData); ok {
			h.currentAttempt = d.AttemptNum
			h.maxAttempts = d.MaxAttempts
		}
	case "item_started":
		if d, ok := data.(engine.ItemStartedData); ok {
			h.currentItem = d.ItemNumber
		}
	case "mods_tracked":
		if d, ok := data.(engine.ModsTrackedData); ok {
			h.lastOCRText = d.OCRText
		}
	}
}

// sendCurrentState sends the current state snapshot to a newly connected client
func (h *WSHub) sendCurrentState(client *WSClient) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stateMsg, _ := engine.MarshalWSMessage("state_change", engine.StateChangeData{State: h.state})
	select {
	case client.send <- stateMsg:
	default:
	}

	if h.state == "running" && h.activeSession != nil {
		duration := time.Since(h.activeSession.StartTime)
		rollsPerMin := 0.0
		if duration.Minutes() > 0 {
			rollsPerMin = float64(h.activeSession.TotalRolls) / duration.Minutes()
		}

		rollMsg, _ := engine.MarshalWSMessage("roll_attempted", engine.RollAttemptedData{
			AttemptNum:  h.currentAttempt,
			MaxAttempts: h.maxAttempts,
			TotalRolls:  h.activeSession.TotalRolls,
			RollsPerMin: rollsPerMin,
		})
		select {
		case client.send <- rollMsg:
		default:
		}

		if len(h.activeSession.ModStats) > 0 {
			modsMsg, _ := engine.MarshalWSMessage("mods_tracked", engine.ModsTrackedData{
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
	}
}
