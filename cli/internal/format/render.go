package format

import (
	"bytes"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// preferredKeyOrder defines the canonical ordering for known frontmatter
// keys when rendering. Keys not listed appear after, in alphabetical order.
// Matches the order used in tolvi-format-v1 spec examples.
var preferredKeyOrder = []string{
	"tags",
	"date",
	"repo",
	"status",
	"ticket",
	"user_impact",
	"product_area",
	"workspace",
	"embedding_model",
	"schema_version",
}

// RenderDocument serializes a frontmatter map and body back to the
// canonical `---\n<yaml>\n---\n\n<body>` shape.
//
// Key ordering is stable: keys appear in preferredKeyOrder when present,
// then any remaining keys in alphabetical order. This means
// parse → render → parse is a fixed point, and renders are reproducible
// (no map-iteration nondeterminism).
func RenderDocument(fm Frontmatter, body []byte) ([]byte, error) {
	ordered := orderedKeys(fm)

	var yamlBuf bytes.Buffer
	enc := yaml.NewEncoder(&yamlBuf)
	enc.SetIndent(2)
	// Build an ordered map representation by emitting a yaml.Node tree
	// (yaml.v3 doesn't preserve insertion order on map types).
	root := &yaml.Node{Kind: yaml.MappingNode}
	for _, k := range ordered {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: k}
		valNode, err := encodeYAMLValue(fm[k])
		if err != nil {
			return nil, fmt.Errorf("render key %q: %w", k, err)
		}
		root.Content = append(root.Content, keyNode, valNode)
	}
	if err := enc.Encode(root); err != nil {
		return nil, fmt.Errorf("yaml encode: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("yaml close: %w", err)
	}

	var out bytes.Buffer
	out.WriteString("---\n")
	out.Write(yamlBuf.Bytes())
	out.WriteString("---\n\n")
	out.Write(body)
	return out.Bytes(), nil
}

func encodeYAMLValue(v any) (*yaml.Node, error) {
	var node yaml.Node
	if err := node.Encode(v); err != nil {
		return nil, err
	}
	// Flow-style for short scalar lists (matches tags: [decision] convention).
	if node.Kind == yaml.SequenceNode && isShortScalarSeq(&node) {
		node.Style = yaml.FlowStyle
	}
	return &node, nil
}

func isShortScalarSeq(n *yaml.Node) bool {
	if n.Kind != yaml.SequenceNode {
		return false
	}
	if len(n.Content) == 0 || len(n.Content) > 6 {
		return false
	}
	for _, c := range n.Content {
		if c.Kind != yaml.ScalarNode {
			return false
		}
	}
	return true
}

func orderedKeys(fm Frontmatter) []string {
	present := map[string]bool{}
	for k := range fm {
		present[k] = true
	}
	var ordered []string
	for _, k := range preferredKeyOrder {
		if present[k] {
			ordered = append(ordered, k)
			delete(present, k)
		}
	}
	var rest []string
	for k := range present {
		rest = append(rest, k)
	}
	sort.Strings(rest)
	return append(ordered, rest...)
}
