package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// RootRole is the part a vault root plays in a repo's chain.
//
// Three roles describe the real topology. A fourth, "personal", was designed
// for a directory that turned out to be an org root rather than a cross-org
// one; it is rejected by LoadRoots rather than left half-declared.
type RootRole string

const (
	// RoleRepo is the repo's own vault/ — committed, world-safe, contributor-fed.
	// It is discovered from the working directory, never declared in roots.json.
	RoleRepo RootRole = "repo"
	// RoleProduct is optional and declared only where sibling repos genuinely
	// share a brain. Private.
	RoleProduct RootRole = "product"
	// RoleOrg is private, one per workspace, and holds sessions for every repo
	// in that workspace plus its internal decisions.
	RoleOrg RootRole = "org"
)

// RootsConfigFileName is the machine-local declaration of every root, read from
// ~/.config/tolvi/. It is never committed: repos commit identity only, so a
// public repo never carries the path of a private vault.
const RootsConfigFileName = "roots.json"

// Root is one declared vault root.
type Root struct {
	Role RootRole `json:"role"`
	Path string   `json:"path"`
	// Workspace names the workspace an org root serves. Required for RoleOrg.
	Workspace string `json:"workspace,omitempty"`
	// Product names the product a product root serves. Required for RoleProduct.
	Product string `json:"product,omitempty"`
}

// private reports whether docs written to this root stay out of the repo's
// git history. Every role but the repo's own vault is private.
func (r Root) private() bool { return r.Role != RoleRepo }

// Identity is what a repo commits about itself in .vault-meta.json: the
// workspace is the container, the repo is the member, and the product is
// optional. The names match the API's model, where uniqueness is keyed on
// (workspace_id, repo_id, doc_type, slug).
type Identity struct {
	Workspace string
	Repo      string
	Product   string
}

// DocScope is the part of a doc's frontmatter that decides which repo it
// belongs to. A doc with no Repo is scoped to whichever container root it
// lives in rather than to any one repo.
type DocScope struct {
	Repo string
}

// Roots is the declared set of roots for this machine.
type Roots struct {
	declared bool
	roots    []Root
}

// Declared reports whether a roots.json was found. Absence is a legitimate
// state, not an error: it is the external contributor, who has only the repo
// vault in front of them.
func (r Roots) Declared() bool { return r.declared }

// LoadRoots reads the machine-local roots declaration.
//
// A missing file yields an undeclared Roots — single-root mode, where every doc
// lands in the repo's own vault and visibility is not consulted. A file that
// exists but cannot be read, parsed, or validated is a hard error: silently
// ignoring it is how an internal note gets written into a public repo.
func LoadRoots(path string) (Roots, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Roots{}, nil
	}
	if err != nil {
		return Roots{}, fmt.Errorf("read %s: %w", path, err)
	}

	var cfg struct {
		Roots []Root `json:"roots"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Roots{}, fmt.Errorf("parse %s: %w", path, err)
	}

	for i, rt := range cfg.Roots {
		switch rt.Role {
		case RoleOrg:
			if rt.Workspace == "" {
				return Roots{}, fmt.Errorf("%s: roots[%d]: an org root requires a workspace", path, i)
			}
		case RoleProduct:
			if rt.Product == "" {
				return Roots{}, fmt.Errorf("%s: roots[%d]: a product root requires a product", path, i)
			}
		case RoleRepo:
			return Roots{}, fmt.Errorf("%s: roots[%d]: the repo root is discovered from the working directory, never declared", path, i)
		default:
			return Roots{}, fmt.Errorf("%s: roots[%d]: unknown role %q (want %q or %q)", path, i, rt.Role, RoleOrg, RoleProduct)
		}
		if rt.Path == "" {
			return Roots{}, fmt.Errorf("%s: roots[%d]: path is required", path, i)
		}
	}

	return Roots{declared: true, roots: cfg.Roots}, nil
}

// Chain selects the roots that apply to one repo, ordered nearest scope first:
// the repo's own vault, then its product root if it declares one, then its
// workspace's org root.
//
// Selecting the org root by workspace is what makes isolation structural: a
// repo declaring an isolated workspace can never resolve another workspace's
// root, with no flag to forget.
func (r Roots) Chain(id Identity, repoVault string) Chain {
	c := Chain{id: id, declared: r.declared}
	c.roots = append(c.roots, Root{Role: RoleRepo, Path: repoVault})

	if id.Product != "" {
		for _, rt := range r.roots {
			if rt.Role == RoleProduct && rt.Product == id.Product {
				c.roots = append(c.roots, rt)
				break
			}
		}
	}
	for _, rt := range r.roots {
		if rt.Role == RoleOrg && rt.Workspace == id.Workspace {
			c.roots = append(c.roots, rt)
			break
		}
	}

	return c
}

// Chain is the ordered set of roots one repo can read from and write to. It is
// the single place the routing rule lives: write (sync, commit), read (recall,
// ask), the hooks, and the skills all resolve through it, so a doc's home
// stops depending on which tool happened to write it.
type Chain struct {
	id       Identity
	roots    []Root
	declared bool
}

// Roots returns the chain in order, nearest scope first.
func (c Chain) Roots() []Root { return c.roots }

// Role returns the chain's root for a role, if it has one.
func (c Chain) Role(role RootRole) (Root, bool) {
	for _, r := range c.roots {
		if r.Role == role {
			return r, true
		}
	}
	return Root{}, false
}

// Target decides where a new doc is written: the root that owns its scope when
// public, and the nearest declared private root above its scope when private.
//
// Sessions always route private, because the rule protecting them is about
// maintainer session context. The one exception is single-root mode, where a
// contributor's note belongs with their PR in the repo's own vault.
//
// A declared config that lacks the root a doc needs refuses the write. The
// asymmetry with single-root mode is deliberate: absent config is a legitimate
// contributor state, while a gap in present config is a misconfiguration, and
// falling back to the public root is how an internal note gets published.
func (c Chain) Target(docType, docVisibility string) (Root, error) {
	repo, _ := c.Role(RoleRepo)

	if !c.declared || !routesPrivate(docType, docVisibility) {
		return repo, nil
	}
	for _, r := range c.roots {
		if r.private() {
			return r, nil
		}
	}
	return Root{}, fmt.Errorf(
		"no private root is declared for workspace %q: refusing to write a %s into the repo's public vault — add an org root to %s",
		c.id.Workspace, docType, RootsConfigFileName,
	)
}

// routesPrivate reports whether a doc is withheld from the repo's own vault.
func routesPrivate(docType, docVisibility string) bool {
	return docType == "session" || docVisibility == "private"
}

// ReadRoots returns the roots a read spans, deduped by path. Two roles may be
// declared over one directory, and reading it twice would double every doc.
func (c Chain) ReadRoots() []Root {
	out := make([]Root, 0, len(c.roots))
	seen := make(map[string]bool, len(c.roots))
	for _, r := range c.roots {
		p := filepath.Clean(r.Path)
		if seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, r)
	}
	return out
}

// Accepts reports whether a doc found in a root belongs to this chain's repo.
//
// The repo's own vault is unambiguously its own, so everything there counts —
// requiring a repo tag would hide the docs a contributor writes. A shared root
// holds many repos' docs, so only those tagged with this repo count.
//
// Accepted consequence, chosen twice: a doc in a shared root with no repo tag
// is container-scoped and surfaces in no repo's recall.
func (c Chain) Accepts(root Root, doc DocScope) bool {
	if root.Role == RoleRepo {
		return true
	}
	return doc.Repo != "" && doc.Repo == c.id.Repo
}

// RootsConfigPath resolves the machine-local roots declaration:
// $XDG_CONFIG_HOME/tolvi/roots.json, else ~/.config/tolvi/roots.json.
func RootsConfigPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "tolvi", RootsConfigFileName)
}

// ChainFor builds the chain for a vault from its identity and the machine's
// declared roots. It is the single entry point every consumer uses, so write,
// read, the hooks, and the skills cannot disagree about where a doc lives.
func ChainFor(m Meta, vaultPath string) (Chain, error) {
	r, err := LoadRoots(RootsConfigPath())
	if err != nil {
		return Chain{}, err
	}
	return r.Chain(m.Identity(), vaultPath), nil
}

// WithPrivateRoot replaces the chain's declared private roots with an explicit
// one, backing the --private-vault flag. The repo root is untouched, so public
// docs still land where they belong.
func (c Chain) WithPrivateRoot(path string) Chain {
	out := Chain{id: c.id, declared: true}
	for _, r := range c.roots {
		if r.Role == RoleRepo {
			out.roots = append(out.roots, r)
		}
	}
	out.roots = append(out.roots, Root{Role: RoleOrg, Path: path, Workspace: c.id.Workspace})
	return out
}

// SessionFileName is what today's session note is called inside root. Notes in
// a shared root carry the repo they belong to, so sibling repos writing on the
// same day cannot collide.
func (c Chain) SessionFileName(root Root, date string) string {
	if root.Role == RoleRepo || c.id.Repo == "" {
		return date + ".md"
	}
	return date + "-" + c.id.Repo + ".md"
}

// SessionNotePath is the absolute path of a date's session note in root. The
// commit gate asks for this rather than computing its own, so the gate and the
// write cannot disagree about which file counts.
func (c Chain) SessionNotePath(root Root, date string) string {
	return filepath.Join(root.Path, "sessions", c.SessionFileName(root, date))
}
