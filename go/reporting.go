package main

import (
	"fmt"
	"image"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ModStat tracks statistics for a specific mod
type ModStat struct {
	ModName    string
	Count      int
	MinValue   int
	MaxValue   int
	AvgValue   float64
	TotalValue int
}

// RoundResult tracks data for a single round/item
type RoundResult struct {
	RoundNumber   int
	Success       bool
	StartPos      image.Point
	EndPos        image.Point
	ModsFound     []string
	TargetHit     bool
	TargetModName string
	TargetValue   int
	ErrorMessage  string
}

// CraftingSession tracks all data during a crafting session
type CraftingSession struct {
	StartTime     time.Time
	EndTime       time.Time
	TotalRolls    int
	ModStats      map[string]*ModStat // Key: mod name
	TargetModHit  bool
	TargetModName string // Which target mod was found
	TargetValue   int
	RoundResults  []RoundResult // Track each individual round
}

// trackMods parses OCR text and tracks all mods found
func trackMods(text string, session *CraftingSession, rollNumber int) {
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

	// Per-Round Details (Batch Mode)
	if len(session.RoundResults) > 0 {
		report.WriteString("ROUND-BY-ROUND DETAILS\n")
		report.WriteString("─────────────────────────────────────────────────\n")
		for _, round := range session.RoundResults {
			report.WriteString(fmt.Sprintf("\n📦 Round #%d\n", round.RoundNumber))
			report.WriteString(fmt.Sprintf("   Start Position: (%d, %d)\n", round.StartPos.X, round.StartPos.Y))
			report.WriteString(fmt.Sprintf("   End Position:   (%d, %d)\n", round.EndPos.X, round.EndPos.Y))

			if round.Success {
				report.WriteString("   Result: ✓ SUCCESS\n")
				if round.TargetHit {
					report.WriteString(fmt.Sprintf("   Target Hit: %s = %d\n", round.TargetModName, round.TargetValue))
				}
			} else {
				report.WriteString("   Result: ○ No target match\n")
			}

			if round.ErrorMessage != "" {
				report.WriteString(fmt.Sprintf("   Error: %s\n", round.ErrorMessage))
			}
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
