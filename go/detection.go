package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/go-vgo/robotgo"
)

// emptyCellReference stores the reference image of an empty cell for comparison
var emptyCellReference image.Image

// loadEmptyCellReference loads the reference image of an empty cell
func loadEmptyCellReference(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open reference image: %w", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode reference image: %w", err)
	}

	emptyCellReference = img
	fmt.Printf("✓ Loaded empty cell reference image: %s (%dx%d)\n",
		path, img.Bounds().Dx(), img.Bounds().Dy())
	return nil
}

// compareImages calculates the similarity between two images using Mean Squared Error (MSE)
// Returns a similarity score from 0.0 (identical) to 1.0 (completely different)
func compareImages(img1, img2 image.Image) float64 {
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	// Images must be same size
	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		return 1.0 // Completely different if sizes don't match
	}

	var sumSquaredDiff float64
	pixelCount := 0

	for y := 0; y < bounds1.Dy(); y++ {
		for x := 0; x < bounds1.Dx(); x++ {
			r1, g1, b1, _ := img1.At(bounds1.Min.X+x, bounds1.Min.Y+y).RGBA()
			r2, g2, b2, _ := img2.At(bounds2.Min.X+x, bounds2.Min.Y+y).RGBA()

			// Convert to 8-bit (0-255)
			r1_8, g1_8, b1_8 := uint8(r1>>8), uint8(g1>>8), uint8(b1>>8)
			r2_8, g2_8, b2_8 := uint8(r2>>8), uint8(g2>>8), uint8(b2>>8)

			// Calculate squared difference for each channel
			dr := float64(int(r1_8) - int(r2_8))
			dg := float64(int(g1_8) - int(g2_8))
			db := float64(int(b1_8) - int(b2_8))

			sumSquaredDiff += dr*dr + dg*dg + db*db
			pixelCount++
		}
	}

	// Calculate MSE and normalize to 0-1 range
	// Max possible diff per channel: 255, so max squared diff = 255*255*3 = 195075 per pixel
	mse := sumSquaredDiff / float64(pixelCount)
	normalizedDiff := mse / 195075.0

	return normalizedDiff
}

// hasItemAtPosition checks if there's an item at the given position by comparing with reference empty cell
func hasItemAtPosition(cfg Config, x, y int) bool {
	// If no reference image loaded, fall back to old method
	if emptyCellReference == nil {
		fmt.Println("     [hasItemAtPosition] WARNING: No reference image loaded, using fallback detection")
		return false
	}

	// Calculate cell dimensions
	totalWidth := cfg.BackpackBottomRight.X - cfg.BackpackTopLeft.X
	totalHeight := cfg.BackpackBottomRight.Y - cfg.BackpackTopLeft.Y
	cellWidth := totalWidth / 12
	cellHeight := totalHeight / 5

	// Move mouse away from the cell to avoid tooltip interference
	// Move to a safe position far from the backpack (e.g., top-left corner of screen)
	robotgo.Move(50, 50)
	time.Sleep(150 * time.Millisecond) // Wait for any tooltip to disappear

	// Capture the actual cell area (item sprite area)
	// Use 80% of cell size to avoid edge artifacts
	captureWidth := int(float64(cellWidth) * 0.8)
	captureHeight := int(float64(cellHeight) * 0.8)
	captureX := x - captureWidth/2
	captureY := y - captureHeight/2

	bitmap := robotgo.CaptureScreen(captureX, captureY, captureWidth, captureHeight)
	img := robotgo.ToImage(bitmap)

	// Compare with reference empty cell image
	diffScore := compareImages(img, emptyCellReference)

	// If difference is above threshold, there's an item
	// Threshold: 0.05 means 5% different from empty cell = has item
	threshold := 0.05
	hasItem := diffScore > threshold

	// Save debug snapshot of captured cell
	seqNum := snapshotCounter.Add(1)
	resultStr := "EMPTY"
	if hasItem {
		resultStr = "HAS_ITEM"
	}
	debugFile := filepath.Join(snapshotsDir, fmt.Sprintf("cell_check_%d_pos_%d_%d_%s_diff%.3f.png",
		seqNum, x, y, resultStr, diffScore))
	saveImage(img, debugFile)

	// Log detection result for debugging
	fmt.Printf("     [hasItemAtPosition] (%d,%d): diff=%.3f (threshold: %.3f) -> %v (saved: %s)\n",
		x, y, diffScore, threshold, hasItem, debugFile)

	return hasItem
}

// findNextItemInArea scans the area and returns the position of the first item found using best-effort strategy
// Strategy: Jumps by item dimensions to efficiently find next item, skipping empty slots
// skippedPositions: map of "x,y" positions to skip (already processed items)
// Returns (x, y, true) if found, (0, 0, false) if no item found
func findNextItemInArea(cfg Config, areaTopLeft image.Point, areaWidth, areaHeight int, skippedPositions map[string]bool) (int, int, bool) {
	cellWidth := (cfg.BackpackBottomRight.X - cfg.BackpackTopLeft.X) / 12
	cellHeight := (cfg.BackpackBottomRight.Y - cfg.BackpackTopLeft.Y) / 5

	fmt.Println("  [findNextItemInArea] Scanning pending area for next item...")
	positionsChecked := 0
	positionsSkipped := 0

	// Best-effort scan: jump by item dimensions to find items efficiently
	// For a 2x3 item, this checks positions: (0,0), (2,0), (4,0), ... then (0,3), (2,3), ...
	for row := 0; row < areaHeight; row += cfg.ItemHeight {
		for col := 0; col < areaWidth; col += cfg.ItemWidth {
			// Calculate absolute position for this potential item (top-left corner)
			x := areaTopLeft.X + (col * cellWidth)
			y := areaTopLeft.Y + (row * cellHeight)

			// Create position key for tracking
			posKey := fmt.Sprintf("%d,%d", x, y)

			// Skip if this position was already processed
			if skippedPositions[posKey] {
				// Already moved - loop automatically jumps by itemWidth to next position
				positionsSkipped++
				continue
			}

			// Check if there's an item at this position
			positionsChecked++
			if hasItemAtPosition(cfg, x, y) {
				// Found an item! Return immediately
				fmt.Printf("  [findNextItemInArea] ✓ Found item at (%d,%d) after checking %d positions (skipped %d)\n",
					x, y, positionsChecked, positionsSkipped)
				return x, y, true
			}
			// Empty slot - loop automatically jumps by itemWidth to next potential position
		}
	}

	fmt.Printf("  [findNextItemInArea] ✗ No items found (checked %d positions, skipped %d)\n",
		positionsChecked, positionsSkipped)
	return 0, 0, false
}

// findEmptySlotInArea finds the first empty slot in an area using best-effort strategy
// Strategy: Jumps by item dimensions when occupied slot found to skip to next potential slot
// Returns (x, y, true) if found, (0, 0, false) if area is full
func findEmptySlotInArea(cfg Config, areaTopLeft image.Point, areaWidth, areaHeight int) (int, int, bool) {
	cellWidth := (cfg.BackpackBottomRight.X - cfg.BackpackTopLeft.X) / 12
	cellHeight := (cfg.BackpackBottomRight.Y - cfg.BackpackTopLeft.Y) / 5

	// Best-effort scan: jump by item dimensions to find empty slots efficiently
	// For a 2x3 item, this checks positions: (0,0), (2,0), (4,0), ... then (0,3), (2,3), ...
	for row := 0; row < areaHeight; row += cfg.ItemHeight {
		for col := 0; col < areaWidth; col += cfg.ItemWidth {
			// Calculate absolute position for this potential slot (top-left corner)
			x := areaTopLeft.X + (col * cellWidth)
			y := areaTopLeft.Y + (row * cellHeight)

			// Check if this slot is empty
			if !hasItemAtPosition(cfg, x, y) {
				// Found empty slot! Return immediately
				return x, y, true
			}
			// Occupied - loop automatically jumps by itemWidth to next potential position
		}
	}

	return 0, 0, false
}
