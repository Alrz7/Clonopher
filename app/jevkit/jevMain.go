package jevkit

import (
	"bytes"
	"clonopher/internal/loggy"
	"encoding/json"
	"io"
	"net/http"
)

func (req *Request) Send() (*Response, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequest(
		http.MethodPost,
		req.Provider.Url,
		bytes.NewReader(b),
	)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

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

	var out Response
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}

	return &out, nil
}