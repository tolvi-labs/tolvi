// Package registry maintains repos.json, the machine-local index of repos
// that have a vault.
//
// It is a cache, never a source of truth. The truth is each repo's
// .vault-meta.json, so an entry is verified when it is read, a repo that moved
// is marked stale rather than trusted or deleted, and a scan can rebuild the
// file from nothing. It carries no schema_version for the same reason
// roots.json carries none: a cache that can be rebuilt does not need a
// migration path.
//
// It is machine-local and never committed, exactly like roots.json: a repo
// commits who it is, and the machine decides what it knows about.
package registry

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

// FileName is the registry's name inside the tolvi config directory.
const FileName = "repos.json"

// Status is what verifying an entry against the disk found.
type Status string

const (
	StatusOK    Status = "ok"
	StatusStale Status = "stale"
)

// Entry is one registered repo. Identity is copied from the vault's meta at
// registration time so a listing can be rendered without reading every vault,
// and re-read when a caller needs more than the summary.
type Entry struct {
	Path       string `json:"path"`
	Workspace  string `json:"workspace"`
	Repo       string `json:"repo,omitempty"`
	Product    string `json:"product,omitempty"`
	Registered string `json:"registered"`
}

// File is the registry's on-disk shape.
type File struct {
	Repos []Entry `json:"repos"`
}

// ConfigPath resolves the machine-local registry, beside roots.json:
// $XDG_CONFIG_HOME/tolvi/repos.json, else ~/.config/tolvi/repos.json.
func ConfigPath() string {
	roots := vault.RootsConfigPath()
	if roots == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(roots), FileName)
}

// Load reads the registry. An absent file is an empty registry, not an error:
// a machine that has never registered a repo is a legitimate state.
func Load(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return File{}, nil
		}
		return File{}, fmt.Errorf("read %s: %w", path, err)
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return File{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return f, nil
}

// Save writes the registry, creating the config directory if needed.
func Save(path string, f File) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// EntryFor reads a repo's own identity, which is what makes it registrable.
// A directory with no readable vault meta is not a tolvi repo.
func EntryFor(repoPath string) (Entry, error) {
	abs, err := filepath.Abs(repoPath)
	if err != nil {
		return Entry{}, err
	}
	meta, err := vault.ReadMeta(filepath.Join(abs, "vault"))
	if err != nil {
		return Entry{}, fmt.Errorf("%s is not a tolvi repo: %w", abs, err)
	}
	return Entry{
		Path:       abs,
		Workspace:  meta.Workspace,
		Repo:       meta.Repo,
		Product:    meta.Product,
		Registered: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// Add registers a repo. Registering one already present is a no-op that keeps
// the original registration time, so `tolvi init --repair` and a re-run of
// init do not churn the file.
func Add(path, repoPath string) error {
	entry, err := EntryFor(repoPath)
	if err != nil {
		return err
	}
	f, err := Load(path)
	if err != nil {
		return err
	}
	for _, e := range f.Repos {
		if e.Path == entry.Path {
			return nil
		}
	}
	f.Repos = append(f.Repos, entry)
	sortEntries(f.Repos)
	return Save(path, f)
}

// Forget drops an entry by path, and reports whether one was there. Forgetting
// an unregistered path is not an error: the caller asked for it to be absent,
// and it is.
func Forget(path, repoPath string) (bool, error) {
	abs, err := filepath.Abs(repoPath)
	if err != nil {
		return false, err
	}
	f, err := Load(path)
	if err != nil {
		return false, err
	}
	kept := make([]Entry, 0, len(f.Repos))
	for _, e := range f.Repos {
		if e.Path != abs {
			kept = append(kept, e)
		}
	}
	if len(kept) == len(f.Repos) {
		return false, nil
	}
	f.Repos = kept
	return true, Save(path, f)
}

// Replace overwrites the registry with exactly these entries, which is what a
// full scan produces.
func Replace(path string, entries []Entry) error {
	sortEntries(entries)
	return Save(path, File{Repos: entries})
}

// Verify checks an entry against the disk. It never rewrites the registry:
// what to do about a stale entry is the caller's call, and the answer is
// always to report it rather than to delete it silently.
func Verify(e Entry) Status {
	if _, err := vault.ReadMeta(filepath.Join(e.Path, "vault")); err != nil {
		return StatusStale
	}
	return StatusOK
}

// skipDirs are directories a scan never descends into. A vendored or
// checked-out copy of another repo is not this machine's repo.
var skipDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	"dist":         true,
	"build":        true,
	"target":       true,
}

// Scan walks the given directories and returns an entry for every repo that
// has a readable vault meta. Directories are searched, not guessed: with no
// argument the caller passes the declared roots, which is the set the machine
// already knows about.
func Scan(dirs []string) ([]Entry, error) {
	var found []Entry
	seen := map[string]bool{}
	for _, dir := range dirs {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		err = filepath.WalkDir(abs, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				// An unreadable directory is skipped rather than fatal: a scan
				// that dies on one permission error finds nothing.
				if d != nil && d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			if !d.IsDir() {
				return nil
			}
			name := d.Name()
			if p != abs && (skipDirs[name] || (len(name) > 1 && name[0] == '.')) {
				return fs.SkipDir
			}
			if name != "vault" {
				return nil
			}
			// A vault directory: its parent is the repo.
			repo := filepath.Dir(p)
			if seen[repo] {
				return fs.SkipDir
			}
			entry, err := EntryFor(repo)
			if err != nil {
				return fs.SkipDir
			}
			seen[repo] = true
			found = append(found, entry)
			return fs.SkipDir
		})
		if err != nil {
			return nil, err
		}
	}
	sortEntries(found)
	return found, nil
}

func sortEntries(entries []Entry) {
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
}
