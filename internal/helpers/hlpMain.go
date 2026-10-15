package helpers

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

func ToYAML(data map[string]any) (string, error) {
	b, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func IsNotOneOf[T comparable](obj T, src ...T) bool {
	return !slices.Contains(src, obj)
}

func ExtractContent(r string) (string, error) {
	var data map[string]any

	if err := json.Unmarshal([]byte(r), &data); err != nil {
		return "", fmt.Errorf("decode AI response: %w", err)
	}

	choices, ok := data["choices"].([]any)
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("AI response has no choices")
	}

	choice, ok := choices[0].(map[string]any)
	if !ok {
		return "", fmt.Errorf("invalid choice format")
	}

	message, ok := choice["message"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("AI response has no valid message")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("AI response has no text content")
	}

	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("AI response content is empty")
	}

	return content, nil
}
