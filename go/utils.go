package main

import (
	"fmt"
	"image"
	"image/png"
	"math/rand"
	"os"
	"time"
)

// Random delay to simulate human behavior (base ± variation in ms)
func humanDelay(baseMs int, variationMs int) {
	delay := baseMs + rand.Intn(variationMs*2) - variationMs
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// saveImage saves an image to a file
func saveImage(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
