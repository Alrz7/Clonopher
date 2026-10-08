package jevkit

import (
	"bytes"
	"clonopher/internal/helpers"
	"clonopher/internal/loggy"
	"encoding/json"
	"io"
	"net/http"
)

func (r *Request) SendTo(model JevModel) (*Response, error) {
	r.Model = model
	err := helpers.DoesSupportModel(r.Provider.Name, r.Provider.Models, string(r.Model))
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequest(
		http.MethodPost,
		r.Provider.Url,
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
