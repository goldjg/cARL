package review

import (
	"fmt"
	"sort"

	"github.com/goldjg/carl/internal/semantic"
)

// Decision is the provider-independent cARL review outcome.
type Decision string

const (
	DecisionPass           Decision = "pass"
	DecisionWarn           Decision = "warn"
	DecisionReviewRequired Decision = "review_required"
	DecisionFail           Decision = "fail"
)

// FindingSeverity identifies cARL policy severity.
type FindingSeverity string

const (
	SeverityInfo    FindingSeverity = "info"
	SeverityWarning FindingSeverity = "warning"
	SeverityReview  FindingSeverity = "review"
	SeverityError   FindingSeverity = "error"
)

// Finding is a cARL policy finding derived from semantic signals.
type Finding struct {
	ID          string          `json:"id"`
	Severity    FindingSeverity `json:"severity"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	SignalIDs   []string        `json:"signalIds"`
	Files       []string        `json:"files,omitempty"`
}

// Result is the normalized output of cARL review.
type Result struct {
	Signals     []semantic.Signal     `json:"signals"`
	Findings    []Finding             `json:"findings"`
	Decision    Decision              `json:"decision"`
	Diagnostics []semantic.Diagnostic `json:"diagnostics,omitempty"`
}

type signalPolicy struct {
	severity    FindingSeverity
	title       string
	description string
}

var reviewSignalPolicies = map[string]signalPolicy{
	"security_sensitive": {
		severity:    SeverityWarning,
		title:       "Security-sensitive change",
		description: "The semantic evaluator identified security-sensitive change characteristics. cARL treats this as advisory context for review.",
	},
	"scope_expansion": {
		severity:    SeverityWarning,
		title:       "Possible scope expansion",
		description: "The semantic evaluator identified changes that may exceed the expected task scope.",
	},
	"architectural_change": {
		severity:    SeverityWarning,
		title:       "Architectural change signal",
		description: "The semantic evaluator identified architecture-level change characteristics.",
	},
	"authentication_or_authorisation_change": {
		severity:    SeverityReview,
		title:       "Authentication or authorization change",
		description: "The semantic evaluator identified possible authentication or authorization behavior changes.",
	},
	"trust_boundary_change": {
		severity:    SeverityReview,
		title:       "Trust-boundary change",
		description: "The semantic evaluator identified a possible trust-boundary or external data-flow change.",
	},
	"security_control_reduction": {
		severity:    SeverityReview,
		title:       "Possible security control reduction",
		description: "The semantic evaluator identified a possible reduction in validation, authorization, or another security control.",
	},
	"breaking_change": {
		severity:    SeverityWarning,
		title:       "Possible breaking change",
		description: "The semantic evaluator identified a possible public-contract or compatibility change.",
	},
	"new_external_dependency": {
		severity:    SeverityWarning,
		title:       "New external dependency signal",
		description: "The semantic evaluator identified a possible new external dependency or integration.",
	},
	"human_review_warranted": {
		severity:    SeverityReview,
		title:       "Human review warranted",
		description: "The semantic evaluator scored this change above the cARL human-review threshold.",
	},
}

// ApplyPolicy maps semantic signals to cARL-owned findings and decisions.
func ApplyPolicy(eval *semantic.EvaluationResult, changedFiles []semantic.ChangedFile) Result {
	if eval == nil {
		return Result{Decision: DecisionPass}
	}
	result := Result{
		Signals:     append([]semantic.Signal(nil), eval.Signals...),
		Diagnostics: append([]semantic.Diagnostic(nil), eval.Diagnostics...),
		Decision:    DecisionPass,
	}
	files := make([]string, 0, len(changedFiles))
	for _, f := range changedFiles {
		files = append(files, f.Path)
	}
	sort.Strings(files)

	for _, sig := range eval.Signals {
		policy, ok := reviewSignalPolicies[sig.ID]
		if !ok || !signalTriggered(sig) {
			continue
		}
		result.Findings = append(result.Findings, Finding{
			ID:          "semantic_" + sig.ID,
			Severity:    policy.severity,
			Title:       policy.title,
			Description: findingDescription(policy.description, sig.Confidence),
			SignalIDs:   []string{sig.ID},
			Files:       files,
		})
		result.Decision = maxDecision(result.Decision, decisionForSeverity(policy.severity))
	}
	return result
}

func signalTriggered(sig semantic.Signal) bool {
	switch v := sig.Value.(type) {
	case bool:
		return v
	case float64:
		return v >= thresholdForSignal(sig.ID)
	case string:
		return v == "true" || v == "yes" || v == "high" || v == "review" || v == "review_required"
	default:
		return false
	}
}

func thresholdForSignal(id string) float64 {
	if id == "human_review_warranted" {
		return 0.7
	}
	return 1
}

func findingDescription(base string, confidence *float64) string {
	if confidence == nil {
		return base
	}
	return fmt.Sprintf("%s Confidence: %.2f.", base, *confidence)
}

func decisionForSeverity(sev FindingSeverity) Decision {
	switch sev {
	case SeverityError:
		return DecisionFail
	case SeverityReview:
		return DecisionReviewRequired
	case SeverityWarning:
		return DecisionWarn
	default:
		return DecisionPass
	}
}

func maxDecision(a, b Decision) Decision {
	if decisionRank(b) > decisionRank(a) {
		return b
	}
	return a
}

func decisionRank(d Decision) int {
	switch d {
	case DecisionFail:
		return 3
	case DecisionReviewRequired:
		return 2
	case DecisionWarn:
		return 1
	default:
		return 0
	}
}
