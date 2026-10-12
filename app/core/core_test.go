package core

import (
	"clonopher/app/burrow"
	"clonopher/app/jevkit"
	"clonopher/app/services/composer"
	"testing"
)

func TestServePayload(t *testing.T) {
	ollamaTev := &Options{
		Provider: &jevkit.Provider{Name: "ollama",
			Url:    "http://127.0.0.1:11434/v1/systemone",
			Models: []string{"tev1:0.8b"}},
		Model: "tev1:0.8b",
	}

	Payload := Payload{
		Name: "golangGinRateLimiter.md",
		Head: "/home/alireza/GoPrj/Clonopher/tmp/testing",
		Src:  "/home/alireza/GoPrj/Clonopher/tmp/golangGinRateLimiter.md",
		Info: &composer.FileInfo{
			Description: "A Go-based rate-limiting middleware for Gin APIs, covering request throttling, HTTP 429 responses, and per-route rate limits.",
		},
		Stage:   burrow.New("/home/alireza/GoPrj/Clonopher/tmp/testing"),
		Options: ollamaTev,
	}
	err := Payload.Serve()
	if err != nil {
		t.Log(err)
	}
}
