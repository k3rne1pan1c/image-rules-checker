package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/k3rne1pan1c/image-checker/internal/hash"
	"github.com/k3rne1pan1c/image-checker/internal/model"
	"github.com/k3rne1pan1c/image-checker/internal/parser"
	"github.com/k3rne1pan1c/image-checker/internal/rules"
)

var (
	configPath = flag.String("config", "config/rules.yaml", "Path to rule configuration file")
	outputPath = flag.String("output", "", "Path to output JSON file (default: stdout)")
	logLevel   = flag.String("log-level", "info", "Log level: debug, info, warn, error")
	recursive  = flag.Bool("recursive", false, "Process images recursively in subdirectories")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <folder>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	folderPath := flag.Arg(0)

	// Setup logging
	logger := setupLogger(*logLevel)

	// Load rule configuration
	var ruleConfig *rules.Config
	var err error
	if *configPath != "" {
		ruleConfig, err = rules.LoadConfig(*configPath)
		if err != nil {
			logger.Warn("Failed to load config file, using defaults", "error", err)
			ruleConfig = rules.DefaultConfig()
		} else {
			logger.Info("Loaded rule configuration", "path", *configPath)
		}
	} else {
		ruleConfig = rules.DefaultConfig()
		logger.Info("Using default rule configuration")
	}

	// Create rule engine
	engine := rules.NewEngine(ruleConfig)

	// Scan folder for images
	imageFiles, err := scanImages(folderPath, *recursive, logger)
	if err != nil {
		logger.Error("Failed to scan folder", "error", err)
		os.Exit(1)
	}

	logger.Info("Found images to process", "count", len(imageFiles))

	// Process images
	result := &model.Result{
		Processed: 0,
		Passed:    0,
		Failed:    0,
		Images:    []model.Image{},
	}

	for _, imgPath := range imageFiles {
		logger.Debug("Processing image", "path", imgPath)
		imgResult := processImage(imgPath, engine, logger)
		result.Images = append(result.Images, *imgResult)
		result.Processed++

		// Determine status
		allPassed := true
		for _, check := range imgResult.Rules {
			if !check.Passed {
				allPassed = false
				break
			}
		}

		if allPassed {
			imgResult.Status = "passed"
			result.Passed++
		} else {
			imgResult.Status = "failed"
			result.Failed++
		}
	}

	// Output JSON
	outputJSON(result, *outputPath, logger)
}

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	return slog.New(handler)
}

func scanImages(folderPath string, recursive bool, logger *slog.Logger) ([]string, error) {
	var imageFiles []string

	extensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".JPG":  true,
		".JPEG": true,
		".PNG":  true,
	}

	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Warn("Error accessing path", "path", path, "error", err)
			return nil
		}

		if info.IsDir() {
			if !recursive && path != folderPath {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if extensions[ext] {
			imageFiles = append(imageFiles, path)
		}

		return nil
	}

	err := filepath.Walk(folderPath, visit)
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return imageFiles, nil
}

func processImage(imgPath string, engine *rules.Engine, logger *slog.Logger) *model.Image {
	img := &model.Image{
		Path: imgPath,
	}

	// Compute hash
	hashValue, err := hash.ComputeSHA256(imgPath)
	if err != nil {
		logger.Warn("Failed to compute hash", "path", imgPath, "error", err)
	} else {
		img.Hash = hashValue
		logger.Debug("Computed hash", "path", imgPath, "hash", hashValue[:16]+"...")
	}

	// Parse EXIF
	exifData, err := parser.ParseEXIF(imgPath)
	if err != nil {
		logger.Debug("Failed to parse EXIF", "path", imgPath, "error", err)
	} else if exifData != nil {
		img.EXIF = exifData
		logger.Debug("Parsed EXIF", "path", imgPath, "maker", exifData.Maker, "timestamp", exifData.Timestamp)
	}

	// Parse image internals (JPEG or PNG)
	ext := strings.ToLower(filepath.Ext(imgPath))
	var imageData *model.ImageData

	switch ext {
	case ".jpg", ".jpeg":
		imageData, err = parser.ParseJPEG(imgPath)
		if err != nil {
			logger.Warn("Failed to parse JPEG", "path", imgPath, "error", err)
		} else {
			img.Image = imageData
			logger.Debug("Parsed JPEG", "path", imgPath, "dimensions", fmt.Sprintf("%dx%d", imageData.Dimensions.Width, imageData.Dimensions.Height))
		}
	case ".png":
		imageData, err = parser.ParsePNG(imgPath)
		if err != nil {
			logger.Warn("Failed to parse PNG", "path", imgPath, "error", err)
		} else {
			img.Image = imageData
			logger.Debug("Parsed PNG", "path", imgPath, "dimensions", fmt.Sprintf("%dx%d", imageData.Dimensions.Width, imageData.Dimensions.Height))
		}
	default:
		logger.Warn("Unsupported image format", "path", imgPath, "ext", ext)
	}

	// Apply rules
	img.Rules = engine.CheckRules(img)
	logger.Debug("Applied rules", "path", imgPath, "rule_count", len(img.Rules))

	return img
}

func outputJSON(result *model.Result, outputPath string, logger *slog.Logger) {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal JSON", "error", err)
		os.Exit(1)
	}

	if outputPath == "" {
		// Output to stdout
		fmt.Println(string(jsonData))
	} else {
		// Output to file
		err := os.WriteFile(outputPath, jsonData, 0644)
		if err != nil {
			logger.Error("Failed to write output file", "path", outputPath, "error", err)
			os.Exit(1)
		}
		logger.Info("Output written to file", "path", outputPath)
	}
}

