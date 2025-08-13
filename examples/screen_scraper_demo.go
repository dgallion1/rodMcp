package main

import (
	"encoding/json"
	"fmt"
	"os"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/internal/webtools"
)

func main() {
	// Initialize logger
	logConfig := logger.Config{
		LogLevel:    "info",
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

	fmt.Println("🕷️  Screen Scraper Demo")
	fmt.Println("=======================")

	// Demo 1: Scrape a simple webpage (single item)
	fmt.Println("\n1. Single Item Scraping Demo")
	singleArgs := map[string]interface{}{
		"url": "https://httpbin.org/html",
		"selectors": map[string]interface{}{
			"title":       "h1",
			"description": "p",
		},
		"wait_timeout": 5,
	}

	result, err := scraper.Execute(singleArgs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else if len(result.Content) > 0 {
		data, _ := json.MarshalIndent(result.Content[0].Data, "", "  ")
		fmt.Printf("Result: %s\n", string(data))
	}

	// Demo 2: Multiple items scraping with GitHub
	fmt.Println("\n2. Multiple Items Scraping Demo (GitHub Issues)")
	multipleArgs := map[string]interface{}{
		"url":                "https://github.com/microsoft/vscode/issues",
		"extract_type":       "multiple",
		"container_selector": ".js-issue-row",
		"selectors": map[string]interface{}{
			"title":  ".Link--primary",
			"labels": ".IssueLabel",
			"author": ".opened-by a",
		},
		"wait_for":      ".js-issue-row",
		"wait_timeout":  10,
		"scroll_to_load": true,
	}

	result, err = scraper.Execute(multipleArgs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else if len(result.Content) > 0 {
		if data := result.Content[0].Data; data != nil {
			if dataMap, ok := data.(map[string]interface{}); ok {
				if items, ok := dataMap["data"].([]map[string]interface{}); ok {
					fmt.Printf("Found %d items:\n", len(items))
					for i, item := range items {
						if i >= 3 {
							fmt.Printf("... and %d more items\n", len(items)-3)
							break
						}
						itemData, _ := json.MarshalIndent(item, "", "  ")
						fmt.Printf("Item %d: %s\n", i+1, string(itemData))
					}
				}
			}
		}
	}

	// Demo 3: Custom JavaScript before scraping
	fmt.Println("\n3. Custom Script Demo")
	customArgs := map[string]interface{}{
		"url": "https://httpbin.org/html",
		"custom_script": `
			// Add some dynamic content
			const newDiv = document.createElement('div');
			newDiv.id = 'dynamic-content';
			newDiv.textContent = 'This was added by JavaScript!';
			document.body.appendChild(newDiv);
		`,
		"selectors": map[string]interface{}{
			"title":           "h1",
			"dynamic_content": "#dynamic-content",
		},
	}

	result, err = scraper.Execute(customArgs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else if len(result.Content) > 0 {
		data, _ := json.MarshalIndent(result.Content[0].Data, "", "  ")
		fmt.Printf("Result with custom script: %s\n", string(data))
	}

	fmt.Println("\n✅ Screen Scraper Demo Complete!")
}