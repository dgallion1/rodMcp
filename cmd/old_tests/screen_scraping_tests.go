package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"rodmcp/internal/browser"
	"rodmcp/internal/logger"
	"rodmcp/internal/webtools"
	"strings"
	"time"
)

// TestResult represents a single test result
type TestResult struct {
	TestName    string        `json:"test_name"`
	Category    string        `json:"category"`
	Passed      bool          `json:"passed"`
	Duration    time.Duration `json:"duration"`
	Error       string        `json:"error,omitempty"`
	Description string        `json:"description"`
}

// TestSuite manages screen scraping tests
type TestSuite struct {
	logger     *logger.Logger
	browserMgr *browser.Manager
	scraper    *webtools.ScreenScrapeTool
	testFile   string
	results    []TestResult
}

func NewTestSuite() (*TestSuite, error) {
	// Initialize logger
	logConfig := logger.Config{
		LogLevel:    "error", // Minimize logging noise during tests
		LogDir:      "logs",
		MaxSize:     10,
		MaxBackups:  3,
		MaxAge:      7,
		Compress:    true,
		Development: false,
	}

	log, err := logger.New(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

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
		return nil, fmt.Errorf("failed to start browser: %w", err)
	}

	// Get absolute path to test file
	testFile, _ := filepath.Abs("test_data/scraping_test.html")

	return &TestSuite{
		logger:     log,
		browserMgr: browserMgr,
		scraper:    webtools.NewScreenScrapeTool(log, browserMgr),
		testFile:   "file://" + testFile,
		results:    make([]TestResult, 0),
	}, nil
}

func (ts *TestSuite) Cleanup() {
	if ts.browserMgr != nil {
		ts.browserMgr.Stop()
	}
	if ts.logger != nil {
		ts.logger.Sync()
	}
}

func (ts *TestSuite) runTest(testName, category, description string, testFunc func() error) {
	fmt.Printf("  %-35s ", testName+"...")
	
	start := time.Now()
	err := testFunc()
	duration := time.Since(start)

	result := TestResult{
		TestName:    testName,
		Category:    category,
		Passed:      err == nil,
		Duration:    duration,
		Description: description,
	}

	if err != nil {
		result.Error = err.Error()
		fmt.Printf("❌ FAILED (%.2fs) - %s\n", duration.Seconds(), err.Error())
	} else {
		fmt.Printf("✅ PASSED (%.2fs)\n", duration.Seconds())
	}

	ts.results = append(ts.results, result)
}

// Single Item Scraping Tests
func (ts *TestSuite) testSingleItemBasic() error {
	args := map[string]interface{}{
		"url": ts.testFile,
		"selectors": map[string]interface{}{
			"title":       "#main-title",
			"description": ".site-description",
		},
		"include_metadata": true,
	}

	result, err := ts.scraper.Execute(args)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	if len(result.Content) == 0 {
		return fmt.Errorf("no content returned")
	}

	data := result.Content[0].Data
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected data format")
	}

	// Validate metadata
	if _, exists := dataMap["metadata"]; !exists {
		return fmt.Errorf("metadata missing")
	}

	// Validate scraped data
	scrapedData, exists := dataMap["data"]
	if !exists {
		return fmt.Errorf("scraped data missing")
	}

	scrapedMap, ok := scrapedData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("scraped data not a map")
	}

	// Check title extraction
	if _, exists := scrapedMap["title"]; !exists {
		return fmt.Errorf("title not extracted")
	}

	// Check description extraction
	if _, exists := scrapedMap["description"]; !exists {
		return fmt.Errorf("description not extracted")
	}

	return nil
}

func (ts *TestSuite) testSingleItemImages() error {
	args := map[string]interface{}{
		"url": ts.testFile,
		"selectors": map[string]interface{}{
			"hero_image": "#hero-img",
		},
	}

	result, err := ts.scraper.Execute(args)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	data := result.Content[0].Data.(map[string]interface{})
	scrapedData := data["data"].(map[string]interface{})
	
	heroImage, exists := scrapedData["hero_image"]
	if !exists {
		return fmt.Errorf("hero_image not extracted")
	}

	// The heroImage should be a proper data structure with value and attributes
	// Let's validate it by checking if it contains the expected image data
	heroImageStr := fmt.Sprintf("%v", heroImage)
	
	// Check for image-specific content
	if !strings.Contains(heroImageStr, "src") {
		return fmt.Errorf("image src not found in extracted data")
	}
	if !strings.Contains(heroImageStr, "alt") {
		return fmt.Errorf("image alt not found in extracted data")
	}
	if !strings.Contains(heroImageStr, "Hero Banner") {
		return fmt.Errorf("expected alt text not found")
	}

	return nil
}

func (ts *TestSuite) testSingleItemLinks() error {
	args := map[string]interface{}{
		"url": ts.testFile,
		"selectors": map[string]interface{}{
			"nav_links": ".nav-link",
		},
	}

	result, err := ts.scraper.Execute(args)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	data := result.Content[0].Data.(map[string]interface{})
	scrapedData := data["data"].(map[string]interface{})
	
	navLinks, exists := scrapedData["nav_links"]
	if !exists {
		return fmt.Errorf("nav_links not extracted")
	}

	// Validate by checking string representation for expected link data
	linkStr := fmt.Sprintf("%v", navLinks)
	
	if !strings.Contains(linkStr, "href") {
		return fmt.Errorf("link href not found in extracted data")
	}
	if !strings.Contains(linkStr, "text") {
		return fmt.Errorf("link text not found in extracted data")
	}
	if !strings.Contains(linkStr, "Home") {
		return fmt.Errorf("expected link text not found")
	}

	return nil
}

func (ts *TestSuite) testSingleItemInputs() error {
	args := map[string]interface{}{
		"url": ts.testFile,
		"selectors": map[string]interface{}{
			"search_input": "#search-input",
			"hidden_input": "#hidden-token",
		},
	}

	result, err := ts.scraper.Execute(args)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	data := result.Content[0].Data.(map[string]interface{})
	scrapedData := data["data"].(map[string]interface{})
	
	// Check search input
	searchInput, exists := scrapedData["search_input"]
	if !exists {
		return fmt.Errorf("search_input not extracted")
	}

	// Validate by checking string representation for expected input data
	inputStr := fmt.Sprintf("%v", searchInput)
	
	if !strings.Contains(inputStr, "value") {
		return fmt.Errorf("input value not found in extracted data")
	}
	if !strings.Contains(inputStr, "placeholder") {
		return fmt.Errorf("input placeholder not found in extracted data")
	}
	if !strings.Contains(inputStr, "test search") {
		return fmt.Errorf("expected input value not found")
	}

	// Check hidden input
	hiddenInput, exists := scrapedData["hidden_input"]
	if !exists {
		return fmt.Errorf("hidden_input not extracted")
	}

	hiddenStr := fmt.Sprintf("%v", hiddenInput)
	if !strings.Contains(hiddenStr, "abc123") {
		return fmt.Errorf("expected hidden input value not found")
	}

	return nil
}

// Multiple Item Scraping Tests
func (ts *TestSuite) testMultipleItemsProducts() error {
	args := map[string]interface{}{
		"url":                ts.testFile,
		"extract_type":       "multiple",
		"container_selector": ".product-card",
		"selectors": map[string]interface{}{
			"title":       ".product-title",
			"description": ".product-description",
			"price":       ".price",
			"rating":      ".rating",
			"link":        ".buy-link",
		},
	}

	result, err := ts.scraper.Execute(args)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	data := result.Content[0].Data.(map[string]interface{})
	scrapedData := data["data"]

	items, ok := scrapedData.([]map[string]interface{})
	if !ok {
		return fmt.Errorf("scraped data is not an array of maps")
	}

	if len(items) != 3 {
		return fmt.Errorf("expected 3 products, got %d", len(items))
	}

	// Validate first product
	product1 := items[0]
	if _, exists := product1["title"]; !exists {
		return fmt.Errorf("product title not extracted")
	}
	if _, exists := product1["price"]; !exists {
		return fmt.Errorf("product price not extracted")
	}
	if _, exists := product1["_index"]; !exists {
		return fmt.Errorf("product index not set")
	}

	return nil
}

func (ts *TestSuite) testMultipleItemsNews() error {
	args := map[string]interface{}{
		"url":                ts.testFile,
		"extract_type":       "multiple",
		"container_selector": ".news-article",
		"selectors": map[string]interface{}{
			"title":    ".article-title",
			"summary":  ".article-summary",
			"date":     ".article-date",
			"author":   ".article-author",
			"read_more": ".read-more",
		},
	}

	result, err := ts.scraper.Execute(args)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	data := result.Content[0].Data.(map[string]interface{})
	scrapedData := data["data"]

	articles, ok := scrapedData.([]map[string]interface{})
	if !ok {
		return fmt.Errorf("scraped data is not an array of maps")
	}

	if len(articles) != 3 {
		return fmt.Errorf("expected 3 articles, got %d", len(articles))
	}

	// Validate first article structure
	article1 := articles[0]
	requiredFields := []string{"title", "summary", "date", "author", "read_more"}
	for _, field := range requiredFields {
		if _, exists := article1[field]; !exists {
			return fmt.Errorf("article field %s not extracted", field)
		}
	}

	return nil
}

// Advanced Feature Tests
func (ts *TestSuite) testWaitForElement() error {
	args := map[string]interface{}{
		"url":          ts.testFile,
		"wait_for":     "#dynamic-content",
		"wait_timeout": 3,
		"selectors": map[string]interface{}{
			"dynamic_text": ".dynamic-text",
		},
	}

	result, err := ts.scraper.Execute(args)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	data := result.Content[0].Data.(map[string]interface{})
	scrapedData := data["data"].(map[string]interface{})
	
	if _, exists := scrapedData["dynamic_text"]; !exists {
		return fmt.Errorf("dynamic text not extracted after wait")
	}

	return nil
}

func (ts *TestSuite) testCustomScript() error {
	customScript := `
		// Add custom data to test extraction
		const testDiv = document.createElement('div');
		testDiv.id = 'custom-test-element';
		testDiv.textContent = 'Custom script executed successfully';
		testDiv.className = 'custom-element';
		document.body.appendChild(testDiv);
	`

	args := map[string]interface{}{
		"url":           ts.testFile,
		"custom_script": customScript,
		"selectors": map[string]interface{}{
			"custom_element": "#custom-test-element",
		},
	}

	result, err := ts.scraper.Execute(args)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	data := result.Content[0].Data.(map[string]interface{})
	scrapedData := data["data"].(map[string]interface{})
	
	customElement, exists := scrapedData["custom_element"]
	if !exists {
		return fmt.Errorf("custom element not extracted")
	}

	// Verify the custom element was created and extracted
	elementStr := fmt.Sprintf("%v", customElement)
	
	if !strings.Contains(elementStr, "Custom script executed successfully") {
		return fmt.Errorf("custom script content not found")
	}

	return nil
}

func (ts *TestSuite) testScrollToLoad() error {
	args := map[string]interface{}{
		"url":            ts.testFile,
		"scroll_to_load": true,
		"selectors": map[string]interface{}{
			"footer": ".site-footer",
		},
	}

	result, err := ts.scraper.Execute(args)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	data := result.Content[0].Data.(map[string]interface{})
	scrapedData := data["data"].(map[string]interface{})
	
	if _, exists := scrapedData["footer"]; !exists {
		return fmt.Errorf("footer not extracted after scroll")
	}

	return nil
}

// Error Handling Tests
func (ts *TestSuite) testMissingElements() error {
	args := map[string]interface{}{
		"url": ts.testFile,
		"selectors": map[string]interface{}{
			"existing":     "#main-title",
			"non_existent": "#does-not-exist",
		},
	}

	result, err := ts.scraper.Execute(args)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	data := result.Content[0].Data.(map[string]interface{})
	scrapedData := data["data"].(map[string]interface{})
	
	// Existing element should be present and have content
	existing, exists := scrapedData["existing"]
	if !exists {
		return fmt.Errorf("existing element not extracted")
	}
	existingStr := fmt.Sprintf("%v", existing)
	if !strings.Contains(existingStr, "Test E-commerce Site") {
		return fmt.Errorf("existing element content not found")
	}

	// Non-existent element should be null - this is handled correctly by the scraper
	nonExistent, exists := scrapedData["non_existent"]
	if !exists {
		return fmt.Errorf("non_existent field not in results")
	}
	
	// nil should be represented as <nil> in string format, or be actually nil
	if nonExistent != nil && fmt.Sprintf("%v", nonExistent) != "<nil>" {
		return fmt.Errorf("non_existent element should be nil, got: %v (type: %T)", nonExistent, nonExistent)
	}

	return nil
}

func (ts *TestSuite) testInvalidContainerSelector() error {
	args := map[string]interface{}{
		"url":                ts.testFile,
		"extract_type":       "multiple",
		"container_selector": ".does-not-exist",
		"selectors": map[string]interface{}{
			"title": "h1",
		},
	}

	result, err := ts.scraper.Execute(args)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	data := result.Content[0].Data.(map[string]interface{})
	scrapedData := data["data"]

	items, ok := scrapedData.([]map[string]interface{})
	if !ok {
		return fmt.Errorf("scraped data is not an array")
	}

	// Should return empty array for non-existent container
	if len(items) != 0 {
		return fmt.Errorf("expected empty array for non-existent container, got %d items", len(items))
	}

	return nil
}

func (ts *TestSuite) testMetadataDisabled() error {
	args := map[string]interface{}{
		"url": ts.testFile,
		"selectors": map[string]interface{}{
			"title": "#main-title",
		},
		"include_metadata": false,
	}

	result, err := ts.scraper.Execute(args)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	data := result.Content[0].Data.(map[string]interface{})
	
	// Should only contain data, not metadata
	if _, exists := data["metadata"]; exists {
		return fmt.Errorf("metadata should not be included when disabled")
	}

	if _, exists := data["data"]; !exists {
		return fmt.Errorf("data field should still be present")
	}

	return nil
}

func (ts *TestSuite) RunAllTests() {
	fmt.Println("🕷️  Screen Scraping Test Suite")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println()

	// Single Item Tests
	fmt.Println("📊 Single Item Extraction Tests")
	fmt.Println("-" + strings.Repeat("-", 40))
	ts.runTest("Basic Single Item", "Single Item", "Extract title and description", ts.testSingleItemBasic)
	ts.runTest("Image Extraction", "Single Item", "Extract image attributes", ts.testSingleItemImages)
	ts.runTest("Link Extraction", "Single Item", "Extract link attributes", ts.testSingleItemLinks)
	ts.runTest("Input Extraction", "Single Item", "Extract input field attributes", ts.testSingleItemInputs)
	fmt.Println()

	// Multiple Item Tests
	fmt.Println("📋 Multiple Item Extraction Tests")
	fmt.Println("-" + strings.Repeat("-", 40))
	ts.runTest("Product Cards", "Multiple Items", "Extract product information", ts.testMultipleItemsProducts)
	ts.runTest("News Articles", "Multiple Items", "Extract article information", ts.testMultipleItemsNews)
	fmt.Println()

	// Advanced Feature Tests
	fmt.Println("⚙️  Advanced Feature Tests")
	fmt.Println("-" + strings.Repeat("-", 40))
	ts.runTest("Wait For Element", "Advanced", "Wait for dynamic content", ts.testWaitForElement)
	ts.runTest("Custom Script", "Advanced", "Execute custom JavaScript", ts.testCustomScript)
	ts.runTest("Scroll To Load", "Advanced", "Scroll to trigger content loading", ts.testScrollToLoad)
	fmt.Println()

	// Error Handling Tests
	fmt.Println("🛡️  Error Handling Tests")
	fmt.Println("-" + strings.Repeat("-", 40))
	ts.runTest("Missing Elements", "Error Handling", "Handle non-existent elements", ts.testMissingElements)
	ts.runTest("Invalid Container", "Error Handling", "Handle invalid container selector", ts.testInvalidContainerSelector)
	ts.runTest("Metadata Disabled", "Error Handling", "Test without metadata", ts.testMetadataDisabled)
	fmt.Println()

	ts.printSummary()
}

func (ts *TestSuite) printSummary() {
	categories := make(map[string][]TestResult)
	totalTests := len(ts.results)
	passedTests := 0
	var totalDuration time.Duration

	// Group results by category
	for _, result := range ts.results {
		categories[result.Category] = append(categories[result.Category], result)
		if result.Passed {
			passedTests++
		}
		totalDuration += result.Duration
	}

	fmt.Println("📈 Test Summary")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("Total Tests: %d | Passed: %d | Failed: %d | Success Rate: %.1f%%\n", 
		totalTests, passedTests, totalTests-passedTests, float64(passedTests)/float64(totalTests)*100)
	fmt.Printf("Total Duration: %.2f seconds\n", totalDuration.Seconds())
	fmt.Println()

	// Category breakdown
	for category, results := range categories {
		passed := 0
		for _, r := range results {
			if r.Passed {
				passed++
			}
		}
		fmt.Printf("%-20s: %d/%d tests passed (%.1f%%)\n", 
			category, passed, len(results), float64(passed)/float64(len(results))*100)
	}

	fmt.Println()

	// Show failed tests if any
	failedTests := make([]TestResult, 0)
	for _, result := range ts.results {
		if !result.Passed {
			failedTests = append(failedTests, result)
		}
	}

	if len(failedTests) > 0 {
		fmt.Println("❌ Failed Tests:")
		fmt.Println("-" + strings.Repeat("-", 40))
		for _, failed := range failedTests {
			fmt.Printf("  %s (%s): %s\n", failed.TestName, failed.Category, failed.Error)
		}
		fmt.Println()
	}

	// Export detailed results
	if err := ts.exportResults(); err != nil {
		fmt.Printf("⚠️  Warning: Could not export detailed results: %v\n", err)
	} else {
		fmt.Println("📄 Detailed results exported to: screen_scraping_test_results.json")
	}
}

func (ts *TestSuite) exportResults() error {
	summary := map[string]interface{}{
		"test_suite":    "Screen Scraping Tests",
		"timestamp":     time.Now().Format(time.RFC3339),
		"total_tests":   len(ts.results),
		"passed_tests":  0,
		"failed_tests":  0,
		"success_rate":  0.0,
		"test_results":  ts.results,
	}

	passedCount := 0
	for _, result := range ts.results {
		if result.Passed {
			passedCount++
		}
	}

	summary["passed_tests"] = passedCount
	summary["failed_tests"] = len(ts.results) - passedCount
	summary["success_rate"] = float64(passedCount) / float64(len(ts.results)) * 100

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("screen_scraping_test_results.json", data, 0644)
}

func main() {
	fmt.Println("Initializing Screen Scraping Test Suite...")
	
	suite, err := NewTestSuite()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize test suite: %v\n", err)
		os.Exit(1)
	}
	defer suite.Cleanup()

	suite.RunAllTests()

	// Exit with non-zero code if any tests failed
	passedTests := 0
	for _, result := range suite.results {
		if result.Passed {
			passedTests++
		}
	}

	if passedTests != len(suite.results) {
		os.Exit(1)
	}
}