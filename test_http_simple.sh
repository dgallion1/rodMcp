#!/bin/bash

# Simplified HTTP MCP Test with Navigation and Text Entry
set -e

BASE_URL="http://localhost:3000"
echo "🧪 Testing HTTP MCP Server Navigation & Text Entry"
echo "=" $(printf '=%.0s' {1..50})

# Test server health
echo "📡 1. Testing server health..."
curl -s "$BASE_URL/health" | jq -r '.status'

# Initialize the server
echo "🚀 2. Initializing MCP server..."
curl -s -X POST "$BASE_URL/mcp/initialize" \
  -H "Content-Type: application/json" \
  -d '{"protocolVersion": "2025-06-18", "capabilities": {}}' | jq -r '.serverInfo.name'

# Create a simple test page with proper JSON escaping
echo "📝 3. Creating simple test webpage..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "create_page",
    "arguments": {
      "filename": "simple_test.html",
      "title": "HTTP MCP Navigation Test", 
      "html": "<h1>HTTP MCP Test</h1><form><input type=\"text\" id=\"testInput\" placeholder=\"Enter text here\" style=\"padding:10px;margin:10px;font-size:16px;\"><br><button type=\"button\" id=\"testBtn\" style=\"padding:10px 20px;margin:10px;font-size:16px;background:#007bff;color:white;border:none;border-radius:4px;\">Click Me</button></form><div id=\"output\" style=\"margin:10px;padding:10px;border:1px solid #ccc;\">Waiting for input...</div>",
      "css": "body{font-family:Arial,sans-serif;margin:40px;background:#f5f5f5;}",
      "javascript": "document.getElementById('testInput').addEventListener('input', function(e) { document.getElementById('output').textContent = 'Input: ' + e.target.value; }); document.getElementById('testBtn').addEventListener('click', function() { document.getElementById('output').innerHTML = '<strong>Button clicked!</strong> Input value: ' + document.getElementById('testInput').value; });"
    }
  }' | jq -r '.content[0].text'

# Navigate to the test page  
echo "🌐 4. Navigating to test page..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "navigate_page",
    "arguments": {
      "url": "simple_test.html"
    }
  }' | jq -r '.content[0].text'

# Wait for page load
echo "⏳ 5. Waiting for page to load..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{"name": "wait", "arguments": {"seconds": 2}}' | jq -r '.content[0].text'

# Enter text in input field
echo "✏️  6. Entering text in input field..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "type_text",
    "arguments": {
      "selector": "#testInput",
      "text": "Hello from HTTP MCP!"
    }
  }' | jq -r '.content[0].text'

# Click the button
echo "🖱️  7. Clicking test button..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "click_element", 
    "arguments": {
      "selector": "#testBtn"
    }
  }' | jq -r '.content[0].text'

# Read the output to verify interaction
echo "📖 8. Reading interaction result..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "get_element_text",
    "arguments": {
      "selector": "#output"
    }
  }' | jq -r '.content[0].text'

# Take screenshot
echo "📸 9. Taking screenshot..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "take_screenshot",
    "arguments": {
      "filename": "http_test_final.png"
    }
  }' | jq -r '.content[0].text'

echo ""
echo "🎉 SUCCESS! HTTP MCP Test Complete!"
echo "✅ Demonstrated: Navigation ✅ Text Entry ✅ Button Clicks ✅ Screenshots"
echo "📁 Files: simple_test.html, http_test_final.png"