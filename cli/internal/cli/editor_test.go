package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCaptureBody_FromEditor(t *testing.T) {
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "tmpl.md")
	_ = os.WriteFile(tmplPath, []byte("# heading\n\nbody\n"), 0o644)

	called := false
	stub := func(path string) error {
		called = true
		return os.WriteFile(path, []byte("# heading\n\nUSER EDITED BODY\n"), 0o644)
	}
	got, err := CaptureBody(CaptureOpts{
		InitialContent: []byte("# heading\n\nbody\n"),
		RunEditor:      stub,
	})
	if err != nil {
		t.Fatalf("CaptureBody: %v", err)
	}
	if !called {
		t.Error("editor stub not invoked")
	}
	if string(got) != "# heading\n\nUSER EDITED BODY\n" {
		t.Errorf("body = %q", got)
	}
}

func TestCaptureBody_FromFlag(t *testing.T) {
	got, err := CaptureBody(CaptureOpts{
		BodyFlag: "from --body flag\n",
	})
	if err != nil {
		t.Fatalf("CaptureBody: %v", err)
	}
	if string(got) != "from --body flag\n" {
		t.Errorf("body = %q", got)
	}
}

func TestCaptureBody_FromStdin(t *testing.T) {
	got, err := CaptureBody(CaptureOpts{
		StdinReader: &eofReader{data: []byte("piped body content")},
	})
	if err != nil {
		t.Fatalf("CaptureBody: %v", err)
	}
	if string(got) != "piped body content" {
		t.Errorf("body = %q", got)
	}
}

func TestCaptureBody_EmptyBodyError(t *testing.T) {
	stub := func(path string) error {
		// Editor stub writes only whitespace.
		return os.WriteFile(path, []byte("\n   \n"), 0o644)
	}
	_, err := CaptureBody(CaptureOpts{
		InitialContent: []byte(""),
		RunEditor:      stub,
	})
	if err == nil {
		t.Fatal("expected error on empty body")
	}
}

// eofReader returns the buffered data on a single Read, then io.EOF.
type eofReader struct {
	data []byte
	done bool
}

func (r *eofReader) Read(p []byte) (int, error) {
	if r.done {
		return 0, errEOF
	}
	n := copy(p, r.data)
	r.done = true
	return n, nil
}

var errEOF = errors.New("EOF")
