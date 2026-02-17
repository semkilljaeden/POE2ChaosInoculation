package engine

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"poe2-chaos-crafter/internal/config"

	"github.com/go-vgo/robotgo"
)

// LoadEmptyCellReference loads the reference image of an empty cell
func (e *Engine) LoadEmptyCellReference(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open reference image: %w", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode reference image: %w", err)
	}

	e.EmptyCellReference = img
	fmt.Printf("✓ Loaded empty cell reference image: %s (%dx%d)\n",
		path, img.Bounds().Dx(), img.Bounds().Dy())
	return nil
}

// CompareImages calculates the similarity between two images using Mean Squared Error (MSE)
// Returns a similarity score from 0.0 (identical) to 1.0 (completely different)
func CompareImages(img1, img2 image.Image) float64 {
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		return 1.0
	}

	var sumSquaredDiff float64
	pixelCount := 0

	for y := 0; y < bounds1.Dy(); y++ {
		for x := 0; x < bounds1.Dx(); x++ {
			r1, g1, b1, _ := img1.At(bounds1.Min.X+x, bounds1.Min.Y+y).RGBA()
			r2, g2, b2, _ := img2.At(bounds2.Min.X+x, bounds2.Min.Y+y).RGBA()

			r1_8, g1_8, b1_8 := uint8(r1>>8), uint8(g1>>8), uint8(b1>>8)
			r2_8, g2_8, b2_8 := uint8(r2>>8), uint8(g2>>8), uint8(b2>>8)

			dr := float64(int(r1_8) - int(r2_8))
			dg := float64(int(g1_8) - int(g2_8))
			db := float64(int(b1_8) - int(b2_8))

			sumSquaredDiff += dr*dr + dg*dg + db*db
			pixelCount++
		}
	}

	mse := sumSquaredDiff / float64(pixelCount)
	normalizedDiff := mse / 195075.0

	return normalizedDiff
}

// HasItemAtPosition checks if there's an item at the given position
func (e *Engine) HasItemAtPosition(cfg config.Config, x, y int) bool {
	if e.EmptyCellReference == nil {
		fmt.Println("     [hasItemAtPosition] WARNING: No reference image loaded, using fallback detection")
		return false
	}

	totalWidth := cfg.BackpackBottomRight.X - cfg.BackpackTopLeft.X
	totalHeight := cfg.BackpackBottomRight.Y - cfg.BackpackTopLeft.Y
	cellWidth := totalWidth / 12
	cellHeight := totalHeight / 5

	robotgo.Move(50, 50)
	time.Sleep(150 * time.Millisecond)

	captureWidth := int(float64(cellWidth) * 0.8)
	captureHeight := int(float64(cellHeight) * 0.8)
	captureX := x - captureWidth/2
	captureY := y - captureHeight/2

	bitmap := robotgo.CaptureScreen(captureX, captureY, captureWidth, captureHeight)
	img := robotgo.ToImage(bitmap)

	diffScore := CompareImages(img, e.EmptyCellReference)

	threshold := 0.05
	hasItem := diffScore > threshold

	if e.DebugMode {
		seqNum := e.SnapshotCounter.Add(1)
		resultStr := "EMPTY"
		if hasItem {
			resultStr = "HAS_ITEM"
		}
		debugFile := filepath.Join(config.SnapshotsDir, fmt.Sprintf("cell_check_%d_pos_%d_%d_%s_diff%.3f.png",
			seqNum, x, y, resultStr, diffScore))
		SaveImage(img, debugFile)
		fmt.Printf("     [hasItemAtPosition] (%d,%d): diff=%.3f (threshold: %.3f) -> %v (saved: %s)\n",
			x, y, diffScore, threshold, hasItem, debugFile)
	} else {
		fmt.Printf("     [hasItemAtPosition] (%d,%d): diff=%.3f -> %v\n", x, y, diffScore, hasItem)
	}

	return hasItem
}

// FindNextItemInArea scans the area and returns the position of the first item found
func (e *Engine) FindNextItemInArea(cfg config.Config, areaTopLeft image.Point, areaWidth, areaHeight int, skippedPositions map[string]bool) (int, int, bool) {
	cellWidth := (cfg.BackpackBottomRight.X - cfg.BackpackTopLeft.X) / 12
	cellHeight := (cfg.BackpackBottomRight.Y - cfg.BackpackTopLeft.Y) / 5

	fmt.Println("  [findNextItemInArea] Scanning pending area for next item...")
	positionsChecked := 0
	positionsSkipped := 0

	for row := 0; row < areaHeight; row += cfg.ItemHeight {
		for col := 0; col < areaWidth; col += cfg.ItemWidth {
			x := areaTopLeft.X + (col * cellWidth)
			y := areaTopLeft.Y + (row * cellHeight)

			posKey := fmt.Sprintf("%d,%d", x, y)

			if skippedPositions[posKey] {
				positionsSkipped++
				continue
			}

			positionsChecked++
			if e.HasItemAtPosition(cfg, x, y) {
				fmt.Printf("  [findNextItemInArea] ✓ Found item at (%d,%d) after checking %d positions (skipped %d)\n",
					x, y, positionsChecked, positionsSkipped)
				return x, y, true
			}
		}
	}

	fmt.Printf("  [findNextItemInArea] ✗ No items found (checked %d positions, skipped %d)\n",
		positionsChecked, positionsSkipped)
	return 0, 0, false
}

// FindEmptySlotInArea finds the first empty slot in an area
func (e *Engine) FindEmptySlotInArea(cfg config.Config, areaTopLeft image.Point, areaWidth, areaHeight int) (int, int, bool) {
	cellWidth := (cfg.BackpackBottomRight.X - cfg.BackpackTopLeft.X) / 12
	cellHeight := (cfg.BackpackBottomRight.Y - cfg.BackpackTopLeft.Y) / 5

	for row := 0; row < areaHeight; row += cfg.ItemHeight {
		for col := 0; col < areaWidth; col += cfg.ItemWidth {
			x := areaTopLeft.X + (col * cellWidth)
			y := areaTopLeft.Y + (row * cellHeight)

			if !e.HasItemAtPosition(cfg, x, y) {
				return x, y, true
			}
		}
	}

	return 0, 0, false
}
