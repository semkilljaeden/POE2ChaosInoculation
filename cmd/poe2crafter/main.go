package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"poe2-chaos-crafter/internal/engine"
	"poe2-chaos-crafter/internal/server"
)

func main() {
	// Parse command-line flags
	webMode := false
	webPort := 8080
	debugMode := false
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

	eng := engine.NewEngine(debugMode)

	if webMode {
		server.StartWebServer(webPort, eng)
		return
	}

	// CLI mode
	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║      POE2 Chaos Crafter - Multi Mod          ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")
	fmt.Println()

	// Setup (config is saved inside setupWizard if newly created)
	cfg := eng.SetupWizard()

	fmt.Println("\n✓ Looking for ANY of these mods:")
	for i, mod := range cfg.TargetMods {
		fmt.Printf("   %d. %s\n", i+1, mod.Description)
	}
	fmt.Println("\nStarting in 5 seconds... Switch to POE2 now!")
	time.Sleep(5 * time.Second)

	// Run the crafter
	eng.Craft(cfg)
}
