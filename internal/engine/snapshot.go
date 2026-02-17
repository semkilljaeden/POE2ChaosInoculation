package engine

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"poe2-chaos-crafter/internal/config"

	"github.com/go-vgo/robotgo"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// HumanDelay adds a random delay to simulate human behavior (base ± variation in ms)
func HumanDelay(baseMs int, variationMs int) {
	delay := baseMs + rand.Intn(variationMs*2) - variationMs
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// SaveImage saves an image to a file
func SaveImage(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// CleanupDebugSnapshots removes old snapshots folder and creates a fresh one
func CleanupDebugSnapshots() {
	// Remove entire snapshots directory if it exists
	if _, err := os.Stat(config.SnapshotsDir); err == nil {
		if err := os.RemoveAll(config.SnapshotsDir); err == nil {
			fmt.Printf("🧹 Cleaned up previous snapshots\n")
		}
	}

	// Create fresh snapshots directory
	if err := os.MkdirAll(config.SnapshotsDir, 0755); err != nil {
		fmt.Printf("⚠ Warning: Could not create snapshots directory: %v\n", err)
	}
}

// DrawBackpackGrid creates a debug image with the backpack grid overlay
func DrawBackpackGrid(cfg config.Config) error {
	// Capture the backpack area
	width := cfg.BackpackBottomRight.X - cfg.BackpackTopLeft.X
	height := cfg.BackpackBottomRight.Y - cfg.BackpackTopLeft.Y

	bitmap := robotgo.CaptureScreen(cfg.BackpackTopLeft.X, cfg.BackpackTopLeft.Y, width, height)
	img := robotgo.ToImage(bitmap)

	// Create a new RGBA image for drawing
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	cellWidth := width / 12
	cellHeight := height / 5

	// Draw grid lines
	gridColor := color.RGBA{0, 255, 0, 255} // Green

	// Vertical lines
	for i := 0; i <= 12; i++ {
		x := i * cellWidth
		for y := 0; y < height; y++ {
			if x < width {
				rgba.Set(x, y, gridColor)
			}
		}
	}

	// Horizontal lines
	for i := 0; i <= 5; i++ {
		y := i * cellHeight
		for x := 0; x < width; x++ {
			if y < height {
				rgba.Set(x, y, gridColor)
			}
		}
	}

	// Draw cell labels (Row,Col format)
	labelColor := color.RGBA{255, 255, 0, 255} // Yellow

	for row := 0; row < 5; row++ {
		for col := 0; col < 12; col++ {
			labelX := col*cellWidth + 5
			labelY := row*cellHeight + 5
			label := fmt.Sprintf("%d,%d", row, col)
			drawString(rgba, labelX, labelY, label, labelColor)
		}
	}

	// Save debug snapshot
	debugFile := filepath.Join(config.SnapshotsDir, "backpack_grid_debug.png")
	if err := SaveImage(rgba, debugFile); err != nil {
		return fmt.Errorf("failed to save backpack grid debug image: %w", err)
	}

	fmt.Printf("✓ Backpack grid debug saved: %s\n", debugFile)
	return nil
}

// drawString draws a string on an image at the specified position
func drawString(img *image.RGBA, x, y int, label string, col color.Color) {
	point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

// DrawFullScreenDebugSnapshot captures the entire screen and labels all important areas
func DrawFullScreenDebugSnapshot(cfg config.Config, itemNum int, stepName string, itemX, itemY, resultX, resultY int) error {
	// Get screen dimensions
	screenWidth, screenHeight := robotgo.GetScreenSize()
	fmt.Printf("     Screen size: %dx%d\n", screenWidth, screenHeight)

	// Capture entire screen
	bitmap := robotgo.CaptureScreen()
	img := robotgo.ToImage(bitmap)

	// Get actual captured dimensions
	bounds := img.Bounds()
	actualWidth := bounds.Dx()
	actualHeight := bounds.Dy()
	fmt.Printf("     Captured: %dx%d\n", actualWidth, actualHeight)

	// Use actual captured dimensions
	screenWidth = actualWidth
	screenHeight = actualHeight

	// Create RGBA image for drawing
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Helper to draw thick rectangle with label
	drawLabeledRect := func(x1, y1, x2, y2 int, col color.RGBA, label string, thickness int) {
		for t := 0; t < thickness; t++ {
			for x := x1 - t; x <= x2+t; x++ {
				if x >= 0 && x < screenWidth {
					if y1-t >= 0 && y1-t < screenHeight {
						rgba.Set(x, y1-t, col)
					}
					if y2+t >= 0 && y2+t < screenHeight {
						rgba.Set(x, y2+t, col)
					}
				}
			}
			for y := y1 - t; y <= y2+t; y++ {
				if y >= 0 && y < screenHeight {
					if x1-t >= 0 && x1-t < screenWidth {
						rgba.Set(x1-t, y, col)
					}
					if x2+t >= 0 && x2+t < screenWidth {
						rgba.Set(x2+t, y, col)
					}
				}
			}
		}
		drawString(rgba, x1+5, y1-15, label, col)
	}

	// Helper to draw a circle
	drawCircle := func(centerX, centerY, radius int, col color.RGBA) {
		for y := centerY - radius; y <= centerY+radius; y++ {
			for x := centerX - radius; x <= centerX+radius; x++ {
				if x >= 0 && x < screenWidth && y >= 0 && y < screenHeight {
					dx := x - centerX
					dy := y - centerY
					if dx*dx+dy*dy <= radius*radius {
						rgba.Set(x, y, col)
					}
				}
			}
		}
	}

	// Calculate cell dimensions
	cellWidth := (cfg.BackpackBottomRight.X - cfg.BackpackTopLeft.X) / 12
	cellHeight := (cfg.BackpackBottomRight.Y - cfg.BackpackTopLeft.Y) / 5

	// 1. Draw backpack grid outline (WHITE)
	drawLabeledRect(
		cfg.BackpackTopLeft.X, cfg.BackpackTopLeft.Y,
		cfg.BackpackBottomRight.X, cfg.BackpackBottomRight.Y,
		color.RGBA{255, 255, 255, 255}, "BACKPACK GRID (5x12)", 5)

	// 2. Draw chaos orb position (RED circle)
	drawCircle(cfg.ChaosPos.X, cfg.ChaosPos.Y, 15, color.RGBA{255, 0, 0, 255})
	drawString(rgba, cfg.ChaosPos.X+20, cfg.ChaosPos.Y, "CHAOS ORB", color.RGBA{255, 0, 0, 255})

	// 3. Draw pending area (CYAN)
	pendingX1 := cfg.PendingAreaTopLeft.X - cellWidth/2
	pendingY1 := cfg.PendingAreaTopLeft.Y - cellHeight/2
	pendingX2 := pendingX1 + (cfg.PendingAreaWidth * cellWidth)
	pendingY2 := pendingY1 + (cfg.PendingAreaHeight * cellHeight)
	drawLabeledRect(pendingX1, pendingY1, pendingX2, pendingY2,
		color.RGBA{0, 255, 255, 255}, fmt.Sprintf("PENDING AREA (%dx%d cells)", cfg.PendingAreaWidth, cfg.PendingAreaHeight), 6)

	// 4. Draw workbench (ORANGE)
	workbenchX1 := cfg.WorkbenchTopLeft.X - cellWidth/2
	workbenchY1 := cfg.WorkbenchTopLeft.Y - cellHeight/2
	workbenchX2 := workbenchX1 + (cfg.ItemWidth * cellWidth)
	workbenchY2 := workbenchY1 + (cfg.ItemHeight * cellHeight)
	drawLabeledRect(workbenchX1, workbenchY1, workbenchX2, workbenchY2,
		color.RGBA{255, 165, 0, 255}, fmt.Sprintf("WORKBENCH (%dx%d)", cfg.ItemWidth, cfg.ItemHeight), 6)

	// 5. Draw result area (MAGENTA)
	resultAreaX1 := cfg.ResultAreaTopLeft.X - cellWidth/2
	resultAreaY1 := cfg.ResultAreaTopLeft.Y - cellHeight/2
	resultAreaX2 := resultAreaX1 + (cfg.ResultAreaWidth * cellWidth)
	resultAreaY2 := resultAreaY1 + (cfg.ResultAreaHeight * cellHeight)
	drawLabeledRect(resultAreaX1, resultAreaY1, resultAreaX2, resultAreaY2,
		color.RGBA{255, 0, 255, 255}, fmt.Sprintf("RESULT AREA (%dx%d cells)", cfg.ResultAreaWidth, cfg.ResultAreaHeight), 6)

	// 6. Draw tooltip area (LIGHT BLUE)
	if cfg.TooltipRect.Min.X != 0 && cfg.TooltipRect.Min.Y != 0 {
		drawLabeledRect(
			cfg.TooltipRect.Min.X, cfg.TooltipRect.Min.Y,
			cfg.TooltipRect.Max.X, cfg.TooltipRect.Max.Y,
			color.RGBA{100, 200, 255, 255}, fmt.Sprintf("TOOLTIP AREA (%dx%d)",
				cfg.TooltipRect.Dx(), cfg.TooltipRect.Dy()), 4)
	}

	// 7. Highlight current item to be moved (YELLOW)
	if itemX != 0 && itemY != 0 {
		itemX1 := itemX - cellWidth/2
		itemY1 := itemY - cellHeight/2
		itemX2 := itemX1 + (cfg.ItemWidth * cellWidth)
		itemY2 := itemY1 + (cfg.ItemHeight * cellHeight)
		drawLabeledRect(itemX1, itemY1, itemX2, itemY2,
			color.RGBA{255, 255, 0, 255}, ">> ITEM TO MOVE <<", 8)
	}

	// 8. Highlight target result slot (GREEN)
	if resultX != 0 && resultY != 0 {
		resultSlotX1 := resultX - cellWidth/2
		resultSlotY1 := resultY - cellHeight/2
		resultSlotX2 := resultSlotX1 + (cfg.ItemWidth * cellWidth)
		resultSlotY2 := resultSlotY1 + (cfg.ItemHeight * cellHeight)
		drawLabeledRect(resultSlotX1, resultSlotY1, resultSlotX2, resultSlotY2,
			color.RGBA{0, 255, 0, 255}, ">> TARGET SLOT <<", 8)
	}

	// Draw title and instructions
	titleText := fmt.Sprintf("=== ROUND #%d - %s ===", itemNum, stepName)
	drawString(rgba, 20, 30, titleText, color.RGBA{255, 255, 255, 255})
	drawString(rgba, 20, 50, "Flow: PENDING (cyan) -> WORKBENCH (orange) -> RESULT (magenta) | TOOLTIP (light blue)", color.RGBA{200, 200, 200, 255})

	// Save snapshot
	debugFile := filepath.Join(config.SnapshotsDir, fmt.Sprintf("round%d_%s.png", itemNum, stepName))
	if err := SaveImage(rgba, debugFile); err != nil {
		return fmt.Errorf("failed to save full screen debug snapshot: %w", err)
	}

	fmt.Printf("✓ Full screen debug snapshot: %s\n", debugFile)

	// Emit snapshot event for web GUI
	// Note: e is not available here as this is a standalone function
	// The caller should emit the event
	return nil
}
