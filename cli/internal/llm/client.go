// Package llm wraps the Anthropic Go SDK with the prompt-caching and
// streaming patterns specific to Tolvi's CAG flow.
package llm

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClientOpts configures the LLM client.
type ClientOpts struct {
	APIKey  string // required
	Model   string // defaults to DefaultModel
	BaseURL string // optional; for tests with a fake transport
}

// DefaultModel is the model used when no override is supplied.
const DefaultModel = "claude-sonnet-4-7"

// Client wraps the Anthropic SDK with prompt-caching defaults.
type Client struct {
	Model string

	inner *anthropic.Client
}

// NewClient validates options and constructs a Client.
func NewClient(opts ClientOpts) (*Client, error) {
	if opts.APIKey == "" {
		return nil, fmt.Errorf("missing Anthropic API key — set ANTHROPIC_API_KEY or write ~/.config/tolvi/config.yaml")
	}
	if opts.Model == "" {
		opts.Model = DefaultModel
	}
	sdkOpts := []option.RequestOption{option.WithAPIKey(opts.APIKey)}
	if opts.BaseURL != "" {
		sdkOpts = append(sdkOpts, option.WithBaseURL(opts.BaseURL))
	}
	inner := anthropic.NewClient(sdkOpts...)
	return &Client{Model: opts.Model, inner: &inner}, nil
}

// StreamResult holds the assembled stream output + token totals.
type StreamResult struct {
	Text          string
	Model         string
	InputTokens   int64
	OutputTokens  int64
	CacheReadHits int64
	StopReason    string
}

// StreamCompletion executes a single streaming completion.
//
// systemPrompt is sent as a cached text block (cache_control: ephemeral).
// userMessage is the user's question, uncached.
// maxTokens caps the output (default 1024 if zero).
// onText is called for each incremental text chunk; the caller is
// responsible for printing to stdout if streaming UX is desired.
//
// Returns the assembled text + token usage after the stream completes.
// An empty text result is an error.
func (c *Client) StreamCompletion(
	ctx context.Context,
	systemPrompt, userMessage string,
	maxTokens int64,
	onText func(string),
) (StreamResult, error) {
	if maxTokens == 0 {
		maxTokens = 1024
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(c.Model),
		MaxTokens: maxTokens,
		System: []anthropic.TextBlockParam{{
			Text:         systemPrompt,
			CacheControl: anthropic.NewCacheControlEphemeralParam(),
		}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)),
		},
	}

	stream := c.inner.Messages.NewStreaming(ctx, params)
	defer stream.Close()

	var (
		assembled    []byte
		result       StreamResult
		gotTextBlock bool
	)
	result.Model = c.Model

	for stream.Next() {
		event := stream.Current()
		switch event.Type {
		case "message_start":
			start := event.AsMessageStart()
			result.InputTokens = start.Message.Usage.InputTokens
			result.CacheReadHits = start.Message.Usage.CacheReadInputTokens
		case "content_block_delta":
			delta := event.AsContentBlockDelta()
			if delta.Delta.Type == "text_delta" {
				text := delta.Delta.Text
				if text != "" {
					gotTextBlock = true
					assembled = append(assembled, text...)
					if onText != nil {
						onText(text)
					}
				}
			}
		case "message_delta":
			md := event.AsMessageDelta()
			result.OutputTokens = md.Usage.OutputTokens
			result.StopReason = string(md.Delta.StopReason)
		}
	}

	if err := stream.Err(); err != nil {
		return result, fmt.Errorf("anthropic stream: %w", err)
	}

	result.Text = string(assembled)

	if !gotTextBlock || result.Text == "" {
		return result, fmt.Errorf("anthropic returned no text (stop_reason=%q)", result.StopReason)
	}

	return result, nil
}
