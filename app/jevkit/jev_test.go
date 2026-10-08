package jevkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestJev(t *testing.T) {
	payload := Request{
		Model: "tev1:0.8b",
		State: "golangGinMiddlewares.md - Gin middleware for logging",
		Questions: map[string]Question{
			"folder": {
				Type:         "choice",
				Instructions: "Choose the most semantically appropriate folder for the supplied state. Prefer the most specific existing concept. Do not choose a parent folder when a more specific child matches.",
				Criteria: map[string]string{
					"programming/golang/web/gin/middleware": `
This folder contains middleware implementations and middleware-related patterns specifically built for the Gin web framework.
Typical contents include authentication middleware, rate limiting, logging, CORS, request tracing, recovery handlers, middleware chaining, and Gin Context based request processing.
Prefer this folder when the file is fundamentally about creating or configuring Gin middleware rather than generic HTTP middleware.
`,

					"programming/golang/web/gin/routing": `
This folder focuses on Gin's routing layer and route organization.
Typical contents include route groups, path parameters, handlers, route registration, HTTP methods, nested routers, and route-level configuration.
Middleware may appear here only when routing itself is the main subject.
`,

					"programming/golang/web/gin/api": `
This folder contains Gin-based API implementation patterns at the application level.
Typical contents include controllers, handlers, request/response models, API endpoints, validation, JSON responses, and API structure.
Use this when the file is about building an API with Gin rather than a specific middleware mechanism.
`,

					"programming/golang/web/net-http/middleware": `
This folder contains generic HTTP middleware built around Go's standard net/http package.
Typical contents include http.Handler wrappers, request logging, authentication, rate limiting, timeout handling, recovery, headers, and middleware composition that is not tied to a specific web framework.
Use this when the concepts should work independently of Gin, Fiber, Echo, or another framework.
`,

					"programming/golang/security/rate-limiting": `
This folder focuses on rate limiting as a security and abuse-prevention mechanism.
Typical contents include token bucket algorithms, leaky bucket algorithms, request quotas, throttling policies, IP-based limits, user-based limits, distributed rate limiting, and protection against excessive requests.
The implementation may use a web framework, but the main concept is rate limiting itself.
`,

					"programming/backend/api/middleware": `
This folder contains general backend API middleware concepts that are not strongly tied to Go or a specific framework.
Typical contents include authentication layers, authorization, logging, tracing, CORS, request validation, throttling, error handling, and request lifecycle processing.
Use this for framework-agnostic backend architecture discussions and implementations.
`,

					"programming/golang/concurrency": `
This folder focuses on Go's concurrency model and primitives.
Typical contents include goroutines, channels, mutexes, atomic operations, worker pools, synchronization, concurrent data structures, cancellation, and race-condition prevention.
A web-related file belongs here only when concurrency is the primary technical subject rather than merely an implementation detail.
`,
				},
			},
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("marshal:", err)
		return
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"http://127.0.0.1:11434/v1/systemone",
		bytes.NewReader(b),
	)
	if err != nil {
		fmt.Println("request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("ollama:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read:", err)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("HTTP %d: %s\n", resp.StatusCode, body)
		return
	}

	var out Response
	if err := json.Unmarshal(body, &out); err != nil {
		fmt.Println("unmarshal:", err)
		fmt.Println(string(body))
		return
	}

	answer, ok := out.Answers["folder"]
	if !ok {
		fmt.Println("answer not found")
		return
	}

	fmt.Println("Model:", out.Model)
	fmt.Println("Choice:", answer.Choice)
	fmt.Println("Confidence:", answer.Confidence)
	fmt.Println("Probabilities:")

	for option, probability := range answer.Probabilities {
		fmt.Printf("  %-22s %.4f\n", option, probability)
	}

	fmt.Printf(
		"\nTokens: input=%d output=%d\n",
		out.Usage.InputTokens,
		out.Usage.OutputTokens,
	)
}

func TestSend(t *testing.T) {
	payload := Request{
		Provider: &Provider{Name: "ollama",
			Url:    "http://127.0.0.1:11434/v1/systemone",
			Models: []string{"tev1:0.8b"}},
		Model: "tev1:0.8b",
		State: "golangGinMiddlewares.md - Gin middleware for logging",
		Questions: map[string]Question{
			"folder": {
				Type:         "choice",
				Instructions: "Which folder best matches the supplied state?",
				Criteria: map[string]string{
					"programming/golang/api":         "Go programming, backend development, Gin, middleware, web development.",
					"programming/golang/middlewares": "golang programming api-middlewares development, web apis",
					"programming/golang/gin/":        "golang-gin api documents...",
				},
			},
		},
	}
	resp, err := payload.SendTo("tev1:0.8b")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)
}
