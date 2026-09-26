package review

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/goldjg/carl/internal/semantic"
)

type OutputFormat string

const (
	FormatText     OutputFormat = "text"
	FormatJSON     OutputFormat = "json"
	FormatMarkdown OutputFormat = "markdown"
)

type outputDocument struct {
	SchemaVersion int             `json:"schemaVersion"`
	Result        Result          `json:"result"`
	DryRun        *dryRunDocument `json:"dryRun,omitempty"`
}

type dryRunDocument struct {
	Provider       string                   `json:"provider"`
	Questions      []semantic.Question      `json:"questions"`
	State          semantic.EvaluationState `json:"state"`
	DataCategories []semantic.DataCategory  `json:"dataCategories"`
}

func writeResult(w io.Writer, format OutputFormat, result Result, dryRun *dryRunDocument) error {
	switch format {
	case FormatJSON:
		data, err := json.MarshalIndent(outputDocument{SchemaVersion: 1, Result: result, DryRun: dryRun}, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal review JSON: %w", err)
		}
		_, err = fmt.Fprintln(w, string(data))
		return err
	case FormatMarkdown:
		writeMarkdown(w, result, dryRun)
		return nil
	default:
		writeText(w, result, dryRun)
		return nil
	}
}

func writeText(w io.Writer, result Result, dryRun *dryRunDocument) {
	fmt.Fprintln(w, "cARL review")
	fmt.Fprintf(w, "Decision: %s\n", result.Decision)
	if dryRun != nil {
		fmt.Fprintln(w, "Dry run: provider was not called")
		fmt.Fprintf(w, "Provider: %s\n", dryRun.Provider)
		fmt.Fprintln(w, "Data categories:")
		for _, cat := range dryRun.DataCategories {
			fmt.Fprintf(w, "  - %s: included=%t redacted=%t truncated=%t\n", cat.Name, cat.Included, cat.Redacted, cat.Truncated)
		}
		fmt.Fprintln(w, "Questions:")
		for _, q := range dryRun.Questions {
			fmt.Fprintf(w, "  - %s (%s): %s\n", q.ID, q.Kind, q.Description)
		}
		return
	}
	if len(result.Diagnostics) > 0 {
		fmt.Fprintln(w, "Diagnostics:")
		for _, d := range result.Diagnostics {
			fmt.Fprintf(w, "  - %s [%s] %s\n", strings.ToUpper(d.Severity), d.Code, d.Message)
		}
	}
	if len(result.Findings) == 0 {
		fmt.Fprintln(w, "Findings: none")
		return
	}
	fmt.Fprintln(w, "Findings:")
	for _, f := range result.Findings {
		fmt.Fprintf(w, "  - %s [%s] %s\n", f.ID, f.Severity, f.Title)
		fmt.Fprintf(w, "    %s\n", f.Description)
	}
}

func writeMarkdown(w io.Writer, result Result, dryRun *dryRunDocument) {
	fmt.Fprintln(w, "# cARL review")
	fmt.Fprintf(w, "\n**Decision:** `%s`\n", result.Decision)
	if dryRun != nil {
		fmt.Fprintln(w, "\n**Dry run:** provider was not called")
		fmt.Fprintf(w, "\n**Provider:** `%s`\n", dryRun.Provider)
		fmt.Fprintln(w, "\n## Data categories")
		for _, cat := range dryRun.DataCategories {
			fmt.Fprintf(w, "- `%s` — included=%t, redacted=%t, truncated=%t\n", cat.Name, cat.Included, cat.Redacted, cat.Truncated)
		}
		fmt.Fprintln(w, "\n## Semantic questions")
		for _, q := range dryRun.Questions {
			fmt.Fprintf(w, "- `%s` (`%s`) — %s\n", q.ID, q.Kind, q.Description)
		}
		return
	}
	if len(result.Diagnostics) > 0 {
		fmt.Fprintln(w, "\n## Diagnostics")
		for _, d := range result.Diagnostics {
			fmt.Fprintf(w, "- **%s** `%s` — %s\n", strings.ToUpper(d.Severity), d.Code, d.Message)
		}
	}
	fmt.Fprintln(w, "\n## Findings")
	if len(result.Findings) == 0 {
		fmt.Fprintln(w, "No findings.")
		return
	}
	for _, f := range result.Findings {
		fmt.Fprintf(w, "- **%s** `%s` — %s\n", f.Severity, f.ID, f.Title)
		fmt.Fprintf(w, "  %s\n", f.Description)
	}
}
