package pack

import (
	"fmt"
	"sort"
	"strings"
)

// EffectivePack is one entry in the computed effective pack set.
type EffectivePack struct {
	ID           string   `json:"id"`
	Version      string   `json:"version"`
	Priority     int      `json:"priority"`
	Mode         string   `json:"mode"`
	Reasons      []string `json:"reasons"`
	OverriddenBy []string `json:"overriddenBy,omitempty"`
}

// Conflict describes a composition conflict detected while computing the
// effective pack set.
type Conflict struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// EffectiveSet is the computed effective pack set: explicit selection plus
// transitively expanded required dependencies, ordered by explicit
// precedence (priority descending, then pack ID) — never load order.
type EffectiveSet struct {
	Packs     []EffectivePack `json:"packs"`
	Conflicts []Conflict      `json:"conflicts,omitempty"`
}

const (
	defaultPriority = 0
	defaultMode     = "additive"
)

// ComputeEffectiveSet computes the effective pack set from the discovered
// pack universe. Composition is conservative: packs add constraints, no pack
// is ever removed from the set by an override, and override authority comes
// only from explicit metadata.
func ComputeEffectiveSet(packs []PackMetadata) (*EffectiveSet, error) {
	index := map[string]PackMetadata{}
	for _, p := range packs {
		index[p.ID] = p
	}

	// Seed with explicitly active packs, deterministically. Repositories
	// without profiles.json mark every selected pack active as a compatibility
	// fallback.
	var queue []string
	reasons := map[string][]string{}
	inSet := map[string]bool{}
	for _, p := range packs {
		if p.State.Active {
			queue = append(queue, p.ID)
			inSet[p.ID] = true
			if len(p.State.ActiveReasons) == 0 {
				reasons[p.ID] = append(reasons[p.ID], "selected")
			} else {
				reasons[p.ID] = append(reasons[p.ID], p.State.ActiveReasons...)
			}
		}
	}
	sort.Strings(queue)

	// Expand required dependencies transitively.
	var conflicts []Conflict
	for i := 0; i < len(queue); i++ {
		id := queue[i]
		p, ok := index[id]
		if !ok {
			conflicts = append(conflicts, Conflict{
				Code:    "missing_pack",
				Message: fmt.Sprintf("selected pack %s is not available", id),
			})
			continue
		}
		deps := append([]PackDependency(nil), p.Dependencies...)
		sort.Slice(deps, func(a, b int) bool { return deps[a].ID < deps[b].ID })
		for _, d := range deps {
			if !d.Required {
				continue
			}
			if _, ok := index[d.ID]; !ok {
				conflicts = append(conflicts, Conflict{
					Code:    "missing_dependency",
					Message: fmt.Sprintf("%s requires %s, which is not available", id, d.ID),
				})
				continue
			}
			reason := fmt.Sprintf("dependency of %s", id)
			if !containsString(reasons[d.ID], reason) {
				reasons[d.ID] = append(reasons[d.ID], reason)
			}
			if !inSet[d.ID] {
				inSet[d.ID] = true
				queue = append(queue, d.ID)
			}
		}
	}

	// Explicit override handling. An override is honoured only when the
	// overriding pack declares it in explicit metadata and the target pack
	// declares mode "overridable". Overridden packs stay in the set and are
	// flagged — no pack silently disables another.
	overriddenBy := map[string][]string{}
	ids := make([]string, 0, len(inSet))
	for id := range inSet {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		p, ok := index[id]
		if !ok || p.Precedence == nil {
			continue
		}
		for _, target := range p.Precedence.Overrides {
			if !inSet[target] {
				continue
			}
			tp := index[target]
			targetMode := defaultMode
			if tp.Precedence != nil && tp.Precedence.Mode != "" {
				targetMode = tp.Precedence.Mode
			}
			if targetMode != "overridable" {
				conflicts = append(conflicts, Conflict{
					Code: "override_not_permitted",
					Message: fmt.Sprintf(
						"%s declares an override of %s, but %s has mode %q (only mode \"overridable\" may be overridden)",
						id, target, target, targetMode),
				})
				continue
			}
			if declaresOverride(index[target], id) {
				conflicts = append(conflicts, Conflict{
					Code:    "mutual_override",
					Message: fmt.Sprintf("%s and %s declare overrides of each other", id, target),
				})
				continue
			}
			overriddenBy[target] = append(overriddenBy[target], id)
		}
	}

	out := &EffectiveSet{}
	for _, id := range ids {
		p := index[id]
		priority := defaultPriority
		mode := defaultMode
		if p.Precedence != nil {
			if p.Precedence.Priority != nil {
				priority = *p.Precedence.Priority
			}
			if p.Precedence.Mode != "" {
				mode = p.Precedence.Mode
			}
		}
		sort.Strings(reasons[id])
		sort.Strings(overriddenBy[id])
		out.Packs = append(out.Packs, EffectivePack{
			ID:           id,
			Version:      p.Version,
			Priority:     priority,
			Mode:         mode,
			Reasons:      reasons[id],
			OverriddenBy: overriddenBy[id],
		})
	}

	// Precedence order: explicit priority descending, ties broken by pack
	// ID — never filesystem or load order.
	sort.SliceStable(out.Packs, func(a, b int) bool {
		if out.Packs[a].Priority != out.Packs[b].Priority {
			return out.Packs[a].Priority > out.Packs[b].Priority
		}
		return out.Packs[a].ID < out.Packs[b].ID
	})

	sort.SliceStable(conflicts, func(a, b int) bool {
		if conflicts[a].Code != conflicts[b].Code {
			return conflicts[a].Code < conflicts[b].Code
		}
		return conflicts[a].Message < conflicts[b].Message
	})
	out.Conflicts = dedupeConflicts(conflicts)
	return out, nil
}

func declaresOverride(p PackMetadata, target string) bool {
	if p.Precedence == nil {
		return false
	}
	return containsString(p.Precedence.Overrides, target)
}

func containsString(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func dedupeConflicts(conflicts []Conflict) []Conflict {
	var out []Conflict
	seen := map[string]bool{}
	for _, c := range conflicts {
		key := c.Code + "\x00" + c.Message
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, c)
	}
	return out
}

// ConflictSummary renders conflicts as a human-readable bullet list.
func ConflictSummary(conflicts []Conflict) string {
	var lines []string
	for _, c := range conflicts {
		lines = append(lines, fmt.Sprintf("[%s] %s", c.Code, c.Message))
	}
	return strings.Join(lines, "\n- ")
}
