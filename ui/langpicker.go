package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"pinakatype.sh/engine"
)

// LangTab represents a category tab in the language picker modal.
type LangTab int

const (
	TabPopular LangTab = iota
	TabCode
	TabAll
)

// LangPicker is an interactive modal for searching and selecting word lists / languages.
type LangPicker struct {
	ActiveTab    LangTab
	SearchQuery  string
	SelectedIndex int
	ScrollOffset int
	Filtered     []string
	SelectedLang string
	applied      bool
	closed       bool
}

// NewLangPicker initializes a new language picker modal.
func NewLangPicker(currentLang string) LangPicker {
	if currentLang == "" {
		currentLang = "english"
	}
	lp := LangPicker{
		ActiveTab:    TabPopular,
		SelectedLang: currentLang,
	}
	lp.refreshFiltered()
	lp.locateSelected()
	return lp
}

func (lp *LangPicker) locateSelected() {
	for i, item := range lp.Filtered {
		if item == lp.SelectedLang {
			lp.SelectedIndex = i
			if lp.SelectedIndex >= 10 {
				lp.ScrollOffset = lp.SelectedIndex - 4
			}
			return
		}
	}
	lp.SelectedIndex = 0
	lp.ScrollOffset = 0
}

func (lp *LangPicker) getSourceList() []string {
	al := engine.GetAssetLoader()
	switch lp.ActiveTab {
	case TabPopular:
		return engine.CuratedPopularLanguages()
	case TabCode:
		return engine.CuratedCodeLanguages()
	case TabAll:
		return al.ListLanguages()
	default:
		return al.ListLanguages()
	}
}

func (lp *LangPicker) refreshFiltered() {
	sources := lp.getSourceList()
	q := strings.ToLower(strings.TrimSpace(lp.SearchQuery))

	if q == "" {
		lp.Filtered = sources
	} else {
		var res []string
		for _, item := range sources {
			if strings.Contains(strings.ToLower(item), q) || strings.Contains(strings.ToLower(engine.FormatLanguageName(item)), q) {
				res = append(res, item)
			}
		}
		lp.Filtered = res
	}

	if lp.SelectedIndex >= len(lp.Filtered) {
		lp.SelectedIndex = len(lp.Filtered) - 1
	}
	if lp.SelectedIndex < 0 {
		lp.SelectedIndex = 0
	}
	lp.adjustScroll()
}

func (lp *LangPicker) adjustScroll() {
	visibleCount := 10
	if lp.SelectedIndex < lp.ScrollOffset {
		lp.ScrollOffset = lp.SelectedIndex
	}
	if lp.SelectedIndex >= lp.ScrollOffset+visibleCount {
		lp.ScrollOffset = lp.SelectedIndex - visibleCount + 1
	}
	if lp.ScrollOffset < 0 {
		lp.ScrollOffset = 0
	}
}

// Update handles keyboard navigation and search inside the modal.
func (lp *LangPicker) Update(msg tea.Msg) (bool, bool) {
	// returns (applied, closed)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			lp.closed = true
			return false, true

		case "enter":
			if len(lp.Filtered) > 0 && lp.SelectedIndex < len(lp.Filtered) {
				lp.SelectedLang = lp.Filtered[lp.SelectedIndex]
				lp.applied = true
				lp.closed = true
				return true, true
			}
			lp.closed = true
			return false, true

		case "tab":
			lp.ActiveTab = (lp.ActiveTab + 1) % 3
			lp.refreshFiltered()
			lp.SelectedIndex = 0
			lp.ScrollOffset = 0
			return false, false

		case "shift+tab":
			lp.ActiveTab = (lp.ActiveTab + 2) % 3
			lp.refreshFiltered()
			lp.SelectedIndex = 0
			lp.ScrollOffset = 0
			return false, false

		case "up", "ctrl+p":
			if lp.SelectedIndex > 0 {
				lp.SelectedIndex--
				lp.adjustScroll()
			}
			return false, false

		case "down", "ctrl+n":
			if lp.SelectedIndex < len(lp.Filtered)-1 {
				lp.SelectedIndex++
				lp.adjustScroll()
			}
			return false, false

		case "backspace", "ctrl+h":
			if len(lp.SearchQuery) > 0 {
				runes := []rune(lp.SearchQuery)
				lp.SearchQuery = string(runes[:len(runes)-1])
				lp.refreshFiltered()
			}
			return false, false

		default:
			if len(msg.Runes) > 0 {
				r := msg.Runes[0]
				// Only accept printable search characters
				if r >= 32 && r != 127 {
					lp.SearchQuery += string(r)
					lp.refreshFiltered()
					return false, false
				}
			}
		}
	}
	return false, false
}

// View overlays the language picker modal.
func (lp *LangPicker) View(styles Styles, screenWidth, screenHeight int) string {
	modalWidth := 60
	if modalWidth > screenWidth-4 {
		modalWidth = screenWidth - 4
	}

	var sb strings.Builder

	// Header / Tabs
	tabs := []struct {
		tab  LangTab
		name string
	}{
		{TabPopular, "Popular"},
		{TabCode, "Code"},
		{TabAll, "All Languages"},
	}

	var tabHeaders []string
	for _, t := range tabs {
		if lp.ActiveTab == t.tab {
			tabHeaders = append(tabHeaders, styles.ActiveOption.Render(fmt.Sprintf("[%s]", t.name)))
		} else {
			tabHeaders = append(tabHeaders, styles.InactiveOption.Render(fmt.Sprintf(" %s ", t.name)))
		}
	}
	sb.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(modalWidth).Render(strings.Join(tabHeaders, "   ")))
	sb.WriteString("\n\n")

	// Search Box
	searchBox := fmt.Sprintf("Search: %s", lp.SearchQuery)
	cursor := styles.Stats.Render("█")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#d1d0c5")).Padding(0, 1).Render(searchBox + cursor))
	sb.WriteString("\n")
	sb.WriteString(styles.Separator.Render(strings.Repeat("─", modalWidth)))
	sb.WriteString("\n")

	// List
	visibleRows := 10
	if len(lp.Filtered) == 0 {
		sb.WriteString(styles.InactiveOption.Padding(2, 2).Render("No matching languages or word lists found."))
		sb.WriteString("\n")
	} else {
		endIdx := lp.ScrollOffset + visibleRows
		if endIdx > len(lp.Filtered) {
			endIdx = len(lp.Filtered)
		}

		for i := lp.ScrollOffset; i < endIdx; i++ {
			langKey := lp.Filtered[i]
			title := engine.FormatLanguageName(langKey)
			isCurrent := (langKey == lp.SelectedLang)
			isSelected := (i == lp.SelectedIndex)

			prefix := "  "
			if isSelected {
				prefix = "> "
			}

			tag := ""
			if isCurrent {
				tag = " (active)"
			}

			lineText := fmt.Sprintf("%s%-34s %-12s%s", prefix, title, fmt.Sprintf("(%s)", langKey), tag)
			if len(lineText) > modalWidth-2 {
				lineText = lineText[:modalWidth-2]
			}

			if isSelected {
				sb.WriteString(styles.ActiveOption.Bold(true).Render(lineText))
			} else if isCurrent {
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#d1d0c5")).Render(lineText))
			} else {
				sb.WriteString(styles.InactiveOption.Render(lineText))
			}
			sb.WriteString("\n")
		}

		// Fill blank lines if list is shorter than visibleRows
		for i := endIdx - lp.ScrollOffset; i < visibleRows; i++ {
			sb.WriteString("\n")
		}
	}

	sb.WriteString(styles.Separator.Render(strings.Repeat("─", modalWidth)))
	sb.WriteString("\n")

	// Footer hints
	footer := styles.InactiveOption.Render("[Tab] Category  [↑/↓] Navigate  [Enter] Select  [Esc] Cancel")
	sb.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(modalWidth).Render(footer))

	// Modal Border Box
	modalBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#e2b714")).
		Background(lipgloss.Color("#1e1e1e")).
		Padding(1, 1).
		Width(modalWidth).
		Render(sb.String())

	// Center on screen
	return lipgloss.Place(
		screenWidth,
		screenHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalBox,
	)
}
