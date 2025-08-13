package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/internal/webtools"
)

func main() {
	// Initialize logger
	logConfig := logger.Config{
		LogLevel:    "debug", // Enable debug logging to see what's happening
		LogDir:      "logs",
		MaxSize:     10,
		MaxBackups:  3,
		MaxAge:      7,
		Compress:    true,
		Development: false,
	}

	log, err := logger.New(logConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// Initialize browser manager
	browserConfig := browser.Config{
		Headless:     true,
		Debug:        false,
		SlowMotion:   0,
		WindowWidth:  1920,
		WindowHeight: 1080,
	}

	browserMgr := browser.NewManager(log, browserConfig)
	if err := browserMgr.Start(browserConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start browser: %v\n", err)
		os.Exit(1)
	}
	defer browserMgr.Stop()

	// Create screen scraper tool
	scraper := webtools.NewScreenScrapeTool(log, browserMgr)

	fmt.Println("🕷️  Quick Screen Scraper Test")
	fmt.Println("=============================")

	// Get absolute path to test file
	testFile, _ := filepath.Abs("test_page.html")
	fileURL := "file://" + testFile

	// Test 1: Single item scraping
	fmt.Println("\n1. Single Item Scraping")
	singleArgs := map[string]interface{}{
		"url": fileURL,
		"selectors": map[string]interface{}{
			"title":       "h1",
			"description": "p",
		},
	}

	result, err := scraper.Execute(singleArgs)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else if len(result.Content) > 0 {
		fmt.Printf("✅ Success! Scraped data:\n")
		data, _ := json.MarshalIndent(result.Content[0].Data, "", "  ")
		fmt.Printf("%s\n", string(data))
	}

	// Test 2: Multiple items scraping
	fmt.Println("\n2. Multiple Items Scraping")
	multipleArgs := map[string]interface{}{
		"url":                fileURL,
		"extract_type":       "multiple",
		"container_selector": ".article",
		"selectors": map[string]interface{}{
			"title":   "h2",
			"content": ".content",
			"link":    "a.link",
		},
	}

	result, err = scraper.Execute(multipleArgs)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else if len(result.Content) > 0 {
		fmt.Printf("✅ Success! Multiple items scraped:\n")
		if data := result.Content[0].Data; data != nil {
			if dataMap, ok := data.(map[string]interface{}); ok {
				if items, ok := dataMap["data"].([]map[string]interface{}); ok {
					fmt.Printf("Found %d articles:\n", len(items))
					for i, item := range items {
						itemData, _ := json.MarshalIndent(item, "", "  ")
						fmt.Printf("Article %d: %s\n", i+1, string(itemData))
					}
				}
			}
		}
	}

	fmt.Println("\n✅ Quick Screen Scraper Test Complete!")
}