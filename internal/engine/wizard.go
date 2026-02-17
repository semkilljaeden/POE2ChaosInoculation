package engine

import (
	"bufio"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"poe2-chaos-crafter/internal/config"

	"github.com/go-vgo/robotgo"
)

// CaptureWithCountdown captures a position with a countdown
func CaptureWithCountdown(prompt string) (int, int) {
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

// ValidateTooltipArea tests if the tooltip area can be read by OCR
func (e *Engine) ValidateTooltipArea(cfg *config.Config, scanner *bufio.Scanner) bool {
	x1 := cfg.TooltipRect.Min.X
	y1 := cfg.TooltipRect.Min.Y
	x2 := cfg.TooltipRect.Max.X
	y2 := cfg.TooltipRect.Max.Y

	fmt.Printf("\nTooltip region: %dx%d pixels\n", x2-x1, y2-y1)

	fmt.Println("\n📸 Capturing and testing tooltip area...")
	time.Sleep(500 * time.Millisecond)
	tooltipBitmap := robotgo.CaptureScreen(x1, y1, x2-x1, y2-y1)
	tooltipImg := robotgo.ToImage(tooltipBitmap)

	tooltipSnapshotFile := filepath.Join(config.SnapshotsDir, "tooltip_area_validation.png")
	SaveImage(tooltipImg, tooltipSnapshotFile)
	fmt.Printf("\n✓ Snapshot saved: %s\n", tooltipSnapshotFile)

	fmt.Println("\n🔍 Running OCR test...")
	tempDir := filepath.Join(os.TempDir(), "poe2_crafter_setup")
	os.MkdirAll(tempDir, 0755)

	ocrText, err := e.RunTesseractOCR(tooltipImg, tempDir, "")
	if err != nil {
		fmt.Printf("\n❌ OCR Error: %v\n", err)
		return false
	}

	fmt.Println("\n📝 OCR Results:")
	fmt.Println("----------------------------------------")
	fmt.Println(ocrText)
	fmt.Println("----------------------------------------")

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

// SetupWizardPreviousConfig handles loading and modifying previous configuration
func SetupWizardPreviousConfig(prevConfig config.Config, scanner *bufio.Scanner) (config.Config, bool) {
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
	fmt.Printf("  - Chaos per round: %d\n", prevConfig.ChaosPerRound)

	fmt.Println("  - Batch Mode: ENABLED")
	wbRow, wbCol := config.GetGridCell(prevConfig, prevConfig.WorkbenchTopLeft.X, prevConfig.WorkbenchTopLeft.Y)
	fmt.Printf("    • Workbench: cell (%d, %d)\n", wbRow, wbCol)
	pRow, pCol := config.GetGridCell(prevConfig, prevConfig.PendingAreaTopLeft.X, prevConfig.PendingAreaTopLeft.Y)
	fmt.Printf("    • Pending area: cell (%d, %d) [%dx%d cells]\n",
		pRow, pCol, prevConfig.PendingAreaWidth, prevConfig.PendingAreaHeight)
	rRow, rCol := config.GetGridCell(prevConfig, prevConfig.ResultAreaTopLeft.X, prevConfig.ResultAreaTopLeft.Y)
	fmt.Printf("    • Result area: cell (%d, %d) [%dx%d cells]\n",
		rRow, rCol, prevConfig.ResultAreaWidth, prevConfig.ResultAreaHeight)

	fmt.Print("\nAny modifications needed? (y/n, default n): ")
	scanner.Scan()
	needsMods := strings.ToLower(strings.TrimSpace(scanner.Text())) == "y"

	return prevConfig, needsMods
}

// SetupWizardSelectiveModifications handles selective modifications to existing config
func SetupWizardSelectiveModifications(cfg config.Config, scanner *bufio.Scanner) config.Config {
	fmt.Println("\n=== SELECTIVE SETUP ===")

	fmt.Print("\nUpdate chaos orb position? (y/n): ")
	scanner.Scan()
	if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
		cfg.ChaosPos.X, cfg.ChaosPos.Y = CaptureWithCountdown(
			"Position for CHAOS ORB in stash")
		fmt.Printf("✓ Chaos position: (%d, %d)\n", cfg.ChaosPos.X, cfg.ChaosPos.Y)
	}

	fmt.Print("\nUpdate backpack grid? (y/n): ")
	scanner.Scan()
	if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
		cfg.BackpackTopLeft.X, cfg.BackpackTopLeft.Y = CaptureWithCountdown(
			"BACKPACK TOP-LEFT corner")
		cfg.BackpackBottomRight.X, cfg.BackpackBottomRight.Y = CaptureWithCountdown(
			"BACKPACK BOTTOM-RIGHT corner")
		fmt.Printf("✓ Grid: (%d, %d) to (%d, %d)\n",
			cfg.BackpackTopLeft.X, cfg.BackpackTopLeft.Y,
			cfg.BackpackBottomRight.X, cfg.BackpackBottomRight.Y)

		fmt.Println("\n📸 Generating grid snapshot...")
		time.Sleep(300 * time.Millisecond)
		if err := DrawBackpackGrid(cfg); err != nil {
			fmt.Printf("⚠ Warning: Could not create grid snapshot: %v\n", err)
		} else {
			fmt.Println("✓ Grid snapshot: backpack_grid_debug.png")
		}
	}

	cfg = setupWizardUpdateTargetMods(cfg, scanner)
	cfg = setupWizardUpdateChaosPerRound(cfg, scanner)
	cfg = setupWizardUpdateLogging(cfg, scanner)
	cfg = setupWizardUpdateItemDimensions(cfg, scanner)
	cfg = setupWizardUpdateBatchAreas(cfg, scanner)
	cfg = setupWizardUpdateTooltip(cfg, scanner)

	if cfg.UseBatchMode {
		cfg.ItemPos = cfg.WorkbenchTopLeft
	}

	return cfg
}

func setupWizardUpdateTargetMods(cfg config.Config, scanner *bufio.Scanner) config.Config {
	fmt.Print("\nUpdate target mods? (y/n): ")
	scanner.Scan()
	if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
		cfg.TargetMods = []config.ModRequirement{}
		fmt.Println("\nEnter mods (format: <mod> <value>, empty to finish):")
		modNum := 1
		for {
			fmt.Printf("Mod #%d: ", modNum)
			scanner.Scan()
			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				if len(cfg.TargetMods) == 0 {
					fmt.Println("❌ Need at least one mod")
					continue
				}
				break
			}
			mod := config.ParseModInput(input, "")
			if mod.Pattern != "" {
				cfg.TargetMods = append(cfg.TargetMods, mod)
				fmt.Printf("✓ Added: %s\n", mod.Description)
				modNum++
			} else {
				fmt.Println("❌ Invalid format. Try 'life 80'")
			}
		}
	}
	return cfg
}

func setupWizardUpdateChaosPerRound(cfg config.Config, scanner *bufio.Scanner) config.Config {
	fmt.Print("\nUpdate chaos orbs per round? (y/n): ")
	scanner.Scan()
	if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
		for {
			fmt.Printf("Chaos orbs per round/item (current: %d, default 10): ", cfg.ChaosPerRound)
			scanner.Scan()
			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				break
			}
			if n, err := strconv.Atoi(input); err == nil && n > 0 {
				cfg.ChaosPerRound = n
				fmt.Printf("✓ Chaos per round: %d\n", cfg.ChaosPerRound)
				break
			}
			fmt.Println("❌ Invalid. Must be a positive number.")
		}
	}
	return cfg
}

func setupWizardUpdateLogging(cfg config.Config, scanner *bufio.Scanner) config.Config {
	fmt.Print("\nUpdate logging/snapshot options? (y/n): ")
	scanner.Scan()
	if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
		fmt.Print("Enable OCR text logging? (y/n): ")
		scanner.Scan()
		cfg.Debug = strings.ToLower(strings.TrimSpace(scanner.Text())) == "y"

		fmt.Print("Save all snapshots? (y/n): ")
		scanner.Scan()
		cfg.SaveAllSnapshots = strings.ToLower(strings.TrimSpace(scanner.Text())) == "y"
	}
	return cfg
}

func setupWizardUpdateItemDimensions(cfg config.Config, scanner *bufio.Scanner) config.Config {
	fmt.Print("\nUpdate item dimensions? (y/n): ")
	scanner.Scan()
	if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
		fmt.Println("\n📏 Item Dimensions:")
		for {
			fmt.Print("Item width in cells (1-12, default 1): ")
			scanner.Scan()
			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				cfg.ItemWidth = 1
				break
			}
			if w, err := strconv.Atoi(input); err == nil && w >= 1 && w <= 12 {
				cfg.ItemWidth = w
				break
			}
			fmt.Println("❌ Invalid. Must be 1-12.")
		}
		for {
			fmt.Print("Item height in cells (1-5, default 1): ")
			scanner.Scan()
			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				cfg.ItemHeight = 1
				break
			}
			if h, err := strconv.Atoi(input); err == nil && h >= 1 && h <= 5 {
				cfg.ItemHeight = h
				break
			}
			fmt.Println("❌ Invalid. Must be 1-5.")
		}
		fmt.Printf("✓ Item size: %dx%d cells\n", cfg.ItemWidth, cfg.ItemHeight)
	}
	return cfg
}

func setupWizardUpdateBatchAreas(cfg config.Config, scanner *bufio.Scanner) config.Config {
	fmt.Print("\nUpdate batch crafting areas? (y/n): ")
	scanner.Scan()
	if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
		fmt.Println("\n🔄 Batch Crafting Areas")
		fmt.Println("You'll specify:")
		fmt.Println("  - Workbench: where items are crafted")
		fmt.Println("  - Pending area: holds items waiting to be crafted")
		fmt.Println("  - Result area: holds finished items")

		cfg = setupWizardConfigureWorkbench(cfg, scanner)
		cfg = setupWizardConfigurePendingArea(cfg, scanner)
		cfg = setupWizardConfigureResultArea(cfg, scanner)
	}
	return cfg
}

func setupWizardConfigureWorkbench(cfg config.Config, scanner *bufio.Scanner) config.Config {
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
	cfg.WorkbenchTopLeft.X, cfg.WorkbenchTopLeft.Y = config.GetCellCenter(cfg, wbRow, wbCol)
	fmt.Printf("✓ Workbench at cell (%d,%d)\n", wbRow, wbCol)
	return cfg
}

func setupWizardConfigurePendingArea(cfg config.Config, scanner *bufio.Scanner) config.Config {
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
	cfg.PendingAreaTopLeft.X, cfg.PendingAreaTopLeft.Y = config.GetCellCenter(cfg, pRow, pCol)
	for {
		fmt.Print("  Width in cells (1-12): ")
		scanner.Scan()
		if w, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && w >= 1 && w <= 12 {
			cfg.PendingAreaWidth = w
			break
		}
		fmt.Println("  ❌ Invalid. Must be 1-12.")
	}
	for {
		fmt.Print("  Height in cells (1-5): ")
		scanner.Scan()
		if h, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && h >= 1 && h <= 5 {
			cfg.PendingAreaHeight = h
			break
		}
		fmt.Println("  ❌ Invalid. Must be 1-5.")
	}
	fmt.Printf("✓ Pending area: (%d,%d) size %dx%d\n", pRow, pCol, cfg.PendingAreaWidth, cfg.PendingAreaHeight)
	return cfg
}

func setupWizardConfigureResultArea(cfg config.Config, scanner *bufio.Scanner) config.Config {
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
	cfg.ResultAreaTopLeft.X, cfg.ResultAreaTopLeft.Y = config.GetCellCenter(cfg, rRow, rCol)
	for {
		fmt.Print("  Width in cells (1-12): ")
		scanner.Scan()
		if w, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && w >= 1 && w <= 12 {
			cfg.ResultAreaWidth = w
			break
		}
		fmt.Println("  ❌ Invalid. Must be 1-12.")
	}
	for {
		fmt.Print("  Height in cells (1-5): ")
		scanner.Scan()
		if h, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil && h >= 1 && h <= 5 {
			cfg.ResultAreaHeight = h
			break
		}
		fmt.Println("  ❌ Invalid. Must be 1-5.")
	}
	fmt.Printf("✓ Result area: (%d,%d) size %dx%d\n", rRow, rCol, cfg.ResultAreaWidth, cfg.ResultAreaHeight)
	return cfg
}

func setupWizardUpdateTooltip(cfg config.Config, scanner *bufio.Scanner) config.Config {
	fmt.Print("\nUpdate tooltip position? (y/n): ")
	scanner.Scan()
	if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
		fmt.Println("\n📝 Tooltip Configuration")
		fmt.Println("Position mouse over item and use countdown to capture tooltip area")
		fmt.Println("")

		x1, y1 := CaptureWithCountdown("Tooltip TOP-LEFT corner")
		x2, y2 := CaptureWithCountdown("Tooltip BOTTOM-RIGHT corner")

		cfg.TooltipRect = image.Rectangle{
			Min: image.Point{X: x1, Y: y1},
			Max: image.Point{X: x2, Y: y2},
		}

		refPos := cfg.WorkbenchTopLeft
		if refPos.X == 0 && refPos.Y == 0 {
			refPos = image.Point{
				X: (cfg.BackpackTopLeft.X + cfg.BackpackBottomRight.X) / 2,
				Y: (cfg.BackpackTopLeft.Y + cfg.BackpackBottomRight.Y) / 2,
			}
		}

		cfg.TooltipOffset = image.Point{
			X: x1 - refPos.X,
			Y: y1 - refPos.Y,
		}
		cfg.TooltipSize = image.Point{
			X: x2 - x1,
			Y: y2 - y1,
		}

		fmt.Printf("✓ Tooltip: (%d, %d) to (%d, %d) [%dx%d]\n",
			x1, y1, x2, y2, cfg.TooltipSize.X, cfg.TooltipSize.Y)
	}
	return cfg
}

// SetupWizardFullSetup performs full setup for first-time users
func SetupWizardFullSetup(cfg config.Config, scanner *bufio.Scanner) config.Config {
	fmt.Println("=== QUICK SETUP ===")
	fmt.Println()

	fmt.Println("Step 1: Backpack Grid Configuration")
	fmt.Println("------------------------------------")
	fmt.Println("The backpack is a 5x12 grid (5 rows, 12 columns = 60 cells)")
	fmt.Println("You'll specify the top-left and bottom-right corners,")
	fmt.Println("and then reference items by cell coordinates (row, col)")
	fmt.Println()

	cfg.BackpackTopLeft.X, cfg.BackpackTopLeft.Y = CaptureWithCountdown(
		"Step 1a: Position for BACKPACK TOP-LEFT corner")

	cfg.BackpackBottomRight.X, cfg.BackpackBottomRight.Y = CaptureWithCountdown(
		"Step 1b: Position for BACKPACK BOTTOM-RIGHT corner")

	fmt.Println("\n\nStep 2: Other Positions")
	fmt.Println("--------------------------")
	fmt.Println("(Tip: Keep POE2 in windowed mode for easier Alt-Tab)")
	fmt.Println()

	cfg.ChaosPos.X, cfg.ChaosPos.Y = CaptureWithCountdown(
		"Step 2a: Position for CHAOS ORB in stash")

	cfg = setupWizardConfigureItemDimensions(cfg, scanner)
	cfg = setupWizardConfigureBatchMode(cfg, scanner)

	return cfg
}

func setupWizardConfigureItemDimensions(cfg config.Config, scanner *bufio.Scanner) config.Config {
	fmt.Println("\n📏 Item Dimensions:")
	for {
		fmt.Print("Item width in cells (1-12, default 1): ")
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			cfg.ItemWidth = 1
			break
		}
		if w, err := strconv.Atoi(input); err == nil && w >= 1 && w <= 12 {
			cfg.ItemWidth = w
			break
		}
		fmt.Println("❌ Invalid. Must be 1-12.")
	}
	for {
		fmt.Print("Item height in cells (1-5, default 1): ")
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			cfg.ItemHeight = 1
			break
		}
		if h, err := strconv.Atoi(input); err == nil && h >= 1 && h <= 5 {
			cfg.ItemHeight = h
			break
		}
		fmt.Println("❌ Invalid. Must be 1-5.")
	}
	fmt.Printf("✓ Item size: %dx%d cells\n", cfg.ItemWidth, cfg.ItemHeight)
	return cfg
}

func setupWizardConfigureBatchMode(cfg config.Config, scanner *bufio.Scanner) config.Config {
	fmt.Println("\n🔄 Batch Mode Configuration (always enabled)")
	fmt.Println("You'll specify:")
	fmt.Println("  - Workbench: where items are crafted")
	fmt.Println("  - Pending area: holds items waiting to be crafted")
	fmt.Println("  - Result area: holds finished items")

	cfg = setupWizardConfigureWorkbench(cfg, scanner)
	cfg = setupWizardConfigurePendingArea(cfg, scanner)
	cfg = setupWizardConfigureResultArea(cfg, scanner)
	cfg.ItemPos = cfg.WorkbenchTopLeft

	return cfg
}

// SetupWizardConfigureTooltip configures tooltip area with validation
func (e *Engine) SetupWizardConfigureTooltip(cfg config.Config, scanner *bufio.Scanner) config.Config {
	fmt.Println("\n\nStep 3: Tooltip Area")
	fmt.Println("--------------------")
	fmt.Println("⚠️  IMPORTANT: Before capturing corners, hover over an item to show the tooltip!")

	x1, y1 := CaptureWithCountdown("Step 3a: TOP-LEFT corner of tooltip")
	x2, y2 := CaptureWithCountdown("Step 3b: BOTTOM-RIGHT corner of tooltip")

	for {
		cfg.TooltipRect = image.Rectangle{
			Min: image.Point{X: x1, Y: y1},
			Max: image.Point{X: x2, Y: y2},
		}

		cfg.TooltipOffset = image.Point{
			X: x1 - cfg.ItemPos.X,
			Y: y1 - cfg.ItemPos.Y,
		}
		cfg.TooltipSize = image.Point{
			X: x2 - x1,
			Y: y2 - y1,
		}

		fmt.Printf("\n✓ Tooltip region: %dx%d pixels\n", x2-x1, y2-y1)
		fmt.Printf("✓ Offset from item: (%d, %d)\n", cfg.TooltipOffset.X, cfg.TooltipOffset.Y)

		fmt.Println("\n📸 Capturing and testing tooltip area...")
		time.Sleep(500 * time.Millisecond)
		tooltipBitmap := robotgo.CaptureScreen(x1, y1, x2-x1, y2-y1)
		tooltipImg := robotgo.ToImage(tooltipBitmap)

		tooltipSnapshotFile := filepath.Join(config.SnapshotsDir, "tooltip_area_setup.png")
		SaveImage(tooltipImg, tooltipSnapshotFile)
		fmt.Printf("\n✓ Snapshot saved: %s\n", tooltipSnapshotFile)

		fmt.Println("\n🔍 Running OCR test...")
		tempDir := filepath.Join(os.TempDir(), "poe2_crafter_setup")
		os.MkdirAll(tempDir, 0755)

		ocrText, err := e.RunTesseractOCR(tooltipImg, tempDir, "")
		if err != nil {
			fmt.Printf("\n❌ OCR Error: %v\n", err)
			fmt.Print("\nRetry tooltip selection? (y/n): ")
			scanner.Scan()
			if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
				x1, y1 = CaptureWithCountdown("Re-capture TOP-LEFT corner of tooltip")
				x2, y2 = CaptureWithCountdown("Re-capture BOTTOM-RIGHT corner of tooltip")
				continue
			}
			break
		}

		fmt.Println("\n📝 OCR Results:")
		fmt.Println("----------------------------------------")
		fmt.Println(ocrText)
		fmt.Println("----------------------------------------")

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
				x1, y1 = CaptureWithCountdown("Re-capture TOP-LEFT corner of tooltip")
				x2, y2 = CaptureWithCountdown("Re-capture BOTTOM-RIGHT corner of tooltip")
				continue
			}

			fmt.Println("\n⚠  Proceeding with current selection (may cause issues during crafting)")
			break
		}
	}

	return cfg
}

// SetupWizardConfigureModsAndOptions configures target mods and other options
func SetupWizardConfigureModsAndOptions(cfg config.Config, scanner *bufio.Scanner) config.Config {
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

	cfg.TargetMods = []config.ModRequirement{}
	modNum := 1
	for {
		fmt.Printf("Mod #%d (or press Enter if done): ", modNum)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			if len(cfg.TargetMods) == 0 {
				fmt.Println("❌ Please enter at least one mod")
			fmt.Println()
				continue
			}
			break
		}

		mod := config.ParseModInput(input, "")
		if mod.Pattern != "" {
			cfg.TargetMods = append(cfg.TargetMods, mod)
			fmt.Printf("✓ Added: %s\n", mod.Description)
			modNum++
		} else {
			fmt.Println("❌ Invalid format. Try 'life 80' or 'fire-res 30'")
			fmt.Println()
		}
	}

	fmt.Printf("\n✓ Total mods to search: %d\n", len(cfg.TargetMods))

	fmt.Println("\n\nStep 5: Options")
	fmt.Println("----------------")

	fmt.Print("\nChaos orbs per round/item (default 10): ")
	scanner.Scan()
	if chaos := scanner.Text(); chaos != "" {
		if n, err := strconv.Atoi(chaos); err == nil && n > 0 {
			cfg.ChaosPerRound = n
		}
	}

	fmt.Print("Enable OCR text logging? (y/n, default n): ")
	scanner.Scan()
	cfg.Debug = strings.ToLower(scanner.Text()) == "y"

	fmt.Print("Save all snapshots for every attempt? (y/n, default n): ")
	scanner.Scan()
	cfg.SaveAllSnapshots = strings.ToLower(scanner.Text()) == "y"

	return cfg
}

// SetupWizard is the main setup wizard function
func (e *Engine) SetupWizard() config.Config {
	scanner := bufio.NewScanner(os.Stdin)
	cfg := config.Config{
		ItemWidth:        1,
		ItemHeight:       1,
		ChaosPerRound:    10,
		UseBatchMode:     true,
		Delay:            75 * time.Millisecond,
		Debug:            false,
		SaveAllSnapshots: false,
	}

	if prevConfig, err := config.LoadConfig(); err == nil {
		var needsMods bool
		prevConfig, needsMods = SetupWizardPreviousConfig(prevConfig, scanner)

		cfg = prevConfig

		if cfg.ItemWidth == 0 {
			cfg.ItemWidth = 1
		}
		if cfg.ItemHeight == 0 {
			cfg.ItemHeight = 1
		}
		cfg.UseBatchMode = true
		if cfg.ChaosPerRound == 0 {
			cfg.ChaosPerRound = 10
		}

		if !needsMods {
			fmt.Println("\n✓ Using existing configuration")
		} else {
			cfg = SetupWizardSelectiveModifications(cfg, scanner)
		}

		cfg.TooltipRect = image.Rectangle{
			Min: image.Point{
				X: cfg.ItemPos.X + cfg.TooltipOffset.X,
				Y: cfg.ItemPos.Y + cfg.TooltipOffset.Y,
			},
			Max: image.Point{
				X: cfg.ItemPos.X + cfg.TooltipOffset.X + cfg.TooltipSize.X,
				Y: cfg.ItemPos.Y + cfg.TooltipOffset.Y + cfg.TooltipSize.Y,
			},
		}

		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("⚠ Could not save: %v\n", err)
		} else {
			fmt.Println("✓ Config saved")
		}

		return cfg
	}

	needsFullSetup := cfg.ChaosPos.X == 0 && cfg.ChaosPos.Y == 0

	if needsFullSetup {
		cfg = SetupWizardFullSetup(cfg, scanner)
	}

	cfg = e.SetupWizardConfigureTooltip(cfg, scanner)

	if needsFullSetup {
		cfg = SetupWizardConfigureModsAndOptions(cfg, scanner)
	}

	fmt.Println("\n💾 Saving configuration...")
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Printf("⚠ Warning: Could not save config: %v\n", err)
		fmt.Println("   (You'll need to reconfigure next time)")
	} else {
		fmt.Printf("✓ Configuration saved to: %s\n", config.GetConfigPath())
		fmt.Println("   (Will be auto-loaded next time)")
	}

	return cfg
}
