package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"pinakatype.sh/engine"
)

// ModeSelector handles mode selection UI and state.
type ModeSelector struct {
	Mode          engine.TestMode
	Language      string
	WordOptionIdx int // index into engine.WordOptions
	TimeOptionIdx int // index into engine.TimeOptions
	Punctuation   bool
	Numbers       bool
}

// NewModeSelector returns a mode selector with sensible defaults.
func NewModeSelector() ModeSelector {
	return ModeSelector{
		Mode:          engine.ModeWords,
		Language:      "english",
		WordOptionIdx: 1, // 25 words
		TimeOptionIdx: 1, // 30 seconds
	}
}

// ToConfig converts the current selection into an engine.ModeConfig.
func (ms *ModeSelector) ToConfig() engine.ModeConfig {
	if ms.Language == "" {
		ms.Language = "english"
	}
	config := engine.ModeConfig{
		Mode:        ms.Mode,
		Language:    ms.Language,
		Punctuation: ms.Punctuation,
		Numbers:     ms.Numbers,
	}

	switch ms.Mode {
	case engine.ModeTime:
		config.TimeLimit = engine.TimeOptions[ms.TimeOptionIdx]
	case engine.ModeWords:
		config.WordCount = engine.WordOptions[ms.WordOptionIdx]
	case engine.ModeQuote:
		q := engine.GetRandomQuoteForLanguage(ms.Language)
		config.QuoteText = q.Text
		config.QuoteSource = q.Source
	}

	return config
}

// HandleKey processes a key event for the mode selector. Returns true if handled.
func (ms *ModeSelector) HandleKey(key string) bool {
	switch key {
	case "p":
		ms.Punctuation = !ms.Punctuation
		return true
	case "n":
		ms.Numbers = !ms.Numbers
		return true
	case "left":
		ms.cycleSubOption(-1)
		return true
	case "right":
		ms.cycleSubOption(1)
		return true
	case "1":
		ms.Mode = engine.ModeTime
		return true
	case "2":
		ms.Mode = engine.ModeWords
		return true
	case "3":
		ms.Mode = engine.ModeQuote
		return true
	case "4":
		ms.Mode = engine.ModeZen
		return true
	}
	return false
}

func (ms *ModeSelector) cycleSubOption(delta int) {
	switch ms.Mode {
	case engine.ModeTime:
		ms.TimeOptionIdx = (ms.TimeOptionIdx + delta + len(engine.TimeOptions)) % len(engine.TimeOptions)
	case engine.ModeWords:
		ms.WordOptionIdx = (ms.WordOptionIdx + delta + len(engine.WordOptions)) % len(engine.WordOptions)
	}
}

// View renders the mode selection bar.
func (ms *ModeSelector) View(styles Styles, width int) string {
	var parts []string

	// Toggles
	if ms.Punctuation {
		parts = append(parts, styles.ActiveOption.Render("@ punctuation"))
	} else {
		parts = append(parts, styles.InactiveOption.Render("@ punctuation"))
	}

	if ms.Numbers {
		parts = append(parts, styles.ActiveOption.Render("# numbers"))
	} else {
		parts = append(parts, styles.InactiveOption.Render("# numbers"))
	}

	parts = append(parts, styles.Separator.Render("|"))

	// Language / Dictionary indicator
	displayLang := ms.Language
	if displayLang == "" {
		displayLang = "english"
	}
	parts = append(parts, styles.InactiveOption.Render(fmt.Sprintf("🌐 %s", displayLang)))

	// Mode selection
	modes := []struct {
		mode engine.TestMode
		name string
		key  string
	}{
		{engine.ModeTime, "time", "1"},
		{engine.ModeWords, "words", "2"},
		{engine.ModeQuote, "quote", "3"},
		{engine.ModeZen, "zen", "4"},
	}

	for _, m := range modes {
		if ms.Mode == m.mode {
			parts = append(parts, styles.ActiveOption.Render(m.name))
		} else {
			parts = append(parts, styles.InactiveOption.Render(m.name))
		}
	}

	// Sub-options (only for time and words modes)
	switch ms.Mode {
	case engine.ModeTime:
		parts = append(parts, styles.Separator.Render("|"))
		for i, opt := range engine.TimeOptions {
			label := fmt.Sprintf("%d", opt)
			if i == ms.TimeOptionIdx {
				parts = append(parts, styles.ActiveOption.Render(label))
			} else {
				parts = append(parts, styles.InactiveOption.Render(label))
			}
		}
	case engine.ModeWords:
		parts = append(parts, styles.Separator.Render("|"))
		for i, opt := range engine.WordOptions {
			label := fmt.Sprintf("%d", opt)
			if i == ms.WordOptionIdx {
				parts = append(parts, styles.ActiveOption.Render(label))
			} else {
				parts = append(parts, styles.InactiveOption.Render(label))
			}
		}
	}

	bar := strings.Join(parts, "  ")

	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(bar)
}
