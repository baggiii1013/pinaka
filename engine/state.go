package engine

import (
	"math"
	"time"
)

// Word represents a single word in the typing test.
type Word struct {
	Text  string
	Input string // User's input so far
	State int    // 0 = pending, 1 = active, 2 = correct, 3 = incorrect
}

// TypingState encapsulates the core typing test logic and progress.
type TypingState struct {
	Words       []Word
	ActiveIndex int
	StartTime   time.Time
	EndTime     time.Time
	Status      string // "Waiting", "InProgress", "Finished"
	Config      ModeConfig
	WPMHistory  []float64 // per-second WPM snapshots
	ElapsedSec  int       // seconds elapsed
}

// NewTypingState creates a fresh test session (legacy, uses default word count).
func NewTypingState(wordCount int) *TypingState {
	return NewTypingStateWithConfig(ModeConfig{
		Mode:      ModeWords,
		WordCount: wordCount,
	})
}

// NewTypingStateWithConfig creates a test session from a mode config.
func NewTypingStateWithConfig(config ModeConfig) *TypingState {
	if config.Language == "" {
		config.Language = "english"
	}

	var targetWords []string

	switch config.Mode {
	case ModeTime:
		// Generate enough words for fast typists over the full duration
		targetWords = GenerateWordsWithOptionsForLanguage(config.Language, 200, config.Punctuation, config.Numbers)
	case ModeWords:
		targetWords = GenerateWordsWithOptionsForLanguage(config.Language, config.WordCount, config.Punctuation, config.Numbers)
	case ModeQuote:
		if config.QuoteText == "" {
			q := GetRandomQuoteForLanguage(config.Language)
			config.QuoteText = q.Text
			config.QuoteSource = q.Source
		}
		targetWords = SplitQuote(config.QuoteText)
	case ModeZen:
		targetWords = GenerateWordsWithOptionsForLanguage(config.Language, 100, config.Punctuation, config.Numbers)
	default:
		targetWords = GenerateWordsWithOptionsForLanguage(config.Language, 25, config.Punctuation, config.Numbers)
	}

	var words []Word
	for i, w := range targetWords {
		state := 0
		if i == 0 {
			state = 1
		}
		words = append(words, Word{
			Text:  w,
			Input: "",
			State: state,
		})
	}

	return &TypingState{
		Words:       words,
		ActiveIndex: 0,
		Status:      "Waiting",
		Config:      config,
		WPMHistory:  []float64{},
	}
}

// HandleKey processes a successfully typed character.
func (s *TypingState) HandleKey(r rune) {
	if s.Status == "Finished" {
		return
	}

	if s.Status == "Waiting" {
		s.Status = "InProgress"
		s.StartTime = time.Now()
	}

	if r == ' ' {
		// Space typed: transition word state
		if len(s.Words[s.ActiveIndex].Input) == 0 {
			return // Do nothing if input length is 0
		}

		// Evaluate correctness
		if s.Words[s.ActiveIndex].Input == s.Words[s.ActiveIndex].Text {
			s.Words[s.ActiveIndex].State = 2 // correct
		} else {
			s.Words[s.ActiveIndex].State = 3 // incorrect
		}

		s.ActiveIndex++

		// Check if at the end of the word list
		if s.ActiveIndex >= len(s.Words) {
			switch s.Config.Mode {
			case ModeTime, ModeZen:
				// Generate more words — don't finish
				s.extendWords()
			default:
				// Words/Quote mode: finish the test
				s.Status = "Finished"
				s.EndTime = time.Now()
				return
			}
		}

		// Mark next word as active
		s.Words[s.ActiveIndex].State = 1
	} else {
		// Regular character typed: handle appending and over-typing constraint
		word := &s.Words[s.ActiveIndex]
		if len([]rune(word.Input)) < len([]rune(word.Text))+5 {
			word.Input += string(r)
		}
	}
}

// extendWords adds more words for time/zen modes when the user reaches the end.
func (s *TypingState) extendWords() {
	lang := s.Config.Language
	if lang == "" {
		lang = "english"
	}
	newWords := GenerateWordsWithOptionsForLanguage(lang, 50, s.Config.Punctuation, s.Config.Numbers)
	for _, w := range newWords {
		s.Words = append(s.Words, Word{
			Text:  w,
			Input: "",
			State: 0,
		})
	}
}

// HandleBackspace safely removes the last character entered, with rewind rules.
func (s *TypingState) HandleBackspace() {
	if s.Status == "Finished" {
		return
	}

	word := &s.Words[s.ActiveIndex]
	runes := []rune(word.Input)

	if len(runes) > 0 {
		// Middle of a word: delete last character
		word.Input = string(runes[:len(runes)-1])
	} else if len(runes) == 0 && s.ActiveIndex > 0 {
		// Handled empty word logic: attempt a rewind
		prevWord := &s.Words[s.ActiveIndex-1]
		if prevWord.State == 3 {
			// Rewind is allowed for incorrect words
			word.State = 0
			s.ActiveIndex--
			s.Words[s.ActiveIndex].State = 1
		}
	}
}

// WPM calculates strict Words Per Minute based on standardized lengths (5 chars = 1 word).
func (s *TypingState) WPM() float64 {
	if s.Status == "Waiting" {
		return 0.0
	}

	var totalTime time.Duration
	if s.Status == "Finished" {
		totalTime = s.EndTime.Sub(s.StartTime)
	} else {
		totalTime = time.Since(s.StartTime)
	}

	// Avoid divide-by-zero on immediate measurements
	if totalTime.Seconds() < 0.1 {
		return 0.0
	}

	minutes := totalTime.Minutes()
	var correctChars int

	// Count only correctly typed characters
	for i, w := range s.Words {
		if i < s.ActiveIndex {
			inputRunes := []rune(w.Input)
			targetRunes := []rune(w.Text)
			for j, r := range inputRunes {
				if j < len(targetRunes) && r == targetRunes[j] {
					correctChars++
				}
			}
			correctChars++ // space
		} else if i == s.ActiveIndex {
			inputRunes := []rune(w.Input)
			targetRunes := []rune(w.Text)
			for j, r := range inputRunes {
				if j < len(targetRunes) && r == targetRunes[j] {
					correctChars++
				}
			}
		}
	}

	return (float64(correctChars) / 5.0) / minutes
}

// RawWPM calculates WPM including all keystrokes (no error deduction).
func (s *TypingState) RawWPM() float64 {
	if s.Status == "Waiting" {
		return 0.0
	}

	var totalTime time.Duration
	if s.Status == "Finished" {
		totalTime = s.EndTime.Sub(s.StartTime)
	} else {
		totalTime = time.Since(s.StartTime)
	}

	if totalTime.Seconds() < 0.1 {
		return 0.0
	}

	minutes := totalTime.Minutes()
	var charsTyped int

	for i, w := range s.Words {
		if i < s.ActiveIndex {
			charsTyped += len([]rune(w.Input)) + 1
		} else if i == s.ActiveIndex {
			charsTyped += len([]rune(w.Input))
		}
	}

	return (float64(charsTyped) / 5.0) / minutes
}

// Accuracy computes the percentage of correctly typed characters.
func (s *TypingState) Accuracy() float64 {
	var correct, total int

	for i, w := range s.Words {
		if i <= s.ActiveIndex {
			inputRunes := []rune(w.Input)
			targetRunes := []rune(w.Text)

			for j, r := range inputRunes {
				total++
				if j < len(targetRunes) && r == targetRunes[j] {
					correct++
				}
			}

			if i < s.ActiveIndex {
				total++
				correct++ // Implicitly correct since they successfully pressed space
			}
		}
	}

	if total == 0 {
		return 100.0
	}

	return (float64(correct) / float64(total)) * 100.0
}

// RecordSnapshot saves the current WPM to history (called once per second by UI tick).
func (s *TypingState) RecordSnapshot() {
	s.ElapsedSec++
	s.WPMHistory = append(s.WPMHistory, s.WPM())
}

// TimeRemaining returns seconds left for time mode, -1 for other modes.
func (s *TypingState) TimeRemaining() int {
	if s.Config.Mode != ModeTime {
		return -1
	}
	remaining := s.Config.TimeLimit - s.ElapsedSec
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsTimeUp checks if the elapsed time has exceeded the time limit.
func (s *TypingState) IsTimeUp() bool {
	if s.Config.Mode != ModeTime {
		return false
	}
	return s.ElapsedSec >= s.Config.TimeLimit
}

// WordProgress returns (completed, total) for word/quote modes.
func (s *TypingState) WordProgress() (int, int) {
	return s.ActiveIndex, len(s.Words)
}

// Consistency calculates typing consistency as a percentage (100 = perfectly consistent).
func (s *TypingState) Consistency() float64 {
	if len(s.WPMHistory) < 2 {
		return 100.0
	}

	// Calculate mean
	var sum float64
	for _, v := range s.WPMHistory {
		sum += v
	}
	mean := sum / float64(len(s.WPMHistory))

	if mean == 0 {
		return 0
	}

	// Calculate standard deviation
	var varianceSum float64
	for _, v := range s.WPMHistory {
		diff := v - mean
		varianceSum += diff * diff
	}
	stdev := math.Sqrt(varianceSum / float64(len(s.WPMHistory)))

	// Coefficient of variation, inverted to a 0-100 scale
	cv := (stdev / mean) * 100
	consistency := 100 - cv
	if consistency < 0 {
		consistency = 0
	}
	return consistency
}

// CharStats returns (correct, incorrect, extra, missed) character counts.
func (s *TypingState) CharStats() (int, int, int, int) {
	var correct, incorrect, extra, missed int

	limit := s.ActiveIndex
	if s.Status == "Finished" {
		limit = len(s.Words)
	}

	for i := 0; i < limit; i++ {
		w := s.Words[i]
		inputRunes := []rune(w.Input)
		targetRunes := []rune(w.Text)

		minLen := len(inputRunes)
		if len(targetRunes) < minLen {
			minLen = len(targetRunes)
		}

		for j := 0; j < minLen; j++ {
			if inputRunes[j] == targetRunes[j] {
				correct++
			} else {
				incorrect++
			}
		}

		if len(inputRunes) > len(targetRunes) {
			extra += len(inputRunes) - len(targetRunes)
		}
		if len(targetRunes) > len(inputRunes) {
			missed += len(targetRunes) - len(inputRunes)
		}
	}

	return correct, incorrect, extra, missed
}

// ElapsedTime returns the total test duration in seconds.
func (s *TypingState) ElapsedTime() float64 {
	if s.Status == "Waiting" {
		return 0
	}
	if s.Status == "Finished" {
		return s.EndTime.Sub(s.StartTime).Seconds()
	}
	return time.Since(s.StartTime).Seconds()
}
