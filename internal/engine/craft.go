package engine

import (
	"fmt"
	"image"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"poe2-chaos-crafter/internal/config"

	"github.com/go-vgo/robotgo"
)

// Craft is the main crafting function that handles batch mode processing
func (e *Engine) Craft(cfg config.Config) {
	// Initialize cooldown time and snapshot counter
	e.PauseToggleCooldown.Store(time.Now())
	e.SnapshotCounter.Store(0)

	// Initialize crafting session for tracking
	session := &CraftingSession{
		StartTime: time.Now(),
		ModStats:  make(map[string]*ModStat),
	}

	// Register session with hub for web GUI status
	if e.SessionManager != nil {
		e.SessionManager.OnSessionStart(session, &cfg)
	}
	defer func() {
		if e.SessionManager != nil {
			e.SessionManager.OnSessionEnd()
		}
	}()

	// Setup signal handler for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\n[DEBUG] Signal received (Ctrl+C)")
		fmt.Println("🛑 Stop requested... Exiting safely.")
		e.StopRequested.Store(true)
		fmt.Println("[DEBUG] stopRequested flag set to true")
	}()

	// Ensure report is generated even if interrupted
	defer func() {
		session.EndTime = time.Now()
		e.GenerateReport(session, cfg)
	}()

	// Clean up old debug snapshots from previous runs (only in debug mode)
	if e.DebugMode {
		CleanupDebugSnapshots()
	}

	// Create resource directory if it doesn't exist
	if err := os.MkdirAll(config.ResourceDir, 0755); err != nil {
		fmt.Printf("⚠ WARNING: Could not create resource directory: %v\n", err)
	}

	// Load empty cell reference image for item detection
	emptyCellRefPath := filepath.Join(config.ResourceDir, "empty_cell_reference.png")
	fmt.Printf("\n📸 Loading empty cell reference image from: %s\n", emptyCellRefPath)
	if err := e.LoadEmptyCellReference(emptyCellRefPath); err != nil {
		fmt.Printf("⚠ WARNING: Could not load empty cell reference: %v\n", err)
		fmt.Println("   Please save an empty cell snapshot as 'resource/empty_cell_reference.png'")
		fmt.Println("   Item detection may not work correctly without it!")
	}

	// Generate grid snapshot at start of crafting (debug mode only)
	if e.DebugMode && cfg.BackpackTopLeft.X != 0 && cfg.BackpackBottomRight.X != 0 {
		fmt.Println("\n📸 Generating grid snapshot...")
		if err := DrawBackpackGrid(cfg); err != nil {
			fmt.Printf("⚠ Warning: Could not create grid snapshot: %v\n", err)
		} else {
			fmt.Println("✓ Grid snapshot: backpack_grid_debug.png")
		}
	}

	// Ensure snapshots directory exists
	os.MkdirAll(config.SnapshotsDir, 0755)

	// Create temp directory for OCR
	tempDir := filepath.Join(os.TempDir(), "poe2_crafter")
	os.MkdirAll(tempDir, 0755)

	// Batch mode: process multiple items from pending area
	if cfg.UseBatchMode {
		fmt.Println("\n🔄 BATCH MODE ENABLED")
		fmt.Printf("📦 Pending area: %dx%d cells\n", cfg.PendingAreaWidth, cfg.PendingAreaHeight)
		fmt.Printf("🎯 Workbench: (%d, %d)\n", cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y)
		fmt.Printf("✅ Result area: %dx%d cells\n\n", cfg.ResultAreaWidth, cfg.ResultAreaHeight)

		processedPositions := make(map[string]bool)
		itemCount := 0

		for {
			if e.StopRequested.Load() {
				fmt.Println("\n✓ Stopped by user")
				return
			}

			itemX, itemY, found := e.FindNextItemInArea(cfg, cfg.PendingAreaTopLeft, cfg.PendingAreaWidth, cfg.PendingAreaHeight, processedPositions)
			if !found {
				fmt.Println("\n✓ No more items in pending area")
				break
			}

			itemCount++
			fmt.Printf("\n📦 Processing item #%d from pending area at (%d, %d)...\n", itemCount, itemX, itemY)
			e.Emit("item_started", ItemStartedData{ItemNumber: itemCount, PendingX: itemX, PendingY: itemY})

			roundResult := RoundResult{
				RoundNumber: itemCount,
				StartPos:    image.Point{X: itemX, Y: itemY},
				Success:     false,
			}

			posKey := fmt.Sprintf("%d,%d", itemX, itemY)
			processedPositions[posKey] = true

			resultX, resultY, foundSlot := e.FindEmptySlotInArea(cfg, cfg.ResultAreaTopLeft, cfg.ResultAreaWidth, cfg.ResultAreaHeight)
			if !foundSlot {
				fmt.Println("\n❌ ERROR: Result area is full!")
				if e.DebugMode {
					DrawFullScreenDebugSnapshot(cfg, itemCount, "error_result_full", itemX, itemY, 0, 0)
				}
				fmt.Println("\n⚠ Warning: Please clear result area and restart.")
				return
			}

			fmt.Printf("     Pending: (%d, %d), Workbench: (%d, %d), Result: (%d, %d)\n",
				itemX, itemY, cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y, resultX, resultY)

			if e.DebugMode {
				if err := os.MkdirAll(config.SnapshotsDir, 0755); err != nil {
					fmt.Printf("⚠ Warning: Could not create snapshots directory: %v\n", err)
				}

				fmt.Println("  📸 [1/2] Saving fullscreen debug before move to workbench...")
				if err := DrawFullScreenDebugSnapshot(cfg, itemCount, "1_before_move_to_workbench", itemX, itemY, cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y); err != nil {
					fmt.Printf("❌ ERROR: Could not create debug snapshot: %v\n", err)
				}
				time.Sleep(500 * time.Millisecond)
			}

			if e.StopRequested.Load() {
				fmt.Println("\n✓ Stopped by user")
				return
			}
			fmt.Println("  → Moving to workbench...")
			if !e.MoveItem(itemX, itemY, cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y) {
				fmt.Println("\n✓ Stopped by user during move")
				return
			}
			time.Sleep(200 * time.Millisecond)

			if !e.StopRequested.Load() && !e.HasItemAtPosition(cfg, cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y) {
				fmt.Println("\n❌ ERROR: Failed to move item to workbench!")
				fmt.Println("   Source: pending area")
				fmt.Printf("   Destination: workbench (%d, %d)\n", cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y)

				if e.DebugMode {
					DrawFullScreenDebugSnapshot(cfg, itemCount, "error_move_to_workbench_failed", itemX, itemY, resultX, resultY)
				}

				fmt.Println("\n⚠  PAUSED - Please manually move the item to workbench")
				PlayVictorySound()
				fmt.Print("   Press Enter to continue after fixing...")
				fmt.Scanln()
			}

			if e.StopRequested.Load() {
				fmt.Println("\n✓ Stopped by user")
				return
			}

			cfg.ItemPos = cfg.WorkbenchTopLeft

			fmt.Println("  → Starting crafting...")
			craftSuccess := e.CraftSingleItem(&cfg, session, tempDir)

			if e.StopRequested.Load() {
				fmt.Println("\n✓ Stopped by user")
				return
			}

			if e.DebugMode {
				fmt.Println("  📸 [2/2] Saving fullscreen debug before move to result area...")
				if err := DrawFullScreenDebugSnapshot(cfg, itemCount, "2_before_move_to_result", cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y, resultX, resultY); err != nil {
					fmt.Printf("❌ ERROR: Could not create debug snapshot: %v\n", err)
				}
				time.Sleep(500 * time.Millisecond)
			}

			if e.StopRequested.Load() {
				fmt.Println("\n✓ Stopped by user")
				return
			}
			fmt.Println("  → Moving to result area...")
			if !e.MoveItem(cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y, resultX, resultY) {
				fmt.Println("\n✓ Stopped by user during move")
				return
			}
			time.Sleep(200 * time.Millisecond)

			if !e.StopRequested.Load() && !e.HasItemAtPosition(cfg, resultX, resultY) {
				fmt.Println("\n❌ ERROR: Failed to move item to result area!")
				fmt.Println("   Source: workbench")
				fmt.Printf("   Destination: result area (%d, %d)\n", resultX, resultY)

				if e.DebugMode {
					DrawFullScreenDebugSnapshot(cfg, itemCount, "error_move_to_result_failed", itemX, itemY, resultX, resultY)
				}

				fmt.Println("\n⚠  PAUSED - Please manually move the item to result area")
				PlayVictorySound()
				fmt.Print("   Press Enter to continue after fixing...")
				fmt.Scanln()
			}

			roundResult.EndPos = image.Point{X: resultX, Y: resultY}
			roundResult.Success = craftSuccess
			if session.TargetModHit {
				roundResult.TargetHit = true
				roundResult.TargetModName = session.TargetModName
				roundResult.TargetValue = session.TargetValue
			}

			session.RoundResults = append(session.RoundResults, roundResult)

			e.Emit("item_completed", ItemCompletedData{
				ItemNumber: itemCount,
				Success:    craftSuccess,
				ResultX:    resultX,
				ResultY:    resultY,
			})

			if craftSuccess {
				fmt.Printf("  ✓ Item #%d completed!\n", itemCount)
			} else {
				fmt.Printf("  ✓ Item #%d processed (no target match)\n", itemCount)
			}

			fmt.Println("  ✓ Ready for next item")
		}

		fmt.Printf("\n🎉 Batch crafting complete! Processed %d items.\n", itemCount)
		return
	}
}

// CraftSingleItem performs the crafting loop for a single item
func (e *Engine) CraftSingleItem(cfg *config.Config, session *CraftingSession, tempDir string) bool {
	fmt.Println("\nPicking up chaos orb...")
	robotgo.MoveSmooth(cfg.ChaosPos.X, cfg.ChaosPos.Y, 0.1, 0.1)
	HumanDelay(20, 10)
	robotgo.Click("right", false)
	HumanDelay(50, 10)

	robotgo.KeyToggle("shift", "down")
	HumanDelay(20, 5)

	robotgo.MoveSmooth(cfg.ItemPos.X, cfg.ItemPos.Y, 0.1, 0.1)
	HumanDelay(30, 10)

	defer func() {
		robotgo.KeyToggle("shift", "up")
	}()

	for attempt := 1; attempt <= cfg.ChaosPerRound; attempt++ {
		session.TotalRolls++

		{
			duration := time.Since(session.StartTime)
			rollsPerMin := 0.0
			if duration.Minutes() > 0 {
				rollsPerMin = float64(session.TotalRolls) / duration.Minutes()
			}
			e.Emit("roll_attempted", RollAttemptedData{
				AttemptNum:  attempt,
				MaxAttempts: cfg.ChaosPerRound,
				TotalRolls:  session.TotalRolls,
				RollsPerMin: rollsPerMin,
			})
		}

		if e.StopRequested.Load() {
			fmt.Println("\n[DEBUG] Stop flag detected in main loop")
			fmt.Println("\n✓ Stopped by user")
			return false
		}

		e.CheckMiddleMouseButton()

		if e.PauseRequested.Load() {
			fmt.Print("\n[DEBUG] Pause flag detected in main loop")
			fmt.Print("\n\n⏸  PAUSED - Press F12 to resume or Ctrl+C to exit... ")
			robotgo.KeyToggle("shift", "up")

			for e.PauseRequested.Load() && !e.StopRequested.Load() {
				time.Sleep(100 * time.Millisecond)
				e.CheckMiddleMouseButton()
			}

			if e.StopRequested.Load() {
				fmt.Println("\n✓ Stopped by user")
				return false
			}

			fmt.Println("\n▶  RESUMING in 5 seconds... Switch to game now!")
			for i := 5; i > 0; i-- {
				fmt.Printf("\r%d... ", i)
				time.Sleep(1 * time.Second)
			}
			fmt.Println("\r▶  RESUMED   ")
			robotgo.MoveSmooth(cfg.ChaosPos.X, cfg.ChaosPos.Y, 0.1, 0.1)
			HumanDelay(20, 10)
			robotgo.Click("right", false)
			HumanDelay(50, 10)
			robotgo.KeyToggle("shift", "down")
			HumanDelay(20, 5)
			robotgo.MoveSmooth(cfg.ItemPos.X, cfg.ItemPos.Y, 0.1, 0.1)
			HumanDelay(30, 10)
		}

		fmt.Printf("\r[%d/%d] Crafting... ", attempt, cfg.ChaosPerRound)

		robotgo.Click("left", false)
		HumanDelay(int(cfg.Delay.Milliseconds())/3, 10)

		robotgo.MoveSmooth(cfg.ItemPos.X+2, cfg.ItemPos.Y+2, 0.05, 0.05)
		HumanDelay(20, 5)
		robotgo.MoveSmooth(cfg.ItemPos.X, cfg.ItemPos.Y, 0.05, 0.05)
		HumanDelay(60, 20)

		bitmap := robotgo.CaptureScreen(
			cfg.TooltipRect.Min.X, cfg.TooltipRect.Min.Y,
			cfg.TooltipRect.Dx(), cfg.TooltipRect.Dy(),
		)
		img := robotgo.ToImage(bitmap)

		SaveImage(img, filepath.Join(config.SnapshotsDir, "current_tooltip.png"))
		e.Emit("tooltip_captured", TooltipCapturedData{Timestamp: time.Now().UnixMilli()})

		text, err := e.RunTesseractOCR(img, tempDir, cfg.GameLanguage)
		if err != nil {
			seqNum := e.SnapshotCounter.Load()
			fmt.Printf("\n\n❌ OCR ERROR #%d: %v\n", seqNum, err)
			fmt.Println("   Tooltip snapshot saved: snapshots/current_tooltip.png")
			fmt.Println("\n⚠  PAUSED - OCR failed to read item tooltip")
			fmt.Println("   Please check the tooltip snapshot and fix any issues")
			PlayVictorySound()
			fmt.Print("   Press Enter to retry this item...")
			fmt.Scanln()
			continue
		}

		if cfg.Debug {
			parsedLines := []string{}
			for _, line := range strings.Split(text, "\n") {
				line = strings.TrimSpace(line)
				if len(line) > 5 && regexp.MustCompile(`\d+`).MatchString(line) {
					parsedLines = append(parsedLines, line)
				}
			}
			if len(parsedLines) > 0 {
				fmt.Printf("\n[#%d Parsed] ", attempt)
				for _, line := range parsedLines {
					fmt.Printf("%s | ", line)
				}
				fmt.Println()
			}
		}

		if len(strings.TrimSpace(text)) < 10 {
			seqNum := e.SnapshotCounter.Load()
			fmt.Printf("\n⚠ Warning: OCR #%d incomplete (%d chars)\n", seqNum, len(text))
		}

		TrackMods(text, session, session.TotalRolls)

		e.Emit("mods_tracked", ModsTrackedData{
			OCRText:    text,
			ModStats:   session.ModStats,
			TotalRolls: session.TotalRolls,
		})

		matched, matchedMod, value := CheckAnyMod(text, cfg.TargetMods)

		if value == -1 {
			seqNum := e.SnapshotCounter.Load()
			fmt.Printf("\n\n⚠️  OCR FAILED #%d - Auto-pausing", seqNum)
			fmt.Printf("\n   Text: %s\n", strings.TrimSpace(text))

			e.PauseRequested.Store(true)
			fmt.Print("\n⏸  AUTO-PAUSED - Press F12 to resume or Ctrl+C to stop\n")

			robotgo.KeyToggle("shift", "up")

			for e.PauseRequested.Load() && !e.StopRequested.Load() {
				time.Sleep(100 * time.Millisecond)
				e.CheckMiddleMouseButton()
			}

			if e.StopRequested.Load() {
				fmt.Println("\n✓ Stopped by user")
				return false
			}

			fmt.Println("\n▶  RESUMING in 5 seconds... Switch to game now!")
			for i := 5; i > 0; i-- {
				fmt.Printf("\r%d... ", i)
				time.Sleep(1 * time.Second)
			}
			fmt.Println("\r▶  RESUMED   ")
			robotgo.MoveSmooth(cfg.ChaosPos.X, cfg.ChaosPos.Y, 0.1, 0.1)
			HumanDelay(20, 10)
			robotgo.Click("right", false)
			HumanDelay(50, 10)
			robotgo.KeyToggle("shift", "down")
			HumanDelay(20, 5)
			robotgo.MoveSmooth(cfg.ItemPos.X, cfg.ItemPos.Y, 0.1, 0.1)
			HumanDelay(30, 10)

			continue
		}

		if matched {
			seqNum := e.SnapshotCounter.Load()
			fmt.Printf("\n\n🎉 SUCCESS #%d (attempt %d)!\n", seqNum, attempt)
			fmt.Printf("   Found: %s = %d\n", matchedMod.Description, value)

			session.TargetModHit = true
			session.TargetModName = matchedMod.Description
			session.TargetValue = value

			e.Emit("target_found", TargetFoundData{
				ModName:    matchedMod.Description,
				Value:      value,
				AttemptNum: attempt,
				TotalRolls: session.TotalRolls,
			})

			PlayVictorySound()
			return true
		}
	}

	fmt.Printf("\n\n○ Used all %d chaos orbs for this round without finding target mod\n", cfg.ChaosPerRound)
	return false
}
