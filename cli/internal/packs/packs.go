// Package packs serves the role packs this binary carries.
//
// tolvi-solo owns the packs. `2026-06-05-tolvi-solo-as-separate-product` is
// active and rejects folding them into this repo, so they are vendored here
// rather than moved: the files under files/ are a copy of tolvi-solo's
// packs/<name>/{pack.json,templates/}, pinned byte-identical by
// .github/scripts/pack-copy-parity-check.sh. Vendoring is what lets `tolvi
// init --pack` work from a cask-installed binary with no sibling checkout.
package packs

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed all:files
var files embed.FS

// Template is one template file and what it is for.
type Template struct {
	File string `json:"file"`
	Use  string `json:"use"`
}

// Pack is a vendored pack's manifest, as tolvi-solo declares it.
type Pack struct {
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	Summary   string     `json:"summary"`
	Verticals []string   `json:"verticals"`
	Templates []Template `json:"templates"`
}

// List returns every vendored pack, sorted by name.
func List() ([]Pack, error) {
	entries, err := fs.ReadDir(files, "files")
	if err != nil {
		return nil, err
	}
	var out []Pack
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		raw, err := files.ReadFile(filepath.Join("files", e.Name(), "pack.json"))
		if err != nil {
			return nil, fmt.Errorf("pack %s has no manifest: %w", e.Name(), err)
		}
		var p Pack
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("pack %s: %w", e.Name(), err)
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Names lists the vendored pack names, for a message that has to name them.
func Names() []string {
	packs, err := List()
	if err != nil {
		return nil
	}
	var names []string
	for _, p := range packs {
		names = append(names, p.Name)
	}
	return names
}

// Install writes a pack's templates into dest and returns what it wrote. A
// template already there is left alone and not reported: people edit these,
// and a re-run that silently replaced an edited template would lose work.
func Install(name, dest string) ([]string, error) {
	dir := filepath.Join("files", name, "templates")
	entries, err := fs.ReadDir(files, dir)
	if err != nil {
		return nil, fmt.Errorf("no pack %q (available: %s)", name, strings.Join(Names(), ", "))
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return nil, err
	}

	var written []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		target := filepath.Join(dest, e.Name())
		if _, err := os.Stat(target); err == nil {
			continue
		}
		data, err := files.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return written, err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return written, err
		}
		written = append(written, e.Name())
	}
	return written, nil
}
