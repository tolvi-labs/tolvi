package llm

import (
	"strings"
	"testing"
)

func TestEstimateTokens(t *testing.T) {
	if got := EstimateTokens("hello world"); got != 2 { // 11 chars / 4 = 2 (integer div)
		t.Errorf("EstimateTokens(11 chars) = %d, want 2", got)
	}
	if got := EstimateTokens(strings.Repeat("x", 400)); got != 100 {
		t.Errorf("EstimateTokens(400 chars) = %d, want 100", got)
	}
}

func TestGuardLevel(t *testing.T) {
	tests := []struct {
		tokens int
		want   GuardLevel
	}{
		{0, GuardOK},
		{50_000, GuardOK},
		{99_999, GuardOK},
		{100_000, GuardWarn},
		{150_000, GuardWarn},
		{180_000, GuardError},
		{200_000, GuardError},
	}
	for _, tt := range tests {
		if got := EvaluateGuard(tt.tokens); got != tt.want {
			t.Errorf("EvaluateGuard(%d) = %d, want %d", tt.tokens, got, tt.want)
		}
	}
}
