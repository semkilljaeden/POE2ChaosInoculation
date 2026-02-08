package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"path/filepath"

	"github.com/go-vgo/robotgo"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// drawBackpackGrid creates a debug image with the backpack grid overlay
func drawBackpackGrid(cfg Config) error {
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
			// Calculate label position (top-left of cell with small offset)
			labelX := col*cellWidth + 5
			labelY := row*cellHeight + 5

			// Draw label text: "R,C" format
			label := fmt.Sprintf("%d,%d", row, col)
			drawString(rgba, labelX, labelY, label, labelColor)
		}
	}

	// Save debug snapshot
	debugFile := filepath.Join(snapshotsDir, "backpack_grid_debug.png")
	if err := saveImage(rgba, debugFile); err != nil {
		return fmt.Errorf("failed to save backpack grid debug image: %w", err)
	}

	fmt.Printf("✓ Backpack grid debug saved: %s\n", debugFile)
	return nil
}

// drawBatchWorkflowSnapshot creates a debug image showing the batch crafting workflow
func drawBatchWorkflowSnapshot(cfg Config, pendingItemX, pendingItemY, resultX, resultY int) error {
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

	// Draw grid lines (green)
	gridColor := color.RGBA{0, 255, 0, 255}
	for i := 0; i <= 12; i++ {
		x := i * cellWidth
		for y := 0; y < height; y++ {
			if x < width {
				rgba.Set(x, y, gridColor)
			}
		}
	}
	for i := 0; i <= 5; i++ {
		y := i * cellHeight
		for x := 0; x < width; x++ {
			if y < height {
				rgba.Set(x, y, gridColor)
			}
		}
	}

	// Helper to draw a thick rectangle border with semi-transparent fill
	drawThickRect := func(x, y, w, h int, borderCol color.RGBA, thickness int, fillAlpha uint8) {
		// Draw semi-transparent fill
		fillCol := color.RGBA{borderCol.R, borderCol.G, borderCol.B, fillAlpha}
		for py := y; py < y+h; py++ {
			for px := x; px < x+w; px++ {
				if px >= 0 && px < width && py >= 0 && py < height {
					// Blend with existing pixel
					existing := rgba.At(px, py)
					er, eg, eb, _ := existing.RGBA()
					fr, fg, fb, fa := fillCol.RGBA()

					// Simple alpha blending
					alpha := float64(fa) / 65535.0
					nr := uint8((float64(er>>8)*(1-alpha) + float64(fr>>8)*alpha))
					ng := uint8((float64(eg>>8)*(1-alpha) + float64(fg>>8)*alpha))
					nb := uint8((float64(eb>>8)*(1-alpha) + float64(fb>>8)*alpha))

					rgba.Set(px, py, color.RGBA{nr, ng, nb, 255})
				}
			}
		}

		// Draw thick border
		for t := 0; t < thickness; t++ {
			// Top and bottom
			for i := x - t; i < x+w+t; i++ {
				if i >= 0 && i < width {
					if y-t >= 0 && y-t < height {
						rgba.Set(i, y-t, borderCol)
					}
					if y+h+t >= 0 && y+h+t < height {
						rgba.Set(i, y+h+t, borderCol)
					}
				}
			}
			// Left and right
			for j := y - t; j < y+h+t; j++ {
				if j >= 0 && j < height {
					if x-t >= 0 && x-t < width {
						rgba.Set(x-t, j, borderCol)
					}
					if x+w+t >= 0 && x+w+t < width {
						rgba.Set(x+w+t, j, borderCol)
					}
				}
			}
		}
	}

	// Convert absolute coordinates to relative coordinates within the screenshot
	toRelativeX := func(absX int) int { return absX - cfg.BackpackTopLeft.X }
	toRelativeY := func(absY int) int { return absY - cfg.BackpackTopLeft.Y }

	itemWidth := cfg.ItemWidth * cellWidth
	itemHeight := cfg.ItemHeight * cellHeight

	// STEP 1: Highlight pending item (cyan/light blue - source)
	if pendingItemX != 0 && pendingItemY != 0 {
		relX := toRelativeX(pendingItemX) - cellWidth/2
		relY := toRelativeY(pendingItemY) - cellHeight/2
		drawThickRect(relX, relY, itemWidth, itemHeight, color.RGBA{0, 255, 255, 255}, 8, 60)
		// Draw large label
		drawString(rgba, relX+10, relY+20, "STEP 1: PENDING", color.RGBA{0, 255, 255, 255})
		drawString(rgba, relX+10, relY+35, fmt.Sprintf("Pos: %d,%d", pendingItemX, pendingItemY), color.RGBA{255, 255, 255, 255})
	}

	// STEP 2: Highlight workbench (orange - crafting location)
	relWbX := toRelativeX(cfg.WorkbenchTopLeft.X) - cellWidth/2
	relWbY := toRelativeY(cfg.WorkbenchTopLeft.Y) - cellHeight/2
	drawThickRect(relWbX, relWbY, itemWidth, itemHeight, color.RGBA{255, 165, 0, 255}, 8, 60)
	// Draw large label
	drawString(rgba, relWbX+10, relWbY+20, "STEP 2: WORKBENCH", color.RGBA{255, 165, 0, 255})
	drawString(rgba, relWbX+10, relWbY+35, fmt.Sprintf("Pos: %d,%d", cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y), color.RGBA{255, 255, 255, 255})

	// STEP 3: Highlight result area slot (magenta - destination)
	if resultX != 0 && resultY != 0 {
		relResX := toRelativeX(resultX) - cellWidth/2
		relResY := toRelativeY(resultY) - cellHeight/2
		drawThickRect(relResX, relResY, itemWidth, itemHeight, color.RGBA{255, 0, 255, 255}, 8, 60)
		// Draw large label
		drawString(rgba, relResX+10, relResY+20, "STEP 3: RESULT", color.RGBA{255, 0, 255, 255})
		drawString(rgba, relResX+10, relResY+35, fmt.Sprintf("Pos: %d,%d", resultX, resultY), color.RGBA{255, 255, 255, 255})
	}

	// Draw cell labels (yellow)
	labelColor := color.RGBA{255, 255, 0, 255}
	for row := 0; row < 5; row++ {
		for col := 0; col < 12; col++ {
			labelX := col*cellWidth + 5
			labelY := row*cellHeight + 5
			label := fmt.Sprintf("%d,%d", row, col)
			drawString(rgba, labelX, labelY, label, labelColor)
		}
	}

	// Draw title and legend
	drawString(rgba, 10, 20, "=== BATCH CRAFTING WORKFLOW ===", color.RGBA{255, 255, 255, 255})
	drawString(rgba, 10, height-75, "Legend:", color.RGBA{255, 255, 255, 255})
	drawString(rgba, 10, height-60, "CYAN = Item to Craft (Pending)", color.RGBA{0, 255, 255, 255})
	drawString(rgba, 10, height-45, "ORANGE = Crafting Station (Workbench)", color.RGBA{255, 165, 0, 255})
	drawString(rgba, 10, height-30, "MAGENTA = Destination (Result Area)", color.RGBA{255, 0, 255, 255})

	// Save debug snapshot
	debugFile := filepath.Join(snapshotsDir, "batch_workflow.png")
	if err := saveImage(rgba, debugFile); err != nil {
		return fmt.Errorf("failed to save batch workflow snapshot: %w", err)
	}

	fmt.Printf("✓ Batch workflow snapshot: %s\n", debugFile)
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

// drawFullScreenDebugSnapshot captures the entire screen and labels all important areas
func drawFullScreenDebugSnapshot(cfg Config, itemNum int, stepName string, itemX, itemY, resultX, resultY int) error {
	// Get screen dimensions
	screenWidth, screenHeight := robotgo.GetScreenSize()
	fmt.Printf("     Screen size: %dx%d\n", screenWidth, screenHeight)

	// Capture entire screen (try with no position to capture all displays)
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
		// Draw rectangle border
		for t := 0; t < thickness; t++ {
			// Top and bottom
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
			// Left and right
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
		// Draw label above the rectangle
		drawString(rgba, x1+5, y1-15, label, col)
	}

	// Helper to draw a circle (for chaos orb position)
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

	// 7. Highlight current item to be moved (YELLOW with pulsing effect)
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

	// Draw title and instructions with round info
	titleText := fmt.Sprintf("=== ROUND #%d - %s ===", itemNum, stepName)
	drawString(rgba, 20, 30, titleText, color.RGBA{255, 255, 255, 255})
	drawString(rgba, 20, 50, "Flow: PENDING (cyan) -> WORKBENCH (orange) -> RESULT (magenta) | TOOLTIP (light blue)", color.RGBA{200, 200, 200, 255})

	// Save snapshot with round number and step name
	debugFile := filepath.Join(snapshotsDir, fmt.Sprintf("round%d_%s.png", itemNum, stepName))
	if err := saveImage(rgba, debugFile); err != nil {
		return fmt.Errorf("failed to save full screen debug snapshot: %w", err)
	}

	fmt.Printf("✓ Full screen debug snapshot: %s\n", debugFile)
	return nil
}
