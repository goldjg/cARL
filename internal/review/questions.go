package review

import "github.com/goldjg/carl/internal/semantic"

// DefaultQuestions returns cARL-owned bounded semantic questions for review.
func DefaultQuestions() []semantic.Question {
	scoreThreshold := 0.7
	return []semantic.Question{
		{
			ID:          "security_sensitive",
			Kind:        semantic.BooleanQuestion,
			Description: "Does this change appear security-sensitive?",
		},
		{
			ID:          "scope_expansion",
			Kind:        semantic.BooleanQuestion,
			Description: "Does this change appear to expand approved scope or modify unrelated surfaces?",
		},
		{
			ID:          "architectural_change",
			Kind:        semantic.BooleanQuestion,
			Description: "Does this change appear to introduce or alter architecture-level structure or boundaries?",
		},
		{
			ID:          "authentication_or_authorisation_change",
			Kind:        semantic.BooleanQuestion,
			Description: "Does this change appear to affect authentication or authorization behavior?",
		},
		{
			ID:          "trust_boundary_change",
			Kind:        semantic.BooleanQuestion,
			Description: "Does this change appear to alter a trust boundary or external data flow?",
		},
		{
			ID:          "security_control_reduction",
			Kind:        semantic.BooleanQuestion,
			Description: "Does this change appear to reduce a security control or validation gate?",
		},
		{
			ID:          "breaking_change",
			Kind:        semantic.BooleanQuestion,
			Description: "Does this change appear to alter a documented public contract or compatibility boundary?",
		},
		{
			ID:          "new_external_dependency",
			Kind:        semantic.BooleanQuestion,
			Description: "Does this change appear to add a new external runtime, build, or CI dependency?",
		},
		{
			ID:          "human_review_warranted",
			Kind:        semantic.ScoreQuestion,
			Description: "How strongly does the change warrant human review beyond routine automated checks?",
			Threshold:   &scoreThreshold,
		},
	}
}
