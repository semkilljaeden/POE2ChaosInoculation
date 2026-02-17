package config

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const SnapshotsDir = "snapshots"
const ResourceDir = "resource"

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
	SaveAllSnapshots bool   // Save every attempt's screenshot
	GameLanguage     string // Game client language for OCR ("en" or "zh-CN")
}

// GetConfigPath returns the config file path
func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".poe2_crafter_config.json")
}

// SaveConfig saves the configuration to a JSON file
func SaveConfig(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(GetConfigPath(), data, 0644)
}

// LoadConfig loads the configuration from a JSON file
func LoadConfig() (Config, error) {
	data, err := os.ReadFile(GetConfigPath())
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

// ParseModInput parses user input and creates a ModRequirement
// gameLang selects which regex patterns to generate ("zh-CN" for Chinese, anything else for English)
func ParseModInput(input string, gameLang string) ModRequirement {
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

	type modTemplate struct {
		pattern string
		desc    string
	}

	var templates map[string]modTemplate

	if gameLang == "zh-CN" {
		templates = map[string]modTemplate{
			"life":        {`\+?(\d+)(?:\(\d+-\d+\))?\s*最大生命`, "生命 %d+"},
			"mana":        {`\+?(\d+)(?:\(\d+-\d+\))?\s*最大魔力`, "魔力 %d+"},
			"str":         {`\+?(\d+)(?:\(\d+-\d+\))?\s*力量`, "力量 %d+"},
			"dex":         {`\+?(\d+)(?:\(\d+-\d+\))?\s*敏捷`, "敏捷 %d+"},
			"int":         {`\+?(\d+)(?:\(\d+-\d+\))?\s*智慧`, "智慧 %d+"},
			"spirit":      {`\+?(\d+)(?:\(\d+-\d+\))?\s*精魂`, "精魂 %d+"},
			"spell-level": {`\+(\d+)\s*(?:所有)?法术技能等级`, "+%d 法术技能等级"},
			"proj-level":  {`\+(\d+)\s*(?:所有)?投射物技能等级`, "+%d 投射物技能等级"},
			"crit-dmg":    {`(\d+)(?:\(\d+-\d+\))?%?\s*暴击伤害加成`, "%d%% 暴击伤害加成"},
			"fire-res":    {`(\d+)(?:\(\d+-\d+\))?%?\s*火焰抗性`, "火焰抗性 %d+%%"},
			"cold-res":    {`(\d+)(?:\(\d+-\d+\))?%?\s*冰冷抗性`, "冰冷抗性 %d+%%"},
			"light-res":   {`(\d+)(?:\(\d+-\d+\))?%?\s*闪电抗性`, "闪电抗性 %d+%%"},
			"chaos-res":   {`(\d+)(?:\(\d+-\d+\))?%?\s*混沌抗性`, "混沌抗性 %d+%%"},
			"armor":       {`(\d+)(?:\(\d+-\d+\))?\s*护甲`, "护甲 %d+"},
			"evasion":     {`(\d+)(?:\(\d+-\d+\))?\s*闪避`, "闪避 %d+"},
			"es":          {`\+?(\d+)(?:\(\d+-\d+\))?\s*最大能量护盾`, "能量护盾 %d+"},
			"movespeed":   {`(\d+)(?:\(\d+-\d+\))?%?\s*移动速度`, "移动速度 %d+%%"},
			"attackspeed": {`(\d+)(?:\(\d+-\d+\))?%?\s*攻击速度`, "攻击速度 %d+%%"},
			"castspeed":   {`(\d+)(?:\(\d+-\d+\))?%?\s*施放速度`, "施放速度 %d+%%"},
		}
	} else {
		templates = map[string]modTemplate{
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
			Description: "Custom: " + input[:Min(len(input), 30)],
		}
	}

	return ModRequirement{}
}

// GetCellCenter calculates the pixel coordinates of the center of a backpack cell
// row: 0-4 (5 rows), col: 0-11 (12 columns)
func GetCellCenter(cfg Config, row int, col int) (int, int) {
	totalWidth := cfg.BackpackBottomRight.X - cfg.BackpackTopLeft.X
	totalHeight := cfg.BackpackBottomRight.Y - cfg.BackpackTopLeft.Y

	cellWidth := totalWidth / 12
	cellHeight := totalHeight / 5

	// Calculate cell center
	centerX := cfg.BackpackTopLeft.X + (col * cellWidth) + (cellWidth / 2)
	centerY := cfg.BackpackTopLeft.Y + (row * cellHeight) + (cellHeight / 2)

	return centerX, centerY
}

// GetGridCell converts pixel coordinates to cell coordinates (row, col)
func GetGridCell(cfg Config, x, y int) (int, int) {
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

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
