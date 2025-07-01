package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// HeadlessStrategy implements scraping using headless Chrome browser
type HeadlessStrategy struct{}

// NewHeadlessStrategy creates a new headless browser strategy
func NewHeadlessStrategy() *HeadlessStrategy {
	return &HeadlessStrategy{}
}

// Execute performs headless browser-based scraping
func (s *HeadlessStrategy) Execute(ctx context.Context, urlStr string, config *Config) (*ScrapedResult, error) {
	// Create a new chromedp context from the parent context with timeout
	taskCtx, cancel := context.WithTimeout(ctx, config.RequestTimeout)
	defer cancel()

	// Create chromedp context with options that ignore SSL errors
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// Create a temporary, disposable user profile for this scrape
		chromedp.Flag("user-data-dir", os.TempDir()+"/go-scraper-profile"),
		// --- Best Practice Flags for a Clean Run ---
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		// SSL/security flags for test sites
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("ignore-ssl-errors", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(taskCtx, opts...)
	defer cancel()

	chromeCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var title string
	var body string
	var nextURL string

	// Define the sequence of actions the browser will perform
	err := chromedp.Run(chromeCtx,
		// Navigate to the URL
		chromedp.Navigate(urlStr),

		// Wait for the page to load (wait for body to be ready)
		chromedp.WaitReady("body", chromedp.ByQuery),

		// Wait a bit for JavaScript to execute
		chromedp.Sleep(3*time.Second),

		// Extract the page title
		chromedp.Title(&title),

		// Extract the full HTML body
		chromedp.OuterHTML("html", &body),
	)

	if err != nil {
		return nil, NewScraperError(urlStr, "Headless execution failed", err)
	}

	// Try to find the "Next" button for pagination (optional, non-blocking)
	// First check if the element exists to avoid infinite loops
	var elementExists bool
	_ = chromedp.Run(chromeCtx,
		chromedp.Evaluate(`document.querySelector('li.next a') !== null`, &elementExists),
	)

	// Only try to get the href if the element exists
	if elementExists {
		_ = chromedp.Run(chromeCtx,
			chromedp.AttributeValue("li.next a", "href", &nextURL, nil, chromedp.ByQuery),
		)
	}

	// If we found a next URL, make it absolute
	if nextURL != "" {
		baseURL, _ := url.Parse(urlStr)
		if nextURLRef, err := url.Parse(nextURL); err == nil {
			nextURL = baseURL.ResolveReference(nextURLRef).String()
		}
	}

	// Extract a meaningful title from the content if the page title is generic
	if title == "" || strings.Contains(strings.ToLower(title), "quotes") {
		title = s.extractTitleFromContent(body)
	}

	return &ScrapedResult{
		Title:      title,
		Body:       body,
		StatusCode: 200, // Chromedp doesn't easily expose status, 200 is safe on success
		NextURL:    nextURL,
	}, nil
}

// extractTitleFromContent extracts a meaningful title from the HTML content
func (s *HeadlessStrategy) extractTitleFromContent(html string) string {
	// For quotes.toscrape.com, try to extract the first quote as title
	// This is a simple extraction - in a real implementation, you might use a proper HTML parser

	// Look for quote text in the content
	if strings.Contains(html, "class=\"text\"") {
		// Simple extraction - find the first quote text
		start := strings.Index(html, "class=\"text\"")
		if start != -1 {
			// Find the opening and closing tags
			openTag := strings.Index(html[start:], ">")
			if openTag != -1 {
				contentStart := start + openTag + 1
				closeTag := strings.Index(html[contentStart:], "</div>")
				if closeTag != -1 {
					quote := html[contentStart : contentStart+closeTag]
					// Clean up the quote (remove HTML entities, trim whitespace)
					quote = strings.TrimSpace(quote)
					if len(quote) > 0 {
						// Limit length for title
						if len(quote) > 100 {
							quote = quote[:97] + "..."
						}
						return fmt.Sprintf("Quotes - %s", quote)
					}
				}
			}
		}
	}

	return "JavaScript-rendered page"
}
