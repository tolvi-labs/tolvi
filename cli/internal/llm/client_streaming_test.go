package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// cannedStreamHandler returns SSE frames mimicking Anthropic's streaming
// response for messages.create. The `parts` slice is split into N
// content_block_delta events, each containing one part as text_delta.
func cannedStreamHandler(parts []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		write := func(evt, data string) {
			_, _ = w.Write([]byte("event: " + evt + "\ndata: " + data + "\n\n"))
			if flusher != nil {
				flusher.Flush()
			}
		}

		write("message_start", `{"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4-7","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":100,"output_tokens":0,"cache_read_input_tokens":50}}}`)
		write("content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`)
		for _, p := range parts {
			write("content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":`+stringJSON(p)+`}}`)
		}
		write("content_block_stop", `{"type":"content_block_stop","index":0}`)
		write("message_delta", `{"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"input_tokens":0,"output_tokens":42,"cache_read_input_tokens":0}}`)
		write("message_stop", `{"type":"message_stop"}`)
	})
}

// stringJSON returns a JSON-encoded string literal of s.
func stringJSON(s string) string {
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

func TestStreamCompletion_AssemblesText(t *testing.T) {
	srv := httptest.NewServer(cannedStreamHandler([]string{"Hello, ", "world!"}))
	defer srv.Close()

	c, err := NewClient(ClientOpts{
		APIKey:  "sk-ant-stub",
		Model:   "claude-sonnet-4-7",
		BaseURL: srv.URL,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	var streamed string
	res, err := c.StreamCompletion(context.Background(), "system prompt", "user query", 100, func(s string) {
		streamed += s
	})
	if err != nil {
		t.Fatalf("StreamCompletion: %v", err)
	}
	if res.Text != "Hello, world!" {
		t.Errorf("assembled text = %q", res.Text)
	}
	if streamed != "Hello, world!" {
		t.Errorf("streamed text = %q", streamed)
	}
	if res.OutputTokens != 42 {
		t.Errorf("OutputTokens = %d, want 42", res.OutputTokens)
	}
	if res.CacheReadHits != 50 {
		t.Errorf("CacheReadHits = %d, want 50", res.CacheReadHits)
	}
}

func TestStreamCompletion_EmptyTextErrors(t *testing.T) {
	srv := httptest.NewServer(cannedStreamHandler([]string{}))
	defer srv.Close()

	c, _ := NewClient(ClientOpts{APIKey: "sk-ant-stub", Model: "claude-sonnet-4-7", BaseURL: srv.URL})
	_, err := c.StreamCompletion(context.Background(), "sys", "user", 100, nil)
	if err == nil {
		t.Fatal("expected error on empty text")
	}
	if !strings.Contains(err.Error(), "no text") {
		t.Errorf("error message: %v", err)
	}
}
