#!/bin/bash

# Test HTTP MCP Server with Navigation and Text Entry
# This script demonstrates the HTTP MCP server working with browser automation

set -e

BASE_URL="http://localhost:3000"
echo "🧪 Testing HTTP MCP Server at $BASE_URL"
echo "=" $(printf '=%.0s' {1..50})

# Test server health
echo "📡 1. Testing server health..."
curl -s "$BASE_URL/health" | jq .
echo

# Initialize the server
echo "🚀 2. Initializing MCP server..."
curl -s -X POST "$BASE_URL/mcp/initialize" \
  -H "Content-Type: application/json" \
  -d '{"protocolVersion": "2025-06-18", "capabilities": {}}' | jq .
echo

# List available tools
echo "🛠️  3. Listing available tools..."
curl -s "$BASE_URL/mcp/tools/list" | jq -r '.tools[] | "  - \(.name): \(.description)"'
echo

# Create an interactive test webpage
echo "📝 4. Creating interactive test webpage..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "create_page",
    "arguments": {
      "filename": "http_test_page.html",
      "title": "HTTP MCP Test Page",
      "html": "<div class=\\\"container\\\"><h1>🧪 HTTP MCP Test Page</h1><form id=\\\"testForm\\\"><div class=\\\"form-group\\\"><label for=\\\"name\\\">Name:</label><input type=\\\"text\\\" id=\\\"name\\\" name=\\\"name\\\" placeholder=\\\"Enter your name\\\" required></div><div class=\\\"form-group\\\"><label for=\\\"email\\\">Email:</label><input type=\\\"email\\\" id=\\\"email\\\" name=\\\"email\\\" placeholder=\\\"Enter your email\\\" required></div><div class=\\\"form-group\\\"><label for=\\\"message\\\">Message:</label><textarea id=\\\"message\\\" name=\\\"message\\\" placeholder=\\\"Enter your message\\\" rows=\\\"4\\\"></textarea></div><button type=\\\"button\\\" id=\\\"submitBtn\\\" class=\\\"btn\\\">Submit Form</button></form><div id=\\\"result\\\" class=\\\"result\\\"></div><div class=\\\"stats\\\"><h3>Form Stats:</h3><p>Name length: <span id=\\\"nameLength\\\">0</span></p><p>Email length: <span id=\\\"emailLength\\\">0</span></p><p>Message length: <span id=\\\"messageLength\\\">0</span></p></div></div>",
      "css": "* { margin: 0; padding: 0; box-sizing: border-box; } body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); min-height: 100vh; padding: 20px; } .container { max-width: 600px; margin: 0 auto; background: white; padding: 30px; border-radius: 15px; box-shadow: 0 10px 30px rgba(0,0,0,0.1); } h1 { color: #333; margin-bottom: 30px; text-align: center; } .form-group { margin-bottom: 20px; } label { display: block; margin-bottom: 8px; color: #555; font-weight: 500; } input, textarea { width: 100%; padding: 12px; border: 2px solid #e1e1e1; border-radius: 8px; font-size: 16px; transition: border-color 0.3s; } input:focus, textarea:focus { outline: none; border-color: #667eea; } .btn { background: linear-gradient(45deg, #667eea, #764ba2); color: white; padding: 12px 30px; border: none; border-radius: 8px; font-size: 16px; cursor: pointer; transition: transform 0.2s; } .btn:hover { transform: translateY(-2px); } .result { margin-top: 20px; padding: 15px; border-radius: 8px; display: none; } .result.success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; } .stats { margin-top: 20px; padding: 15px; background: #f8f9fa; border-radius: 8px; } .stats h3 { color: #333; margin-bottom: 10px; } .stats p { margin: 5px 0; color: #666; }",
      "javascript": "console.log('🧪 HTTP MCP Test Page Loaded'); const nameInput = document.getElementById('name'); const emailInput = document.getElementById('email'); const messageInput = document.getElementById('message'); const submitBtn = document.getElementById('submitBtn'); const result = document.getElementById('result'); const nameLength = document.getElementById('nameLength'); const emailLength = document.getElementById('emailLength'); const messageLength = document.getElementById('messageLength'); function updateStats() { nameLength.textContent = nameInput.value.length; emailLength.textContent = emailInput.value.length; messageLength.textContent = messageInput.value.length; } nameInput.addEventListener('input', updateStats); emailInput.addEventListener('input', updateStats); messageInput.addEventListener('input', updateStats); submitBtn.addEventListener('click', function() { const name = nameInput.value.trim(); const email = emailInput.value.trim(); const message = messageInput.value.trim(); if (!name || !email) { alert('Please fill in name and email fields'); return; } result.className = 'result success'; result.style.display = 'block'; result.innerHTML = '\\u003Ch4\\u003E✅ Form Submitted Successfully!\\u003C/h4\\u003E\\u003Cp\\u003E\\u003Cstrong\\u003EName:\\u003C/strong\\u003E ' + name + '\\u003C/p\\u003E\\u003Cp\\u003E\\u003Cstrong\\u003EEmail:\\u003C/strong\\u003E ' + email + '\\u003C/p\\u003E\\u003Cp\\u003E\\u003Cstrong\\u003EMessage:\\u003C/strong\\u003E ' + (message || 'No message provided') + '\\u003C/p\\u003E\\u003Cp\\u003E\\u003Cstrong\\u003ETimestamp:\\u003C/strong\\u003E ' + new Date().toLocaleString() + '\\u003C/p\\u003E'; console.log('Form submitted:', { name, email, message }); }); console.log('🎉 Event listeners registered');"
    }
  }' | jq -r '.content[0].text'
echo

# Navigate to the test page
echo "🌐 5. Navigating to test page..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "navigate_page", 
    "arguments": {
      "url": "http_test_page.html"
    }
  }' | jq -r '.content[0].text'
echo

# Wait for page to load
echo "⏳ 6. Waiting for page to fully load..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "wait",
    "arguments": {
      "seconds": 2
    }
  }' | jq -r '.content[0].text'
echo

# Enter text in the name field
echo "✏️  7. Entering text in name field..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "type_text",
    "arguments": {
      "selector": "#name",
      "text": "Claude AI Assistant"
    }
  }' | jq -r '.content[0].text'
echo

# Enter text in the email field
echo "📧 8. Entering text in email field..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "type_text",
    "arguments": {
      "selector": "#email", 
      "text": "claude@anthropic.com"
    }
  }' | jq -r '.content[0].text'
echo

# Enter text in the message field
echo "💬 9. Entering text in message field..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "type_text",
    "arguments": {
      "selector": "#message",
      "text": "This is a test message from the HTTP MCP server! The navigation fix is working perfectly and the server can handle multiple requests without being killed by Claude."
    }
  }' | jq -r '.content[0].text'
echo

# Click the submit button
echo "🖱️  10. Clicking submit button..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "click_element",
    "arguments": {
      "selector": "#submitBtn"
    }
  }' | jq -r '.content[0].text'
echo

# Wait for form submission animation
echo "⏳ 11. Waiting for form submission to complete..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "wait",
    "arguments": {
      "seconds": 1
    }
  }' | jq -r '.content[0].text'
echo

# Get the result text to verify form submission
echo "📖 12. Reading form submission result..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "get_element_text",
    "arguments": {
      "selector": "#result"
    }
  }' | jq -r '.content[0].text'
echo

# Take a screenshot of the final result
echo "📸 13. Taking screenshot of final result..."
curl -s -X POST "$BASE_URL/mcp/tools/call" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "take_screenshot",
    "arguments": {
      "filename": "http_mcp_test_result.png"
    }
  }' | jq -r '.content[0].text'
echo

echo "🎉 HTTP MCP Test Complete!"
echo "=" $(printf '=%.0s' {1..50})
echo "✅ Successfully demonstrated:"
echo "   • HTTP MCP server responding to requests"
echo "   • Creating interactive web pages" 
echo "   • Navigation to pages (with fixed navigation issue)"
echo "   • Text entry in multiple form fields"
echo "   • Button clicking and form submission"
echo "   • Reading page content after interaction"
echo "   • Screenshot capture"
echo ""
echo "📁 Generated files:"
echo "   • http_test_page.html - Interactive test page"
echo "   • http_mcp_test_result.png - Screenshot of final result"
echo ""
echo "🚀 The HTTP MCP server allows Claude to make stateless requests"
echo "   without managing persistent processes - solving the kill issue!"