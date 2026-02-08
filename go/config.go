package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
)

const snapshotsDir = "snapshots"
const resourceDir = "resource"

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
	ItemPos             image.Point     // Legacy: single item position (for backward compatibility)
	ItemWidth           int             // Item width in cells (e.g., 1 for 1x1, 2 for 2x3)
	ItemHeight          int             // Item height in cells (e.g., 1 for 1x1, 3 for 2x3)
	TooltipOffset       image.Point     // Offset from ItemPos to tooltip top-left
	TooltipSize         image.Point     // Width and height of tooltip
	TooltipRect         image.Rectangle `json:"-"` // Runtime only, calculated from ItemPos + Offset
	BackpackTopLeft     image.Point     // Top-left corner of backpack grid
	BackpackBottomRight image.Point     // Bottom-right corner of backpack grid

	// Batch crafting areas
	WorkbenchTopLeft   image.Point // Top-left of workbench (exact match to item dimensions)
	PendingAreaTopLeft image.Point // Top-left of pending items area
	PendingAreaWidth   int         // Width of pending area in cells
	PendingAreaHeight  int         // Height of pending area in cells
	ResultAreaTopLeft  image.Point // Top-left of result items area
	ResultAreaWidth    int         // Width of result area in cells
	ResultAreaHeight   int         // Height of result area in cells
	UseBatchMode       bool        // Enable batch crafting workflow

	TargetMods       []ModRequirement // Support multiple target mods
	ChaosPerRound    int              // Number of chaos orbs to use per item/round
	Delay            time.Duration
	Debug            bool
	SaveAllSnapshots bool // Save every attempt's screenshot
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

// parseModInput parses user input and creates a ModRequirement
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
		"life":        {`(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+MAXIMUM\s+LIFE`, "Life %d+"},
		"mana":        {`(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+MAXIMUM\s+MANA`, "Mana %d+"},
		"str":         {`(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+STRENGTH`, "Strength %d+"},
		"dex":         {`(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+DEXTERITY`, "Dexterity %d+"},
		"int":         {`(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+INTELLIGENCE`, "Intelligence %d+"},
		"spirit":      {`(?i)[+#]?(\d+)(?:\(\d+-\d+\))?\s+TO\s+SPIRIT`, "Spirit %d+"},
		"spell-level": {`\+(\d+)\s+TO\s+LEVEL\s+OF\s+ALL\s+SPELL\s+SKILLS`, "+%d to Level of all Spell Skills (or higher)"},
		"proj-level":  {`\+(\d+)\s+TO\s+LEVEL\s+OF\s+ALL\s+PROJECTILE\s+SKILLS`, "+%d to Level of all Projectile Skills (or higher)"},
		"crit-dmg":    {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*INCREASED\s+CRITICAL\s+DAMAGE\s+BONUS`, "%d%%+ increased Critical Damage Bonus"},
		"fire-res":    {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?FIRE\s+RESISTANCE`, "Fire Res %d+%%"},
		"cold-res":    {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?COLD\s+RESISTANCE`, "Cold Res %d+%%"},
		"light-res":   {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?LIGHTNING\s+RESISTANCE`, "Lightning Res %d+%%"},
		"chaos-res":   {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?CHAOS\s+RESISTANCE`, "Chaos Res %d+%%"},
		"armor":       {`(?i)(\d+)(?:\(\d+-\d+\))?\s+(?:INCREASED\s+)?ARMOUR`, "Armour %d+"},
		"evasion":     {`(?i)(\d+)(?:\(\d+-\d+\))?\s+(?:INCREASED\s+)?EVASION`, "Evasion %d+"},
		"es":          {`(?i)\+(\d+)(?:\(\d+-\d+\))?\s+TO\s+MAXIMUM\s+ENERGY\s+SHIELD`, "Energy Shield %d+"},
		"movespeed":   {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?MOVEMENT\s+SPEED`, "Movement Speed %d+%%"},
		"attackspeed": {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?ATTACK\s+SPEED`, "Attack Speed %d+%%"},
		"castspeed":   {`(?i)(\d+)(?:\(\d+-\d+\))?%?\s*(?:INCREASED\s+)?CAST\s+SPEED`, "Cast Speed %d+%%"},
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

// captureWithCountdown captures a position with a countdown
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

// setupWizardPreviousConfig handles loading and modifying previous configuration
func setupWizardPreviousConfig(prevConfig Config, scanner *bufio.Scanner) (Config, bool) {
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

	// Display batch mode info (always enabled now)
	fmt.Println("  - Batch Mode: ENABLED")
	wbRow, wbCol := getGridCell(prevConfig, prevConfig.WorkbenchTopLeft.X, prevConfig.WorkbenchTopLeft.Y)
	fmt.Printf("    • Workbench: cell (%d, %d)\n", wbRow, wbCol)
	pRow, pCol := getGridCell(prevConfig, prevConfig.PendingAreaTopLeft.X, prevConfig.PendingAreaTopLeft.Y)
	fmt.Printf("    • Pending area: cell (%d, %d) [%dx%d cells]\n",
		pRow, pCol, prevConfig.PendingAreaWidth, prevConfig.PendingAreaHeight)
	rRow, rCol := getGridCell(prevConfig, prevConfig.ResultAreaTopLeft.X, prevConfig.ResultAreaTopLeft.Y)
	fmt.Printf("    • Result area: cell (%d, %d) [%dx%d cells]\n",
		rRow, rCol, prevConfig.ResultAreaWidth, prevConfig.ResultAreaHeight)

	// Quick start option
	fmt.Print("\nAny modifications needed? (y/n, default n): ")
	scanner.Scan()
	needsMods := strings.ToLower(strings.TrimSpace(scanner.Text())) == "y"

	return prevConfig, needsMods
}

// setupWizardSelectiveModifications handles selective modifications to existing config
func setupWizardSelectiveModifications(config Config, scanner *bufio.Scanner) Config {
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
	config = setupWizardUpdateTargetMods(config, scanner)

	// 4. Chaos per round
	config = setupWizardUpdateChaosPerRound(config, scanner)

	// 5. Logging and snapshots
	config = setupWizardUpdateLogging(config, scanner)

	// 6. Item dimensions
	config = setupWizardUpdateItemDimensions(config, scanner)

	// 7. Batch crafting areas (always enabled)
	config = setupWizardUpdateBatchAreas(config, scanner)

	// 8. Tooltip position
	config = setupWizardUpdateTooltip(config, scanner)

	// Set ItemPos to workbench for batch mode, or keep existing for single mode
	if config.UseBatchMode {
		config.ItemPos = config.WorkbenchTopLeft
	}

	return config
}

// setupWizardUpdateTargetMods updates target mods
func setupWizardUpdateTargetMods(config Config, scanner *bufio.Scanner) Config {
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
	return config
}

// setupWizardUpdateChaosPerRound updates chaos per round setting
func setupWizardUpdateChaosPerRound(config Config, scanner *bufio.Scanner) Config {
	fmt.Print("\nUpdate chaos orbs per round? (y/n): ")
	scanner.Scan()
	if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
		for {
			fmt.Printf("Chaos orbs per round/item (current: %d, default 10): ", config.ChaosPerRound)
			scanner.Scan()
			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				break // Keep current value
			}
			if n, err := strconv.Atoi(input); err == nil && n > 0 {
				config.ChaosPerRound = n
				fmt.Printf("✓ Chaos per round: %d\n", config.ChaosPerRound)
				break
			}
			fmt.Println("❌ Invalid. Must be a positive number.")
		}
	}
	return config
}

// setupWizardUpdateLogging updates logging and snapshot options
func setupWizardUpdateLogging(config Config, scanner *bufio.Scanner) Config {
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
	return config
}

// setupWizardUpdateItemDimensions updates item dimensions
func setupWizardUpdateItemDimensions(config Config, scanner *bufio.Scanner) Config {
	fmt.Print("\nUpdate item dimensions? (y/n): ")
	scanner.Scan()
	if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
		fmt.Println("\n📏 Item Dimensions:")
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
		fmt.Printf("✓ Item size: %dx%d cells\n", config.ItemWidth, config.ItemHeight)
	}
	return config
}

// setupWizardUpdateBatchAreas updates batch crafting areas
func setupWizardUpdateBatchAreas(config Config, scanner *bufio.Scanner) Config {
	fmt.Print("\nUpdate batch crafting areas? (y/n): ")
	scanner.Scan()
	if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
		fmt.Println("\n🔄 Batch Crafting Areas")
		fmt.Println("You'll specify:")
		fmt.Println("  - Workbench: where items are crafted")
		fmt.Println("  - Pending area: holds items waiting to be crafted")
		fmt.Println("  - Result area: holds finished items")

		// Workbench
		config = setupWizardConfigureWorkbench(config, scanner)

		// Pending area
		config = setupWizardConfigurePendingArea(config, scanner)

		// Result area
		config = setupWizardConfigureResultArea(config, scanner)
	}
	return config
}

// setupWizardConfigureWorkbench configures the workbench position
func setupWizardConfigureWorkbench(config Config, scanner *bufio.Scanner) Config {
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
	return config
}

// setupWizardConfigurePendingArea configures the pending area
func setupWizardConfigurePendingArea(config Config, scanner *bufio.Scanner) Config {
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
	return config
}

// setupWizardConfigureResultArea configures the result area
func setupWizardConfigureResultArea(config Config, scanner *bufio.Scanner) Config {
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
	return config
}

// setupWizardUpdateTooltip updates tooltip position
func setupWizardUpdateTooltip(config Config, scanner *bufio.Scanner) Config {
	fmt.Print("\nUpdate tooltip position? (y/n): ")
	scanner.Scan()
	if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
		fmt.Println("\n📝 Tooltip Configuration")
		fmt.Println("Position mouse over item and use countdown to capture tooltip area")
		fmt.Println("")

		x1, y1 := captureWithCountdown("Tooltip TOP-LEFT corner")
		x2, y2 := captureWithCountdown("Tooltip BOTTOM-RIGHT corner")

		config.TooltipRect = image.Rectangle{
			Min: image.Point{X: x1, Y: y1},
			Max: image.Point{X: x2, Y: y2},
		}

		// Calculate offset and size (will be relative to workbench in batch mode)
		refPos := config.WorkbenchTopLeft
		if refPos.X == 0 && refPos.Y == 0 {
			// Use backpack center as fallback
			refPos = image.Point{
				X: (config.BackpackTopLeft.X + config.BackpackBottomRight.X) / 2,
				Y: (config.BackpackTopLeft.Y + config.BackpackBottomRight.Y) / 2,
			}
		}

		config.TooltipOffset = image.Point{
			X: x1 - refPos.X,
			Y: y1 - refPos.Y,
		}
		config.TooltipSize = image.Point{
			X: x2 - x1,
			Y: y2 - y1,
		}

		fmt.Printf("✓ Tooltip: (%d, %d) to (%d, %d) [%dx%d]\n",
			x1, y1, x2, y2, config.TooltipSize.X, config.TooltipSize.Y)
	}
	return config
}

// setupWizardFullSetup performs full setup for first-time users
func setupWizardFullSetup(config Config, scanner *bufio.Scanner) Config {
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

	// Step 2b: Item dimensions
	config = setupWizardConfigureItemDimensions(config, scanner)

	// Step 2c: Batch mode configuration (always enabled)
	config = setupWizardConfigureBatchMode(config, scanner)

	return config
}

// setupWizardConfigureItemDimensions configures item dimensions
func setupWizardConfigureItemDimensions(config Config, scanner *bufio.Scanner) Config {
	fmt.Println("\n📏 Item Dimensions:")
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
	fmt.Printf("✓ Item size: %dx%d cells\n", config.ItemWidth, config.ItemHeight)
	return config
}

// setupWizardConfigureBatchMode configures batch mode areas
func setupWizardConfigureBatchMode(config Config, scanner *bufio.Scanner) Config {
	fmt.Println("\n🔄 Batch Mode Configuration (always enabled)")
	fmt.Println("You'll specify:")
	fmt.Println("  - Workbench: where items are crafted")
	fmt.Println("  - Pending area: holds items waiting to be crafted")
	fmt.Println("  - Result area: holds finished items")

	// Workbench
	config = setupWizardConfigureWorkbench(config, scanner)

	// Pending area
	config = setupWizardConfigurePendingArea(config, scanner)

	// Result area
	config = setupWizardConfigureResultArea(config, scanner)

	// Set ItemPos to workbench for batch mode
	config.ItemPos = config.WorkbenchTopLeft

	return config
}

// setupWizardConfigureTooltip configures tooltip area with validation
func setupWizardConfigureTooltip(config Config, scanner *bufio.Scanner) Config {
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

	return config
}

// setupWizardConfigureModsAndOptions configures target mods and other options
func setupWizardConfigureModsAndOptions(config Config, scanner *bufio.Scanner) Config {
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

	fmt.Print("\nChaos orbs per round/item (default 10): ")
	scanner.Scan()
	if chaos := scanner.Text(); chaos != "" {
		if n, err := strconv.Atoi(chaos); err == nil && n > 0 {
			config.ChaosPerRound = n
		}
	}

	fmt.Print("Enable OCR text logging? (y/n, default n): ")
	scanner.Scan()
	config.Debug = strings.ToLower(scanner.Text()) == "y"

	fmt.Print("Save all snapshots for every attempt? (y/n, default n): ")
	scanner.Scan()
	config.SaveAllSnapshots = strings.ToLower(scanner.Text()) == "y"

	return config
}

// setupWizard is the main setup wizard function
func setupWizard() Config {
	scanner := bufio.NewScanner(os.Stdin)
	config := Config{
		ItemWidth:        1, // Default to 1x1 item
		ItemHeight:       1,
		ChaosPerRound:    10,                    // Default 10 chaos orbs per item
		UseBatchMode:     true,                  // Always use batch mode
		Delay:            75 * time.Millisecond, // Very fast default delay
		Debug:            false,
		SaveAllSnapshots: false,
	}

	// Check for previous config
	if prevConfig, err := loadConfig(); err == nil {
		var needsMods bool
		prevConfig, needsMods = setupWizardPreviousConfig(prevConfig, scanner)

		// Copy all settings from previous config
		config = prevConfig

		// Ensure item dimensions have valid defaults (backward compatibility)
		if config.ItemWidth == 0 {
			config.ItemWidth = 1
		}
		if config.ItemHeight == 0 {
			config.ItemHeight = 1
		}

		// Ensure batch mode is always enabled (backward compatibility)
		config.UseBatchMode = true

		// Ensure chaos per round has a valid default
		if config.ChaosPerRound == 0 {
			config.ChaosPerRound = 10
		}

		if !needsMods {
			// Quick start - no modifications
			fmt.Println("\n✓ Using existing configuration")
		} else {
			// Selective modifications
			config = setupWizardSelectiveModifications(config, scanner)
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
		config = setupWizardFullSetup(config, scanner)
	}

	// Configure tooltip
	config = setupWizardConfigureTooltip(config, scanner)

	// Only configure mod and options if doing full setup
	if needsFullSetup {
		config = setupWizardConfigureModsAndOptions(config, scanner)
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
