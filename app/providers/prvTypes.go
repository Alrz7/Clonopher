package providers

type Provider struct {
	Url         string   `koanf:"url"`
	Models      []string `koanf:"models"`
	ApiKey      string   `koanf:"api_key,omitempty"`
	ActiveModel string   `koanf:"active_model,omitempty"`
	// AdditionalHeaders map[string]string `koanf:"headers"`
}
type JevQaTypes string

type JevQuestion struct {
	Type         JevQaTypes        `json:"type"`
	Instructions string            `json:"instructions"`
	Criteria     map[string]string `json:"criteria"`
}

type LlmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type JevRequest struct {
	Provider  *Provider              `json:"-"`
	Model     string               `json:"model"`
	State     string                 `json:"state"`
	Questions map[string]JevQuestion `json:"questions"`
}

type LlmRequest struct {
	Provider *Provider             `json:"-"`
	Model    string              `json:"model"`
	Stream   bool                  `json:"stream"`
	Messages map[string]LlmMessage `json:"messages"`
}

type JevAnswer struct {
	Type          string             `json:"type"`
	Choice        string             `json:"choice"`
	Probabilities map[string]float64 `json:"probabilities"`
	Confidence    float64            `json:"confidence"`
}

type JevResponse struct {
	Model   string             `json:"model"`
	Answers map[string]JevAnswer `json:"answers"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type LlmResponse string

func (r JevResponse) IsValid() bool {
	return r.Model != "" && len(r.Answers) > 0
}
