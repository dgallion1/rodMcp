# Screen Scraping Tool - LLM Usage Examples

## Basic Single Item Extraction
```json
{
  "url": "https://example.com/product/123",
  "selectors": {
    "title": "h1.product-title",
    "price": ".price-current",
    "description": ".product-description p",
    "image": "img.hero-image",
    "rating": "[data-rating]"
  }
}
```

## Multiple Items (Product List)
```json
{
  "url": "https://store.com/products",
  "extract_type": "multiple",
  "container_selector": ".product-card",
  "selectors": {
    "name": ".product-name",
    "price": ".price",
    "link": "a[href]",
    "image": "img[src]"
  }
}
```

## Dynamic Content with Wait
```json
{
  "url": "https://spa-app.com/data",
  "selectors": {
    "content": ".ajax-loaded-content"
  },
  "wait_for": ".loading-complete",
  "wait_timeout": 15
}
```

## Lazy Loading with Scroll
```json
{
  "url": "https://infinite-scroll.com",
  "extract_type": "multiple",
  "container_selector": ".item",
  "selectors": {
    "title": "h3",
    "image": "img[data-src]"
  },
  "scroll_to_load": true
}
```

## Custom JavaScript + Scraping
```json
{
  "url": "https://complex-site.com",
  "selectors": {
    "data": ".hidden-data"
  },
  "custom_script": "document.querySelector('.load-more').click(); document.querySelector('.show-hidden').click();"
}
```

## CSS Selector Reference
- `#id` - Element with ID
- `.class` - Elements with class
- `[attribute]` - Elements with attribute
- `[href^="https"]` - Links starting with https
- `:nth-child(2)` - Second child element
- `div > p` - Direct child paragraphs
- `h1, h2, h3` - Multiple selectors