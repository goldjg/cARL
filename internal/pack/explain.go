package pack

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goldjg/carl/internal/manifest"
)

// PolicyExplanationNotice states the intentional epistemic boundary of the
// Phase 5 explanation surface.
const PolicyExplanationNotice = "Pack-level policy provenance only. This output does not interpret individual natural-language rules or expose prompts, hidden model reasoning, or chain-of-thought. Instruction availability or loading does not prove model adherence."

// PolicyContext identifies the repository artefact and optional profile
// context that supplied active pack seeds.
type PolicyContext struct {
	Mode    string `json:"mode"`
	Source  string `json:"source"`
	Profile string `json:"profile,omitempty"`
	Role    string `json:"role,omitempty"`
	Task    string `json:"task,omitempty"`
}

// PolicyActivation is one structured reason why a pack entered the effective
// policy set.
type PolicyActivation struct {
	Kind        string `json:"kind"`
	Description string `json:"description"`
	Source      string `json:"source"`
	RelatedPack string `json:"relatedPack,omitempty"`
	Profile     string `json:"profile,omitempty"`
	Role        string `json:"role,omitempty"`
	Task        string `json:"task,omitempty"`
}

// PolicyEffect describes conservative pack-level composition effects.
// Non-overridden effective packs add constraints; resolved overrides and
// overridden packs remain explicit for provenance.
type PolicyEffect struct {
	AddsConstraints   bool     `json:"addsConstraints"`
	DeclaredOverrides []string `json:"declaredOverrides"`
	ResolvedOverrides []string `json:"resolvedOverrides"`
	OverriddenBy      []string `json:"overriddenBy"`
}

// PolicyExplanation explains one discoverable pack relative to the current
// effective policy evaluation.
type PolicyExplanation struct {
	ID                  string             `json:"id"`
	Version             string             `json:"version"`
	Title               string             `json:"title,omitempty"`
	Applied             bool               `json:"applied"`
	Status              string             `json:"status"`
	Source              string             `json:"source"`
	CanonicalDefinition string             `json:"canonicalDefinition"`
	RegistryProvenance  *PackProvenance    `json:"registryProvenance,omitempty"`
	Selected            bool               `json:"selected"`
	ActiveSeed          bool               `json:"activeSeed"`
	Order               *int               `json:"order,omitempty"`
	Priority            int                `json:"priority"`
	Mode                string             `json:"mode"`
	Activation          []PolicyActivation `json:"activation"`
	Requires            []string           `json:"requires"`
	RequiredBy          []string           `json:"requiredBy"`
	Effect              PolicyEffect       `json:"effect"`
}

// PolicyDecision records an observable evaluation decision and the canonical
// input that supports it.
type PolicyDecision struct {
	Kind    string `json:"kind"`
	Outcome string `json:"outcome"`
	Subject string `json:"subject,omitempty"`
	Target  string `json:"target,omitempty"`
	Code    string `json:"code,omitempty"`
	Source  string `json:"source,omitempty"`
	Reason  string `json:"reason"`
}

// PolicyTrace is the complete pack-level effective-policy evaluation trace.
type PolicyTrace struct {
	SchemaVersion int                 `json:"schemaVersion"`
	Notice        string              `json:"notice"`
	Context       PolicyContext       `json:"context"`
	Policies      []PolicyExplanation `json:"policies"`
	Decisions     []PolicyDecision    `json:"decisions"`
	Conflicts     []Conflict          `json:"conflicts"`
}

type policyEvaluation struct {
	Trace        PolicyTrace
	Explanations map[string]PolicyExplanation
}

func policyContextForRoot(rootDir string) (PolicyContext, error) {
	profiles, err := ReadProfiles(rootDir)
	if err != nil {
		return PolicyContext{}, err
	}
	if profiles != nil {
		return PolicyContext{
			Mode:    "profiles",
			Source:  ProfileFileName,
			Profile: profiles.Active.Profile,
			Role:    profiles.Active.Role,
			Task:    profiles.Active.Task,
		}, nil
	}

	selectionPath := filepath.Join(rootDir, filepath.FromSlash(SelectionFileName))
	if _, err := os.Stat(selectionPath); err == nil {
		return PolicyContext{
			Mode:   "legacy-selection",
			Source: SelectionFileName,
		}, nil
	} else if !os.IsNotExist(err) {
		return PolicyContext{}, fmt.Errorf("inspect %s: %w", SelectionFileName, err)
	}
	if manifest.Exists(rootDir) {
		return PolicyContext{
			Mode:   "legacy-selection",
			Source: manifest.FileName,
		}, nil
	}
	return PolicyContext{Mode: "none", Source: "none"}, nil
}

func buildPolicyEvaluation(packs []PackMetadata, context PolicyContext) (*policyEvaluation, error) {
	set, err := ComputeEffectiveSet(packs)
	if err != nil {
		return nil, err
	}

	index := make(map[string]PackMetadata, len(packs))
	explanations := make(map[string]PolicyExplanation, len(packs))
	for _, pack := range packs {
		index[pack.ID] = pack
		priority, mode := packPrecedence(pack)
		explanations[pack.ID] = PolicyExplanation{
			ID:                  pack.ID,
			Version:             pack.Version,
			Title:               pack.Title,
			Status:              "inactive",
			Source:              pack.Source,
			CanonicalDefinition: expectedPackPath(pack.ID),
			RegistryProvenance:  pack.Provenance,
			Selected:            pack.State.Selected,
			ActiveSeed:          pack.State.Active,
			Priority:            priority,
			Mode:                mode,
			Activation:          []PolicyActivation{},
			Requires:            requiredPackIDs(pack),
			RequiredBy:          []string{},
			Effect: PolicyEffect{
				DeclaredOverrides: declaredOverrides(pack),
				ResolvedOverrides: []string{},
				OverriddenBy:      []string{},
			},
		}
	}

	resolvedOverrides := map[string][]string{}
	for _, effective := range set.Packs {
		for _, overridingPack := range effective.OverriddenBy {
			resolvedOverrides[overridingPack] = append(
				resolvedOverrides[overridingPack],
				effective.ID,
			)
		}
	}
	for id := range resolvedOverrides {
		sort.Strings(resolvedOverrides[id])
	}

	policies := make([]PolicyExplanation, 0, len(set.Packs))
	decisions := make([]PolicyDecision, 0, len(set.Packs)*3)
	for i, effective := range set.Packs {
		if _, ok := index[effective.ID]; !ok {
			return nil, fmt.Errorf("effective pack %s is not discoverable", effective.ID)
		}
		explanation := explanations[effective.ID]
		order := i + 1
		overridden := len(effective.OverriddenBy) > 0
		explanation.Applied = !overridden
		explanation.Status = "effective"
		explanation.Order = &order
		explanation.Priority = effective.Priority
		explanation.Mode = effective.Mode
		explanation.Activation = activationsFromReasons(effective.Reasons, context)
		explanation.RequiredBy = relatedPacks(explanation.Activation)
		explanation.Effect.AddsConstraints = !overridden
		explanation.Effect.ResolvedOverrides = append(
			[]string(nil),
			resolvedOverrides[effective.ID]...,
		)
		explanation.Effect.OverriddenBy = append(
			[]string(nil),
			effective.OverriddenBy...,
		)
		if overridden {
			explanation.Status = "overridden"
		}
		explanations[effective.ID] = explanation
		policies = append(policies, explanation)

		for _, activation := range explanation.Activation {
			decision := PolicyDecision{
				Kind:    activation.Kind,
				Outcome: "included",
				Subject: effective.ID,
				Source:  activation.Source,
				Reason:  activation.Description,
			}
			if activation.Kind == "dependency" {
				decision.Target = effective.ID
				decision.Subject = activation.RelatedPack
			}
			decisions = append(decisions, decision)
		}
		decisions = append(decisions, PolicyDecision{
			Kind:    "precedence",
			Outcome: "ordered",
			Subject: effective.ID,
			Source:  expectedPackPath(effective.ID),
			Reason: fmt.Sprintf(
				"position %d uses priority %d with pack-ID tie-breaking",
				order,
				effective.Priority,
			),
		})
		if overridden {
			decisions = append(decisions, PolicyDecision{
				Kind:    "constraint",
				Outcome: "not-applied",
				Subject: effective.ID,
				Source:  expectedPackPath(effective.ID),
				Reason:  "the pack remains visible for provenance but its instruction definition is overridden and not applied",
			})
		} else {
			decisions = append(decisions, PolicyDecision{
				Kind:    "constraint",
				Outcome: "strengthens",
				Subject: effective.ID,
				Source:  expectedPackPath(effective.ID),
				Reason:  "the effective pack adds constraints at pack level; individual prose rules are not interpreted",
			})
		}
	}

	for _, policy := range policies {
		for _, target := range policy.Effect.ResolvedOverrides {
			decisions = append(decisions, PolicyDecision{
				Kind:    "override",
				Outcome: "resolved",
				Subject: policy.ID,
				Target:  target,
				Source:  policy.CanonicalDefinition,
				Reason: fmt.Sprintf(
					"%s explicitly declares the override and %s declares mode \"overridable\"",
					policy.ID,
					target,
				),
			})
		}
	}
	for _, conflict := range set.Conflicts {
		decisions = append(decisions, PolicyDecision{
			Kind:    "conflict",
			Outcome: "unresolved",
			Code:    conflict.Code,
			Reason:  conflict.Message,
		})
	}

	return &policyEvaluation{
		Trace: PolicyTrace{
			SchemaVersion: metadataSchemaVersion,
			Notice:        PolicyExplanationNotice,
			Context:       context,
			Policies:      policies,
			Decisions:     decisions,
			Conflicts:     append([]Conflict(nil), set.Conflicts...),
		},
		Explanations: explanations,
	}, nil
}

func packPrecedence(pack PackMetadata) (int, string) {
	priority := defaultPriority
	mode := defaultMode
	if pack.Precedence != nil {
		if pack.Precedence.Priority != nil {
			priority = *pack.Precedence.Priority
		}
		if pack.Precedence.Mode != "" {
			mode = pack.Precedence.Mode
		}
	}
	return priority, mode
}

func requiredPackIDs(pack PackMetadata) []string {
	ids := []string{}
	for _, dependency := range pack.Dependencies {
		if dependency.Required {
			ids = append(ids, dependency.ID)
		}
	}
	sort.Strings(ids)
	return ids
}

func declaredOverrides(pack PackMetadata) []string {
	if pack.Precedence == nil {
		return []string{}
	}
	overrides := append([]string(nil), pack.Precedence.Overrides...)
	sort.Strings(overrides)
	return overrides
}

func activationsFromReasons(reasons []string, context PolicyContext) []PolicyActivation {
	activations := make([]PolicyActivation, 0, len(reasons))
	for _, reason := range reasons {
		activation := PolicyActivation{
			Kind:        "activation",
			Description: reason,
			Source:      context.Source,
		}
		switch {
		case reason == "selected":
			activation.Kind = "selection"
		case reason == "organization default":
			activation.Kind = "organization-default"
		case reason == "repository default":
			activation.Kind = "repository-default"
		case strings.HasPrefix(reason, "profile "):
			activation.Kind = "profile"
			activation.Profile = strings.TrimPrefix(reason, "profile ")
		case strings.HasPrefix(reason, "role "):
			activation.Kind = "role"
			role, profile := splitOverlayReason(
				strings.TrimPrefix(reason, "role "),
				" in profile ",
			)
			activation.Role = role
			activation.Profile = profile
		case strings.HasPrefix(reason, "task "):
			activation.Kind = "task"
			task, profile := splitOverlayReason(
				strings.TrimPrefix(reason, "task "),
				" in profile ",
			)
			activation.Task = task
			activation.Profile = profile
		case strings.HasPrefix(reason, "dependency of "):
			activation.Kind = "dependency"
			activation.RelatedPack = strings.TrimPrefix(reason, "dependency of ")
			activation.Source = expectedPackPath(activation.RelatedPack)
		}
		activations = append(activations, activation)
	}
	return activations
}

func splitOverlayReason(value, separator string) (string, string) {
	parts := strings.SplitN(value, separator, 2)
	if len(parts) != 2 {
		return value, ""
	}
	return parts[0], parts[1]
}

func relatedPacks(activations []PolicyActivation) []string {
	related := []string{}
	for _, activation := range activations {
		if activation.Kind == "dependency" && activation.RelatedPack != "" {
			related = append(related, activation.RelatedPack)
		}
	}
	sort.Strings(related)
	return related
}

func decisionsForPolicy(decisions []PolicyDecision, packID string) []PolicyDecision {
	out := []PolicyDecision{}
	for _, decision := range decisions {
		if decision.Subject == packID || decision.Target == packID ||
			decision.Kind == "conflict" {
			out = append(out, decision)
		}
	}
	return out
}
