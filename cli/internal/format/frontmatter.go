package format

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Frontmatter is a parsed YAML frontmatter map. Values are kept as the
// raw types yaml.v3 produces (string, int, []interface{}, map[string]interface{}).
// Callers reach for type-specific accessors below.
type Frontmatter map[string]any

// String returns fm[key] coerced to string. Missing or non-string keys
// return "".
func (fm Frontmatter) String(key string) string {
	v, ok := fm[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// Tags returns fm["tags"] as []string. The second return is false when
// the key is missing or the value isn't a list.
func (fm Frontmatter) Tags() ([]string, bool) {
	v, ok := fm["tags"]
	if !ok {
		return nil, false
	}
	list, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		s, ok := item.(string)
		if !ok {
			continue
		}
		out = append(out, s)
	}
	return out, true
}

// ParseFrontmatter splits a Markdown document into its YAML frontmatter
// (delimited by lines containing exactly `---`) and body.
//
// Errors when:
//   - no opening `---` on the first non-empty line
//   - no closing `---` before EOF
//   - frontmatter YAML is malformed
func ParseFrontmatter(src []byte) (Frontmatter, []byte, error) {
	delim := []byte("---\n")

	if !bytes.HasPrefix(src, delim) && !bytes.HasPrefix(src, []byte("---\r\n")) {
		return nil, nil, fmt.Errorf("missing opening frontmatter delimiter on first line")
	}

	// Move past the opening delimiter.
	after := bytes.TrimPrefix(src, delim)
	after = bytes.TrimPrefix(after, []byte("---\r\n"))

	// Find the closing delimiter on its own line.
	closeIdx := bytes.Index(after, []byte("\n---\n"))
	if closeIdx < 0 {
		closeIdx = bytes.Index(after, []byte("\n---\r\n"))
		if closeIdx < 0 {
			return nil, nil, fmt.Errorf("missing closing frontmatter delimiter")
		}
	}

	rawYAML := after[:closeIdx+1] // include the trailing newline before ---
	body := after[closeIdx:]
	body = bytes.TrimPrefix(body, []byte("\n---\n"))
	body = bytes.TrimPrefix(body, []byte("\n---\r\n"))
	// Trim leading blank line if present.
	body = bytes.TrimPrefix(body, []byte("\n"))

	var root yaml.Node
	if err := yaml.Unmarshal(rawYAML, &root); err != nil {
		return nil, nil, fmt.Errorf("invalid frontmatter YAML: %w", err)
	}
	fm, err := decodeMapping(&root)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid frontmatter YAML: %w", err)
	}
	return fm, body, nil
}

// decodeMapping walks a yaml.Node and returns a Frontmatter where scalar
// values without an explicit type tag are kept as strings. This avoids
// yaml.v3's default !!timestamp resolution turning `date: 2026-04-12`
// into a time.Time.
func decodeMapping(n *yaml.Node) (Frontmatter, error) {
	if n.Kind == yaml.DocumentNode {
		if len(n.Content) == 0 {
			return Frontmatter{}, nil
		}
		return decodeMapping(n.Content[0])
	}
	if n.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("frontmatter must be a YAML mapping")
	}
	fm := make(Frontmatter, len(n.Content)/2)
	for i := 0; i < len(n.Content); i += 2 {
		keyNode := n.Content[i]
		valNode := n.Content[i+1]
		v, err := decodeNode(valNode)
		if err != nil {
			return nil, err
		}
		fm[keyNode.Value] = v
	}
	return fm, nil
}

func decodeNode(n *yaml.Node) (any, error) {
	switch n.Kind {
	case yaml.ScalarNode:
		// Preserve the literal source string for implicit-tagged scalars
		// (e.g. dates) so callers see exactly what was written. Explicit
		// tags still go through yaml's resolver.
		if n.Tag == "" || n.Tag == "!!str" || n.Tag == "!!timestamp" {
			return n.Value, nil
		}
		var v any
		if err := n.Decode(&v); err != nil {
			return nil, err
		}
		return v, nil
	case yaml.SequenceNode:
		out := make([]any, 0, len(n.Content))
		for _, c := range n.Content {
			v, err := decodeNode(c)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case yaml.MappingNode:
		out := make(map[string]any, len(n.Content)/2)
		for i := 0; i < len(n.Content); i += 2 {
			k := n.Content[i].Value
			v, err := decodeNode(n.Content[i+1])
			if err != nil {
				return nil, err
			}
			out[k] = v
		}
		return out, nil
	case yaml.AliasNode:
		return decodeNode(n.Alias)
	}
	return nil, nil
}
