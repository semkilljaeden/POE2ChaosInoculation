package main

import (
	"fmt"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
)

// Windows API for key state checking and sound
var (
	user32          = syscall.NewLazyDLL("user32.dll")
	procGetKeyState = user32.NewProc("GetKeyState")
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procBeep        = kernel32.NewProc("Beep")
)

// moveItem moves an item from one position to another.
// Returns true if completed, false if aborted by stop request.
func moveItem(fromX, fromY, toX, toY int) bool {
	fmt.Printf("     [moveItem] Starting move from (%d,%d) to (%d,%d)\n", fromX, fromY, toX, toY)

	// Step 1: Move cursor to source
	if stopRequested.Load() {
		fmt.Println("     [moveItem] Aborted (stop requested)")
		return false
	}
	fmt.Printf("     [moveItem] Step 1: Moving cursor to source (%d,%d)\n", fromX, fromY)
	robotgo.Move(fromX, fromY)
	time.Sleep(100 * time.Millisecond)
	actualX, actualY := robotgo.Location()
	fmt.Printf("     [moveItem] Step 1: Cursor at (%d,%d)\n", actualX, actualY)

	// Step 2: Click to grab item (button down + up)
	if stopRequested.Load() {
		fmt.Println("     [moveItem] Aborted (stop requested)")
		return false
	}
	fmt.Println("     [moveItem] Step 2: LEFT CLICK to grab item")
	fmt.Println("     [moveItem]   - Button DOWN")
	robotgo.Toggle("left", "down")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("     [moveItem]   - Button UP")
	robotgo.Toggle("left", "up")
	time.Sleep(200 * time.Millisecond)
	fmt.Println("     [moveItem] Step 2: Item grabbed (cursor should show item)")

	// Step 3: Move cursor to destination
	if stopRequested.Load() {
		fmt.Println("     [moveItem] Aborted after grab (stop requested)")
		return false
	}
	fmt.Printf("     [moveItem] Step 3: Moving cursor to destination (%d,%d)\n", toX, toY)
	robotgo.MoveSmooth(toX, toY, 0.5, 0.5)
	time.Sleep(100 * time.Millisecond)
	actualX, actualY = robotgo.Location()
	fmt.Printf("     [moveItem] Step 3: Cursor at (%d,%d)\n", actualX, actualY)

	// Step 4: Click to drop item (button down + up)
	if stopRequested.Load() {
		fmt.Println("     [moveItem] Aborted after move (stop requested)")
		return false
	}
	fmt.Println("     [moveItem] Step 4: LEFT CLICK to drop item")
	fmt.Println("     [moveItem]   - Button DOWN")
	robotgo.Toggle("left", "down")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("     [moveItem]   - Button UP")
	robotgo.Toggle("left", "up")
	time.Sleep(200 * time.Millisecond)
	fmt.Println("     [moveItem] Step 4: Item dropped at destination")
	fmt.Println("     [moveItem] Move complete")
	return true
}

// getKeyState returns the state of a virtual key
func getKeyState(vKey int) int16 {
	ret, _, _ := procGetKeyState.Call(uintptr(vKey))
	return int16(ret)
}

// playBeep plays a beep sound with specified frequency and duration
func playBeep(frequency int, durationMs int) {
	procBeep.Call(uintptr(frequency), uintptr(durationMs))
}

// playVictorySound plays a triumphant victory melody
func playVictorySound() {
	// Victory fanfare melody (inspired by Final Fantasy victory theme)
	// Notes: C5, C5, C5, C5, G#4, A#4, C5, A#4, C5
	notes := []struct {
		freq int // Frequency in Hz
		dur  int // Duration in milliseconds
	}{
		{523, 150}, // C5
		{523, 150}, // C5
		{523, 150}, // C5
		{523, 400}, // C5 (longer)
		{415, 350}, // G#4
		{466, 350}, // A#4
		{523, 150}, // C5
		{466, 150}, // A#4
		{523, 600}, // C5 (final note, longest)
	}

	// Play the melody in a goroutine so it doesn't block
	go func() {
		for _, note := range notes {
			playBeep(note.freq, note.dur)
			time.Sleep(time.Duration(note.dur) * time.Millisecond)
		}
	}()
}

// checkPauseToggle checks for F12 key state to toggle pause
func checkPauseToggle() {
	// Check cooldown to prevent rapid toggling
	cooldown := pauseToggleCooldown.Load().(time.Time)
	if !time.Now().After(cooldown) {
		return
	}

	// VK_F12 = 0x7B = 123
	// Check if F12 is currently pressed (negative value means key is down)
	keyState := getKeyState(0x7B)
	f12Pressed := keyState < 0

	// Detect state change (key press, not toggle)
	lastState := lastPauseKeyState.Load()
	if f12Pressed != lastState {
		fmt.Printf("\n[DEBUG] F12 state changed: %v -> %v", lastState, f12Pressed)
		lastPauseKeyState.Store(f12Pressed)

		// Only toggle on key press (not release)
		if f12Pressed {
			pauseToggleCooldown.Store(time.Now().Add(300 * time.Millisecond))

			// Toggle pause state
			currentPause := pauseRequested.Load()
			pauseRequested.Store(!currentPause)

			if !currentPause {
				fmt.Print("\n[DEBUG] pauseRequested flag set to true")
				fmt.Print("\n⏸  PAUSED - Press F12 to resume or Ctrl+C to stop")
				emit("state_change", StateChangeData{State: "paused"})
			} else {
				fmt.Print("\n[DEBUG] pauseRequested flag set to false")
				fmt.Print("\n▶  RESUMED")
				emit("state_change", StateChangeData{State: "running"})
			}
		}
	}
}

// checkMiddleMouseButton is now an alias for checkPauseToggle
func checkMiddleMouseButton() {
	checkPauseToggle()
}
