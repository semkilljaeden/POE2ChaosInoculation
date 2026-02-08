package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
)

// craft is the main crafting function that handles batch mode processing
func craft(cfg Config) {
	// Initialize cooldown time and snapshot counter
	pauseToggleCooldown.Store(time.Now())
	snapshotCounter.Store(0)

	// Initialize crafting session for tracking
	session := &CraftingSession{
		StartTime: time.Now(),
		ModStats:  make(map[string]*ModStat),
	}

	// Setup signal handler for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\n[DEBUG] Signal received (Ctrl+C)")
		fmt.Println("🛑 Stop requested... Exiting safely.")
		stopRequested.Store(true)
		fmt.Println("[DEBUG] stopRequested flag set to true")
	}()

	// Ensure report is generated even if interrupted
	defer func() {
		session.EndTime = time.Now()
		generateReport(session, cfg)
	}()

	// Clean up old debug snapshots from previous runs
	cleanupDebugSnapshots()

	// Create resource directory if it doesn't exist
	if err := os.MkdirAll(resourceDir, 0755); err != nil {
		fmt.Printf("⚠ WARNING: Could not create resource directory: %v\n", err)
	}

	// Load empty cell reference image for item detection
	emptyCellRefPath := filepath.Join(resourceDir, "empty_cell_reference.png")
	fmt.Printf("\n📸 Loading empty cell reference image from: %s\n", emptyCellRefPath)
	if err := loadEmptyCellReference(emptyCellRefPath); err != nil {
		fmt.Printf("⚠ WARNING: Could not load empty cell reference: %v\n", err)
		fmt.Println("   Please save an empty cell snapshot as 'resource/empty_cell_reference.png'")
		fmt.Println("   Item detection may not work correctly without it!")
	}

	// Generate grid snapshot at start of crafting
	if cfg.BackpackTopLeft.X != 0 && cfg.BackpackBottomRight.X != 0 {
		fmt.Println("\n📸 Generating grid snapshot...")
		if err := drawBackpackGrid(cfg); err != nil {
			fmt.Printf("⚠ Warning: Could not create grid snapshot: %v\n", err)
		} else {
			fmt.Println("✓ Grid snapshot: backpack_grid_debug.png")
		}
	}

	// Create temp directory for OCR
	tempDir := filepath.Join(os.TempDir(), "poe2_crafter")
	os.MkdirAll(tempDir, 0755)

	// Batch mode: process multiple items from pending area
	if cfg.UseBatchMode {
		fmt.Println("\n🔄 BATCH MODE ENABLED")
		fmt.Printf("📦 Pending area: %dx%d cells\n", cfg.PendingAreaWidth, cfg.PendingAreaHeight)
		fmt.Printf("🎯 Workbench: (%d, %d)\n", cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y)
		fmt.Printf("✅ Result area: %dx%d cells\n\n", cfg.ResultAreaWidth, cfg.ResultAreaHeight)

		// Track processed positions to avoid re-detecting moved items
		processedPositions := make(map[string]bool)
		itemCount := 0

		for {
			// Check if stop requested
			if stopRequested.Load() {
				fmt.Println("\n✓ Stopped by user")
				return
			}

			// Find next item in pending area (skip already processed positions)
			itemX, itemY, found := findNextItemInArea(cfg, cfg.PendingAreaTopLeft, cfg.PendingAreaWidth, cfg.PendingAreaHeight, processedPositions)
			if !found {
				fmt.Println("\n✓ No more items in pending area")
				break
			}

			itemCount++
			fmt.Printf("\n📦 Processing item #%d from pending area at (%d, %d)...\n", itemCount, itemX, itemY)

			// Create round result for tracking
			roundResult := RoundResult{
				RoundNumber: itemCount,
				StartPos:    image.Point{X: itemX, Y: itemY},
				Success:     false, // Will update at end
			}

			// Mark this position as processed
			posKey := fmt.Sprintf("%d,%d", itemX, itemY)
			processedPositions[posKey] = true

			// Find empty slot in result area FIRST (before moving anything)
			resultX, resultY, foundSlot := findEmptySlotInArea(cfg, cfg.ResultAreaTopLeft, cfg.ResultAreaWidth, cfg.ResultAreaHeight)
			if !foundSlot {
				fmt.Println("\n❌ ERROR: Result area is full!")
				fmt.Println("   📸 Saving error snapshot...")
				drawFullScreenDebugSnapshot(cfg, itemCount, "error_result_full", itemX, itemY, 0, 0)
				fmt.Println("\n⚠ Warning: Please clear result area and restart.")
				return
			}

			// Generate debug snapshots BEFORE moving - shows the plan
			fmt.Println("  → Generating debug snapshots...")
			fmt.Printf("     Pending: (%d, %d), Workbench: (%d, %d), Result: (%d, %d)\n",
				itemX, itemY, cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y, resultX, resultY)

			// Ensure snapshots directory exists
			if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
				fmt.Printf("⚠ Warning: Could not create snapshots directory: %v\n", err)
			}

			// Generate fullscreen debug snapshot before move to workbench
			fmt.Println("  📸 [1/2] Saving fullscreen debug before move to workbench...")
			if err := drawFullScreenDebugSnapshot(cfg, itemCount, "1_before_move_to_workbench", itemX, itemY, cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y); err != nil {
				fmt.Printf("❌ ERROR: Could not create debug snapshot: %v\n", err)
			} else {
				fmt.Println("  ✓ Fullscreen debug snapshot saved")
			}
			time.Sleep(500 * time.Millisecond) // Give user time to see the snapshot

			// Move item from pending to workbench
			fmt.Println("  → Moving to workbench...")
			moveItem(itemX, itemY, cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y)
			time.Sleep(200 * time.Millisecond)

			// Verify the item was moved to workbench
			if !hasItemAtPosition(cfg, cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y) {
				fmt.Println("\n❌ ERROR: Failed to move item to workbench!")
				fmt.Println("   Source: pending area")
				fmt.Printf("   Destination: workbench (%d, %d)\n", cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y)

				// Save error snapshot
				fmt.Println("   📸 Saving error snapshot...")
				drawFullScreenDebugSnapshot(cfg, itemCount, "error_move_to_workbench_failed", itemX, itemY, resultX, resultY)

				fmt.Println("\n⚠  PAUSED - Please manually move the item to workbench")
				playVictorySound() // Alert sound
				fmt.Print("   Press Enter to continue after fixing...")
				fmt.Scanln()
			}

			// Update ItemPos to workbench for crafting
			cfg.ItemPos = cfg.WorkbenchTopLeft

			// Craft this item (use existing single-item logic below)
			fmt.Println("  → Starting crafting...")
			craftSuccess := craftSingleItem(&cfg, session, tempDir)

			// Generate fullscreen debug snapshot before move to result area
			fmt.Println("  📸 [2/2] Saving fullscreen debug before move to result area...")
			if err := drawFullScreenDebugSnapshot(cfg, itemCount, "2_before_move_to_result", cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y, resultX, resultY); err != nil {
				fmt.Printf("❌ ERROR: Could not create debug snapshot: %v\n", err)
			} else {
				fmt.Println("  ✓ Fullscreen debug snapshot saved")
			}
			time.Sleep(500 * time.Millisecond) // Give user time to see the snapshot

			// Move item from workbench to result area
			fmt.Println("  → Moving to result area...")
			moveItem(cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y, resultX, resultY)
			time.Sleep(200 * time.Millisecond)

			// Verify the item was moved to result area
			if !hasItemAtPosition(cfg, resultX, resultY) {
				fmt.Println("\n❌ ERROR: Failed to move item to result area!")
				fmt.Println("   Source: workbench")
				fmt.Printf("   Destination: result area (%d, %d)\n", resultX, resultY)

				// Save error snapshot
				fmt.Println("   📸 Saving error snapshot...")
				drawFullScreenDebugSnapshot(cfg, itemCount, "error_move_to_result_failed", itemX, itemY, resultX, resultY)

				fmt.Println("\n⚠  PAUSED - Please manually move the item to result area")
				playVictorySound() // Alert sound
				fmt.Print("   Press Enter to continue after fixing...")
				fmt.Scanln()
			}

			// Finalize round result
			roundResult.EndPos = image.Point{X: resultX, Y: resultY}
			roundResult.Success = craftSuccess
			if session.TargetModHit {
				roundResult.TargetHit = true
				roundResult.TargetModName = session.TargetModName
				roundResult.TargetValue = session.TargetValue
			}

			// Save round result
			session.RoundResults = append(session.RoundResults, roundResult)

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

// craftSingleItem performs the crafting loop for a single item
func craftSingleItem(cfg *Config, session *CraftingSession, tempDir string) bool {
	// Initial setup: Right-click chaos orb once
	fmt.Println("\nPicking up chaos orb...")
	robotgo.MoveSmooth(cfg.ChaosPos.X, cfg.ChaosPos.Y, 0.1, 0.1) // Very fast movement
	humanDelay(20, 10)
	robotgo.Click("right", false)
	humanDelay(50, 10)

	// Hold Shift for continuous crafting
	robotgo.KeyToggle("shift", "down")
	humanDelay(20, 5)

	// Move to item position
	robotgo.MoveSmooth(cfg.ItemPos.X, cfg.ItemPos.Y, 0.1, 0.1)
	humanDelay(30, 10)

	defer func() {
		// Always release Shift when exiting
		robotgo.KeyToggle("shift", "up")
	}()

	for attempt := 1; attempt <= cfg.ChaosPerRound; attempt++ {
		session.TotalRolls++

		// Check if stop requested
		if stopRequested.Load() {
			fmt.Println("\n[DEBUG] Stop flag detected in main loop")
			fmt.Println("\n✓ Stopped by user")
			return false
		}

		// Check for pause toggle
		checkMiddleMouseButton()

		// Check if pause requested
		if pauseRequested.Load() {
			fmt.Print("\n[DEBUG] Pause flag detected in main loop")
			fmt.Print("\n\n⏸  PAUSED - Press F12 to resume or Ctrl+C to exit... ")
			// Release shift while paused
			robotgo.KeyToggle("shift", "up")

			// Wait until pause is released
			for pauseRequested.Load() && !stopRequested.Load() {
				time.Sleep(100 * time.Millisecond)
				checkMiddleMouseButton() // Check for F12 to resume
			}

			if stopRequested.Load() {
				fmt.Println("\n✓ Stopped by user")
				return false
			}

			// Resume - countdown and re-grab chaos
			fmt.Println("\n▶  RESUMING in 5 seconds... Switch to game now!")
			for i := 5; i > 0; i-- {
				fmt.Printf("\r%d... ", i)
				time.Sleep(1 * time.Second)
			}
			fmt.Println("\r▶  RESUMED   ")
			robotgo.MoveSmooth(cfg.ChaosPos.X, cfg.ChaosPos.Y, 0.1, 0.1)
			humanDelay(20, 10)
			robotgo.Click("right", false)
			humanDelay(50, 10)
			robotgo.KeyToggle("shift", "down")
			humanDelay(20, 5)
			robotgo.MoveSmooth(cfg.ItemPos.X, cfg.ItemPos.Y, 0.1, 0.1)
			humanDelay(30, 10)
		}

		fmt.Printf("\r[%d/%d] Crafting... ", attempt, cfg.ChaosPerRound)

		// Left-click item to apply chaos (Shift is already held)
		robotgo.Click("left", false)
		humanDelay(int(cfg.Delay.Milliseconds())/3, 10) // Even faster with minimal variation

		// Small movement to ensure tooltip refresh
		robotgo.MoveSmooth(cfg.ItemPos.X+2, cfg.ItemPos.Y+2, 0.05, 0.05)
		humanDelay(20, 5)
		robotgo.MoveSmooth(cfg.ItemPos.X, cfg.ItemPos.Y, 0.05, 0.05)
		humanDelay(60, 20) // Even faster tooltip wait

		// Capture tooltip (no Alt key needed)
		bitmap := robotgo.CaptureScreen(
			cfg.TooltipRect.Min.X, cfg.TooltipRect.Min.Y,
			cfg.TooltipRect.Dx(), cfg.TooltipRect.Dy(),
		)
		img := robotgo.ToImage(bitmap)

		// Save current tooltip for debugging (always updated, no counter)
		saveImage(img, filepath.Join(snapshotsDir, "current_tooltip.png"))

		// OCR using command-line Tesseract
		text, err := runTesseractOCR(img, tempDir)
		if err != nil {
			seqNum := snapshotCounter.Load()
			fmt.Printf("\n\n❌ OCR ERROR #%d: %v\n", seqNum, err)
			fmt.Println("   Tooltip snapshot saved: snapshots/current_tooltip.png")
			fmt.Println("\n⚠  PAUSED - OCR failed to read item tooltip")
			fmt.Println("   Please check the tooltip snapshot and fix any issues")
			playVictorySound() // Alert sound
			fmt.Print("   Press Enter to retry this item...")
			fmt.Scanln()
			continue
		}

		// Log OCR text if debug enabled
		if cfg.Debug {
			// Extract lines that contain numbers (the actual mod values)
			parsedLines := []string{}
			for _, line := range strings.Split(text, "\n") {
				line = strings.TrimSpace(line)
				// Only show lines with digits (mod values)
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

		// Check if text looks incomplete (very short or empty)
		if len(strings.TrimSpace(text)) < 10 {
			seqNum := snapshotCounter.Load()
			fmt.Printf("\n⚠ Warning: OCR #%d incomplete (%d chars)\n", seqNum, len(text))
		}

		// Track all mods found in this roll
		trackMods(text, session)

		// Check if any of the target mods matched
		matched, matchedMod, value := checkAnyMod(text, cfg.TargetMods)

		// If value is -1, it means no valid mod pattern was detected
		if value == -1 {
			seqNum := snapshotCounter.Load()
			fmt.Printf("\n\n⚠️  OCR FAILED #%d - Auto-pausing", seqNum)
			fmt.Printf("\n   Text: %s\n", strings.TrimSpace(text))

			// Auto-pause for debugging
			pauseRequested.Store(true)
			fmt.Print("\n⏸  AUTO-PAUSED - Press F12 to resume or Ctrl+C to stop\n")

			// Release shift while paused
			robotgo.KeyToggle("shift", "up")

			// Wait until pause is released
			for pauseRequested.Load() && !stopRequested.Load() {
				time.Sleep(100 * time.Millisecond)
				checkMiddleMouseButton()
			}

			if stopRequested.Load() {
				fmt.Println("\n✓ Stopped by user")
				return false
			}

			// Resume - countdown and re-grab chaos
			fmt.Println("\n▶  RESUMING in 5 seconds... Switch to game now!")
			for i := 5; i > 0; i-- {
				fmt.Printf("\r%d... ", i)
				time.Sleep(1 * time.Second)
			}
			fmt.Println("\r▶  RESUMED   ")
			robotgo.MoveSmooth(cfg.ChaosPos.X, cfg.ChaosPos.Y, 0.1, 0.1)
			humanDelay(20, 10)
			robotgo.Click("right", false)
			humanDelay(50, 10)
			robotgo.KeyToggle("shift", "down")
			humanDelay(20, 5)
			robotgo.MoveSmooth(cfg.ItemPos.X, cfg.ItemPos.Y, 0.1, 0.1)
			humanDelay(30, 10)

			continue
		}

		if matched {
			seqNum := snapshotCounter.Load()
			fmt.Printf("\n\n🎉 SUCCESS #%d (attempt %d)!\n", seqNum, attempt)
			fmt.Printf("   Found: %s = %d\n", matchedMod.Description, value)

			session.TargetModHit = true
			session.TargetModName = matchedMod.Description
			session.TargetValue = value

			// Play victory melody
			playVictorySound()
			return true
		}
	}

	fmt.Printf("\n\n○ Used all %d chaos orbs for this round without finding target mod\n", cfg.ChaosPerRound)
	return false
}

// preprocessForOCR improves image quality for better OCR accuracy
func preprocessForOCR(img image.Image) image.Image {
	bounds := img.Bounds()

	// Step 1: Convert to grayscale
	grayImg := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}

	// Step 2: More aggressive contrast enhancement and binarization
	contrastImg := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldColor := grayImg.GrayAt(x, y)
			brightness := oldColor.Y

			// More aggressive thresholding for clearer text
			var newBrightness uint8
			if brightness > 80 {
				// Make bright pixels pure white (text)
				enhanced := float64(brightness) * 1.5
				if enhanced > 255 {
					enhanced = 255
				}
				newBrightness = uint8(enhanced)
			} else {
				// Make dark pixels darker (background)
				enhanced := float64(brightness) * 0.3
				newBrightness = uint8(enhanced)
			}

			contrastImg.SetGray(x, y, color.Gray{Y: newBrightness})
		}
	}

	// Step 3: Scale up 3x for even better OCR (higher resolution = better accuracy)
	scaledWidth := bounds.Dx() * 3
	scaledHeight := bounds.Dy() * 3
	scaledImg := image.NewGray(image.Rect(0, 0, scaledWidth, scaledHeight))

	// Simple nearest-neighbor scaling
	for y := 0; y < scaledHeight; y++ {
		for x := 0; x < scaledWidth; x++ {
			srcX := bounds.Min.X + x/3
			srcY := bounds.Min.Y + y/3
			scaledImg.Set(x, y, contrastImg.At(srcX, srcY))
		}
	}

	return scaledImg
}

// cropImage creates a cropped version of the image
func cropImage(img image.Image, cropPercent int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate crop amount (remove cropPercent from each edge)
	cropX := (width * cropPercent) / 100
	cropY := (height * cropPercent) / 100

	// Ensure we don't crop too much
	if cropX*2 >= width || cropY*2 >= height {
		return img
	}

	// Create new rectangle with cropped bounds
	newRect := image.Rect(
		bounds.Min.X+cropX,
		bounds.Min.Y+cropY,
		bounds.Max.X-cropX,
		bounds.Max.Y-cropY,
	)

	// Create new image with cropped bounds
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}

	if si, ok := img.(subImager); ok {
		return si.SubImage(newRect)
	}

	return img
}

// runTesseractOCRSingle runs OCR with specific settings
func runTesseractOCRSingle(img image.Image, tempDir string, suffix string, psm int, usePreprocess bool) (string, error) {
	var processedImg image.Image
	if usePreprocess {
		processedImg = preprocessForOCR(img)
	} else {
		processedImg = img
	}

	// Save preprocessed image to temp file
	tempImg := filepath.Join(tempDir, fmt.Sprintf("temp_ocr_%s.png", suffix))
	if err := saveImage(processedImg, tempImg); err != nil {
		return "", fmt.Errorf("failed to save temp image: %w", err)
	}
	defer os.Remove(tempImg)

	// Output file (tesseract adds .txt automatically)
	tempOut := filepath.Join(tempDir, fmt.Sprintf("temp_ocr_%s", suffix))
	tempOutTxt := tempOut + ".txt"
	defer os.Remove(tempOutTxt)

	// Run tesseract with specified PSM mode
	// Character whitelist: English letters, numbers, spaces, and common symbols in POE2
	cmd := exec.Command("tesseract", tempImg, tempOut, "-l", "eng",
		"--psm", fmt.Sprintf("%d", psm),
		"--oem", "1",
		"-c", "tessedit_char_whitelist=ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789 +-()%#")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tesseract failed: %w", err)
	}

	// Read output
	data, err := os.ReadFile(tempOutTxt)
	if err != nil {
		return "", fmt.Errorf("failed to read OCR output: %w", err)
	}

	return string(data), nil
}

// runTesseractOCR runs OCR with multiple strategies and returns the best result
func runTesseractOCR(img image.Image, tempDir string) (string, error) {
	seqNum := snapshotCounter.Add(1)

	// Save original and preprocessed snapshots
	debugOriginalFile := filepath.Join(snapshotsDir, fmt.Sprintf("snap_%d_raw.png", seqNum))
	debugProcessedFile := filepath.Join(snapshotsDir, fmt.Sprintf("snap_%d_processed.png", seqNum))
	saveImage(img, debugOriginalFile)
	saveImage(preprocessForOCR(img), debugProcessedFile)

	type ocrStrategy struct {
		name          string
		psm           int
		usePreprocess bool
	}

	// Fast path: Try most promising strategies first with early exit
	fastStrategies := []ocrStrategy{
		{"PSM6_raw", 6, false},         // Fast: raw image, no preprocessing
		{"PSM6_preprocessed", 6, true}, // Fallback: with preprocessing
	}

	bestText := ""
	bestScore := 0

	for _, strategy := range fastStrategies {
		text, err := runTesseractOCRSingle(img, tempDir, strategy.name, strategy.psm, strategy.usePreprocess)
		if err != nil {
			continue
		}

		// Score by: text length + bonus for having numbers (mod values)
		textLen := len(strings.TrimSpace(text))
		hasNumbers := regexp.MustCompile(`\d+`).MatchString(text)
		score := textLen
		if hasNumbers {
			score += 50 // Bonus for having numbers (likely mod values)
		}

		// Keep the best result
		if score > bestScore {
			bestText = text
			bestScore = score
		}

		// Early exit if we got good results (saves time)
		if bestScore >= 80 {
			return bestText, nil
		}
	}

	// If we got decent results, return it
	if bestScore >= 30 {
		return bestText, nil
	}

	// Slow path: Only if fast strategies failed, try alternative PSM modes
	fmt.Print(" [Trying alternatives...]")
	slowStrategies := []ocrStrategy{
		{"PSM4_preprocessed", 4, true},   // Single column
		{"PSM11_preprocessed", 11, true}, // Sparse text
	}

	for _, strategy := range slowStrategies {
		text, err := runTesseractOCRSingle(img, tempDir, strategy.name, strategy.psm, strategy.usePreprocess)
		if err != nil {
			continue
		}

		textLen := len(strings.TrimSpace(text))
		hasNumbers := regexp.MustCompile(`\d+`).MatchString(text)
		score := textLen
		if hasNumbers {
			score += 50
		}

		if score > bestScore {
			bestText = text
			bestScore = score
		}

		// Early exit if we got good results
		if bestScore >= 80 {
			return bestText, nil
		}
	}

	// Return best result found (even if not perfect)
	if bestScore > 0 {
		return bestText, nil
	}

	// If everything failed, return empty with note
	return "", fmt.Errorf("all OCR strategies failed")
}

// extractModLines extracts mod lines matching the pattern: prefix/suffix + spaces + mod text + spaces + tier
func extractModLines(text string) []string {
	// Pattern: (PREFIX|SUFFIX) followed by spaces, mod text, spaces, and tier (T1, T2, etc.)
	// Example: "PREFIX     +50 TO MAXIMUM LIFE     T2"
	modLinePattern := regexp.MustCompile(`(?i)(PREFIX|SUFFIX)\s{2,}.+?\s{2,}T\d+`)
	matches := modLinePattern.FindAllString(text, -1)
	return matches
}

// checkMod checks if a specific mod appears in the OCR text
func checkMod(text string, mod ModRequirement) (bool, int) {
	// Search for the mod pattern directly in the OCR text
	re := regexp.MustCompile(mod.Pattern)
	matches := re.FindAllStringSubmatch(text, -1)

	// If no matches at all, might be an OCR issue
	if len(matches) == 0 {
		// Check if text is suspiciously short (might indicate OCR failure)
		if len(strings.TrimSpace(text)) < 10 {
			fmt.Printf("\n⚠ WARNING: OCR text seems incomplete or empty")
			return false, -1 // -1 signals pattern mismatch for auto-pause
		}
		// Text exists but mod not found - this is normal, keep crafting
		return false, 0
	}

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		// Extract number from the match
		numStr := strings.TrimSpace(match[1])
		value, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}

		// Check if meets minimum value requirement
		if value >= mod.MinValue {
			return true, value
		}
	}

	return false, 0
}

// checkAnyMod checks if any of the target mods appear in the text
func checkAnyMod(text string, mods []ModRequirement) (bool, ModRequirement, int) {
	for _, mod := range mods {
		matched, value := checkMod(text, mod)
		if matched {
			return true, mod, value
		}
		// If value is -1, OCR failed - propagate this
		if value == -1 {
			return false, ModRequirement{}, -1
		}
	}
	return false, ModRequirement{}, 0
}
