package providers

import (
	"bytes"
	"clonopher/internal/loggy"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
)

func (req *JevRequest) Send(ctx context.Context) (*JevResponse, error) {
	err := DoesSupportModel(req.Provider.Url, req.Provider.Models, string(req.Model))
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		req.Provider.Url,
		bytes.NewReader(b),
	)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if req.Provider.ApiKey != "" {
		httpReq.Header.Set("Authorization", req.Provider.ApiKey)
	}

	// loggy.Info("req:")
	// loggy.Info(req)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// loggy.Info("resp:")
	// loggy.Info(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, loggy.Sayf("HTTP %d: %s\n", resp.StatusCode, body)
	}

	var out JevResponse

	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}

	if !out.IsValid() {
		var retry struct {
			Result JevResponse `json:"result"`
		}
		if err := json.Unmarshal(body, &retry); err != nil {
			return nil, err
		}
		out = retry.Result
	}
	if !out.IsValid() {
		return nil, loggy.Sayf("jev Response was not valid, Response Body: <<%v>>", body)
	}

	return &out, nil
}

func DoesSupportModel(provider string, supportedModels []string, targetModel string) error {
	if !slices.Contains(supportedModels, targetModel) {
		return fmt.Errorf("provider [%v] does Not contain model [%v] inside config!", provider, targetModel)
	}
	return nil
}
