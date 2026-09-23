package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/tolvi-labs/tolvi/cli/internal/registry"
	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

// ReposOpts carries what the repos commands need. Every path is injectable so
// the commands are testable without touching the developer's real config.
type ReposOpts struct {
	RegistryPath string
	// Dirs is what scan searches, or the single path forget drops. Empty for
	// scan means the declared roots, which is the set the machine already
	// knows about.
	Dirs    []string
	JSON    bool
	Stdout  io.Writer
	Today   string
	Version string
}

// RunReposList prints the registered repos, verified against the disk.
func RunReposList(opts ReposOpts) error {
	opts = reposDefaults(opts)
	f, err := registry.Load(opts.RegistryPath)
	if err != nil {
		return err
	}

	summaries := make([]repoSummary, 0, len(f.Repos))
	for _, e := range f.Repos {
		summaries = append(summaries, summarize(e, opts.Today))
	}

	if opts.JSON {
		return printReposListJSON(opts.Stdout, opts.Version, summaries)
	}

	if len(summaries) == 0 {
		fmt.Fprintf(opts.Stdout, "No repos registered. `tolvi init` registers a repo, and `tolvi repos scan <dir>` finds existing ones.\n")
		return nil
	}
	for _, s := range summaries {
		mark := " "
		if s.Status == string(registry.StatusStale) {
			mark = "!"
		}
		fmt.Fprintf(opts.Stdout, "%s %-40s %s", mark, s.Path, s.Workspace)
		if s.Repo != "" {
			fmt.Fprintf(opts.Stdout, "/%s", s.Repo)
		}
		if s.Status == string(registry.StatusStale) {
			fmt.Fprint(opts.Stdout, "   (stale: no vault here now; `tolvi repos forget` drops it)")
		}
		fmt.Fprintln(opts.Stdout)
	}
	return nil
}

// RunReposScan rebuilds the registry from what is on disk.
func RunReposScan(opts ReposOpts) error {
	opts = reposDefaults(opts)
	dirs := opts.Dirs
	if len(dirs) == 0 {
		dirs = declaredRootDirs()
		if len(dirs) == 0 {
			return fmt.Errorf("no directory given and no roots declared: `tolvi repos scan <dir>` says where to look")
		}
	}
	found, err := registry.Scan(dirs)
	if err != nil {
		return err
	}
	if err := registry.Replace(opts.RegistryPath, found); err != nil {
		return err
	}
	fmt.Fprintf(opts.Stdout, "Registered %d repo(s) from %d director(ies).\n", len(found), len(dirs))
	return nil
}

// RunReposForget drops one entry by path.
func RunReposForget(opts ReposOpts) error {
	opts = reposDefaults(opts)
	if len(opts.Dirs) != 1 {
		return fmt.Errorf("forget takes exactly one path")
	}
	removed, err := registry.Forget(opts.RegistryPath, opts.Dirs[0])
	if err != nil {
		return err
	}
	if !removed {
		fmt.Fprintf(opts.Stdout, "%s is not registered; nothing to forget.\n", opts.Dirs[0])
		return nil
	}
	fmt.Fprintf(opts.Stdout, "Forgot %s.\n", opts.Dirs[0])
	return nil
}

// RegisterAfterInit records a freshly provisioned repo. It is best-effort by
// design: the vault is what init delivers, and the registry is a cache a scan
// can rebuild, so a failure here warns and names the repair rather than
// failing a command that already did its job.
func RegisterAfterInit(stderr io.Writer, registryPath, repoPath string) {
	if registryPath == "" {
		return
	}
	if err := registry.Add(registryPath, repoPath); err != nil {
		fmt.Fprintf(stderr, "warning: could not record this repo in %s (%v)\n         run `tolvi repos scan %s` to repair the index; the vault itself is fine\n",
			registryPath, err, repoPath)
	}
}

// repoSummary is one row of `repos list`, with the routing answers computed
// here by the resolver rather than left for a consumer to re-derive.
type repoSummary struct {
	Path        string
	Workspace   string
	Repo        string
	Product     string
	Status      string
	VaultPath   string
	SessionNote string
	Registered  string
}

func summarize(e registry.Entry, today string) repoSummary {
	s := repoSummary{
		Path:       e.Path,
		Workspace:  e.Workspace,
		Repo:       e.Repo,
		Product:    e.Product,
		Status:     string(registry.Verify(e)),
		Registered: e.Registered,
	}
	if s.Status != string(registry.StatusOK) {
		// A repo that is not there has no routing answers to give, and
		// inventing them from the cached identity would be a guess.
		return s
	}
	vaultPath := filepath.Join(e.Path, "vault")
	s.VaultPath = vaultPath
	meta, err := vault.ReadMeta(vaultPath)
	if err != nil {
		return s
	}
	chain, err := vault.ChainFor(meta, vaultPath)
	if err != nil {
		return s
	}
	target, err := chain.Target("session", "")
	if err != nil {
		return s
	}
	s.SessionNote = chain.SessionNotePath(target, today)
	return s
}

// declaredRootDirs is the default search set: the roots this machine already
// declares, plus nothing invented. A machine with no roots.json is told to
// name a directory instead.
func declaredRootDirs() []string {
	roots, err := vault.LoadRoots(vault.RootsConfigPath())
	if err != nil || !roots.Declared() {
		return nil
	}
	var dirs []string
	seen := map[string]bool{}
	for _, r := range roots.All() {
		// A root holds vaults; the repos that own them sit alongside it, so
		// the search starts one level up.
		dir := filepath.Dir(r.Path)
		if dir == "" || seen[dir] {
			continue
		}
		seen[dir] = true
		dirs = append(dirs, dir)
	}
	return dirs
}

func reposDefaults(opts ReposOpts) ReposOpts {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Today == "" {
		opts.Today = time.Now().Format("2006-01-02")
	}
	if opts.RegistryPath == "" {
		opts.RegistryPath = registry.ConfigPath()
	}
	return opts
}
