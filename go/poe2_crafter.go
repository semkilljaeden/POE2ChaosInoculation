package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const snapshotsDir = "snapshots"

// ModRequirement defines what mod to look for
type ModRequirement struct {
	Pattern     string // Regex pattern for the mod name
	MinValue    int    // Minimum acceptable value (legacy, 0 = tier mode)
	TierLevel   string // Tier to match (e.g., "T1", "T2"), empty = value mode
	Description string // What this is
}

// Config for the crafter
type Config struct {
	ChaosPos            image.Point
	ItemPos             image.Point      // Legacy: single item position (for backward compatibility)
	ItemWidth           int              // Item width in cells (e.g., 1 for 1x1, 2 for 2x3)
	ItemHeight          int              // Item height in cells (e.g., 1 for 1x1, 3 for 2x3)
	TooltipOffset       image.Point      // Offset from ItemPos to tooltip top-left
	TooltipSize         image.Point      // Width and height of tooltip
	TooltipRect         image.Rectangle  `json:"-"` // Runtime only, calculated from ItemPos + Offset
	BackpackTopLeft     image.Point      // Top-left corner of backpack grid
	BackpackBottomRight image.Point      // Bottom-right corner of backpack grid

	// Batch crafting areas
	WorkbenchTopLeft    image.Point      // Top-left of workbench (exact match to item dimensions)
	PendingAreaTopLeft  image.Point      // Top-left of pending items area
	PendingAreaWidth    int              // Width of pending area in cells
	PendingAreaHeight   int              // Height of pending area in cells
	ResultAreaTopLeft   image.Point      // Top-left of result items area
	ResultAreaWidth     int              // Width of result area in cells
	ResultAreaHeight    int              // Height of result area in cells
	UseBatchMode        bool             // Enable batch crafting workflow

	TargetMods          []ModRequirement // Support multiple target mods
	MaxAttempts         int
	Delay               time.Duration
	Debug               bool
	SaveAllSnapshots    bool // Save every attempt's screenshot
}

// ModStat tracks statistics for a specific mod
type ModStat struct {
	ModName    string
	Count      int
	MinValue   int
	MaxValue   int
	AvgValue   float64
	TotalValue int
}

// CraftingSession tracks all data during a crafting session
type CraftingSession struct {
	StartTime       time.Time
	EndTime         time.Time
	TotalRolls      int
	ModStats        map[string]*ModStat // Key: mod name
	TargetModHit    bool
	TargetModName   string // Which target mod was found
	TargetValue     int
}

// Global control flags with atomic access for thread safety
var (
	stopRequested       atomic.Bool
	pauseRequested      atomic.Bool
	pauseToggleCooldown atomic.Value // stores time.Time
	lastPauseKeyState   atomic.Bool
	snapshotCounter     atomic.Int32 // Sequential counter for snapshot naming
)

// Random delay to simulate human behavior (base ± variation in ms)
func humanDelay(baseMs int, variationMs int) {
	delay := baseMs + rand.Intn(variationMs*2) - variationMs
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

func main() {
	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║      POE2 Chaos Crafter - Multi Mod          ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")
	fmt.Println()

	// Setup (config is saved inside setupWizard if newly created)
	config := setupWizard()

	fmt.Println("\n✓ Looking for ANY of these mods:")
	for i, mod := range config.TargetMods {
		fmt.Printf("   %d. %s\n", i+1, mod.Description)
	}
	fmt.Println("\nStarting in 5 seconds... Switch to POE2 now!")
	time.Sleep(5 * time.Second)

	// Run the crafter
	craft(config)
}

func captureWithCountdown(prompt string) (int, int) {
	fmt.Printf("\n%s", prompt)
	fmt.Print("\nPress any key, then you have 5 seconds to position mouse... ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	for i := 5; i > 0; i-- {
		fmt.Printf("\r%d... ", i)
		time.Sleep(1 * time.Second)
	}

	x, y := robotgo.GetMousePos()
	fmt.Printf("\r✓ Captured at (%d, %d)   \n", x, y)
	return x, y
}

// validateTooltipArea tests if the tooltip area can be read by OCR
func validateTooltipArea(cfg *Config, scanner *bufio.Scanner) bool {
	x1 := cfg.TooltipRect.Min.X
	y1 := cfg.TooltipRect.Min.Y
	x2 := cfg.TooltipRect.Max.X
	y2 := cfg.TooltipRect.Max.Y

	fmt.Printf("\nTooltip region: %dx%d pixels\n", x2-x1, y2-y1)

	// Capture tooltip
	fmt.Println("\n📸 Capturing and testing tooltip area...")
	time.Sleep(500 * time.Millisecond) // Brief pause to ensure tooltip is visible
	tooltipBitmap := robotgo.CaptureScreen(x1, y1, x2-x1, y2-y1)
	tooltipImg := robotgo.ToImage(tooltipBitmap)

	// Save snapshot
	tooltipSnapshotFile := filepath.Join(snapshotsDir, "tooltip_area_validation.png")
	saveImage(tooltipImg, tooltipSnapshotFile)
	fmt.Printf("\n✓ Snapshot saved: %s\n", tooltipSnapshotFile)

	// Run OCR test
	fmt.Println("\n🔍 Running OCR test...")
	tempDir := filepath.Join(os.TempDir(), "poe2_crafter_setup")
	os.MkdirAll(tempDir, 0755)

	ocrText, err := runTesseractOCR(tooltipImg, tempDir)
	if err != nil {
		fmt.Printf("\n❌ OCR Error: %v\n", err)
		return false
	}

	// Display OCR results
	fmt.Println("\n📝 OCR Results:")
	fmt.Println("----------------------------------------")
	fmt.Println(ocrText)
	fmt.Println("----------------------------------------")

	// Check if we got reasonable text (at least some content)
	textLines := strings.Split(strings.TrimSpace(ocrText), "\n")
	validLines := 0
	for _, line := range textLines {
		if len(strings.TrimSpace(line)) > 3 {
			validLines++
		}
	}

	if validLines > 0 {
		fmt.Printf("\n✅ SUCCESS - OCR detected %d line(s) of text\n", validLines)
		return true
	}

	fmt.Println("\n⚠️  WARNING: No text detected in tooltip area!")
	return false
}

func setupWizard() Config {
	scanner := bufio.NewScanner(os.Stdin)
	config := Config{
		ItemWidth:        1, // Default to 1x1 item
		ItemHeight:       1,
		MaxAttempts:      1000,
		Delay:            75 * time.Millisecond, // Very fast default delay
		Debug:            false,
		SaveAllSnapshots: false,
	}

	// Check for previous config
	if prevConfig, err := loadConfig(); err == nil {
		fmt.Printf("📁 Found previous configuration\n")
		fmt.Printf("  - Chaos: (%d, %d)\n", prevConfig.ChaosPos.X, prevConfig.ChaosPos.Y)
		fmt.Printf("  - Grid: (%d, %d) to (%d, %d)\n",
			prevConfig.BackpackTopLeft.X, prevConfig.BackpackTopLeft.Y,
			prevConfig.BackpackBottomRight.X, prevConfig.BackpackBottomRight.Y)
		if prevConfig.ItemWidth > 0 && prevConfig.ItemHeight > 0 {
			fmt.Printf("  - Item size: %dx%d cells\n", prevConfig.ItemWidth, prevConfig.ItemHeight)
		}
		fmt.Printf("  - Mods: ")
		if len(prevConfig.TargetMods) > 0 {
			fmt.Printf("%s", prevConfig.TargetMods[0].Description)
			for i := 1; i < len(prevConfig.TargetMods); i++ {
				fmt.Printf(", %s", prevConfig.TargetMods[i].Description)
			}
			fmt.Println()
		} else {
			fmt.Println("(none)")
		}

		// Quick start option
		fmt.Print("\nAny modifications needed? (y/n, default n): ")
		scanner.Scan()
		needsMods := strings.ToLower(strings.TrimSpace(scanner.Text())) == "y"

		// Copy all settings from previous config
		config.ChaosPos = prevConfig.ChaosPos
		config.ItemPos = prevConfig.ItemPos
		config.ItemWidth = prevConfig.ItemWidth
		config.ItemHeight = prevConfig.ItemHeight
		config.TooltipOffset = prevConfig.TooltipOffset
		config.TooltipSize = prevConfig.TooltipSize
		config.BackpackTopLeft = prevConfig.BackpackTopLeft
		config.BackpackBottomRight = prevConfig.BackpackBottomRight
		config.WorkbenchTopLeft = prevConfig.WorkbenchTopLeft
		config.PendingAreaTopLeft = prevConfig.PendingAreaTopLeft
		config.PendingAreaWidth = prevConfig.PendingAreaWidth
		config.PendingAreaHeight = prevConfig.PendingAreaHeight
		config.ResultAreaTopLeft = prevConfig.ResultAreaTopLeft
		config.ResultAreaWidth = prevConfig.ResultAreaWidth
		config.ResultAreaHeight = prevConfig.ResultAreaHeight
		config.UseBatchMode = prevConfig.UseBatchMode
		config.TargetMods = prevConfig.TargetMods
		config.MaxAttempts = prevConfig.MaxAttempts
		config.Delay = prevConfig.Delay
		config.Debug = prevConfig.Debug
		config.SaveAllSnapshots = prevConfig.SaveAllSnapshots

		// Ensure item dimensions have valid defaults (backward compatibility)
		if config.ItemWidth == 0 {
			config.ItemWidth = 1
		}
		if config.ItemHeight == 0 {
			config.ItemHeight = 1
		}

		if !needsMods {
			// Quick start - no modifications
			fmt.Println("\n✓ Using existing configuration")
		} else {
			// Selective modifications
			fmt.Println("\n=== SELECTIVE SETUP ===")

			// 1. Chaos orb position
			fmt.Print("\nUpdate chaos orb position? (y/n): ")
			scanner.Scan()
			if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
				config.ChaosPos.X, config.ChaosPos.Y = captureWithCountdown(
					"Position for CHAOS ORB in stash")
				fmt.Printf("✓ Chaos position: (%d, %d)\n", config.ChaosPos.X, config.ChaosPos.Y)
			}

			// 2. Backpack grid
			fmt.Print("\nUpdate backpack grid? (y/n): ")
			scanner.Scan()
			if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
				config.BackpackTopLeft.X, config.BackpackTopLeft.Y = captureWithCountdown(
					"BACKPACK TOP-LEFT corner")
				config.BackpackBottomRight.X, config.BackpackBottomRight.Y = captureWithCountdown(
					"BACKPACK BOTTOM-RIGHT corner")
				fmt.Printf("✓ Grid: (%d, %d) to (%d, %d)\n",
					config.BackpackTopLeft.X, config.BackpackTopLeft.Y,
					config.BackpackBottomRight.X, config.BackpackBottomRight.Y)

				// Generate grid snapshot immediately
				fmt.Println("\n📸 Generating grid snapshot...")
				time.Sleep(300 * time.Millisecond)
				if err := drawBackpackGrid(config); err != nil {
					fmt.Printf("⚠ Warning: Could not create grid snapshot: %v\n", err)
				} else {
					fmt.Println("✓ Grid snapshot: backpack_grid_debug.png")
				}
			}

			// 3. Target mods
			fmt.Print("\nUpdate target mods? (y/n): ")
			scanner.Scan()
			if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
				config.TargetMods = []ModRequirement{}
				fmt.Println("\nEnter mods (format: <mod> <value>, empty to finish):")
				modNum := 1
				for {
					fmt.Printf("Mod #%d: ", modNum)
					scanner.Scan()
					input := strings.TrimSpace(scanner.Text())
					if input == "" {
						if len(config.TargetMods) == 0 {
							fmt.Println("❌ Need at least one mod")
							continue
						}
						break
					}
					mod := parseModInput(input)
					if mod.Pattern != "" {
						config.TargetMods = append(config.TargetMods, mod)
						fmt.Printf("✓ Added: %s\n", mod.Description)
						modNum++
					} else {
						fmt.Println("❌ Invalid format. Try 'life 80'")
					}
				}
			}

			// 4. Logging and snapshots
			fmt.Print("\nUpdate logging/snapshot options? (y/n): ")
			scanner.Scan()
			if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
				fmt.Print("Enable OCR text logging? (y/n): ")
				scanner.Scan()
				config.Debug = strings.ToLower(strings.TrimSpace(scanner.Text())) == "y"

				fmt.Print("Save all snapshots? (y/n): ")
				scanner.Scan()
				config.SaveAllSnapshots = strings.ToLower(strings.TrimSpace(scanner.Text())) == "y"
			}

			// 5. Item position and dimensions
			fmt.Print("\nUpdate item position/dimensions? (y/n): ")
			scanner.Scan()
			if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
				fmt.Println("\n📍 Specify item position and dimensions:")
				fmt.Println("Note: Specify the TOP-LEFT corner cell of the item")

				var row, col int
				for {
					fmt.Print("Row (0-4): ")
					scanner.Scan()
					if r, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && r >= 0 && r < 5 {
						row = r
						break
					}
					fmt.Println("❌ Invalid. Must be 0-4.")
				}
				for {
					fmt.Print("Col (0-11): ")
					scanner.Scan()
					if c, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && c >= 0 && c < 12 {
						col = c
						break
					}
					fmt.Println("❌ Invalid. Must be 0-11.")
				}

				// Ask for item dimensions
				for {
					fmt.Print("Item width in cells (1-12, default 1): ")
					scanner.Scan()
					input := strings.TrimSpace(scanner.Text())
					if input == "" {
						config.ItemWidth = 1
						break
					}
					if w, err := strconv.Atoi(input); err == nil && w >= 1 && w <= 12 {
						config.ItemWidth = w
						break
					}
					fmt.Println("❌ Invalid. Must be 1-12.")
				}
				for {
					fmt.Print("Item height in cells (1-5, default 1): ")
					scanner.Scan()
					input := strings.TrimSpace(scanner.Text())
					if input == "" {
						config.ItemHeight = 1
						break
					}
					if h, err := strconv.Atoi(input); err == nil && h >= 1 && h <= 5 {
						config.ItemHeight = h
						break
					}
					fmt.Println("❌ Invalid. Must be 1-5.")
				}

				// Calculate positions
				config.ItemPos.X, config.ItemPos.Y = getCellCenter(config, row, col)
				fmt.Printf("✓ Item at cell (%d,%d) [%dx%d cells] → pixel (%d, %d)\n",
					row, col, config.ItemWidth, config.ItemHeight, config.ItemPos.X, config.ItemPos.Y)
			}

			// 6. Batch mode configuration
			fmt.Print("\nConfigure batch crafting mode? (y/n): ")
			scanner.Scan()
			if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
				config.UseBatchMode = true
				fmt.Println("\n🔄 Batch Mode Configuration")
				fmt.Println("You'll specify:")
				fmt.Println("  - Workbench: where items are crafted")
				fmt.Println("  - Pending area: holds items waiting to be crafted")
				fmt.Println("  - Result area: holds finished items")

				// Workbench
				fmt.Println("Workbench (exact match to item dimensions):")
				var wbRow, wbCol int
				for {
					fmt.Print("  Workbench row (0-4): ")
					scanner.Scan()
					if r, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && r >= 0 && r < 5 {
						wbRow = r
						break
					}
					fmt.Println("  ❌ Invalid. Must be 0-4.")
				}
				for {
					fmt.Print("  Workbench col (0-11): ")
					scanner.Scan()
					if c, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && c >= 0 && c < 12 {
						wbCol = c
						break
					}
					fmt.Println("  ❌ Invalid. Must be 0-11.")
				}
				config.WorkbenchTopLeft.X, config.WorkbenchTopLeft.Y = getCellCenter(config, wbRow, wbCol)
				fmt.Printf("✓ Workbench at cell (%d,%d)\n", wbRow, wbCol)

				// Pending area
				fmt.Println("\nPending area (items waiting to be crafted):")
				var pRow, pCol int
				for {
					fmt.Print("  Top-left row (0-4): ")
					scanner.Scan()
					if r, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && r >= 0 && r < 5 {
						pRow = r
						break
					}
					fmt.Println("  ❌ Invalid. Must be 0-4.")
				}
				for {
					fmt.Print("  Top-left col (0-11): ")
					scanner.Scan()
					if c, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && c >= 0 && c < 12 {
						pCol = c
						break
					}
					fmt.Println("  ❌ Invalid. Must be 0-11.")
				}
				config.PendingAreaTopLeft.X, config.PendingAreaTopLeft.Y = getCellCenter(config, pRow, pCol)
				for {
					fmt.Print("  Width in cells (1-12): ")
					scanner.Scan()
					if w, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && w >= 1 && w <= 12 {
						config.PendingAreaWidth = w
						break
					}
					fmt.Println("  ❌ Invalid. Must be 1-12.")
				}
				for {
					fmt.Print("  Height in cells (1-5): ")
					scanner.Scan()
					if h, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && h >= 1 && h <= 5 {
						config.PendingAreaHeight = h
						break
					}
					fmt.Println("  ❌ Invalid. Must be 1-5.")
				}
				fmt.Printf("✓ Pending area: (%d,%d) size %dx%d\n", pRow, pCol, config.PendingAreaWidth, config.PendingAreaHeight)

				// Result area
				fmt.Println("\nResult area (finished items):")
				var rRow, rCol int
				for {
					fmt.Print("  Top-left row (0-4): ")
					scanner.Scan()
					if r, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && r >= 0 && r < 5 {
						rRow = r
						break
					}
					fmt.Println("  ❌ Invalid. Must be 0-4.")
				}
				for {
					fmt.Print("  Top-left col (0-11): ")
					scanner.Scan()
					if c, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && c >= 0 && c < 12 {
						rCol = c
						break
					}
					fmt.Println("  ❌ Invalid. Must be 0-11.")
				}
				config.ResultAreaTopLeft.X, config.ResultAreaTopLeft.Y = getCellCenter(config, rRow, rCol)
				for {
					fmt.Print("  Width in cells (1-12): ")
					scanner.Scan()
					if w, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && w >= 1 && w <= 12 {
						config.ResultAreaWidth = w
						break
					}
					fmt.Println("  ❌ Invalid. Must be 1-12.")
				}
				for {
					fmt.Print("  Height in cells (1-5): ")
					scanner.Scan()
					if h, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && h >= 1 && h <= 5 {
						config.ResultAreaHeight = h
						break
					}
					fmt.Println("  ❌ Invalid. Must be 1-5.")
				}
				fmt.Printf("✓ Result area: (%d,%d) size %dx%d\n", rRow, rCol, config.ResultAreaWidth, config.ResultAreaHeight)
			} else {
				config.UseBatchMode = false
			}
		}

		// Calculate tooltip rect based on item position
		config.TooltipRect = image.Rectangle{
			Min: image.Point{
				X: config.ItemPos.X + config.TooltipOffset.X,
				Y: config.ItemPos.Y + config.TooltipOffset.Y,
			},
			Max: image.Point{
				X: config.ItemPos.X + config.TooltipOffset.X + config.TooltipSize.X,
				Y: config.ItemPos.Y + config.TooltipOffset.Y + config.TooltipSize.Y,
			},
		}

		// Save updated config
		if err := saveConfig(config); err != nil {
			fmt.Printf("⚠ Could not save: %v\n", err)
		} else {
			fmt.Println("✓ Config saved")
		}

		return config
	}

	// Only do full setup if we don't have partial config from failed validation
	needsFullSetup := config.ChaosPos.X == 0 && config.ChaosPos.Y == 0

	if needsFullSetup {
		fmt.Println("=== QUICK SETUP ===\n")

		// Step 1: Backpack Grid
		fmt.Println("Step 1: Backpack Grid Configuration")
		fmt.Println("------------------------------------")
		fmt.Println("The backpack is a 5x12 grid (5 rows, 12 columns = 60 cells)")
		fmt.Println("You'll specify the top-left and bottom-right corners,")
		fmt.Println("and then reference items by cell coordinates (row, col)\n")

		config.BackpackTopLeft.X, config.BackpackTopLeft.Y = captureWithCountdown(
			"Step 1a: Position for BACKPACK TOP-LEFT corner")

		config.BackpackBottomRight.X, config.BackpackBottomRight.Y = captureWithCountdown(
			"Step 1b: Position for BACKPACK BOTTOM-RIGHT corner")

		// Step 2: Other Positions
		fmt.Println("\n\nStep 2: Other Positions")
		fmt.Println("--------------------------")
		fmt.Println("(Tip: Keep POE2 in windowed mode for easier Alt-Tab)\n")

		config.ChaosPos.X, config.ChaosPos.Y = captureWithCountdown(
			"Step 2a: Position for CHAOS ORB in stash")

		// Get item position using cell coordinates
		fmt.Println("\n📍 Specify item position and dimensions:")
		fmt.Println("Note: Specify the TOP-LEFT corner cell of the item")

		var row, col int
		for {
			fmt.Print("Enter row (0-4): ")
			scanner.Scan()
			if r, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && r >= 0 && r < 5 {
				row = r
				break
			}
			fmt.Println("❌ Invalid row. Must be 0-4.")
		}

		for {
			fmt.Print("Enter column (0-11): ")
			scanner.Scan()
			if c, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && c >= 0 && c < 12 {
				col = c
				break
			}
			fmt.Println("❌ Invalid column. Must be 0-11.")
		}

		// Ask for item dimensions
		for {
			fmt.Print("Item width in cells (1-12, default 1): ")
			scanner.Scan()
			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				config.ItemWidth = 1
				break
			}
			if w, err := strconv.Atoi(input); err == nil && w >= 1 && w <= 12 {
				config.ItemWidth = w
				break
			}
			fmt.Println("❌ Invalid. Must be 1-12.")
		}
		for {
			fmt.Print("Item height in cells (1-5, default 1): ")
			scanner.Scan()
			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				config.ItemHeight = 1
				break
			}
			if h, err := strconv.Atoi(input); err == nil && h >= 1 && h <= 5 {
				config.ItemHeight = h
				break
			}
			fmt.Println("❌ Invalid. Must be 1-5.")
		}

		// Calculate item position from cell coordinates
		config.ItemPos.X, config.ItemPos.Y = getCellCenter(config, row, col)
		fmt.Printf("✓ Item at cell (%d,%d) [%dx%d cells] → pixel (%d, %d)\n",
			row, col, config.ItemWidth, config.ItemHeight, config.ItemPos.X, config.ItemPos.Y)
	}

	fmt.Println("\n\nStep 3: Tooltip Area")
	fmt.Println("--------------------")
	fmt.Println("⚠️  IMPORTANT: Before capturing corners, hover over an item to show the tooltip!")

	x1, y1 := captureWithCountdown(
		"Step 3a: TOP-LEFT corner of tooltip")

	x2, y2 := captureWithCountdown(
		"Step 3b: BOTTOM-RIGHT corner of tooltip")

	// Loop until valid tooltip area is captured
	for {
		config.TooltipRect = image.Rectangle{
			Min: image.Point{X: x1, Y: y1},
			Max: image.Point{X: x2, Y: y2},
		}

		// Calculate and save offset relative to item position
		config.TooltipOffset = image.Point{
			X: x1 - config.ItemPos.X,
			Y: y1 - config.ItemPos.Y,
		}
		config.TooltipSize = image.Point{
			X: x2 - x1,
			Y: y2 - y1,
		}

		fmt.Printf("\n✓ Tooltip region: %dx%d pixels\n", x2-x1, y2-y1)
		fmt.Printf("✓ Offset from item: (%d, %d)\n", config.TooltipOffset.X, config.TooltipOffset.Y)

		// Capture tooltip immediately
		fmt.Println("\n📸 Capturing and testing tooltip area...")
		time.Sleep(500 * time.Millisecond) // Brief pause to ensure tooltip is visible
		tooltipBitmap := robotgo.CaptureScreen(x1, y1, x2-x1, y2-y1)
		tooltipImg := robotgo.ToImage(tooltipBitmap)

		// Save snapshot
		tooltipSnapshotFile := filepath.Join(snapshotsDir, "tooltip_area_setup.png")
		saveImage(tooltipImg, tooltipSnapshotFile)
		fmt.Printf("\n✓ Snapshot saved: %s\n", tooltipSnapshotFile)

		// Run OCR test
		fmt.Println("\n🔍 Running OCR test...")
		tempDir := filepath.Join(os.TempDir(), "poe2_crafter_setup")
		os.MkdirAll(tempDir, 0755)

		ocrText, err := runTesseractOCR(tooltipImg, tempDir)
		if err != nil {
			fmt.Printf("\n❌ OCR Error: %v\n", err)
			fmt.Print("\nRetry tooltip selection? (y/n): ")
			scanner.Scan()
			if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
				x1, y1 = captureWithCountdown("Re-capture TOP-LEFT corner of tooltip")
				x2, y2 = captureWithCountdown("Re-capture BOTTOM-RIGHT corner of tooltip")
				continue
			}
			break
		}

		// Display OCR results
		fmt.Println("\n📝 OCR Results:")
		fmt.Println("----------------------------------------")
		fmt.Println(ocrText)
		fmt.Println("----------------------------------------")

		// Check if we got reasonable text
		textLines := strings.Split(strings.TrimSpace(ocrText), "\n")
		validLines := 0
		for _, line := range textLines {
			if len(strings.TrimSpace(line)) > 3 {
				validLines++
			}
		}

		if validLines > 0 {
			fmt.Printf("\n✅ SUCCESS - OCR detected %d line(s) of text\n", validLines)
			fmt.Println("\n✓ Tooltip area validated successfully!")
			break
		} else {
			fmt.Println("\n⚠️  WARNING: No text detected in tooltip area!")
			fmt.Println("\nPossible issues:")
			fmt.Println("   - Tooltip area too small/large")
			fmt.Println("   - Not hovering over item")
			fmt.Println("   - OCR quality issues")

			fmt.Print("\nRetry tooltip selection? (y/n): ")
			scanner.Scan()
			if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
				x1, y1 = captureWithCountdown("Re-capture TOP-LEFT corner of tooltip")
				x2, y2 = captureWithCountdown("Re-capture BOTTOM-RIGHT corner of tooltip")
				continue
			}

			fmt.Println("\n⚠  Proceeding with current selection (may cause issues during crafting)")
			break
		}
	}

	// Only configure mod and options if doing full setup
	if needsFullSetup {
		// Step 4: Multiple Mods
		fmt.Println("\n\nStep 4: What Mods Are You Looking For?")
		fmt.Println("---------------------------------------")
		fmt.Println("\nFormat: <mod> <min_value>")
		fmt.Println("\nQuick templates:")
		fmt.Println("  life 80         - Life")
		fmt.Println("  mana 60         - Mana")
		fmt.Println("  str 45          - Strength")
		fmt.Println("  dex 45          - Dexterity")
		fmt.Println("  int 45          - Intelligence")
		fmt.Println("  spirit 50       - Spirit")
		fmt.Println("  spell-level 3   - +3 to Level of all Spell Skills")
		fmt.Println("  proj-level 3    - +3 to Level of all Projectile Skills")
		fmt.Println("  crit-dmg 39     - 39% increased Critical Damage Bonus")
		fmt.Println("  fire-res 30     - Fire Resistance")
		fmt.Println("  cold-res 30     - Cold Resistance")
		fmt.Println("  light-res 30    - Lightning Resistance")
		fmt.Println("  chaos-res 20    - Chaos Resistance")
		fmt.Println("\nEnter mods one per line (empty line to finish):")
		fmt.Println()

		config.TargetMods = []ModRequirement{}
		modNum := 1
		for {
			fmt.Printf("Mod #%d (or press Enter if done): ", modNum)
			scanner.Scan()
			input := strings.TrimSpace(scanner.Text())

			if input == "" {
				if len(config.TargetMods) == 0 {
					fmt.Println("❌ Please enter at least one mod\n")
					continue
				}
				break
			}

			mod := parseModInput(input)
			if mod.Pattern != "" {
				config.TargetMods = append(config.TargetMods, mod)
				fmt.Printf("✓ Added: %s\n", mod.Description)
				modNum++
			} else {
				fmt.Println("❌ Invalid format. Try 'life 80' or 'fire-res 30'\n")
			}
		}

		fmt.Printf("\n✓ Total mods to search: %d\n", len(config.TargetMods))

		// Step 5: Options
		fmt.Println("\n\nStep 5: Options")
		fmt.Println("----------------")

		fmt.Print("\nMax attempts (default 1000): ")
		scanner.Scan()
		if attempts := scanner.Text(); attempts != "" {
			if n, err := strconv.Atoi(attempts); err == nil && n > 0 {
				config.MaxAttempts = n
			}
		}

		fmt.Print("Enable OCR text logging? (y/n, default n): ")
		scanner.Scan()
		config.Debug = strings.ToLower(scanner.Text()) == "y"

		fmt.Print("Save all snapshots for every attempt? (y/n, default n): ")
		scanner.Scan()
		config.SaveAllSnapshots = strings.ToLower(scanner.Text()) == "y"
	}

	// Save the new configuration
	fmt.Println("\n💾 Saving configuration...")
	if err := saveConfig(config); err != nil {
		fmt.Printf("⚠ Warning: Could not save config: %v\n", err)
		fmt.Println("   (You'll need to reconfigure next time)")
	} else {
		fmt.Printf("✓ Configuration saved to: %s\n", getConfigPath())
		fmt.Println("   (Will be auto-loaded next time)")
	}

	return config
}

// Config file path
func getConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".poe2_crafter_config.json")
}

// saveConfig saves the configuration to a JSON file
func saveConfig(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getConfigPath(), data, 0644)
}

// loadConfig loads the configuration from a JSON file
func loadConfig() (Config, error) {
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

func parseModInput(input string) ModRequirement {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return ModRequirement{}
	}

	modType := strings.ToLower(parts[0])

	// Parse minimum value
	value, err := strconv.Atoi(parts[1])
	if err != nil {
		return ModRequirement{}
	}

	templates := map[string]struct {
		pattern string
		desc    string
	}{
		// Pattern explanation: (?:\(\d+-\d+\))? = optional range display like (165-179)
		"life":           {`(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+MAXIMUM\s+LIFE`, "Life %d+"},
		"mana":           {`(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+MAXIMUM\s+MANA`, "Mana %d+"},
		"str":            {`(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+STRENGTH`, "Strength %d+"},
		"dex":            {`(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+DEXTERITY`, "Dexterity %d+"},
		"int":            {`(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+INTELLIGENCE`, "Intelligence %d+"},
		"spirit":         {`(?i)[+#]?(\d+)(?:\(\d+-\d+\))?\s+TO\s+SPIRIT`, "Spirit %d+"},
		"spell-level":    {`\+(\d+)\s+TO\s+LEVEL\s+OF\s+ALL\s+SPELL\s+SKILLS`, "+%d to Level of all Spell Skills (or higher)"},
		"proj-level":     {`\+(\d+)\s+TO\s+LEVEL\s+OF\s+ALL\s+PROJECTILE\s+SKILLS`, "+%d to Level of all Projectile Skills (or higher)"},
		"crit-dmg":       {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*INCREASED\s+CRITICAL\s+DAMAGE\s+BONUS`, "%d%%+ increased Critical Damage Bonus"},
		"fire-res":       {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?FIRE\s+RESISTANCE`, "Fire Res %d+%%"},
		"cold-res":       {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?COLD\s+RESISTANCE`, "Cold Res %d+%%"},
		"light-res":      {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?LIGHTNING\s+RESISTANCE`, "Lightning Res %d+%%"},
		"chaos-res":      {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?CHAOS\s+RESISTANCE`, "Chaos Res %d+%%"},
		"armor":          {`(?i)(\d+)(?:\(\d+-\d+\))?\s+(?:INCREASED\s+)?ARMOUR`, "Armour %d+"},
		"evasion":        {`(?i)(\d+)(?:\(\d+-\d+\))?\s+(?:INCREASED\s+)?EVASION`, "Evasion %d+"},
		"es":             {`(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+MAXIMUM\s+ENERGY\s+SHIELD`, "Energy Shield %d+"},
		"movespeed":      {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?MOVEMENT\s+SPEED`, "Movement Speed %d+%%"},
		"attackspeed":    {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?ATTACK\s+SPEED`, "Attack Speed %d+%%"},
		"castspeed":      {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?CAST\s+SPEED`, "Cast Speed %d+%%"},
	}

	if tmpl, exists := templates[modType]; exists {
		return ModRequirement{
			Pattern:     tmpl.pattern,
			MinValue:    value,
			TierLevel:   "",
			Description: fmt.Sprintf(tmpl.desc, value),
		}
	}

	// Custom regex
	if strings.Contains(input, "(\\d+)") || strings.Contains(input, `(\d+)`) {
		return ModRequirement{
			Pattern:     input,
			MinValue:    0,
			Description: "Custom: " + input[:min(len(input), 30)],
		}
	}

	return ModRequirement{}
}

// cleanupDebugSnapshots removes old snapshots folder and creates a fresh one
func cleanupDebugSnapshots() {
	// Remove entire snapshots directory if it exists
	if _, err := os.Stat(snapshotsDir); err == nil {
		if err := os.RemoveAll(snapshotsDir); err == nil {
			fmt.Printf("🧹 Cleaned up previous snapshots\n")
		}
	}

	// Create fresh snapshots directory
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		fmt.Printf("⚠ Warning: Could not create snapshots directory: %v\n", err)
	}
}

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

// hasItemAtPosition checks if there's an item at the given position by hovering and checking for a tooltip
func hasItemAtPosition(cfg Config, x, y int) bool {
	// Move mouse to position
	robotgo.Move(x, y)
	time.Sleep(50 * time.Millisecond)

	// Capture a small area to detect if tooltip appears
	// Simple heuristic: check if there's content in the tooltip area
	bitmap := robotgo.CaptureScreen(
		cfg.TooltipRect.Min.X, cfg.TooltipRect.Min.Y,
		min(cfg.TooltipRect.Dx(), 100), min(cfg.TooltipRect.Dy(), 50),
	)
	img := robotgo.ToImage(bitmap)

	// Check if there's visible content (not just black/empty)
	bounds := img.Bounds()
	pixelCount := 0
	nonBlackCount := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixelCount++
			r, g, b, _ := img.At(x, y).RGBA()
			// Check if pixel is not close to black
			if r > 5000 || g > 5000 || b > 5000 {
				nonBlackCount++
			}
		}
	}

	// If more than 10% of pixels are non-black, assume item exists
	return pixelCount > 0 && float64(nonBlackCount)/float64(pixelCount) > 0.1
}

// findNextItemInArea scans the area and returns the position of the first item found
// Returns (x, y, true) if found, (0, 0, false) if no item found
func findNextItemInArea(cfg Config, areaTopLeft image.Point, areaWidth, areaHeight int) (int, int, bool) {
	cellWidth := (cfg.BackpackBottomRight.X - cfg.BackpackTopLeft.X) / 12
	cellHeight := (cfg.BackpackBottomRight.Y - cfg.BackpackTopLeft.Y) / 5

	// Scan through each cell in the area
	for row := 0; row < areaHeight; row += cfg.ItemHeight {
		for col := 0; col < areaWidth; col += cfg.ItemWidth {
			// Calculate absolute position for this cell
			x := areaTopLeft.X + (col * cellWidth)
			y := areaTopLeft.Y + (row * cellHeight)

			// Check if there's an item at this position
			if hasItemAtPosition(cfg, x, y) {
				return x, y, true
			}
		}
	}

	return 0, 0, false
}

// findEmptySlotInArea finds the first empty slot in an area
// Returns (x, y, true) if found, (0, 0, false) if area is full
func findEmptySlotInArea(cfg Config, areaTopLeft image.Point, areaWidth, areaHeight int) (int, int, bool) {
	cellWidth := (cfg.BackpackBottomRight.X - cfg.BackpackTopLeft.X) / 12
	cellHeight := (cfg.BackpackBottomRight.Y - cfg.BackpackTopLeft.Y) / 5

	// Scan through each cell in the area
	for row := 0; row < areaHeight; row += cfg.ItemHeight {
		for col := 0; col < areaWidth; col += cfg.ItemWidth {
			// Calculate absolute position for this cell
			x := areaTopLeft.X + (col * cellWidth)
			y := areaTopLeft.Y + (row * cellHeight)

			// Check if this slot is empty
			if !hasItemAtPosition(cfg, x, y) {
				return x, y, true
			}
		}
	}

	return 0, 0, false
}

// moveItem moves an item from one position to another
func moveItem(fromX, fromY, toX, toY int) {
	// Pick up the item
	robotgo.Move(fromX, fromY)
	time.Sleep(50 * time.Millisecond)
	robotgo.Click()
	time.Sleep(100 * time.Millisecond)

	// Drop it at the destination
	robotgo.Move(toX, toY)
	time.Sleep(50 * time.Millisecond)
	robotgo.Click()
	time.Sleep(100 * time.Millisecond)
}

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

		itemCount := 0
		for {
			// Check if stop requested
			if stopRequested.Load() {
				fmt.Println("\n✓ Stopped by user")
				return
			}

			// Find next item in pending area
			itemX, itemY, found := findNextItemInArea(cfg, cfg.PendingAreaTopLeft, cfg.PendingAreaWidth, cfg.PendingAreaHeight)
			if !found {
				fmt.Println("\n✓ No more items in pending area")
				break
			}

			itemCount++
			fmt.Printf("\n📦 Processing item #%d from pending area...\n", itemCount)

			// Move item from pending to workbench
			fmt.Println("  → Moving to workbench...")
			moveItem(itemX, itemY, cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y)
			time.Sleep(200 * time.Millisecond)

			// Update ItemPos to workbench for crafting
			cfg.ItemPos = cfg.WorkbenchTopLeft

			// Craft this item (use existing single-item logic below)
			fmt.Println("  → Starting crafting...")
			craftSuccess := craftSingleItem(&cfg, session, tempDir)

			// Find empty slot in result area
			resultX, resultY, foundSlot := findEmptySlotInArea(cfg, cfg.ResultAreaTopLeft, cfg.ResultAreaWidth, cfg.ResultAreaHeight)
			if !foundSlot {
				fmt.Println("\n⚠ Warning: Result area is full! Leaving item on workbench.")
				fmt.Println("   Please clear result area and restart.")
				return
			}

			// Move item from workbench to result area
			fmt.Println("  → Moving to result area...")
			moveItem(cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y, resultX, resultY)
			time.Sleep(200 * time.Millisecond)

			if craftSuccess {
				fmt.Printf("  ✓ Item #%d completed!\n", itemCount)
			} else {
				fmt.Printf("  ✓ Item #%d processed (no target match)\n", itemCount)
			}
		}

		fmt.Printf("\n🎉 Batch crafting complete! Processed %d items.\n", itemCount)
		return
	}

	// Single-item mode (original behavior)
	fmt.Println("\n💡 Tips:")
	fmt.Println("   - Press Ctrl+C to STOP safely")
	fmt.Println("   - Press F12 to PAUSE/RESUME")
	fmt.Println("   - Using Shift-hold method for faster crafting!")

	craftSingleItem(&cfg, session, tempDir)
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

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
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

		fmt.Printf("\r[%d/%d] Crafting... ", attempt, cfg.MaxAttempts)

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
			fmt.Printf("\n⚠ OCR error #%d: %v\n", seqNum, err)
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

	fmt.Printf("\n\n❌ Reached max attempts (%d) without finding any target mod\n", cfg.MaxAttempts)
	return false
}

// Windows API for key state checking and sound
var (
	user32           = syscall.NewLazyDLL("user32.dll")
	procGetKeyState  = user32.NewProc("GetKeyState")
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procBeep         = kernel32.NewProc("Beep")
)

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
		{523, 150},  // C5
		{523, 150},  // C5
		{523, 150},  // C5
		{523, 400},  // C5 (longer)
		{415, 350},  // G#4
		{466, 350},  // A#4
		{523, 150},  // C5
		{466, 150},  // A#4
		{523, 600},  // C5 (final note, longest)
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
			} else {
				fmt.Print("\n[DEBUG] pauseRequested flag set to false")
				fmt.Print("\n▶  RESUMED")
			}
		}
	}
}

// checkMiddleMouseButton is now an alias for checkPauseToggle
func checkMiddleMouseButton() {
	checkPauseToggle()
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

func runTesseractOCR(img image.Image, tempDir string) (string, error) {
	seqNum := snapshotCounter.Add(1)

	// Save original and preprocessed snapshots
	debugOriginalFile := filepath.Join(snapshotsDir, fmt.Sprintf("snap_%d_raw.png", seqNum))
	debugProcessedFile := filepath.Join(snapshotsDir, fmt.Sprintf("snap_%d_processed.png", seqNum))
	saveImage(img, debugOriginalFile)
	saveImage(preprocessForOCR(img), debugProcessedFile)

	type ocrStrategy struct {
		name         string
		psm          int
		usePreprocess bool
	}

	// Fast path: Try most promising strategies first with early exit
	fastStrategies := []ocrStrategy{
		{"PSM6_raw", 6, false},           // Fast: raw image, no preprocessing
		{"PSM6_preprocessed", 6, true},   // Fallback: with preprocessing
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

// trackMods parses OCR text and tracks all mods found
func trackMods(text string, session *CraftingSession) {
	// Common mod patterns to track
	modPatterns := []struct {
		name    string
		pattern string
	}{
		{"Life", `(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+MAXIMUM\s+LIFE`},
		{"Mana", `(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+MAXIMUM\s+MANA`},
		{"Strength", `(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+STRENGTH`},
		{"Dexterity", `(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+DEXTERITY`},
		{"Intelligence", `(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+INTELLIGENCE`},
		{"Spirit", `(?i)[+#]?(\d+)(?:\(\d+-\d+\))?\s+TO\s+SPIRIT`},
		{"Spell Skills Level", `\+(\d+)\s+TO\s+LEVEL\s+OF\s+ALL\s+SPELL\s+SKILLS`},
		{"Projectile Skills Level", `\+(\d+)\s+TO\s+LEVEL\s+OF\s+ALL\s+PROJECTILE\s+SKILLS`},
		{"Critical Damage Bonus", `(?i)(\d+)(?:\(\d+-\d+\))?%?\s*INCREASED\s+CRITICAL\s+DAMAGE\s+BONUS`},
		{"Fire Resistance", `(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?FIRE\s+RESISTANCE`},
		{"Cold Resistance", `(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?COLD\s+RESISTANCE`},
		{"Lightning Resistance", `(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?LIGHTNING\s+RESISTANCE`},
		{"Chaos Resistance", `(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?CHAOS\s+RESISTANCE`},
		{"Armour", `(?i)(\d+)(?:\(\d+-\d+\))?\s+(?:INCREASED\s+)?ARMOUR`},
		{"Evasion", `(?i)(\d+)(?:\(\d+-\d+\))?\s+(?:INCREASED\s+)?EVASION`},
		{"Energy Shield", `(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+MAXIMUM\s+ENERGY\s+SHIELD`},
		{"Movement Speed", `(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?MOVEMENT\s+SPEED`},
		{"Attack Speed", `(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?ATTACK\s+SPEED`},
		{"Cast Speed", `(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?CAST\s+SPEED`},
	}

	for _, mod := range modPatterns {
		re := regexp.MustCompile(mod.pattern)
		matches := re.FindAllStringSubmatch(text, -1)

		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			value, err := strconv.Atoi(match[1])
			if err != nil {
				continue
			}

			// Update or create mod stat
			stat, exists := session.ModStats[mod.name]
			if !exists {
				stat = &ModStat{
					ModName:  mod.name,
					MinValue: value,
					MaxValue: value,
				}
				session.ModStats[mod.name] = stat
			}

			stat.Count++
			stat.TotalValue += value
			stat.AvgValue = float64(stat.TotalValue) / float64(stat.Count)

			if value < stat.MinValue {
				stat.MinValue = value
			}
			if value > stat.MaxValue {
				stat.MaxValue = value
			}
		}
	}
}

// generateReport creates a detailed report of the crafting session
func generateReport(session *CraftingSession, cfg Config) {
	duration := session.EndTime.Sub(session.StartTime)

	// Create report filename with timestamp
	reportFile := fmt.Sprintf("crafting_report_%s.txt", session.StartTime.Format("2006-01-02_15-04-05"))

	var report strings.Builder
	report.WriteString("╔═══════════════════════════════════════════════╗\n")
	report.WriteString("║       POE2 CHAOS CRAFTER - SESSION REPORT     ║\n")
	report.WriteString("╚═══════════════════════════════════════════════╝\n\n")

	// Session Summary
	report.WriteString("SESSION SUMMARY\n")
	report.WriteString("─────────────────────────────────────────────────\n")
	report.WriteString(fmt.Sprintf("Start Time:     %s\n", session.StartTime.Format("2006-01-02 15:04:05")))
	report.WriteString(fmt.Sprintf("End Time:       %s\n", session.EndTime.Format("2006-01-02 15:04:05")))
	report.WriteString(fmt.Sprintf("Duration:       %s\n", duration.Round(time.Second)))
	report.WriteString(fmt.Sprintf("Total Rolls:    %d\n", session.TotalRolls))
	if session.TotalRolls > 0 && duration.Seconds() > 0 {
		rollsPerMin := float64(session.TotalRolls) / duration.Minutes()
		report.WriteString(fmt.Sprintf("Speed:          %.1f rolls/min\n", rollsPerMin))
	}
	report.WriteString("Target Mods:    ")
	if len(cfg.TargetMods) > 0 {
		report.WriteString(cfg.TargetMods[0].Description)
		for i := 1; i < len(cfg.TargetMods); i++ {
			report.WriteString(fmt.Sprintf(", %s", cfg.TargetMods[i].Description))
		}
		report.WriteString("\n")
	} else {
		report.WriteString("(none)\n")
	}
	if session.TargetModHit {
		report.WriteString(fmt.Sprintf("Result:         ✓ SUCCESS - %s (Value: %d)\n", session.TargetModName, session.TargetValue))
	} else {
		report.WriteString("Result:         ✗ Not found\n")
	}
	report.WriteString("\n")

	// Mod Statistics
	if len(session.ModStats) > 0 {
		report.WriteString("MOD STATISTICS\n")
		report.WriteString("─────────────────────────────────────────────────\n")
		report.WriteString(fmt.Sprintf("%-20s %8s %8s %8s %8s\n", "Mod Name", "Count", "Min", "Max", "Avg"))
		report.WriteString("─────────────────────────────────────────────────\n")

		// Sort mods by count (most frequent first)
		type modEntry struct {
			name string
			stat *ModStat
		}
		var entries []modEntry
		for name, stat := range session.ModStats {
			entries = append(entries, modEntry{name, stat})
		}

		// Simple bubble sort by count
		for i := 0; i < len(entries); i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[j].stat.Count > entries[i].stat.Count {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}

		for _, entry := range entries {
			stat := entry.stat
			report.WriteString(fmt.Sprintf("%-20s %8d %8d %8d %8.1f\n",
				stat.ModName, stat.Count, stat.MinValue, stat.MaxValue, stat.AvgValue))
		}
		report.WriteString("\n")
	}

	// Probability Analysis
	if session.TotalRolls > 0 && len(session.ModStats) > 0 {
		report.WriteString("PROBABILITY ANALYSIS\n")
		report.WriteString("─────────────────────────────────────────────────\n")

		// Reuse entries from above (sorted by count)
		type modEntry struct {
			name string
			stat *ModStat
		}
		var probEntries []modEntry
		for name, stat := range session.ModStats {
			probEntries = append(probEntries, modEntry{name, stat})
		}

		// Sort by count
		for i := 0; i < len(probEntries); i++ {
			for j := i + 1; j < len(probEntries); j++ {
				if probEntries[j].stat.Count > probEntries[i].stat.Count {
					probEntries[i], probEntries[j] = probEntries[j], probEntries[i]
				}
			}
		}

		for _, entry := range probEntries {
			probability := float64(entry.stat.Count) / float64(session.TotalRolls) * 100
			report.WriteString(fmt.Sprintf("%-20s: %.2f%% (%d/%d rolls)\n",
				entry.stat.ModName, probability, entry.stat.Count, session.TotalRolls))
		}
		report.WriteString("\n")
	}

	reportText := report.String()

	// Save to file
	if err := os.WriteFile(reportFile, []byte(reportText), 0644); err != nil {
		fmt.Printf("\n⚠ Warning: Could not save report: %v\n", err)
	} else {
		fmt.Printf("\n📊 Report saved: %s\n", reportFile)
	}

	// Also print to console
	fmt.Println("\n" + reportText)
}

func saveImage(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
