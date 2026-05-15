package cli

import (
	"encoding/json"
	"io"

	"github.com/tolvi-labs/tolvi/cli/internal/llm"
	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

type askJSONOutput struct {
	Answer     string        `json:"answer"`
	Citations  []citationOut `json:"citations"`
	Unverified []string      `json:"unverified_citations"`
	Model      string        `json:"model"`
	Tokens     tokensOut     `json:"tokens"`
}

type citationOut struct {
	Slug    string `json:"slug"`
	DocType string `json:"doc_type"`
	Path    string `json:"path"`
	Status  string `json:"status"`
}

type tokensOut struct {
	Input         int64 `json:"input"`
	Output        int64 `json:"output"`
	CacheReadHits int64 `json:"cache_read_hits"`
}

func printAskJSON(stdout, stderr io.Writer, res llm.StreamResult, matched, unmatched []string, docBySlug map[string]vault.Doc) error {
	_ = stderr
	out := askJSONOutput{
		Answer:     res.Text,
		Unverified: unmatched,
		Model:      res.Model,
		Tokens: tokensOut{
			Input:         res.InputTokens,
			Output:        res.OutputTokens,
			CacheReadHits: res.CacheReadHits,
		},
	}
	for _, slug := range matched {
		d := docBySlug[slug]
		out.Citations = append(out.Citations, citationOut{
			Slug:    slug,
			DocType: d.Type,
			Path:    d.Path,
			Status:  d.Frontmatter.String("status"),
		})
	}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
