package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func main() {
	startTime := time.Now()
	inputDir := getEnv("INPUT_DIR", "/data")
	outputDir := getEnv("OUTPUT_DIR", "/output")
	processedDir := getEnv("PROCESSED_DIR", "/processed")
	fps := getEnv("FPS", "24")

	fmt.Printf("[START] Timelapse job initiated at %v\n", startTime.Format("15:04:05"))
	fmt.Printf("[CONFIG] Input: %s | Output: %s | Archive: %s | FPS: %s\n", inputDir, outputDir, processedDir, fps)

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		fmt.Printf("[ERROR] Failed to read input directory: %v\n", err)
		return
	}

	processedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dayFolderName := entry.Name()
		dayPath := filepath.Join(inputDir, dayFolderName)
		folderStartTime := time.Now()

		fmt.Printf("\n--- [%s] Processing Started ---\n", dayFolderName)

		var files []string
		err := filepath.Walk(dayPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("[WARN] Error accessing path %s: %v\n", path, err)
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".jpg") {
				files = append(files, path)
			}
			return nil
		})

		if err != nil {
			fmt.Printf("[ERROR] Walk failed for %s: %v\n", dayFolderName, err)
			continue
		}

		totalFiles := len(files)
		fmt.Printf("[1/4] Discovery: Found %d images\n", totalFiles)

		if totalFiles == 0 {
			fmt.Printf("[SKIP] No images found in %s. Skipping...\n", dayFolderName)
			continue
		}
		sort.Strings(files)

		timestamp := time.Now().Format("2006-01-02-15-04-05")
		outputFileName := fmt.Sprintf("timelapse_%s.mp4", timestamp)
		outputFile := filepath.Join(outputDir, outputFileName)

		tempList := fmt.Sprintf("list_%s.txt", dayFolderName)
		f, _ := os.Create(tempList)
		for _, p := range files {
			fmt.Fprintf(f, "file '%s'\n", p)
		}
		f.Close()

		fmt.Printf("[2/4] Encoding: Generating %s at %s FPS...\n", outputFileName, fps)
		cmd := exec.Command("ffmpeg", "-y", "-r", fps, "-f", "concat", "-safe", "0", "-i", tempList,
			"-c:v", "libx264", "-pix_fmt", "yuv420p", outputFile)

		if output, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("[ERROR] FFmpeg failed for %s: %v\nDebug Info:\n%s\n", dayFolderName, err, string(output))
			os.Remove(tempList)
			continue
		}
		os.Remove(tempList)
		fmt.Printf("[SUCCESS] Video saved: %s\n", outputFile)

		destPath := filepath.Join(processedDir, dayFolderName)
		fmt.Printf("[3/4] Archiving: Copying to %s...\n", destPath)

		cpCmd := exec.Command("cp", "-r", dayPath, destPath)
		if output, err := cpCmd.CombinedOutput(); err != nil {
			fmt.Printf("[ERROR] Copy failed: %v\n%s\n", err, string(output))
			continue
		}

		fmt.Printf("[4/4] Cleanup: Removing original folder %s\n", dayPath)
		rmCmd := exec.Command("rm", "-rf", dayPath)
		if err := rmCmd.Run(); err != nil {
			fmt.Printf("[WARN] Cleanup failed: %v\n", err)
		}

		processedCount++
		fmt.Printf("--- [%s] Completed in %v ---\n", dayFolderName, time.Since(folderStartTime).Round(time.Second))
	}

	fmt.Printf("\n[FINISH] Processed %d folders in %v\n", processedCount, time.Since(startTime).Round(time.Second))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
