package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"pinakatype.sh/engine"
)

// sparklineBlocks are the Unicode block elements for sparkline rendering (8 levels).
var sparklineBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// renderResults produces the full results dashboard view.
func renderResults(eng *engine.TypingState, styles Styles, width, height int) string {
	var sb strings.Builder

	wpm := eng.WPM()
	rawWPM := eng.RawWPM()
	acc := eng.Accuracy()
	consistency := eng.Consistency()
	elapsed := eng.ElapsedTime()
	correct, incorrect, extra, missed := eng.CharStats()

	// 1. Big WPM number
	bigWPM := styles.BigWPM.Render(fmt.Sprintf("%.0f", wpm))
	wpmLabel := styles.ResultLabel.Render("wpm")
	sb.WriteString(lipgloss.JoinVertical(lipgloss.Center, bigWPM, wpmLabel))
	sb.WriteString("\n\n")

	// 2. Stats row
	statsItems := []string{
		styles.ResultLabel.Render("raw ") + styles.ResultValue.Render(fmt.Sprintf("%.0f", rawWPM)),
		styles.ResultLabel.Render("accuracy ") + styles.ResultValue.Render(fmt.Sprintf("%.0f%%", acc)),
		styles.ResultLabel.Render("consistency ") + styles.ResultValue.Render(fmt.Sprintf("%.0f%%", consistency)),
		styles.ResultLabel.Render("time ") + styles.ResultValue.Render(fmt.Sprintf("%.0fs", elapsed)),
	}
	statsRow := strings.Join(statsItems, styles.Separator.Render("  |  "))
	sb.WriteString(statsRow)
	sb.WriteString("\n\n")

	// 3. Character stats
	charStatsStr := strings.Join([]string{
		styles.CorrectText.Render(fmt.Sprintf("correct: %d", correct)),
		styles.WrongText.Render(fmt.Sprintf("incorrect: %d", incorrect)),
		styles.WrongText.Render(fmt.Sprintf("extra: %d", extra)),
		styles.InactiveOption.Render(fmt.Sprintf("missed: %d", missed)),
	}, "  ")
	sb.WriteString(charStatsStr)
	sb.WriteString("\n\n")

	// 4. Test type
	testType := renderTestType(eng.Config, styles)
	sb.WriteString(testType)
	sb.WriteString("\n\n")

	// 5. WPM sparkline
	if len(eng.WPMHistory) > 1 {
		chartWidth := width - 20
		if chartWidth < 20 {
			chartWidth = 20
		}
		if chartWidth > 80 {
			chartWidth = 80
		}
		sparkline := renderSparkline(eng.WPMHistory, styles.ChartBar, chartWidth)
		sb.WriteString(sparkline)
		sb.WriteString("\n\n")
	}

	// 6. Help
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#737373")).Render("tab to restart    esc to quit"))

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		sb.String(),
	)
}

// renderTestType shows what mode/config was used.
func renderTestType(config engine.ModeConfig, styles Styles) string {
	var parts []string

	displayLang := config.Language
	if displayLang == "" {
		displayLang = "english"
	}
	parts = append(parts, displayLang)

	switch config.Mode {
	case engine.ModeTime:
		parts = append(parts, fmt.Sprintf("time %d", config.TimeLimit))
	case engine.ModeWords:
		parts = append(parts, fmt.Sprintf("words %d", config.WordCount))
	case engine.ModeQuote:
		parts = append(parts, "quote")
	case engine.ModeZen:
		parts = append(parts, "zen")
	}

	if config.Punctuation {
		parts = append(parts, "punctuation")
	}
	if config.Numbers {
		parts = append(parts, "numbers")
	}

	return styles.InactiveOption.Render(strings.Join(parts, "  "))
}

// renderSparkline creates a single-row sparkline chart from WPM history values.
func renderSparkline(values []float64, style lipgloss.Style, maxWidth int) string {
	if len(values) == 0 {
		return ""
	}

	// Downsample if needed
	data := values
	if len(data) > maxWidth {
		data = downsample(data, maxWidth)
	}

	// Find min and max for scaling
	minVal, maxVal := data[0], data[0]
	for _, v := range data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	// Build sparkline
	var sb strings.Builder
	valRange := maxVal - minVal
	if valRange == 0 {
		valRange = 1 // avoid divide by zero
	}

	for _, v := range data {
		normalized := (v - minVal) / valRange
		idx := int(math.Round(normalized * float64(len(sparklineBlocks)-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparklineBlocks) {
			idx = len(sparklineBlocks) - 1
		}
		sb.WriteRune(sparklineBlocks[idx])
	}

	return style.Render(sb.String())
}

// downsample reduces a slice of values to the target length by averaging adjacent values.
func downsample(values []float64, targetLen int) []float64 {
	if targetLen >= len(values) {
		return values
	}

	result := make([]float64, targetLen)
	bucketSize := float64(len(values)) / float64(targetLen)

	for i := 0; i < targetLen; i++ {
		start := int(float64(i) * bucketSize)
		end := int(float64(i+1) * bucketSize)
		if end > len(values) {
			end = len(values)
		}

		var sum float64
		count := 0
		for j := start; j < end; j++ {
			sum += values[j]
			count++
		}
		if count > 0 {
			result[i] = sum / float64(count)
		}
	}

	return result
}
