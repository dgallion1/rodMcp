# Using RodMCP with Claude

Once RodMCP is installed and configured, here's how to use it with Claude for web development tasks.

## Quick Test

Ask Claude:
```
What web development tools do you have available?
```

Claude should respond with the 6 RodMCP tools.

## Example Use Cases

### 1. Create a Simple Webpage

```
Create an HTML page called "portfolio.html" with:
- A dark theme
- My name "Alex Smith" as the main heading
- Three sections: About, Skills, Contact
- Some CSS styling to make it look professional
- Take a screenshot when done
```

### 2. Interactive Dashboard

```
Create an interactive dashboard with:
- Real-time clock
- Color-changing buttons
- A form with validation
- Charts or graphs using JavaScript
- Start a live preview server so I can see it
```

### 3. Test and Debug Existing Pages

```
Navigate to my website at http://localhost:3000 and:
- Take a screenshot
- Test the contact form by filling it out
- Check if all buttons work
- Report any JavaScript errors
```

### 4. Rapid Prototyping

```
Create a prototype for a todo app:
- Add/remove todo items
- Mark items as complete
- Local storage persistence
- Clean modern UI
- Test all functionality
```

### 5. Educational Content

```
Create an interactive learning page about:
- CSS Flexbox with live examples
- JavaScript event handling demonstrations
- Responsive design patterns
- Include hands-on exercises
```

## Tool-Specific Examples

### create_page
```
Create a landing page for a coffee shop with:
- Hero section with background image
- Menu grid layout
- Contact information footer
- Warm color scheme (#8B4513, #D2B48C, #F5DEB3)
```

### navigate_page + execute_script
```
Open Google.com and:
- Search for "web development best practices"
- Click on the first result
- Scroll down to read the content
- Take notes on key points
```

### take_screenshot
```
Navigate to my project page and take screenshots of:
- The main header
- The projects gallery
- The contact form
- Save them with descriptive filenames
```

### set_browser_visibility
```
Show me the browser window while you create and test this webpage 
so I can see exactly how the animations work.
```

```
Switch to headless mode for faster execution while running 
the automated test suite.
```

```
I want to see the browser while you debug the CSS layout issues,
then switch back to headless when you're done.
```

### live_preview
```
Create a multi-page website with:
- index.html (home page)
- about.html (about page)
- contact.html (contact form)
- shared styles.css
- Start a live preview server so I can navigate between pages
```

## Advanced Workflows

### A/B Testing
```
Create two versions of the same landing page:
- Version A: Blue color scheme, large hero image
- Version B: Green color scheme, video background
- Take screenshots of both
- Help me decide which looks better
```

### Responsive Design Testing
```
Create a responsive website and test it at different screen sizes:
- Desktop: 1920x1080
- Tablet: 768x1024
- Mobile: 375x667
- Take screenshots at each size
```

### Performance Testing
```
Create a heavy webpage with many images and scripts, then:
- Measure load times
- Test JavaScript performance
- Optimize and compare results
- Document the improvements
```

## Integration Examples

### With External APIs
```
Create a weather dashboard that:
- Uses a public weather API
- Shows current weather and forecast
- Has a search function for different cities
- Updates automatically every 30 seconds
- Style it with a modern glassmorphism design
```

### With Local Files
```
I have a data.json file with product information:
- Create an e-commerce product listing page
- Load the JSON data with JavaScript
- Add search and filter functionality
- Make it mobile-responsive
- Include a shopping cart feature
```

### Form Processing
```
Create a comprehensive contact form with:
- Input validation (email, phone, required fields)
- Success/error messages
- File upload capability
- Form data preview before submission
- Save form data to localStorage for testing
```

## Debugging Workflows

### JavaScript Debugging
```
I have a JavaScript error on my page. Can you:
- Navigate to the page
- Open browser console
- Identify the error
- Fix the JavaScript code
- Test the fix
```

### CSS Layout Issues
```
My CSS grid layout isn't working correctly:
- Take a screenshot of the current state
- Inspect the CSS rules
- Identify the problem
- Fix the layout
- Show before/after comparison
```

### Cross-browser Testing
```
Test my website for browser compatibility:
- Check if all features work in the browser
- Identify any layout issues
- Suggest fallbacks for unsupported features
- Create a compatibility report
```

## Tips for Best Results

### 1. Be Specific
```
❌ "Make a website"
✅ "Create a portfolio website with a dark theme, navigation menu, project gallery, and contact form"
```

### 2. Request Screenshots
```
✅ "Create the page and take a screenshot so I can see the result"
```

### 3. Test Functionality
```
✅ "After creating the form, test it by filling out all fields and submitting"
```

### 4. Iterate and Improve
```
"The page looks good, but can you make the header more prominent and add some animation to the buttons?"
```

### 5. Use Live Preview for Complex Sites
```
"Create a multi-page site and start a live preview server so I can navigate between pages"
```

## Troubleshooting

If Claude can't use the tools:
1. Check that RodMCP is properly installed
2. Verify the MCP configuration
3. Restart Claude Desktop/CLI
4. Ask Claude: "Can you see the rodmcp tools?"

For development mode (visible browser):
```
"Use the browser in non-headless mode so I can watch what's happening"
```

---

🎯 **Remember**: Claude can create, test, and debug web pages programmatically. It's like having a full-stack developer who can build and interact with web applications in real-time!