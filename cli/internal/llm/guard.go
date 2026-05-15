package llm

// Token-estimation heuristic + thresholds per design spec §5.

const (
	// charsPerToken is the cheap English-text approximation; real-world
	// average for Claude is ~3.5, so dividing by 4 systematically
	// over-counts by ~15% — that's the safety margin against the API limit.
	charsPerToken = 4

	WarnThreshold  = 100_000
	ErrorThreshold = 180_000
)

// GuardLevel is the result of EvaluateGuard.
type GuardLevel int

const (
	GuardOK GuardLevel = iota
	GuardWarn
	GuardError
)

// EstimateTokens returns the approximate token count of s using the
// chars/4 heuristic. Faster + deterministic than calling the API's
// count_tokens endpoint; over-counts on average.
func EstimateTokens(s string) int {
	return len(s) / charsPerToken
}

// EvaluateGuard classifies a token count into OK / Warn / Error per the
// thresholds in the design spec.
func EvaluateGuard(tokens int) GuardLevel {
	switch {
	case tokens >= ErrorThreshold:
		return GuardError
	case tokens >= WarnThreshold:
		return GuardWarn
	}
	return GuardOK
}
