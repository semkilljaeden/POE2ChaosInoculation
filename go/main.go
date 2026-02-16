package main

import (
	"fmt"
	"os"
	"strconv"
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
	debugMode           bool         // Save snapshot files to disk (--debug)
)

func main() {
	// Parse command-line flags
	webMode := false
	webPort := 8080
	for i, arg := range os.Args[1:] {
		if arg == "--web" {
			webMode = true
			// Check if next arg is a port number
			if i+2 < len(os.Args) {
				if port, err := strconv.Atoi(os.Args[i+2]); err == nil && port > 0 && port < 65536 {
					webPort = port
				}
			}
		}
		if arg == "--debug" {
			debugMode = true
		}
	}

	if webMode {
		startWebServer(webPort)
		return
	}

	// CLI mode
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
