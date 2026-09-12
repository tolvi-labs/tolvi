package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

// ErrNoSessionNote is returned by RunCommit when no session note exists for
// today. The caller maps it to ExitVaultState. RunCommit prints the
// user-facing guidance itself, so the caller should not reprint.
var ErrNoSessionNote = errors.New("no session note for today")

// CommitOpts controls the commit command.
type CommitOpts struct {
	RepoRoot  string
	VaultPath string
	Message   string    // -m message; if empty, git opens $EDITOR
	Today     string    // YYYY-MM-DD; "" → time.Now()
	Stdin     io.Reader // forwarded to git commit (for $EDITOR flows)
	Stdout    io.Writer
	Stderr    io.Writer

	// Public/private routing overrides (from --open-source/--OS and
	// --private-vault). When public, today's session note lives in the
	// private vault, so the gate looks there instead of the local vault.
	ForcePublic  bool
	PrivateVault string
}

// hasSessionNote reports whether vault/sessions/<date>.md exists and contains
// at least one "## " session block. This mirrors the tolvi-sync pre-commit
// hook's gate exactly.
func hasSessionNote(vaultPath, date string) bool {
	return hasSessionNoteFile(filepath.Join(vaultPath, "sessions", date+".md"))
}

// hasSessionNoteFile reports whether the file at path exists and contains at
// least one "## " session block.
func hasSessionNoteFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "## ") {
			return true
		}
	}
	return false
}

// sessionNotePath returns the absolute path of today's session note that the
// commit gate must check. Under local (non-public) visibility this is
// <vaultPath>/sessions/<date>.md. Under public visibility the note lives in
// the private vault as <private>/sessions/<date>-<workspace>.md.
func sessionNotePath(vaultPath string, m vault.Meta, date string) (string, error) {
	root, routed, err := vault.ResolveDocDestination(vaultPath, m, "session", "")
	if err != nil {
		return "", err
	}
	name := date + ".md"
	if routed {
		name = RoutedSessionFileName(date, m.Workspace)
	}
	return filepath.Join(root, "sessions", name), nil
}

// RunCommit is the mechanical, deterministic commit path. It gates on a
// session note existing for today, auto-stages vault/ so the vault always
// lands in the commit, and runs git commit. It performs NO synthesis — for
// that, use the /tolvi-commit skill in a working session.
func RunCommit(opts CommitOpts) error {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.Stdin == nil {
		opts.Stdin = os.Stdin
	}
	today := opts.Today
	if today == "" {
		today = time.Now().Format("2006-01-02")
	}

	// Resolve where today's session note should live. Under public
	// visibility it lives in the private vault; otherwise in the local vault.
	// If meta is unreadable, fall back to legacy local-only behavior.
	meta, err := vault.ReadMeta(opts.VaultPath)
	if err != nil {
		// A broken routing config must not degrade to local-only: that is
		// how a session note ends up gated against the public vault the
		// config exists to keep it out of. Only a missing or unreadable
		// .vault-meta.json falls back.
		if _, statErr := os.Stat(filepath.Join(opts.VaultPath, vault.RoutingConfigFileName)); statErr == nil {
			return fmt.Errorf("read vault meta: %w", err)
		}
		meta = vault.Meta{}
	}
	if opts.ForcePublic {
		meta.Visibility = "public"
	}
	if opts.PrivateVault != "" {
		meta.PrivateVault = opts.PrivateVault
	}
	notePath, err := sessionNotePath(opts.VaultPath, meta, today)
	if err != nil {
		fmt.Fprintf(opts.Stderr,
			"tolvi: %v.\n"+
				"  Set private_vault in .vault-meta.json or pass --private-vault <path>.\n",
			err)
		return ErrNoSessionNote
	}

	if !hasSessionNoteFile(notePath) {
		fmt.Fprintf(opts.Stderr,
			"tolvi: no session note for %s.\n"+
				"  Capture one first, then commit:\n"+
				"    • controlled:    tolvi sync session \"<title>\"\n"+
				"    • from a session: run the /tolvi-commit skill (synthesizes the note and commits)\n",
			today)
		return ErrNoSessionNote
	}

	// Auto-stage vault/ so it's always part of the commit. Stage only the
	// vault — code the user already staged is committed alongside it. This is
	// the controlled path; it never runs `git add -A`.
	rel, err := filepath.Rel(opts.RepoRoot, opts.VaultPath)
	if err != nil {
		rel = opts.VaultPath
	}
	add := exec.Command("git", "add", "--", rel)
	add.Dir = opts.RepoRoot
	add.Stdout, add.Stderr = opts.Stdout, opts.Stderr
	if err := add.Run(); err != nil {
		return fmt.Errorf("git add %s: %w", rel, err)
	}

	args := []string{"commit"}
	if opts.Message != "" {
		args = append(args, "-m", opts.Message)
	}
	commit := exec.Command("git", args...)
	commit.Dir = opts.RepoRoot
	commit.Stdin, commit.Stdout, commit.Stderr = opts.Stdin, opts.Stdout, opts.Stderr
	if err := commit.Run(); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}
