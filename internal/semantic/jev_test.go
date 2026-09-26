package semantic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTypeSafeJevEvaluatorNormalizesAnswers(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"answers": map[string]any{
				"security_sensitive": map[string]any{"value": true, "confidence": 0.93},
			},
		})
	}))
	defer server.Close()

	eval := NewTypeSafeJevEvaluator(JEVConfig{Endpoint: server.URL}, "test-secret-key", server.Client())
	result, err := eval.Evaluate(context.Background(), EvaluationState{}, []Question{{ID: "security_sensitive", Kind: BooleanQuestion}})
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if gotAuth != "Bearer test-secret-key" {
		t.Fatalf("authorization header not set")
	}
	if len(result.Signals) != 1 || result.Signals[0].ID != "security_sensitive" || result.Signals[0].Source != "typesafe-jev" {
		t.Fatalf("unexpected signals: %+v", result.Signals)
	}
	if result.Signals[0].Value != true {
		t.Fatalf("unexpected signal value: %#v", result.Signals[0].Value)
	}
}

func TestTypeSafeJevEvaluatorProviderErrorsAreSanitized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "token test-secret-key rejected", http.StatusUnauthorized)
	}))
	defer server.Close()

	eval := NewTypeSafeJevEvaluator(JEVConfig{Endpoint: server.URL}, "test-secret-key", server.Client())
	_, err := eval.Evaluate(context.Background(), EvaluationState{}, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "test-secret-key") || strings.Contains(err.Error(), "token") {
		t.Fatalf("provider error leaked secret-bearing response body: %v", err)
	}
}

func TestTypeSafeJevEvaluatorMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{not-json"))
	}))
	defer server.Close()

	eval := NewTypeSafeJevEvaluator(JEVConfig{Endpoint: server.URL}, "test-secret-key", server.Client())
	_, err := eval.Evaluate(context.Background(), EvaluationState{}, nil)
	if err == nil || !strings.Contains(err.Error(), "valid JSON") {
		t.Fatalf("expected malformed response error, got %v", err)
	}
}
