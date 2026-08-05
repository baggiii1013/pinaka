package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"pinakatype.sh/engine"
)

// Screen represents which view is active.
type Screen int

const (
	ScreenTyping  Screen = iota
	ScreenResults
)

// tickMsg fires once per second while the test is in progress.
type tickMsg time.Time

// Model represents the state of our TUI
type Model struct {
	engine       *engine.TypingState
	styles       Styles
	width        int
	height       int
	screen       Screen
	modeSelector ModeSelector
	langPicker   *LangPicker
	ticking      bool
}

// NewModel initializes the TUI
func NewModel() tea.Model {
	ms := NewModeSelector()
	config := ms.ToConfig()
	return Model{
		engine:       engine.NewTypingStateWithConfig(config),
		styles:       DefaultStyles(),
		screen:       ScreenTyping,
		modeSelector: ms,
	}
}

// Init runs initially when the app starts
func (m Model) Init() tea.Cmd {
	return nil
}

func tickEverySecond() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles incoming events (like keypresses)
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		if m.engine.Status == "InProgress" {
			m.engine.RecordSnapshot()

			// Check time-up for time mode
			if m.engine.IsTimeUp() {
				m.engine.Status = "Finished"
				m.engine.EndTime = time.Now()
				m.screen = ScreenResults
				m.ticking = false
				return m, nil
			}

			return m, tickEverySecond()
		}
		m.ticking = false

	case tea.KeyMsg:
		switch m.screen {
		case ScreenResults:
			return m.updateResults(msg)
		case ScreenTyping:
			return m.updateTyping(msg)
		}
	}
	return m, nil
}

// updateResults handles input on the results screen.
func (m Model) updateResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		return m, tea.Quit
	case "tab":
		return m.restart()
	}
	return m, nil
}

// updateTyping handles input on the typing screen.
func (m Model) updateTyping(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If language picker modal is open, delegate input to it
	if m.langPicker != nil {
		applied, closed := m.langPicker.Update(msg)
		if closed {
			if applied {
				m.modeSelector.Language = m.langPicker.SelectedLang
				config := m.modeSelector.ToConfig()
				m.engine = engine.NewTypingStateWithConfig(config)
			}
			m.langPicker = nil
		}
		return m, nil
	}

	key := msg.String()

	switch key {
	case "esc", "ctrl+c":
		return m, tea.Quit

	case "tab":
		return m.restart()

	case "backspace", "ctrl+h":
		m.engine.HandleBackspace()
		return m, nil

	default:
		// Before test starts, check modal triggers or delegate to mode selector
		if m.engine.Status == "Waiting" {
			if key == "l" || key == "d" {
				lp := NewLangPicker(m.modeSelector.Language)
				m.langPicker = &lp
				return m, nil
			}

			if m.modeSelector.HandleKey(key) {
				// Mode changed — regenerate test
				config := m.modeSelector.ToConfig()
				m.engine = engine.NewTypingStateWithConfig(config)
				return m, nil
			}
		}

		// Ensure it's a typed character and not empty/navigational
		if len(msg.Runes) > 0 {
			wasWaiting := m.engine.Status == "Waiting"
			m.engine.HandleKey(msg.Runes[0])

			// Start ticking on first keypress
			if wasWaiting && m.engine.Status == "InProgress" && !m.ticking {
				m.ticking = true
				return m, tickEverySecond()
			}

			// Check if test just finished (word/quote mode)
			if m.engine.Status == "Finished" {
				m.screen = ScreenResults
				m.ticking = false
			}
		}
	}
	return m, nil
}

// restart resets the test with current mode settings.
func (m Model) restart() (tea.Model, tea.Cmd) {
	config := m.modeSelector.ToConfig()
	m.engine = engine.NewTypingStateWithConfig(config)
	m.screen = ScreenTyping
	m.ticking = false
	return m, nil
}

// View returns the string to be rendered in the terminal
func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	if m.langPicker != nil {
		return m.langPicker.View(m.styles, m.width, m.height)
	}

	switch m.screen {
	case ScreenResults:
		return renderResults(m.engine, m.styles, m.width, m.height)
	default:
		return m.renderTypingScreen()
	}
}

// renderTypingScreen renders the main typing test view.
func (m Model) renderTypingScreen() string {
	var sb strings.Builder

	// Mode selector bar (only fully visible when waiting)
	if m.engine.Status == "Waiting" {
		contentWidth := m.width - 20
		if contentWidth < 50 {
			contentWidth = 50
		} else if contentWidth > 120 {
			contentWidth = 120
		}
		sb.WriteString(m.modeSelector.View(m.styles, contentWidth))
		sb.WriteString("\n\n")
	}

	// Stats bar
	sb.WriteString(m.renderStatsBar())
	sb.WriteString("\n\n")

	// Main Typing View
	var textBlock strings.Builder

	for wordIdx, word := range m.engine.Words {
		isActive := (wordIdx == m.engine.ActiveIndex) && (m.engine.Status != "Finished")
		targetRunes := []rune(word.Text)
		inputRunes := []rune(word.Input)

		maxLen := len(targetRunes)
		if len(inputRunes) > maxLen {
			maxLen = len(inputRunes)
		}

		for i := 0; i < maxLen; i++ {
			isCursor := isActive && i == len(inputRunes)

			var charStr string
			var style lipgloss.Style

			if i < len(targetRunes) {
				charStr = string(targetRunes[i])
				if i < len(inputRunes) {
					if inputRunes[i] == targetRunes[i] {
						style = m.styles.CorrectText
					} else {
						style = m.styles.WrongText
					}
				} else {
					style = m.styles.NormalText
				}
			} else {
				charStr = string(inputRunes[i])
				style = m.styles.WrongText.Background(lipgloss.Color("#ca4754"))
			}

			if isCursor {
				style = m.styles.CursorText
			}

			textBlock.WriteString(style.Render(charStr))
		}

		// Cursor at end of word
		if isActive && len(inputRunes) >= len(targetRunes) && len(inputRunes) == maxLen {
			textBlock.WriteString(m.styles.CursorText.Render(" "))
		}

		// Space between words
		if wordIdx < len(m.engine.Words)-1 {
			if wordIdx < m.engine.ActiveIndex && word.State == 3 {
				textBlock.WriteString(m.styles.WrongText.Background(lipgloss.Color("#ca4754")).Render(" "))
			} else {
				textBlock.WriteString(m.styles.NormalText.Render(" "))
			}
		}
	}

	// Dynamic text wrapping
	width := m.width - 20
	if width < 50 {
		width = 50
	} else if width > 120 {
		width = 120
	}

	mainTextBox := lipgloss.NewStyle().Width(width).Render(textBlock.String())
	sb.WriteString(mainTextBox)
	sb.WriteString("\n\n")

	// Bottom help text
	if m.engine.Status == "Waiting" {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#737373")).Render(
			"l/d language  p punctuation  n numbers  1-4 mode  ←→ options  tab restart  esc quit",
		))
	} else {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#737373")).Render("tab to restart    esc to quit"))
	}

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		sb.String(),
	)
}

// renderStatsBar renders the live stats bar based on the current mode.
func (m Model) renderStatsBar() string {
	switch m.engine.Config.Mode {
	case engine.ModeTime:
		if m.engine.Status == "Waiting" {
			return m.styles.TimerText.Render(fmt.Sprintf("%d", m.engine.Config.TimeLimit))
		}
		remaining := m.engine.TimeRemaining()
		return m.styles.TimerText.Render(fmt.Sprintf("%d", remaining))

	case engine.ModeWords:
		wpm := m.engine.WPM()
		completed, total := m.engine.WordProgress()
		return m.styles.Stats.Render(fmt.Sprintf("%.0f WPM  %d/%d", wpm, completed, total))

	case engine.ModeQuote:
		wpm := m.engine.WPM()
		completed, total := m.engine.WordProgress()
		return m.styles.Stats.Render(fmt.Sprintf("%.0f WPM  %d/%d", wpm, completed, total))

	case engine.ModeZen:
		wpm := m.engine.WPM()
		return m.styles.Stats.Render(fmt.Sprintf("%.0f WPM", wpm))

	default:
		wpm := m.engine.WPM()
		acc := m.engine.Accuracy()
		return m.styles.Stats.Render(fmt.Sprintf("%.0f WPM | %.0f%% ACC", wpm, acc))
	}
}
