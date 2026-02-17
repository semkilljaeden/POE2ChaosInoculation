package engine

import (
	"image"
	"sync/atomic"
	"time"

	"poe2-chaos-crafter/internal/config"
)

// EventBroadcaster defines the interface for broadcasting events to WebSocket clients
type EventBroadcaster interface {
	Broadcast(msgType string, data interface{})
}

// SessionManager abstracts the hub's session tracking so engine doesn't import server
type SessionManager interface {
	OnSessionStart(session *CraftingSession, cfg *config.Config)
	OnSessionEnd()
	GetState() string
	SetState(state string)
}

// Engine holds all runtime state that was previously in package-level globals
type Engine struct {
	StopRequested       atomic.Bool
	PauseRequested      atomic.Bool
	PauseToggleCooldown atomic.Value // stores time.Time
	LastPauseKeyState   atomic.Bool
	SnapshotCounter     atomic.Int32 // Sequential counter for snapshot naming
	DebugMode           bool
	EmptyCellReference  image.Image
	Broadcaster         EventBroadcaster // nil in CLI mode
	SessionManager      SessionManager   // nil in CLI mode
}

// NewEngine creates a new Engine with default state
func NewEngine(debugMode bool) *Engine {
	e := &Engine{
		DebugMode: debugMode,
	}
	e.PauseToggleCooldown.Store(time.Now())
	return e
}

// Emit sends an event to all connected WebSocket clients.
// No-op when Broadcaster is nil (CLI mode).
func (e *Engine) Emit(msgType string, data interface{}) {
	if e.Broadcaster != nil {
		e.Broadcaster.Broadcast(msgType, data)
	}
}
