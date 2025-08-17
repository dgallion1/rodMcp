package webtools

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRealisticBrowserOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping realistic browser test in short mode")
	}
	
	// Create test browser manager
	tbm := NewTestBrowserManager(t)
	
	t.Run("CreateAndNavigateFlow", func(t *testing.T) {
		// Create a test page
		testPage := tbm.CreateTestPage(t, "test-flow.html", "")
		
		// Navigate to the page
		page, pageID := tbm.NavigateToPageWithRetry(t, "./"+testPage, 3)
		if page == nil {
			t.Fatal("Failed to navigate to test page")
		}
		
		// Test screenshot with proper timing
		screenshotTool := NewScreenshotTool(tbm.log, tbm.Manager)
		tbm.ExecuteWithTimeout(t, func() error {
			response, err := screenshotTool.Execute(map[string]interface{}{
				"filename": "realistic-test.png",
				"page_id":  pageID,
			})
			if err != nil {
				return err
			}
			if response.IsError {
				return fmt.Errorf("screenshot error: %s", response.Content[0].Text)
			}
			return nil
		}, 10*time.Second, "Screenshot operation")
		
		// Verify screenshot file was created
		if _, err := os.Stat("realistic-test.png"); os.IsNotExist(err) {
			t.Error("Screenshot file was not created")
		}
	})
	
	t.Run("ScriptExecutionFlow", func(t *testing.T) {
		// Create page with interactive content
		content := `<!DOCTYPE html>
<html>
<head><title>Script Test</title></head>
<body>
    <div id="target">Original Content</div>
    <script>
        window.testFunction = function() {
            document.getElementById('target').textContent = 'Modified by script';
            return 'Success';
        };
    </script>
</body>
</html>`
		
		testPage := tbm.CreateTestPage(t, "script-test.html", content)
		page, pageID := tbm.NavigateToPageWithRetry(t, "./"+testPage, 3)
		if page == nil {
			t.Fatal("Failed to navigate to script test page")
		}
		
		// Wait for page to load fully
		time.Sleep(1 * time.Second)
		
		// Execute script with proper timing
		scriptTool := NewExecuteScriptTool(tbm.log, tbm.Manager)
		tbm.ExecuteWithTimeout(t, func() error {
			response, err := scriptTool.Execute(map[string]interface{}{
				"script":  "window.testFunction()",
				"page_id": pageID,
			})
			if err != nil {
				return err
			}
			if response.IsError {
				return fmt.Errorf("script error: %s", response.Content[0].Text)
			}
			
			// Verify script returned expected result
			if !strings.Contains(response.Content[0].Text, "Success") {
				return fmt.Errorf("unexpected script result: %s", response.Content[0].Text)
			}
			return nil
		}, 15*time.Second, "Script execution")
	})
}

func TestRealisticMultiPageWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping realistic multi-page test in short mode")
	}
	
	tbm := NewTestBrowserManager(t)
	
	// Create multiple test pages
	page1Content := `<!DOCTYPE html>
<html><head><title>Page 1</title></head>
<body><h1>Page 1</h1><p>First page</p><a href="./page2.html">Go to Page 2</a></body></html>`
	
	page2Content := `<!DOCTYPE html>
<html><head><title>Page 2</title></head>
<body><h1>Page 2</h1><p>Second page</p><div id="data">Page 2 Data</div></body></html>`
	
	page1 := tbm.CreateTestPage(t, "page1.html", page1Content)
	page2 := tbm.CreateTestPage(t, "page2.html", page2Content)
	
	// Navigate to first page
	_, pageID1 := tbm.NavigateToPageWithRetry(t, "./"+page1, 3)
	
	// Navigate to second page
	_, pageID2 := tbm.NavigateToPageWithRetry(t, "./"+page2, 3)
	
	// Verify both pages exist
	pages := tbm.GetAllPages()
	if len(pages) < 2 {
		t.Logf("Warning: Expected 2 pages, got %d", len(pages))
	}
	
	// Test operations on specific pages
	scriptTool := NewExecuteScriptTool(tbm.log, tbm.Manager)
	
	// Test script on page 2
	tbm.ExecuteWithTimeout(t, func() error {
		response, err := scriptTool.Execute(map[string]interface{}{
			"script":  "document.getElementById('data').textContent",
			"page_id": pageID2,
		})
		if err != nil {
			return err
		}
		if response.IsError {
			return fmt.Errorf("script error: %s", response.Content[0].Text)
		}
		
		if !strings.Contains(response.Content[0].Text, "Page 2 Data") {
			return fmt.Errorf("unexpected content: %s", response.Content[0].Text)
		}
		return nil
	}, 10*time.Second, "Page 2 script execution")
	
	// Test screenshot on page 1
	screenshotTool := NewScreenshotTool(tbm.log, tbm.Manager)
	tbm.ExecuteWithTimeout(t, func() error {
		response, err := screenshotTool.Execute(map[string]interface{}{
			"filename": "page1-final.png",
			"page_id":  pageID1,
		})
		if err != nil {
			return err
		}
		if response.IsError {
			return fmt.Errorf("screenshot error: %s", response.Content[0].Text)
		}
		return nil
	}, 10*time.Second, "Page 1 screenshot")
}

func TestRealisticErrorRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping realistic error recovery test in short mode")
	}
	
	tbm := NewTestBrowserManager(t)
	
	// Test navigation to invalid URL (should handle gracefully)
	navTool := NewNavigatePageTool(tbm.log, tbm.Manager)
	
	t.Run("InvalidURLRecovery", func(t *testing.T) {
		response, err := navTool.Execute(map[string]interface{}{
			"url": "https://definitely-invalid-domain-12345.test",
		})
		
		// Should not panic, should return error gracefully
		if err == nil && !response.IsError {
			t.Error("Expected error for invalid URL, but got success")
		}
		
		// Browser should still be functional after error
		pages := tbm.GetAllPages()
		t.Logf("Pages after invalid navigation: %d", len(pages))
	})
	
	t.Run("RecoveryWithValidPage", func(t *testing.T) {
		// After error, we should be able to navigate to valid page
		testPage := tbm.CreateTestPage(t, "recovery-test.html", "")
		
		_, pageID := tbm.NavigateToPageWithRetry(t, "./"+testPage, 3)
		
		// Verify page is accessible
		scriptTool := NewExecuteScriptTool(tbm.log, tbm.Manager)
		tbm.ExecuteWithTimeout(t, func() error {
			response, err := scriptTool.Execute(map[string]interface{}{
				"script":  "document.title",
				"page_id": pageID,
			})
			if err != nil {
				return err
			}
			if response.IsError {
				return fmt.Errorf("script error: %s", response.Content[0].Text)
			}
			if !strings.Contains(response.Content[0].Text, "Test Page") {
				return fmt.Errorf("unexpected title: %s", response.Content[0].Text)
			}
			return nil
		}, 10*time.Second, "Recovery script execution")
	})
}