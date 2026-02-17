package engine

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"poe2-chaos-crafter/internal/config"
)

// PreprocessForOCR improves image quality for better OCR accuracy
func PreprocessForOCR(img image.Image) image.Image {
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

			var newBrightness uint8
			if brightness > 80 {
				enhanced := float64(brightness) * 1.5
				if enhanced > 255 {
					enhanced = 255
				}
				newBrightness = uint8(enhanced)
			} else {
				enhanced := float64(brightness) * 0.3
				newBrightness = uint8(enhanced)
			}

			contrastImg.SetGray(x, y, color.Gray{Y: newBrightness})
		}
	}

	// Step 3: Scale up 3x for better OCR
	scaledWidth := bounds.Dx() * 3
	scaledHeight := bounds.Dy() * 3
	scaledImg := image.NewGray(image.Rect(0, 0, scaledWidth, scaledHeight))

	for y := 0; y < scaledHeight; y++ {
		for x := 0; x < scaledWidth; x++ {
			srcX := bounds.Min.X + x/3
			srcY := bounds.Min.Y + y/3
			scaledImg.Set(x, y, contrastImg.At(srcX, srcY))
		}
	}

	return scaledImg
}

// RunTesseractOCRSingle runs OCR with specific settings
func RunTesseractOCRSingle(img image.Image, tempDir string, suffix string, psm int, usePreprocess bool, gameLang string) (string, error) {
	var processedImg image.Image
	if usePreprocess {
		processedImg = PreprocessForOCR(img)
	} else {
		processedImg = img
	}

	// Save preprocessed image to temp file
	tempImg := filepath.Join(tempDir, fmt.Sprintf("temp_ocr_%s.png", suffix))
	if err := SaveImage(processedImg, tempImg); err != nil {
		return "", fmt.Errorf("failed to save temp image: %w", err)
	}
	defer os.Remove(tempImg)

	// Output file (tesseract adds .txt automatically)
	tempOut := filepath.Join(tempDir, fmt.Sprintf("temp_ocr_%s", suffix))
	tempOutTxt := tempOut + ".txt"
	defer os.Remove(tempOutTxt)

	// Select language and whitelist based on game language
	tessLang := "eng"
	var tessArgs []string
	if gameLang == "zh-CN" {
		tessLang = "chi_sim"
		tessArgs = []string{tempImg, tempOut, "-l", tessLang,
			"--psm", fmt.Sprintf("%d", psm),
			"--oem", "1"}
	} else {
		tessArgs = []string{tempImg, tempOut, "-l", tessLang,
			"--psm", fmt.Sprintf("%d", psm),
			"--oem", "1",
			"-c", "tessedit_char_whitelist=ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789 +-()%#"}
	}
	cmd := exec.Command("tesseract", tessArgs...)
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

// RunTesseractOCR runs OCR with multiple strategies and returns the best result
func (e *Engine) RunTesseractOCR(img image.Image, tempDir string, gameLang string) (string, error) {
	seqNum := e.SnapshotCounter.Add(1)

	// Save original and preprocessed snapshots
	if e.DebugMode {
		debugOriginalFile := filepath.Join(config.SnapshotsDir, fmt.Sprintf("snap_%d_raw.png", seqNum))
		debugProcessedFile := filepath.Join(config.SnapshotsDir, fmt.Sprintf("snap_%d_processed.png", seqNum))
		SaveImage(img, debugOriginalFile)
		SaveImage(PreprocessForOCR(img), debugProcessedFile)
	}

	type ocrStrategy struct {
		name          string
		psm           int
		usePreprocess bool
	}

	// Fast path: Try most promising strategies first with early exit
	fastStrategies := []ocrStrategy{
		{"PSM6_raw", 6, false},
		{"PSM6_preprocessed", 6, true},
	}

	bestText := ""
	bestScore := 0

	for _, strategy := range fastStrategies {
		text, err := RunTesseractOCRSingle(img, tempDir, strategy.name, strategy.psm, strategy.usePreprocess, gameLang)
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

		if bestScore >= 80 {
			return bestText, nil
		}
	}

	if bestScore >= 30 {
		return bestText, nil
	}

	// Slow path
	fmt.Print(" [Trying alternatives...]")
	slowStrategies := []ocrStrategy{
		{"PSM4_preprocessed", 4, true},
		{"PSM11_preprocessed", 11, true},
	}

	for _, strategy := range slowStrategies {
		text, err := RunTesseractOCRSingle(img, tempDir, strategy.name, strategy.psm, strategy.usePreprocess, gameLang)
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

		if bestScore >= 80 {
			return bestText, nil
		}
	}

	if bestScore > 0 {
		return bestText, nil
	}

	return "", fmt.Errorf("all OCR strategies failed")
}

// CheckMod checks if a specific mod appears in the OCR text
func CheckMod(text string, mod config.ModRequirement) (bool, int) {
	re := regexp.MustCompile(mod.Pattern)
	matches := re.FindAllStringSubmatch(text, -1)

	if len(matches) == 0 {
		if len(strings.TrimSpace(text)) < 10 {
			fmt.Printf("\n⚠ WARNING: OCR text seems incomplete or empty")
			return false, -1
		}
		return false, 0
	}

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		numStr := strings.TrimSpace(match[1])
		value, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}

		if value >= mod.MinValue {
			return true, value
		}
	}

	return false, 0
}

// CheckAnyMod checks if any of the target mods appear in the text
func CheckAnyMod(text string, mods []config.ModRequirement) (bool, config.ModRequirement, int) {
	for _, mod := range mods {
		matched, value := CheckMod(text, mod)
		if matched {
			return true, mod, value
		}
		if value == -1 {
			return false, config.ModRequirement{}, -1
		}
	}
	return false, config.ModRequirement{}, 0
}
