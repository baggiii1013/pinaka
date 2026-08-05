package engine

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode"
)

var commonWords = []string{
	"the", "be", "to", "of", "and", "a", "in", "that", "have", "I",
	"it", "for", "not", "on", "with", "he", "as", "you", "do", "at",
	"this", "but", "his", "by", "from", "they", "we", "say", "her", "she",
	"or", "an", "will", "my", "one", "all", "would", "there", "their", "what",
	"so", "up", "out", "if", "about", "who", "get", "which", "go", "me",
	"when", "make", "can", "like", "time", "no", "just", "him", "know", "take",
	"people", "into", "year", "your", "good", "some", "could", "them", "see", "other",
	"than", "then", "now", "look", "only", "come", "its", "over", "think", "also",
	"back", "after", "use", "two", "how", "our", "work", "first", "well", "way",
	"even", "new", "want", "because", "any", "these", "give", "day", "most", "us",
	"find", "here", "thing", "many", "still", "between", "name", "should", "school", "big",
	"great", "each", "help", "through", "long", "own", "right", "old", "while", "world",
	"same", "last", "might", "never", "need", "too", "much", "mean", "left", "keep",
	"let", "begin", "seem", "country", "hand", "high", "place", "point", "life", "where",
	"turn", "few", "group", "such", "play", "run", "small", "number", "off", "always",
	"move", "night", "live", "close", "nothing", "every", "next", "hard", "open", "start",
	"show", "part", "against", "children", "story", "second", "late", "example", "city", "eye",
	"head", "above", "often", "change", "together", "large", "earth", "add", "food", "under",
	"power", "learn", "plant", "home", "water", "air", "side", "cover", "line", "state",
	"read", "carry", "near", "build", "follow", "house", "door", "paper", "before", "face",
}

var quotes = []string{
	"The only way to do great work is to love what you do.",
	"In the middle of difficulty lies opportunity.",
	"Life is what happens when you are busy making other plans.",
	"The future belongs to those who believe in the beauty of their dreams.",
	"It does not matter how slowly you go as long as you do not stop.",
	"The greatest glory in living lies not in never falling but in rising every time we fall.",
	"The way to get started is to quit talking and begin doing.",
	"If life were predictable it would cease to be life and be without flavor.",
	"In the end it is not the years in your life that count but the life in your years.",
	"Life is really simple but we insist on making it complicated.",
	"The purpose of our lives is to be happy.",
	"You only live once but if you do it right once is enough.",
	"Many of life's failures are people who did not realize how close they were to success when they gave up.",
	"Tell me and I forget. Teach me and I remember. Involve me and I learn.",
	"The best time to plant a tree was twenty years ago. The second best time is now.",
	"Whoever is happy will make others happy too.",
	"Do not let making a living prevent you from making a life.",
	"Spread love everywhere you go. Let no one ever come to you without leaving happier.",
	"The only impossible journey is the one you never begin.",
	"Success is not final and failure is not fatal. It is the courage to continue that counts.",
}

// GenerateWordsForLanguage returns a slice of randomly selected words for a specific language.
func GenerateWordsForLanguage(lang string, count int) []string {
	if count <= 0 {
		return []string{}
	}

	wordList := GetAssetLoader().GetLanguageWords(lang)
	if len(wordList) == 0 {
		wordList = commonWords
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]string, count)
	for i := 0; i < count; i++ {
		idx := r.Intn(len(wordList))
		result[i] = wordList[idx]
	}

	return result
}

// GenerateWords returns a slice of randomly selected common words (defaults to english).
func GenerateWords(count int) []string {
	return GenerateWordsForLanguage("english", count)
}

// GenerateWordsWithOptionsForLanguage returns random words for a specific language with optional punctuation and numbers.
func GenerateWordsWithOptionsForLanguage(lang string, count int, punctuation, numbers bool) []string {
	words := GenerateWordsForLanguage(lang, count)
	if !punctuation && !numbers {
		return words
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	punctMarks := []string{".", ",", "!", "?", ";"}

	for i := range words {
		if numbers && (i+1)%8 == 0 {
			digits := r.Intn(4) + 1 // 1-4 digit number
			max := 1
			for d := 0; d < digits; d++ {
				max *= 10
			}
			words[i] = fmt.Sprintf("%d", r.Intn(max))
			continue
		}

		if punctuation {
			if (i+1)%7 == 0 {
				// Capitalize the word if not in code mode
				if !strings.HasPrefix(lang, "code_") {
					runes := []rune(words[i])
					if len(runes) > 0 {
						runes[0] = unicode.ToUpper(runes[0])
						words[i] = string(runes)
					}
				}
			}
			if (i+1)%5 == 0 {
				mark := punctMarks[r.Intn(len(punctMarks))]
				words[i] = words[i] + mark
			}
		}
	}

	return words
}

// GenerateWordsWithOptions returns random words with optional punctuation and numbers.
func GenerateWordsWithOptions(count int, punctuation, numbers bool) []string {
	return GenerateWordsWithOptionsForLanguage("english", count, punctuation, numbers)
}

// GetRandomQuoteForLanguage returns a random quote entry for a given language.
func GetRandomQuoteForLanguage(lang string) QuoteEntry {
	quoteList := GetAssetLoader().GetQuotes(lang)
	if len(quoteList) == 0 {
		return QuoteEntry{
			ID:     1,
			Text:   "The only way to do great work is to love what you do.",
			Source: "Steve Jobs",
			Length: 52,
		}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return quoteList[r.Intn(len(quoteList))]
}

// GetRandomQuote returns a random quote from the default english quotes list.
func GetRandomQuote() string {
	return GetRandomQuoteForLanguage("english").Text
}

// SplitQuote splits a quote string into individual words.
func SplitQuote(quote string) []string {
	return strings.Fields(quote)
}
