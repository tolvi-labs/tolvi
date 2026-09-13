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
// commit gate must check. The gate asks the chain rather than computing its
// own path, so it cannot disagree with where sync just wrote.
func sessionNotePath(chain vault.Chain, date string) (string, error) {
	root, err := chain.Target("session", "")
	if err != nil {
		return "", err
	}
	return chain.SessionNotePath(root, date), nil
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

	// Resolve where today's session note should live. An unreadable meta is a
	// hard failure rather than a quiet fall back to the repo vault: gating
	// against the public vault is how an internal note ends up committed to it.
	meta, err := vault.ReadMeta(opts.VaultPath)
	if err != nil {
		return fmt.Errorf("read vault meta: %w", err)
	}
	chain, err := vault.ChainFor(meta, opts.VaultPath)
	if err != nil {
		return fmt.Errorf("resolve vault roots: %w", err)
	}
	if opts.PrivateVault != "" {
		chain = chain.WithPrivateRoot(opts.PrivateVault)
	}
	notePath, err := sessionNotePath(chain, today)
	if err != nil {
		fmt.Fprintf(opts.Stderr,
			"tolvi: %v.\n"+
				"  Declare an org root in %s or pass --private-vault <path>.\n",
			err, vault.RootsConfigPath())
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
