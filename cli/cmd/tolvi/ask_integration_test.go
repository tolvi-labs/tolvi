package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// reuse buildToTmp from sync_integration_test.go (same package)

// stubAnthropic returns an httptest.Server that emits a single Anthropic
// SSE response containing answerText. The text becomes the assembled
// streaming response.
func stubAnthropic(answerText string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		writeSSE := func(evt, data string) {
			_, _ = w.Write([]byte("event: " + evt + "\ndata: " + data + "\n\n"))
			if flusher != nil {
				flusher.Flush()
			}
		}
		writeSSE("message_start", `{"type":"message_start","message":{"id":"x","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4-7","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":50,"output_tokens":0,"cache_read_input_tokens":0}}}`)
		writeSSE("content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`)
		writeSSE("content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":`+jsonString(answerText)+`}}`)
		writeSSE("content_block_stop", `{"type":"content_block_stop","index":0}`)
		writeSSE("message_delta", `{"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"input_tokens":0,"output_tokens":10,"cache_read_input_tokens":0}}`)
		writeSSE("message_stop", `{"type":"message_stop"}`)
	}))
}

// jsonString JSON-encodes a string (escapes quotes, backslashes, newlines).
func jsonString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func TestIntegration_Ask_StreamsAnswerAndPrintsSources(t *testing.T) {
	bin := buildToTmp(t)
	work := t.TempDir()

	// 1. tolvi init
	init := exec.Command(bin, "init", "--workspace", "ask-it")
	init.Dir = work
	if out, err := init.CombinedOutput(); err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}

	// 2. tolvi sync decision (auto-generates slug "why-postgres")
	sync := exec.Command(bin, "sync", "decision", "Why Postgres",
		"--body", "# Postgres\n\nbecause pgvector.\n")
	sync.Dir = work
	if out, err := sync.CombinedOutput(); err != nil {
		t.Fatalf("sync: %v\n%s", err, out)
	}

	// 3. Stub Anthropic API; LLM "answers" with the verified citation.
	stub := stubAnthropic("We chose Postgres. [[why-postgres]]")
	defer stub.Close()

	// 4. tolvi ask, pointed at the stub.
	ask := exec.Command(bin, "ask", "why postgres")
	ask.Dir = work
	ask.Env = append(os.Environ(),
		"ANTHROPIC_API_KEY=sk-ant-stub",
		"ANTHROPIC_BASE_URL="+stub.URL,
	)
	out, err := ask.CombinedOutput()
	if err != nil {
		t.Fatalf("ask: %v\n%s", err, out)
	}
	s := string(out)
	if !strings.Contains(s, "We chose Postgres") {
		t.Errorf("answer not printed: %s", s)
	}
	if !strings.Contains(s, "Sources:") {
		t.Errorf("Sources footer missing: %s", s)
	}
	// Slug "why-postgres" should appear in the verified Sources block.
	if !strings.Contains(s, "[[why-postgres]]") {
		t.Errorf("verified citation missing: %s", s)
	}
	// And the path to the decision file should be in the footer.
	expectedPath := filepath.ToSlash(filepath.Join("decisions"))
	if !strings.Contains(s, expectedPath) {
		t.Errorf("Sources path missing: %s", s)
	}
}

func TestIntegration_Ask_UnverifiedCitationFlagged(t *testing.T) {
	bin := buildToTmp(t)
	work := t.TempDir()

	init := exec.Command(bin, "init", "--workspace", "ask-unverified")
	init.Dir = work
	if out, err := init.CombinedOutput(); err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}
	sync := exec.Command(bin, "sync", "decision", "Real Decision",
		"--body", "# Real\n\nbody.\n")
	sync.Dir = work
	if out, err := sync.CombinedOutput(); err != nil {
		t.Fatalf("sync: %v\n%s", err, out)
	}

	stub := stubAnthropic("See [[real-decision]] and also [[hallucinated]].")
	defer stub.Close()

	ask := exec.Command(bin, "ask", "anything")
	ask.Dir = work
	ask.Env = append(os.Environ(),
		"ANTHROPIC_API_KEY=sk-ant-stub",
		"ANTHROPIC_BASE_URL="+stub.URL,
	)
	out, err := ask.CombinedOutput()
	if err != nil {
		t.Fatalf("ask: %v\n%s", err, out)
	}
	s := string(out)
	if !strings.Contains(s, "⚠ unverified") {
		t.Errorf("unverified marker missing: %s", s)
	}
	if !strings.Contains(s, "hallucinated") {
		t.Errorf("hallucinated slug missing from footer: %s", s)
	}
}
