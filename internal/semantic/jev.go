package semantic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const jevSource = "typesafe-jev"

// HTTPDoer is the subset of http.Client used by TypeSafeJevEvaluator.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// TypeSafeJevEvaluator adapts TypeSafe AI JEV responses into cARL semantic
// signals. It never maps signals to cARL policy outcomes.
type TypeSafeJevEvaluator struct {
	cfg    JEVConfig
	apiKey string
	client HTTPDoer
}

// NewTypeSafeJevEvaluator creates a TypeSafe JEV evaluator.
func NewTypeSafeJevEvaluator(cfg JEVConfig, apiKey string, client HTTPDoer) *TypeSafeJevEvaluator {
	if client == nil {
		timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		client = &http.Client{Timeout: timeout}
	}
	return &TypeSafeJevEvaluator{cfg: cfg, apiKey: apiKey, client: client}
}

type jevRequest struct {
	Model     string          `json:"model,omitempty"`
	State     EvaluationState `json:"state"`
	Questions []Question      `json:"questions"`
}

type jevResponse struct {
	Signals []Signal `json:"signals"`
	Answers map[string]struct {
		Value      any            `json:"value"`
		Confidence *float64       `json:"confidence,omitempty"`
		Metadata   map[string]any `json:"metadata,omitempty"`
	} `json:"answers"`
}

// Evaluate sends a bounded semantic payload to TypeSafe JEV and normalizes the
// response into provider-independent cARL signals.
func (e *TypeSafeJevEvaluator) Evaluate(ctx context.Context, state EvaluationState, questions []Question) (*EvaluationResult, error) {
	if e.cfg.Endpoint == "" {
		return nil, &ProviderError{Code: "missing_endpoint", Message: "JEV endpoint is not configured"}
	}
	if e.apiKey == "" {
		return nil, &ProviderError{Code: "missing_api_key", Message: "JEV API key environment variable is not set"}
	}

	body, err := json.Marshal(jevRequest{Model: e.cfg.Model, State: state, Questions: questions})
	if err != nil {
		return nil, fmt.Errorf("marshal JEV request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, &ProviderError{Code: "invalid_endpoint", Message: "JEV endpoint is invalid"}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, &ProviderError{Code: "network_failure", Message: "JEV request failed"}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, statusProviderError(resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, &ProviderError{Code: "provider_response_error", Message: "JEV response could not be read"}
	}
	var decoded jevResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, &ProviderError{Code: "malformed_response", Message: "JEV response was not valid JSON"}
	}
	return normalizeJEVResponse(decoded), nil
}

func normalizeJEVResponse(resp jevResponse) *EvaluationResult {
	result := &EvaluationResult{}
	for _, sig := range resp.Signals {
		sig.Source = jevSource
		result.Signals = append(result.Signals, sig)
	}
	for id, answer := range resp.Answers {
		result.Signals = append(result.Signals, Signal{
			ID:         id,
			Value:      answer.Value,
			Confidence: answer.Confidence,
			Source:     jevSource,
			Metadata:   answer.Metadata,
		})
	}
	return result
}

func statusProviderError(status int) *ProviderError {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return &ProviderError{Code: "invalid_api_key", Message: "JEV rejected the configured API key"}
	case http.StatusTooManyRequests:
		return &ProviderError{Code: "rate_limited", Message: "JEV rate limit exceeded"}
	case http.StatusBadRequest, http.StatusNotFound, http.StatusUnprocessableEntity:
		return &ProviderError{Code: "unsupported_request", Message: "JEV rejected the evaluation request"}
	default:
		if status >= 500 {
			return &ProviderError{Code: "provider_outage", Message: "JEV service is unavailable"}
		}
		return &ProviderError{Code: "provider_error", Message: "JEV request failed"}
	}
}
