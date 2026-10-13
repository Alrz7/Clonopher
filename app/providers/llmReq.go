package providers

import (
	"bytes"
	"clonopher/internal/loggy"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func (r *LlmRequest) SendTo(ctx context.Context, model string) (*LlmResponse, error) {
	r.Model = model
	err := DoesSupportModel(r.Provider.Url, r.Provider.Models, r.Model)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		r.Provider.Url,
		bytes.NewReader(b),
	)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if r.Provider.ApiKey != "" {
		httpReq.Header.Set("Authorization", r.Provider.ApiKey)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, loggy.Sayf("HTTP %d: %s\n", resp.StatusCode, body)
	}

	loggy.Info(body)

	var out LlmResponse

	// if err := json.Unmarshal(body, &out); err != nil {
	// 	return nil, err
	// }

	// if !out.IsValid() {
	// 	var retry struct {
	// 		Result JevResponse `json:"result"`
	// 	}
	// 	if err := json.Unmarshal(body, &retry); err != nil {
	// 		return nil, err
	// 	}
	// 	out = retry.Result
	// }
	// if !out.IsValid() {
	// 	return nil, loggy.Sayf("jev Response was not valid, Response Body: <<%v>>", body)
	// }

	return &out, nil
}
