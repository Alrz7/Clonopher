package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type concurrentUploadPayload struct {
	FileName    string `json:"fileName"`
	Head        string `json:"head"`
	Content     string `json:"content"`
	Description string `json:"description"`
	Policy      string `json:"policy"`
}

type concurrentUploadResponse struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type concurrentUploadResult struct {
	statusCode int
	body       string
	duration   time.Duration
	err        error
}

// TestConcurrentPayloadUpload sends four requests concurrently to an already-running API server.
// Set CLONOPHER_API_URL to override the default endpoint.
func TestConcurrentPayloadUpload(t *testing.T) {
	endpoint := os.Getenv("CLONOPHER_API_URL")
	if endpoint == "" {
		endpoint = "http://localhost:6520/api/v1/payloads/text"
	}

	cases := []struct {
		name    string
		payload concurrentUploadPayload
	}{
		{
			name: "Go rate limiter",
			payload: concurrentUploadPayload{
				FileName:    "go-rate-limiter-test.md",
				Head:        "testing",
				Content:     "Go HTTP rate limiter notes using token bucket middleware for Gin handlers and concurrency-safe counters.",
				Description: "Study notes about implementing a rate limiter in Go web services.",
				Policy:      "notify",
			},
		},
		{
			name: "Gin middleware rate limit",
			payload: concurrentUploadPayload{
				FileName:    "gin-middleware-rate-limit-test.md",
				Head:        "testing",
				Content:     "Building a Go Gin middleware for HTTP rate limiting with a token bucket, per-client request counters, and tests.",
				Description: "Go backend notes about Gin middleware and rate limiting behavior.",
				Policy:      "notify",
			},
		},
		{
			name: "Lentil soup recipe",
			payload: concurrentUploadPayload{
				FileName:    "lentil-soup-recipe-test.md",
				Head:        "testing",
				Content:     "Recipe for lentil soup with onions, carrots, cumin, lemon, and simple cooking steps.",
				Description: "A home cooking recipe with ingredients and preparation notes.",
				Policy:      "notify",
			},
		},
		{
			name: "Flight simulator controls",
			payload: concurrentUploadPayload{
				FileName:    "flight-sim-controls-test.txt",
				Head:        "testing",
				Content:     "Flight simulator control setup for helicopter yaw, collective pitch, cyclic controls, and joystick sensitivity.",
				Description: "Gaming notes about configuring helicopter controls in a flight simulator.",
				Policy:      "notify",
			},
		},
	}

	// Encode payloads before releasing the goroutines, so JSON encoding time is not
	// included in each request's measured duration.
	requestBodies := make([][]byte, len(cases))
	for i := range cases {
		body, err := json.Marshal(cases[i].payload)
		if err != nil {
			t.Fatalf("encode payload %q: %v", cases[i].name, err)
		}
		requestBodies[i] = body
	}

	client := &http.Client{Timeout: 630 * time.Second}
	results := make([]concurrentUploadResult, len(cases))
	start := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(len(cases))
	for i := range cases {
		go func(i int) {
			defer wg.Done()
			<-start // Release all requests together to exercise concurrent handling.

			requestStarted := time.Now()
			defer func() {
				results[i].duration = time.Since(requestStarted)
			}()

			req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(requestBodies[i]))
			if err != nil {
				results[i].err = err
				return
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				results[i].err = err
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			results[i].statusCode = resp.StatusCode
			results[i].body = string(body)
			if err != nil {
				results[i].err = err
			}
		}(i)
	}

	t.Logf("Sending %d concurrent requests to %s", len(cases), endpoint)
	startedAt := time.Now()
	close(start)
	wg.Wait()
	totalDuration := time.Since(startedAt)

	for i, result := range results {
		if result.err != nil {
			t.Errorf("request %d (%s) failed after %s: %v", i+1, cases[i].name, result.duration.Round(time.Millisecond), result.err)
			continue
		}

		if result.statusCode < http.StatusOK || result.statusCode >= http.StatusMultipleChoices {
			t.Errorf("request %d (%s) returned HTTP %d", i+1, cases[i].name, result.statusCode)
		}

		var response concurrentUploadResponse
		if err := json.Unmarshal([]byte(result.body), &response); err == nil {
			t.Logf("request %d: %-26s status=%d duration=%s path=%s message=%q",
				i+1,
				cases[i].name,
				result.statusCode,
				result.duration.Round(time.Millisecond),
				response.Path,
				response.Message,
			)
		} else {
			t.Logf("request %d: %-26s status=%d duration=%s body=%s",
				i+1,
				cases[i].name,
				result.statusCode,
				result.duration.Round(time.Millisecond),
				strings.TrimSpace(result.body),
			)
		}
	}

	t.Logf("Total wall time for all %d requests: %s", len(cases), totalDuration.Round(time.Millisecond))
}
