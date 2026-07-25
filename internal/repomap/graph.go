package repomap

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	evidenceDerived     = "derived"
	evidencePartial     = "partial"
	evidenceUnavailable = "unavailable"
)

// Graph is the evidence-scoped cognitive repository graph emitted by carl map.
type Graph struct {
	Nodes    []GraphNode   `json:"nodes"`
	Edges    []GraphEdge   `json:"edges"`
	Coverage GraphCoverage `json:"coverage"`
}

// GraphNode is a stable repository-relative component or artefact.
type GraphNode struct {
	ID                    string   `json:"id"`
	Kind                  string   `json:"kind"`
	Path                  string   `json:"path"`
	Purpose               string   `json:"purpose,omitempty"`
	Criticality           string   `json:"criticality"`
	TrustBoundary         string   `json:"trust_boundary"`
	PolicyAttachmentPoint bool     `json:"policy_attachment_point"`
	AgentContext          string   `json:"agent_context"`
	Owners                []string `json:"owners,omitempty"`
	ChangeImpact          []string `json:"change_impact,omitempty"`
}

// GraphEdge is a directed, evidence-backed relationship between two nodes.
type GraphEdge struct {
	From     string   `json:"from"`
	To       string   `json:"to"`
	Type     string   `json:"type"`
	Evidence []string `json:"evidence"`
}

// GraphCoverage records how much evidence backs each graph knowledge facet.
type GraphCoverage struct {
	Ownership         GraphFacet `json:"ownership"`
	Dependencies      GraphFacet `json:"dependencies"`
	DataFlows         GraphFacet `json:"data_flows"`
	TrustBoundaries   GraphFacet `json:"trust_boundaries"`
	Criticality       GraphFacet `json:"criticality"`
	PolicyAttachments GraphFacet `json:"policy_attachments"`
	ChangeImpact      GraphFacet `json:"change_impact"`
}

// GraphFacet states whether graph knowledge is derived, partial, or
// unavailable and explains the evidence boundary.
type GraphFacet struct {
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type goPackage struct {
	imports map[string][]string
}

func buildGraph(rootDir string, m *Map) (Graph, error) {
	packages, modulePath, err := discoverGoPackages(rootDir)
	if err != nil {
		return Graph{}, err
	}

	nodes := map[string]GraphNode{}
	addNode := func(node GraphNode) {
		nodes[node.ID] = node
	}

	addNode(GraphNode{
		ID:                    componentID("."),
		Kind:                  "repository",
		Path:                  ".",
		Purpose:               "Repository root",
		Criticality:           "high",
		TrustBoundary:         "repository",
		PolicyAttachmentPoint: true,
		AgentContext:          "Repository orientation anchor; inspect connected components before changing cross-cutting behaviour.",
	})

	directoryPaths := make([]string, 0, len(m.Directories))
	for rel := range m.Directories {
		directoryPaths = append(directoryPaths, rel)
	}
	sort.Strings(directoryPaths)
	for _, rel := range directoryPaths {
		kind := "component"
		if _, ok := packages[rel]; ok {
			kind = "package"
		}
		addNode(componentNode(rel, kind, m.Directories[rel]))
	}

	packagePaths := make([]string, 0, len(packages))
	for rel := range packages {
		packagePaths = append(packagePaths, rel)
	}
	sort.Strings(packagePaths)
	for _, rel := range packagePaths {
		if rel == "." {
			continue
		}
		if _, ok := nodes[componentID(rel)]; !ok {
			addNode(componentNode(rel, "package", ""))
		}
	}

	for _, file := range m.EntryPoints {
		addNode(fileNode("entry_point", "entry-point", file, "high", "repository"))
	}
	for _, file := range m.Workflows {
		addNode(fileNode("workflow", "workflow", file, "high", "automation"))
	}
	for _, file := range m.Governance {
		addNode(fileNode("governance", "governance", file, "high", "governance"))
	}
	for _, file := range m.Documentation {
		addNode(fileNode("documentation", "documentation", file, "low", "repository"))
	}

	policyNodes, err := discoverPolicyNodes(rootDir)
	if err != nil {
		return Graph{}, err
	}
	for _, node := range policyNodes {
		addNode(node)
	}

	edges := make([]GraphEdge, 0, len(nodes))
	for _, node := range nodes {
		if node.ID == componentID(".") {
			continue
		}
		if parentID := nearestContainerID(node.Path, nodes); parentID != "" && parentID != node.ID {
			edges = append(edges, GraphEdge{
				From:     parentID,
				To:       node.ID,
				Type:     "contains",
				Evidence: []string{node.Path},
			})
		}
	}

	dependencyEdges := dependencyEdges(packages, modulePath, nodes)
	edges = append(edges, dependencyEdges...)
	applyDirectChangeImpact(nodes, dependencyEdges)

	nodeList := make([]GraphNode, 0, len(nodes))
	for _, node := range nodes {
		sort.Strings(node.ChangeImpact)
		node.ChangeImpact = deduplicateStrings(node.ChangeImpact)
		nodeList = append(nodeList, node)
	}
	sort.Slice(nodeList, func(i, j int) bool {
		return nodeList[i].ID < nodeList[j].ID
	})
	sortGraphEdges(edges)

	return Graph{
		Nodes: nodeList,
		Edges: edges,
		Coverage: GraphCoverage{
			Ownership: GraphFacet{
				Status: evidenceUnavailable,
				Detail: "Ownership is not inferred. No owner is reported without a supported authoritative ownership source.",
			},
			Dependencies: GraphFacet{
				Status: evidencePartial,
				Detail: "Repository-local Go import declarations are derived statically; dependencies in other languages and dynamic dependencies are not inferred.",
			},
			DataFlows: GraphFacet{
				Status: evidenceUnavailable,
				Detail: "Static imports do not prove runtime data flow, so runtime flows are not inferred.",
			},
			TrustBoundaries: GraphFacet{
				Status: evidenceDerived,
				Detail: "Nodes are classified as repository, governance, policy, or automation surfaces from canonical repository-relative paths.",
			},
			Criticality: GraphFacet{
				Status: evidenceDerived,
				Detail: "Criticality is a deterministic orientation heuristic based on node kind, not a risk assessment.",
			},
			PolicyAttachments: GraphFacet{
				Status: evidencePartial,
				Detail: "Repository components are marked as policy attachment points and policy definitions are discoverable; active policy assignment remains the responsibility of `carl trace`.",
			},
			ChangeImpact: GraphFacet{
				Status: evidencePartial,
				Detail: "Change impact lists direct reverse repository-local Go import dependencies only; it is not transitive or a runtime impact guarantee.",
			},
		},
	}, nil
}

func componentNode(rel, kind, purpose string) GraphNode {
	context := purpose
	if context == "" {
		if kind == "package" {
			context = fmt.Sprintf("Go package at %s.", rel)
		} else {
			context = fmt.Sprintf("Repository component at %s.", rel)
		}
	}
	return GraphNode{
		ID:                    componentID(rel),
		Kind:                  kind,
		Path:                  rel,
		Purpose:               purpose,
		Criticality:           "medium",
		TrustBoundary:         "repository",
		PolicyAttachmentPoint: true,
		AgentContext:          context,
	}
}

func fileNode(idPrefix, kind string, file File, criticality, boundary string) GraphNode {
	context := file.Purpose
	if context == "" {
		context = fmt.Sprintf("%s artefact at %s.", strings.ReplaceAll(kind, "_", " "), file.Path)
	}
	return GraphNode{
		ID:            idPrefix + ":" + file.Path,
		Kind:          kind,
		Path:          file.Path,
		Purpose:       file.Purpose,
		Criticality:   criticality,
		TrustBoundary: boundary,
		AgentContext:  context,
	}
}

func componentID(rel string) string {
	if rel == "." || rel == "" {
		return "repository:."
	}
	return "component:" + filepath.ToSlash(rel)
}

func nearestContainerID(rel string, nodes map[string]GraphNode) string {
	parent := path.Dir(filepath.ToSlash(rel))
	for {
		id := componentID(parent)
		if _, ok := nodes[id]; ok {
			return id
		}
		if parent == "." || parent == "/" {
			return componentID(".")
		}
		parent = path.Dir(parent)
	}
}

func discoverPolicyNodes(rootDir string) ([]GraphNode, error) {
	instructionsDir := filepath.Join(rootDir, ".github", "instructions")
	if _, err := os.Stat(instructionsDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var nodes []GraphNode
	err := filepath.WalkDir(instructionsDir, func(absPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".instructions.md") {
			return nil
		}
		rel, err := filepath.Rel(rootDir, absPath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		packRel, err := filepath.Rel(instructionsDir, absPath)
		if err != nil {
			return err
		}
		packID := strings.TrimSuffix(filepath.ToSlash(packRel), ".instructions.md")
		nodes = append(nodes, GraphNode{
			ID:            "policy:" + packID,
			Kind:          "policy",
			Path:          rel,
			Purpose:       "Instruction pack " + packID,
			Criticality:   "high",
			TrustBoundary: "policy",
			AgentContext:  "Policy definition only; verify active policy evaluation with `carl trace`.",
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})
	return nodes, nil
}

func discoverGoPackages(rootDir string) (map[string]*goPackage, string, error) {
	modulePath := readGoModuleName(filepath.Join(rootDir, "go.mod"))
	packages := map[string]*goPackage{}

	err := filepath.WalkDir(rootDir, func(absPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if absPath != rootDir && skipDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 || !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}

		relFile, err := filepath.Rel(rootDir, absPath)
		if err != nil {
			return err
		}
		relFile = filepath.ToSlash(relFile)
		relDir := path.Dir(relFile)
		pkg := packages[relDir]
		if pkg == nil {
			pkg = &goPackage{imports: map[string][]string{}}
			packages[relDir] = pkg
		}

		parsed, err := parser.ParseFile(token.NewFileSet(), absPath, nil, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("parse Go imports in %s: %w", relFile, err)
		}
		for _, spec := range parsed.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return fmt.Errorf("parse Go import in %s: %w", relFile, err)
			}
			pkg.imports[importPath] = append(pkg.imports[importPath], relFile)
		}
		return nil
	})
	if err != nil {
		return nil, "", err
	}
	return packages, modulePath, nil
}

func dependencyEdges(packages map[string]*goPackage, modulePath string, nodes map[string]GraphNode) []GraphEdge {
	if modulePath == "" {
		return nil
	}
	var edges []GraphEdge
	for rel, pkg := range packages {
		from := componentID(rel)
		if _, ok := nodes[from]; !ok {
			continue
		}
		for importPath, evidence := range pkg.imports {
			var targetRel string
			switch {
			case importPath == modulePath:
				targetRel = "."
			case strings.HasPrefix(importPath, modulePath+"/"):
				targetRel = strings.TrimPrefix(importPath, modulePath+"/")
			default:
				continue
			}
			to := componentID(targetRel)
			if _, ok := nodes[to]; !ok || from == to {
				continue
			}
			sort.Strings(evidence)
			edges = append(edges, GraphEdge{
				From:     from,
				To:       to,
				Type:     "depends_on",
				Evidence: deduplicateStrings(evidence),
			})
		}
	}
	return edges
}

func applyDirectChangeImpact(nodes map[string]GraphNode, edges []GraphEdge) {
	for _, edge := range edges {
		if edge.Type != "depends_on" {
			continue
		}
		target := nodes[edge.To]
		target.ChangeImpact = append(target.ChangeImpact, edge.From)
		nodes[edge.To] = target
	}
}

func sortGraphEdges(edges []GraphEdge) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Type != edges[j].Type {
			return edges[i].Type < edges[j].Type
		}
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		if edges[i].To != edges[j].To {
			return edges[i].To < edges[j].To
		}
		return strings.Join(edges[i].Evidence, "\x00") < strings.Join(edges[j].Evidence, "\x00")
	})
}

func deduplicateStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := values[:0]
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}
