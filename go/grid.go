package main

// getCellCenter calculates the pixel coordinates of the center of a backpack cell
// row: 0-4 (5 rows), col: 0-11 (12 columns)
func getCellCenter(cfg Config, row int, col int) (int, int) {
	totalWidth := cfg.BackpackBottomRight.X - cfg.BackpackTopLeft.X
	totalHeight := cfg.BackpackBottomRight.Y - cfg.BackpackTopLeft.Y

	cellWidth := totalWidth / 12
	cellHeight := totalHeight / 5

	// Calculate cell center
	centerX := cfg.BackpackTopLeft.X + (col * cellWidth) + (cellWidth / 2)
	centerY := cfg.BackpackTopLeft.Y + (row * cellHeight) + (cellHeight / 2)

	return centerX, centerY
}

// getGridCell converts pixel coordinates to cell coordinates (row, col)
func getGridCell(cfg Config, x, y int) (int, int) {
	totalWidth := cfg.BackpackBottomRight.X - cfg.BackpackTopLeft.X
	totalHeight := cfg.BackpackBottomRight.Y - cfg.BackpackTopLeft.Y

	cellWidth := totalWidth / 12
	cellHeight := totalHeight / 5

	// Calculate which cell the pixel is in
	col := (x - cfg.BackpackTopLeft.X) / cellWidth
	row := (y - cfg.BackpackTopLeft.Y) / cellHeight

	// Clamp to valid range
	if col < 0 {
		col = 0
	}
	if col > 11 {
		col = 11
	}
	if row < 0 {
		row = 0
	}
	if row > 4 {
		row = 4
	}

	return row, col
}
