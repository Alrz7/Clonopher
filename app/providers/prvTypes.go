package providers

type Provider struct {
	Name        string   `koanf:"name"`
	Url         string   `koanf:"url"`
	Models      []string `koanf:"models"`
	ApiKey      string   `koanf:"api_key,omitempty"`
	ActiveModel string   `koanf:"active_model,omitempty"`
	// AdditionalHeaders map[string]string `koanf:"headers"`
}
type JevQaTypes string

const (
	JevQAChoice JevQaTypes = "choice"
	JevQAScore  JevQaTypes = "score"
	JevQANoul   JevQaTypes = "noul"
)

type JevQuestion struct {
	Type         JevQaTypes        `json:"type"`
	Instructions string            `json:"instructions"`
	Criteria     map[string]string `json:"criteria,omitempty"`
}

type JevAnswer struct {
	Type          JevQaTypes         `json:"type"`
	Choice        string             `json:"choice,omitempty"`
	Noul          *float64           `json:"noul,omitempty"`
	Score         *float64           `json:"score,omitempty"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
	Confidence    float64            `json:"confidence,omitempty"`
}

type JevRequest struct {
	Provider  *Provider              `json:"-"`
	Model     string                 `json:"model"`
	State     string                 `json:"state"`
	Questions map[string]JevQuestion `json:"questions"`
}

type JevResponse struct {
	Model   string               `json:"model"`
	Answers map[string]JevAnswer `json:"answers"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type LlmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LlmFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type LlmTool struct {
	Type     string      `json:"type"`
	Function LlmFunction `json:"function"`
}

type LlmRequest struct {
	Provider   *Provider    `json:"-"`
	Model      string       `json:"model"`
	Stream     bool         `json:"stream"`
	Messages   []LlmMessage `json:"messages"`
	Tools      []LlmTool    `json:"tools,omitempty"`
	ToolChoice string       `json:"tool_choice,omitempty"`
}

func (r JevResponse) IsValid() bool {
	return r.Model != "" && len(r.Answers) > 0
}
