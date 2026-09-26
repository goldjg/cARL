// Package semantic defines provider-neutral semantic evaluation models.
package semantic

import "context"

// Evaluator evaluates bounded semantic questions against collected repository
// state. Implementations provide signals only; cARL maps signals to policy.
type Evaluator interface {
	Evaluate(ctx context.Context, state EvaluationState, questions []Question) (*EvaluationResult, error)
}

// EvaluationState is the bounded, auditable state sent to an evaluator.
type EvaluationState struct {
	Repository     RepositoryMetadata `json:"repository"`
	Review         ReviewState        `json:"review"`
	Diff           DiffState          `json:"diff"`
	Context        []ContextItem      `json:"context,omitempty"`
	Task           string             `json:"task,omitempty"`
	DataCategories []DataCategory     `json:"dataCategories"`
}

// RepositoryMetadata describes the repository without including credentials.
type RepositoryMetadata struct {
	Name          string `json:"name"`
	Root          string `json:"root,omitempty"`
	CurrentBranch string `json:"currentBranch,omitempty"`
	RemoteOrigin  string `json:"remoteOrigin,omitempty"`
}

// ReviewState describes the diff range and changed files under review.
type ReviewState struct {
	Base         string        `json:"base"`
	Head         string        `json:"head,omitempty"`
	BaseSHA      string        `json:"baseSha,omitempty"`
	HeadSHA      string        `json:"headSha,omitempty"`
	ChangedFiles []ChangedFile `json:"changedFiles"`
}

// ChangedFile records bounded git diff statistics for one path.
type ChangedFile struct {
	Path      string `json:"path"`
	Additions int    `json:"additions,omitempty"`
	Deletions int    `json:"deletions,omitempty"`
	Binary    bool   `json:"binary,omitempty"`
	Redacted  bool   `json:"redacted,omitempty"`
}

// DiffState contains the redacted diff payload.
type DiffState struct {
	Text       string `json:"text"`
	Truncated  bool   `json:"truncated,omitempty"`
	Redacted   bool   `json:"redacted,omitempty"`
	LimitBytes int    `json:"limitBytes"`
}

// ContextItem is a bounded repository/cARL context document.
type ContextItem struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Truncated bool   `json:"truncated,omitempty"`
	Redacted  bool   `json:"redacted,omitempty"`
}

// DataCategory documents one category of data included in evaluator payloads.
type DataCategory struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Included    bool   `json:"included"`
	Redacted    bool   `json:"redacted,omitempty"`
	Truncated   bool   `json:"truncated,omitempty"`
}

// QuestionKind identifies a bounded semantic question primitive.
type QuestionKind string

const (
	BooleanQuestion QuestionKind = "boolean"
	ChoiceQuestion  QuestionKind = "choice"
	ScoreQuestion   QuestionKind = "score"
)

// Question is a cARL-owned semantic question. It is intentionally bounded and
// typed rather than an arbitrary prompt blob.
type Question struct {
	ID          string       `json:"id"`
	Kind        QuestionKind `json:"kind"`
	Description string       `json:"description"`
	Choices     []string     `json:"choices,omitempty"`
	Threshold   *float64     `json:"threshold,omitempty"`
}

// EvaluationResult contains provider-independent semantic signals plus
// non-policy diagnostics.
type EvaluationResult struct {
	Signals     []Signal     `json:"signals"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

// Signal is one normalized evaluator answer.
type Signal struct {
	ID         string         `json:"id"`
	Value      any            `json:"value"`
	Confidence *float64       `json:"confidence,omitempty"`
	Source     string         `json:"source"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// Diagnostic describes evaluator/configuration status without being a policy
// finding.
type Diagnostic struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

// ProviderError is a sanitized evaluator failure. Message must not include
// secrets, request headers, or provider response bodies.
type ProviderError struct {
	Code    string
	Message string
}

func (e *ProviderError) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return e.Message
}
