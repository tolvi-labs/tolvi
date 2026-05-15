package llm

import (
	"strings"
	"testing"
)

func TestNewClient_NoAPIKeyError(t *testing.T) {
	_, err := NewClient(ClientOpts{})
	if err == nil {
		t.Fatal("expected error when no API key")
	}
	if !strings.Contains(err.Error(), "API key") {
		t.Errorf("error message lacks 'API key': %v", err)
	}
}

func TestNewClient_WithKey(t *testing.T) {
	c, err := NewClient(ClientOpts{APIKey: "sk-ant-stub", Model: "claude-sonnet-4-7"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c == nil {
		t.Fatal("nil client returned")
	}
	if c.Model != "claude-sonnet-4-7" {
		t.Errorf("model = %q", c.Model)
	}
}

func TestNewClient_DefaultModel(t *testing.T) {
	c, err := NewClient(ClientOpts{APIKey: "sk-ant-stub"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.Model == "" {
		t.Error("default model not set")
	}
}
