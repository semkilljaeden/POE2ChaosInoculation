package server

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"poe2-chaos-crafter/internal/config"
	"poe2-chaos-crafter/internal/engine"

	"github.com/go-vgo/robotgo"
	"github.com/gorilla/websocket"
	"golang.org/x/image/draw"
)

//go:embed web/*
var webFS embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// StartWebServer starts the web GUI server
func StartWebServer(port int, eng *engine.Engine) {
	hub := NewWSHub()
	eng.Broadcaster = hub
	eng.SessionManager = hub
	go hub.Run()

	mux := http.NewServeMux()

	// Serve embedded static files
	webContent, _ := fs.Sub(webFS, "web")
	fileServer := http.FileServer(http.FS(webContent))
	mux.Handle("/", fileServer)

	// WebSocket endpoint
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, hub)
	})

	// REST API
	mux.HandleFunc("/api/config", handleConfig)
	mux.HandleFunc("/api/config/reload", handleConfigReload)
	mux.HandleFunc("/api/craft/start", func(w http.ResponseWriter, r *http.Request) {
		handleCraftStart(w, r, eng, hub)
	})
	mux.HandleFunc("/api/craft/stop", func(w http.ResponseWriter, r *http.Request) {
		handleCraftStop(w, r, eng)
	})
	mux.HandleFunc("/api/craft/pause", func(w http.ResponseWriter, r *http.Request) {
		handleCraftPause(w, r, eng)
	})
	mux.HandleFunc("/api/craft/status", func(w http.ResponseWriter, r *http.Request) {
		handleCraftStatus(w, r, hub)
	})
	mux.HandleFunc("/api/session", func(w http.ResponseWriter, r *http.Request) {
		handleSession(w, r, hub)
	})
	mux.HandleFunc("/api/wizard/capture", func(w http.ResponseWriter, r *http.Request) {
		handleWizardCapture(w, r, eng)
	})
	mux.HandleFunc("/api/wizard/validate-tooltip", func(w http.ResponseWriter, r *http.Request) {
		handleWizardValidateTooltip(w, r, eng)
	})
	mux.HandleFunc("/api/wizard/parse-mod", handleWizardParseMod)
	mux.HandleFunc("/api/snapshot/current-tooltip", handleCurrentTooltip)
	mux.HandleFunc("/api/snapshot/screen", func(w http.ResponseWriter, r *http.Request) {
		handleScreenCapture(w, r, hub)
	})
	mux.HandleFunc("/api/mod-templates", handleModTemplates)

	lanIP := getLANIP()

	fmt.Println("\n╔═══════════════════════════════════════════════╗")
	fmt.Println("║          POE2 Chaos Crafter - Web GUI         ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")
	fmt.Printf("\n  Local:   http://localhost:%d\n", port)
	if lanIP != "" {
		fmt.Printf("  Network: http://%s:%d\n", lanIP, port)
	}
	fmt.Println("\n  Open in any browser (PC, phone, tablet)")
	fmt.Println("  Press Ctrl+C to stop the server")
	fmt.Println()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		fmt.Println("\nShutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Server stopped.")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, hub *WSHub) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[WS] Upgrade error: %v\n", err)
		return
	}

	client := &WSClient{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}

	hub.register <- client
	go client.writePump()
	go client.readPump()
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		cfg, err := config.LoadConfig()
		if err != nil {
			http.Error(w, `{"error":"no config found"}`, http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(cfg)

	case "POST":
		var cfg config.Config
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, `{"error":"invalid config"}`, http.StatusBadRequest)
			return
		}
		if err := config.SaveConfig(cfg); err != nil {
			http.Error(w, `{"error":"failed to save"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "saved"})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func handleConfigReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		http.Error(w, `{"error":"failed to load config"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

func handleCraftStart(w http.ResponseWriter, r *http.Request, eng *engine.Engine, hub *WSHub) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if hub.GetState() == "running" {
		http.Error(w, `{"error":"already running"}`, http.StatusConflict)
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		http.Error(w, `{"error":"no config found, run wizard first"}`, http.StatusBadRequest)
		return
	}

	cfg.TooltipRect.Min.X = cfg.ItemPos.X + cfg.TooltipOffset.X
	cfg.TooltipRect.Min.Y = cfg.ItemPos.Y + cfg.TooltipOffset.Y
	cfg.TooltipRect.Max.X = cfg.TooltipRect.Min.X + cfg.TooltipSize.X
	cfg.TooltipRect.Max.Y = cfg.TooltipRect.Min.Y + cfg.TooltipSize.Y

	cfg.UseBatchMode = true
	if cfg.ItemWidth == 0 {
		cfg.ItemWidth = 1
	}
	if cfg.ItemHeight == 0 {
		cfg.ItemHeight = 1
	}
	if cfg.ChaosPerRound == 0 {
		cfg.ChaosPerRound = 10
	}

	eng.StopRequested.Store(false)
	eng.PauseRequested.Store(false)

	go func() {
		eng.Emit("state_change", engine.StateChangeData{State: "countdown"})
		for i := 5; i > 0; i-- {
			eng.Emit("craft_countdown", engine.CraftCountdownData{SecondsLeft: i})
			fmt.Printf("\rStarting in %d... ", i)
			time.Sleep(1 * time.Second)
			if eng.StopRequested.Load() {
				eng.Emit("state_change", engine.StateChangeData{State: "idle"})
				return
			}
		}
		fmt.Println("\rStarting crafting!   ")
		eng.Emit("state_change", engine.StateChangeData{State: "running"})
		eng.Craft(cfg)
		eng.Emit("state_change", engine.StateChangeData{State: "idle"})
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func handleCraftStop(w http.ResponseWriter, r *http.Request, eng *engine.Engine) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	eng.StopRequested.Store(true)
	eng.Emit("state_change", engine.StateChangeData{State: "stopped"})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopping"})
}

func handleCraftPause(w http.ResponseWriter, r *http.Request, eng *engine.Engine) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	current := eng.PauseRequested.Load()
	eng.PauseRequested.Store(!current)

	state := "running"
	if !current {
		state = "paused"
	}
	eng.Emit("state_change", engine.StateChangeData{State: state})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": state})
}

func handleCraftStatus(w http.ResponseWriter, r *http.Request, hub *WSHub) {
	if r.Method != "GET" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	data, err := hub.GetStatusJSON()
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func handleSession(w http.ResponseWriter, r *http.Request, hub *WSHub) {
	if r.Method != "GET" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	hub.mu.RLock()
	session := hub.activeSession
	cfg := hub.activeConfig
	hub.mu.RUnlock()

	if session == nil || cfg == nil {
		http.Error(w, `{"error":"no active session"}`, http.StatusNotFound)
		return
	}

	report := engine.BuildReportData(session, *cfg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func handleWizardCapture(w http.ResponseWriter, r *http.Request, eng *engine.Engine) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Field string `json:"field"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	go func() {
		for i := 5; i > 0; i-- {
			eng.Emit("capture_countdown", engine.CaptureCountdownData{
				SecondsLeft: i,
				Field:       req.Field,
			})
			time.Sleep(1 * time.Second)
		}

		x, y := robotgo.GetMousePos()
		eng.Emit("capture_result", engine.CaptureResultData{
			Field: req.Field,
			X:     x,
			Y:     y,
		})
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "capturing"})
}

func handleWizardValidateTooltip(w http.ResponseWriter, r *http.Request, eng *engine.Engine) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		X1           int    `json:"x1"`
		Y1           int    `json:"y1"`
		X2           int    `json:"x2"`
		Y2           int    `json:"y2"`
		GameLanguage string `json:"gameLanguage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	width := req.X2 - req.X1
	height := req.Y2 - req.Y1
	if width <= 0 || height <= 0 {
		http.Error(w, `{"error":"invalid tooltip dimensions"}`, http.StatusBadRequest)
		return
	}

	bitmap := robotgo.CaptureScreen(req.X1, req.Y1, width, height)
	img := robotgo.ToImage(bitmap)

	os.MkdirAll(config.SnapshotsDir, 0755)
	tooltipFile := filepath.Join(config.SnapshotsDir, "tooltip_area_validation.png")
	engine.SaveImage(img, tooltipFile)

	tempDir := filepath.Join(os.TempDir(), "poe2_crafter_setup")
	os.MkdirAll(tempDir, 0755)

	ocrText, err := eng.RunTesseractOCR(img, tempDir, req.GameLanguage)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	validLines := 0
	for _, line := range strings.Split(strings.TrimSpace(ocrText), "\n") {
		if len(strings.TrimSpace(line)) > 3 {
			validLines++
		}
	}

	var imgBase64 string
	var buf strings.Builder
	enc := base64.NewEncoder(base64.StdEncoding, &buf)
	png.Encode(enc, img)
	enc.Close()
	imgBase64 = buf.String()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    validLines > 0,
		"ocrText":    ocrText,
		"validLines": validLines,
		"image":      imgBase64,
	})
}

func handleWizardParseMod(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Input        string `json:"input"`
		GameLanguage string `json:"gameLanguage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	mod := config.ParseModInput(req.Input, req.GameLanguage)
	if mod.Pattern == "" {
		http.Error(w, `{"error":"invalid mod format"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mod)
}

func handleCurrentTooltip(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	filePath := filepath.Join(config.SnapshotsDir, "current_tooltip.png")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, `{"error":"no tooltip yet"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, filePath)
}

func handleScreenCapture(w http.ResponseWriter, r *http.Request, hub *WSHub) {
	if r.Method != "GET" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	state := hub.GetState()
	if state != "running" && state != "paused" && state != "countdown" {
		http.Error(w, `{"error":"crafting not active"}`, http.StatusServiceUnavailable)
		return
	}

	bitmap := robotgo.CaptureScreen()
	img := robotgo.ToImage(bitmap)

	small := downsampleImage(img, 4)

	var buf bytes.Buffer
	if err := png.Encode(&buf, small); err != nil {
		http.Error(w, `{"error":"encode failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	w.Write(buf.Bytes())
}

func downsampleImage(src image.Image, factor int) image.Image {
	bounds := src.Bounds()
	newW := bounds.Dx() / factor
	newH := bounds.Dy() / factor
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)
	return dst
}

func handleModTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	templates := []map[string]string{
		{"key": "life", "name": "Life", "name_zh": "生命", "example": "life 80"},
		{"key": "mana", "name": "Mana", "name_zh": "魔力", "example": "mana 60"},
		{"key": "str", "name": "Strength", "name_zh": "力量", "example": "str 45"},
		{"key": "dex", "name": "Dexterity", "name_zh": "敏捷", "example": "dex 45"},
		{"key": "int", "name": "Intelligence", "name_zh": "智慧", "example": "int 45"},
		{"key": "spirit", "name": "Spirit", "name_zh": "精魂", "example": "spirit 50"},
		{"key": "spell-level", "name": "Spell Skills Level", "name_zh": "法术技能等级", "example": "spell-level 3"},
		{"key": "proj-level", "name": "Projectile Skills Level", "name_zh": "投射物技能等级", "example": "proj-level 3"},
		{"key": "crit-dmg", "name": "Critical Damage Bonus", "name_zh": "暴击伤害加成", "example": "crit-dmg 39"},
		{"key": "fire-res", "name": "Fire Resistance", "name_zh": "火焰抗性", "example": "fire-res 30"},
		{"key": "cold-res", "name": "Cold Resistance", "name_zh": "冰冷抗性", "example": "cold-res 30"},
		{"key": "light-res", "name": "Lightning Resistance", "name_zh": "闪电抗性", "example": "light-res 30"},
		{"key": "chaos-res", "name": "Chaos Resistance", "name_zh": "混沌抗性", "example": "chaos-res 20"},
		{"key": "armor", "name": "Armour", "name_zh": "护甲", "example": "armor 100"},
		{"key": "evasion", "name": "Evasion", "name_zh": "闪避", "example": "evasion 100"},
		{"key": "es", "name": "Energy Shield", "name_zh": "能量护盾", "example": "es 50"},
		{"key": "movespeed", "name": "Movement Speed", "name_zh": "移动速度", "example": "movespeed 20"},
		{"key": "attackspeed", "name": "Attack Speed", "name_zh": "攻击速度", "example": "attackspeed 10"},
		{"key": "castspeed", "name": "Cast Speed", "name_zh": "施放速度", "example": "castspeed 10"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func getLANIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return ""
}
