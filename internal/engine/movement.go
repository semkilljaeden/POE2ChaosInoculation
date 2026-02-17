package engine

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

// MoveItem moves an item from one position to another.
// Returns true if completed, false if aborted by stop request.
func (e *Engine) MoveItem(fromX, fromY, toX, toY int) bool {
	fmt.Printf("     [moveItem] Starting move from (%d,%d) to (%d,%d)\n", fromX, fromY, toX, toY)

	if e.StopRequested.Load() {
		fmt.Println("     [moveItem] Aborted (stop requested)")
		return false
	}
	fmt.Printf("     [moveItem] Step 1: Moving cursor to source (%d,%d)\n", fromX, fromY)
	robotgo.Move(fromX, fromY)
	time.Sleep(100 * time.Millisecond)
	actualX, actualY := robotgo.Location()
	fmt.Printf("     [moveItem] Step 1: Cursor at (%d,%d)\n", actualX, actualY)

	if e.StopRequested.Load() {
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

	if e.StopRequested.Load() {
		fmt.Println("     [moveItem] Aborted after grab (stop requested)")
		return false
	}
	fmt.Printf("     [moveItem] Step 3: Moving cursor to destination (%d,%d)\n", toX, toY)
	robotgo.MoveSmooth(toX, toY, 0.5, 0.5)
	time.Sleep(100 * time.Millisecond)
	actualX, actualY = robotgo.Location()
	fmt.Printf("     [moveItem] Step 3: Cursor at (%d,%d)\n", actualX, actualY)

	if e.StopRequested.Load() {
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

// GetKeyState returns the state of a virtual key
func GetKeyState(vKey int) int16 {
	ret, _, _ := procGetKeyState.Call(uintptr(vKey))
	return int16(ret)
}

// PlayBeep plays a beep sound with specified frequency and duration
func PlayBeep(frequency int, durationMs int) {
	procBeep.Call(uintptr(frequency), uintptr(durationMs))
}

// PlayVictorySound plays a triumphant victory melody
func PlayVictorySound() {
	notes := []struct {
		freq int
		dur  int
	}{
		{523, 150}, {523, 150}, {523, 150}, {523, 400},
		{415, 350}, {466, 350}, {523, 150}, {466, 150}, {523, 600},
	}

	go func() {
		for _, note := range notes {
			PlayBeep(note.freq, note.dur)
			time.Sleep(time.Duration(note.dur) * time.Millisecond)
		}
	}()
}

// CheckPauseToggle checks for F12 key state to toggle pause
func (e *Engine) CheckPauseToggle() {
	cooldown := e.PauseToggleCooldown.Load().(time.Time)
	if !time.Now().After(cooldown) {
		return
	}

	// VK_F12 = 0x7B = 123
	keyState := GetKeyState(0x7B)
	f12Pressed := keyState < 0

	lastState := e.LastPauseKeyState.Load()
	if f12Pressed != lastState {
		fmt.Printf("\n[DEBUG] F12 state changed: %v -> %v", lastState, f12Pressed)
		e.LastPauseKeyState.Store(f12Pressed)

		if f12Pressed {
			e.PauseToggleCooldown.Store(time.Now().Add(300 * time.Millisecond))

			currentPause := e.PauseRequested.Load()
			e.PauseRequested.Store(!currentPause)

			if !currentPause {
				fmt.Print("\n[DEBUG] pauseRequested flag set to true")
				fmt.Print("\n⏸  PAUSED - Press F12 to resume or Ctrl+C to stop")
				e.Emit("state_change", StateChangeData{State: "paused"})
			} else {
				fmt.Print("\n[DEBUG] pauseRequested flag set to false")
				fmt.Print("\n▶  RESUMED")
				e.Emit("state_change", StateChangeData{State: "running"})
			}
		}
	}
}

// CheckMiddleMouseButton is now an alias for CheckPauseToggle
func (e *Engine) CheckMiddleMouseButton() {
	e.CheckPauseToggle()
}
