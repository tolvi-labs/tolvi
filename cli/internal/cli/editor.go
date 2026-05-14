package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CaptureOpts controls how the doc body is captured.
// Precedence: BodyFlag > StdinReader > editor flow.
// RunEditor is an injection point for tests; nil means "use $EDITOR".
type CaptureOpts struct {
	InitialContent []byte
	BodyFlag       string
	StdinReader    io.Reader
	RunEditor      func(path string) error
}

// CaptureBody returns the bytes the user wants stored as the doc body.
//
// Errors when the resulting body is empty (whitespace-only counts).
func CaptureBody(opts CaptureOpts) ([]byte, error) {
	if opts.BodyFlag != "" {
		return assertNonEmpty([]byte(opts.BodyFlag))
	}
	if opts.StdinReader != nil {
		data, err := io.ReadAll(opts.StdinReader)
		if err != nil && err.Error() != "EOF" {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		return assertNonEmpty(data)
	}

	runEditor := opts.RunEditor
	if runEditor == nil {
		runEditor = runSystemEditor
	}

	tmpDir, err := os.MkdirTemp("", "tolvi-sync-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	tmpPath := filepath.Join(tmpDir, "doc.md")
	if err := os.WriteFile(tmpPath, opts.InitialContent, 0o644); err != nil {
		return nil, fmt.Errorf("write template: %w", err)
	}

	if err := runEditor(tmpPath); err != nil {
		return nil, fmt.Errorf("editor: %w", err)
	}

	out, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("read edited file: %w", err)
	}
	return assertNonEmpty(out)
}

func assertNonEmpty(body []byte) ([]byte, error) {
	if strings.TrimSpace(string(body)) == "" {
		return nil, fmt.Errorf("empty body — aborting")
	}
	return body, nil
}

func runSystemEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
