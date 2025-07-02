package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// APIHandler handles HTTP API requests
type APIHandler struct {
	scraper *Scraper
	config  *Config
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(scraper *Scraper, config *Config) *APIHandler {
	return &APIHandler{
		scraper: scraper,
		config:  config,
	}
}

// ScrapeRequest represents a scraping request
type ScrapeRequest struct {
	URLs    []string `json:"urls"`
	SiteURL string   `json:"site_url,omitempty"`
}

// ScrapeResponse represents a scraping response
type ScrapeResponse struct {
	Results []ScrapedData `json:"results"`
	Metrics interface{}   `json:"metrics,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// HandleScrape handles scraping requests
func (h *APIHandler) HandleScrape(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var results []ScrapedData

	if req.SiteURL != "" {
		results = h.scraper.ScrapeSite(req.SiteURL)
	} else if len(req.URLs) > 0 {
		results = h.scraper.ScrapeURLs(req.URLs)
	} else {
		http.Error(w, "No URLs provided", http.StatusBadRequest)
		return
	}

	response := ScrapeResponse{
		Results: results,
	}

	if h.config.EnableMetrics {
		response.Metrics = h.scraper.metrics.GetMetrics()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleHealth handles health check requests
func (h *APIHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "2.0.0",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleMetrics handles metrics requests
func (h *APIHandler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	if !h.config.EnableMetrics {
		http.Error(w, "Metrics disabled", http.StatusServiceUnavailable)
		return
	}

	metrics := h.scraper.metrics.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// StartAPIServer starts the HTTP API server
func StartAPIServer(scraper *Scraper, config *Config, port int) error {
	handler := NewAPIHandler(scraper, config)

	// Set up routes
	http.HandleFunc("/scrape", handler.HandleScrape)
	http.HandleFunc("/health", handler.HandleHealth)
	http.HandleFunc("/metrics", handler.HandleMetrics)

	// Start server
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("🚀 Starting API server on port %d\n", port)
	fmt.Printf("📡 Endpoints:\n")
	fmt.Printf("   POST /scrape - Scrape URLs or site\n")
	fmt.Printf("   GET  /health - Health check\n")
	fmt.Printf("   GET  /metrics - Get metrics\n")

	return http.ListenAndServe(addr, nil)
}
