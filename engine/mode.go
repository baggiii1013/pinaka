package engine

// TestMode represents the type of typing test.
type TestMode int

const (
	ModeTime  TestMode = iota // Fixed duration, unlimited words
	ModeWords                 // Fixed word count
	ModeQuote                 // Type a full quote/sentence
	ModeZen                   // Freeform, no end condition
)

// ModeConfig holds the configuration for a typing test session.
type ModeConfig struct {
	Mode        TestMode
	Language    string // Language / dictionary identifier (e.g. "english", "code_go", "spanish")
	WordCount   int    // for ModeWords: 10/25/50/100
	TimeLimit   int    // for ModeTime: 15/30/60/120 seconds
	QuoteText   string // for ModeQuote: the full quote
	QuoteSource string // for ModeQuote: author or origin
	Punctuation bool
	Numbers     bool
}

// WordOptions are the available word counts for word mode.
var WordOptions = []int{10, 25, 50, 100}

// TimeOptions are the available time limits for time mode.
var TimeOptions = []int{15, 30, 60, 120}

// DefaultConfig returns a sensible default (words mode, 25 words, english language).
func DefaultConfig() ModeConfig {
	return ModeConfig{
		Mode:      ModeWords,
		Language:  "english",
		WordCount: 25,
	}
}
