package review

import (
	"testing"

	"github.com/goldjg/carl/internal/semantic"
)

func TestApplyPolicyMapsSignalsToReviewDecision(t *testing.T) {
	conf := 0.93
	result := ApplyPolicy(&semantic.EvaluationResult{Signals: []semantic.Signal{
		{ID: "security_sensitive", Value: true, Confidence: &conf, Source: "test"},
		{ID: "trust_boundary_change", Value: true, Source: "test"},
	}}, []semantic.ChangedFile{{Path: "internal/review/review.go"}})

	if result.Decision != DecisionReviewRequired {
		t.Fatalf("unexpected decision: %s", result.Decision)
	}
	if len(result.Findings) != 2 {
		t.Fatalf("expected two findings, got %+v", result.Findings)
	}
}

func TestApplyPolicyIgnoresFalseSignals(t *testing.T) {
	result := ApplyPolicy(&semantic.EvaluationResult{Signals: []semantic.Signal{
		{ID: "security_sensitive", Value: false, Source: "test"},
	}}, nil)
	if result.Decision != DecisionPass || len(result.Findings) != 0 {
		t.Fatalf("false signal should not create finding: %+v", result)
	}
}
