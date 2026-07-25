package pack

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/goldjg/carl/internal/cmdutil"
)

// ExplainCommand implements top-level `carl explain`.
type ExplainCommand struct {
	pack *Command
}

// NewExplain returns the top-level policy explanation command.
func NewExplain(arts Artifacts) *ExplainCommand {
	return &ExplainCommand{pack: New(arts)}
}

// Name returns the command name.
func (c *ExplainCommand) Name() string { return "explain" }

// Synopsis returns a short description.
func (c *ExplainCommand) Synopsis() string {
	return "Explain pack-level policy provenance and evaluation"
}

// Run executes `carl explain`.
func (c *ExplainCommand) Run(_ context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing pack ID\n\nUsage: carl explain <pack-id> [--json]")
	}
	if args[0] == "--help" || args[0] == "-h" {
		fmt.Println("Usage: carl explain <pack-id> [--json]")
		return nil
	}
	packID := args[0]
	jsonOut, help, err := parseJSONFlag(args[1:])
	if err != nil {
		return err
	}
	if help {
		fmt.Println("Usage: carl explain <pack-id> [--json]")
		return nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return c.RunExplainInDir(cwd, packID, jsonOut)
}

type explainPayload struct {
	SchemaVersion int               `json:"schemaVersion"`
	Notice        string            `json:"notice"`
	Context       PolicyContext     `json:"context"`
	Policy        PolicyExplanation `json:"policy"`
	Decisions     []PolicyDecision  `json:"decisions"`
	Conflicts     []Conflict        `json:"conflicts"`
}

// RunExplainInDir explains one discoverable pack.
func (c *ExplainCommand) RunExplainInDir(rootDir, packID string, jsonOut bool) error {
	evaluation, err := c.evaluateInDir(rootDir)
	if err != nil {
		return policyInputError(err, jsonOut)
	}
	explanation, ok := evaluation.Explanations[packID]
	if !ok {
		msg := fmt.Sprintf("unknown pack %q", packID)
		if jsonOut {
			return newJSONExitError("pack_not_found", msg)
		}
		return fmt.Errorf("%s", msg)
	}
	payload := explainPayload{
		SchemaVersion: metadataSchemaVersion,
		Notice:        PolicyExplanationNotice,
		Context:       evaluation.Trace.Context,
		Policy:        explanation,
		Decisions:     decisionsForPolicy(evaluation.Trace.Decisions, packID),
		Conflicts:     evaluation.Trace.Conflicts,
	}

	if jsonOut {
		if err := printPolicyJSON(payload, "marshal policy explanation JSON"); err != nil {
			return err
		}
		if len(payload.Conflicts) > 0 {
			return &cmdutil.ExitError{Code: 1}
		}
		return nil
	}

	printPolicyExplanation(payload)
	if len(payload.Conflicts) > 0 {
		return &cmdutil.ExitError{
			Code:    1,
			Message: "policy composition conflicts detected",
		}
	}
	return nil
}

func (c *ExplainCommand) evaluateInDir(rootDir string) (*policyEvaluation, error) {
	packs, err := c.pack.discover(rootDir)
	if err != nil {
		return nil, err
	}
	if issues := validatePackSet(packs); len(issues) > 0 {
		return nil, fmt.Errorf(
			"pack metadata validation failed:\n- %s",
			strings.Join(issues, "\n- "),
		)
	}
	policyContext, err := policyContextForRoot(rootDir)
	if err != nil {
		return nil, err
	}
	return buildPolicyEvaluation(packs, policyContext)
}

// TraceCommand implements top-level `carl trace`.
type TraceCommand struct {
	explain *ExplainCommand
}

// NewTrace returns the top-level policy evaluation trace command.
func NewTrace(arts Artifacts) *TraceCommand {
	return &TraceCommand{explain: NewExplain(arts)}
}

// Name returns the command name.
func (c *TraceCommand) Name() string { return "trace" }

// Synopsis returns a short description.
func (c *TraceCommand) Synopsis() string {
	return "Trace the effective pack-level policy evaluation"
}

// Run executes `carl trace`.
func (c *TraceCommand) Run(_ context.Context, args []string) error {
	jsonOut, help, err := parseJSONFlag(args)
	if err != nil {
		return err
	}
	if help {
		fmt.Println("Usage: carl trace [--json]")
		return nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return c.RunTraceInDir(cwd, jsonOut)
}

// RunTraceInDir prints the complete effective pack policy trace.
func (c *TraceCommand) RunTraceInDir(rootDir string, jsonOut bool) error {
	evaluation, err := c.explain.evaluateInDir(rootDir)
	if err != nil {
		return policyInputError(err, jsonOut)
	}
	trace := evaluation.Trace
	if jsonOut {
		if err := printPolicyJSON(trace, "marshal policy trace JSON"); err != nil {
			return err
		}
		if len(trace.Conflicts) > 0 {
			return &cmdutil.ExitError{Code: 1}
		}
		return nil
	}

	printPolicyTrace(trace)
	if len(trace.Conflicts) > 0 {
		return &cmdutil.ExitError{
			Code:    1,
			Message: "policy composition conflicts detected",
		}
	}
	return nil
}

func printPolicyJSON(payload any, context string) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("%s: %w", context, err)
	}
	fmt.Println(string(data))
	return nil
}

func policyInputError(err error, jsonOut bool) error {
	if jsonOut {
		return newJSONExitError("policy_evaluation_failed", err.Error())
	}
	return err
}

func printPolicyExplanation(payload explainPayload) {
	policy := payload.Policy
	fmt.Printf("Policy Explanation: %s\n\n", policy.ID)
	fmt.Printf("Notice: %s\n\n", payload.Notice)
	fmt.Printf("Applied:              %t\n", policy.Applied)
	fmt.Printf("Status:               %s\n", policy.Status)
	fmt.Printf("Version:              %s\n", policy.Version)
	fmt.Printf("Source:               %s\n", policy.Source)
	fmt.Printf("Canonical definition: %s\n", policy.CanonicalDefinition)
	fmt.Printf("Selected:             %t\n", policy.Selected)
	fmt.Printf("Active seed:          %t\n", policy.ActiveSeed)
	if policy.Order == nil {
		fmt.Println("Effective order:      none")
	} else {
		fmt.Printf("Effective order:      %d\n", *policy.Order)
	}
	fmt.Printf("Priority:             %d\n", policy.Priority)
	fmt.Printf("Mode:                 %s\n", policy.Mode)
	fmt.Printf("Adds constraints:     %t\n", policy.Effect.AddsConstraints)
	printPolicyValues("Requires", policy.Requires)
	printPolicyValues("Required by", policy.RequiredBy)
	printPolicyValues("Declared overrides", policy.Effect.DeclaredOverrides)
	printPolicyValues("Resolved overrides", policy.Effect.ResolvedOverrides)
	printPolicyValues("Overridden by", policy.Effect.OverriddenBy)

	fmt.Println()
	fmt.Println("Activation:")
	if len(policy.Activation) == 0 {
		fmt.Println("  none (pack is not in the effective policy)")
	} else {
		for _, activation := range policy.Activation {
			fmt.Printf(
				"  [%s] %s (source: %s)\n",
				activation.Kind,
				activation.Description,
				activation.Source,
			)
		}
	}
	printPolicyDecisions(payload.Decisions)
	printPolicyConflicts(payload.Conflicts)
}

func printPolicyTrace(trace PolicyTrace) {
	fmt.Println("Policy Evaluation Trace")
	fmt.Println()
	fmt.Printf("Notice: %s\n\n", trace.Notice)
	fmt.Printf("Context: %s (%s)\n", trace.Context.Mode, trace.Context.Source)
	if trace.Context.Profile != "" {
		fmt.Printf("Profile: %s\n", trace.Context.Profile)
		fmt.Printf("Role:    %s\n", valueOrNone(trace.Context.Role))
		fmt.Printf("Task:    %s\n", valueOrNone(trace.Context.Task))
	}

	fmt.Println()
	fmt.Println("Effective Policies (precedence order):")
	if len(trace.Policies) == 0 {
		fmt.Println("  none")
	} else {
		fmt.Println("  Order  ID                                Status      Priority  Mode")
		for _, policy := range trace.Policies {
			fmt.Printf(
				"  %-6d %-33s %-11s %-9d %s\n",
				*policy.Order,
				policy.ID,
				policy.Status,
				policy.Priority,
				policy.Mode,
			)
		}
	}
	printPolicyDecisions(trace.Decisions)
	printPolicyConflicts(trace.Conflicts)
}

func printPolicyValues(label string, values []string) {
	value := "none"
	if len(values) > 0 {
		value = strings.Join(values, ", ")
	}
	fmt.Printf("%-21s %s\n", label+":", value)
}

func printPolicyDecisions(decisions []PolicyDecision) {
	fmt.Println()
	fmt.Println("Decisions:")
	if len(decisions) == 0 {
		fmt.Println("  none")
		return
	}
	for _, decision := range decisions {
		relationship := decision.Subject
		if decision.Target != "" {
			relationship += " -> " + decision.Target
		}
		if relationship != "" {
			relationship += ": "
		}
		fmt.Printf(
			"  [%s/%s] %s%s\n",
			decision.Kind,
			decision.Outcome,
			relationship,
			decision.Reason,
		)
	}
}

func printPolicyConflicts(conflicts []Conflict) {
	if len(conflicts) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("Conflicts:")
	fmt.Printf("- %s\n", ConflictSummary(conflicts))
}
