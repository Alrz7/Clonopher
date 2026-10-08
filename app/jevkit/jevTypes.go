package jevkit

type Provider struct {
	Name   string   `koanf:"name"`
	Url    string   `koanf:"url"`
	Models []string `koanf:"models"`
}
type QaTypes string
type JevModel string

type Question struct {
	Type         QaTypes           `json:"type"`
	Instructions string            `json:"instructions"`
	Criteria     map[string]string `json:"criteria"`
}

type Request struct {
	Provider  *Provider           `json:"-"`
	Model     JevModel              `json:"model"`
	State     string              `json:"state"`
	Questions map[string]Question `json:"questions"`
}

type Answer struct {
	Type          string             `json:"type"`
	Choice        string             `json:"choice"`
	Probabilities map[string]float64 `json:"probabilities"`
	Confidence    float64            `json:"confidence"`
}

type Response struct {
	Model   JevModel          `json:"model"`
	Answers map[string]Answer `json:"answers"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}
