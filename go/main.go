package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Global control flags with atomic access for thread safety
var (
	stopRequested       atomic.Bool
	pauseRequested      atomic.Bool
	pauseToggleCooldown atomic.Value // stores time.Time
	lastPauseKeyState   atomic.Bool
	snapshotCounter     atomic.Int32 // Sequential counter for snapshot naming
)

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
