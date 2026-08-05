package engine

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"

	"pinakatype.sh/data"
)

// LanguageData matches Monkeytype's language JSON schema.
type LanguageData struct {
	Name               string   `json:"name"`
	BCP47              string   `json:"bcp47,omitempty"`
	Words              []string `json:"words"`
	OrderedByFrequency bool     `json:"orderedByFrequency,omitempty"`
}

// QuoteEntry matches individual quotes in Monkeytype's quotes dataset.
type QuoteEntry struct {
	ID     int    `json:"id"`
	Text   string `json:"text"`
	Source string `json:"source"`
	Length int    `json:"length"`
}

// QuotesData matches Monkeytype's quote JSON schema.
type QuotesData struct {
	Language string       `json:"language"`
	Groups   [][]int      `json:"groups,omitempty"`
	Quotes   []QuoteEntry `json:"quotes"`
}

// AssetLoader manages cached access to embedded language and quote resources.
type AssetLoader struct {
	mu         sync.RWMutex
	langCache  map[string][]string
	quoteCache map[string][]QuoteEntry
	languages  []string
	quoteLangs []string
}

var (
	defaultLoader *AssetLoader
	loaderOnce    sync.Once
)

// GetAssetLoader returns the singleton AssetLoader.
func GetAssetLoader() *AssetLoader {
	loaderOnce.Do(func() {
		defaultLoader = &AssetLoader{
			langCache:  make(map[string][]string),
			quoteCache: make(map[string][]QuoteEntry),
		}
		defaultLoader.initIndex()
	})
	return defaultLoader
}

// initIndex discovers all embedded languages and quote datasets.
func (al *AssetLoader) initIndex() {
	langEntries, err := data.LanguagesFS.ReadDir("languages")
	if err == nil {
		for _, entry := range langEntries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				name := strings.TrimSuffix(entry.Name(), ".json")
				al.languages = append(al.languages, name)
			}
		}
		sort.Strings(al.languages)
	}

	quoteEntries, err := data.QuotesFS.ReadDir("quotes")
	if err == nil {
		for _, entry := range quoteEntries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				name := strings.TrimSuffix(entry.Name(), ".json")
				al.quoteLangs = append(al.quoteLangs, name)
			}
		}
		sort.Strings(al.quoteLangs)
	}
}

// ListLanguages returns all available language identifiers.
func (al *AssetLoader) ListLanguages() []string {
	al.mu.RLock()
	defer al.mu.RUnlock()
	res := make([]string, len(al.languages))
	copy(res, al.languages)
	return res
}

// ListQuoteLanguages returns all available quote language identifiers.
func (al *AssetLoader) ListQuoteLanguages() []string {
	al.mu.RLock()
	defer al.mu.RUnlock()
	res := make([]string, len(al.quoteLangs))
	copy(res, al.quoteLangs)
	return res
}

// GetLanguageWords returns the word list for a given language key with fallback to english.
func (al *AssetLoader) GetLanguageWords(lang string) []string {
	if lang == "" {
		lang = "english"
	}

	al.mu.RLock()
	if words, ok := al.langCache[lang]; ok {
		al.mu.RUnlock()
		return words
	}
	al.mu.RUnlock()

	al.mu.Lock()
	defer al.mu.Unlock()

	// Double-check after acquiring write lock
	if words, ok := al.langCache[lang]; ok {
		return words
	}

	words, err := al.loadLanguageFile(lang)
	if err != nil {
		if lang != "english" {
			// Fallback to english
			if engWords, ok := al.langCache["english"]; ok {
				return engWords
			}
			engWords, engErr := al.loadLanguageFile("english")
			if engErr == nil {
				al.langCache["english"] = engWords
				return engWords
			}
		}
		// Hardcoded fallback
		return commonWords
	}

	al.langCache[lang] = words
	return words
}

func (al *AssetLoader) loadLanguageFile(lang string) ([]string, error) {
	filename := path.Join("languages", fmt.Sprintf("%s.json", lang))
	content, err := data.LanguagesFS.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var data LanguageData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	if len(data.Words) == 0 {
		return nil, fmt.Errorf("empty word list in %s", filename)
	}

	return data.Words, nil
}

// GetQuotes returns quote entries for a given language key with fallback to english.
func (al *AssetLoader) GetQuotes(lang string) []QuoteEntry {
	if lang == "" {
		lang = "english"
	}

	al.mu.RLock()
	if q, ok := al.quoteCache[lang]; ok {
		al.mu.RUnlock()
		return q
	}
	al.mu.RUnlock()

	al.mu.Lock()
	defer al.mu.Unlock()

	if q, ok := al.quoteCache[lang]; ok {
		return q
	}

	quoteList, err := al.loadQuotesFile(lang)
	if err != nil {
		if lang != "english" {
			if engQ, ok := al.quoteCache["english"]; ok {
				return engQ
			}
			engQ, engErr := al.loadQuotesFile("english")
			if engErr == nil {
				al.quoteCache["english"] = engQ
				return engQ
			}
		}
		// Convert hardcoded quotes to QuoteEntry
		var fallbackQuotes []QuoteEntry
		for i, q := range quotes {
			fallbackQuotes = append(fallbackQuotes, QuoteEntry{
				ID:     i + 1,
				Text:   q,
				Source: "Unknown",
				Length: len(q),
			})
		}
		return fallbackQuotes
	}

	al.quoteCache[lang] = quoteList
	return quoteList
}

func (al *AssetLoader) loadQuotesFile(lang string) ([]QuoteEntry, error) {
	filename := path.Join("quotes", fmt.Sprintf("%s.json", lang))
	content, err := data.QuotesFS.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var data QuotesData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	if len(data.Quotes) == 0 {
		return nil, fmt.Errorf("empty quotes list in %s", filename)
	}

	return data.Quotes, nil
}

// CuratedPopularLanguages returns high-priority popular vocabularies.
func CuratedPopularLanguages() []string {
	return []string{
		"english",
		"english_1k",
		"english_5k",
		"english_10k",
		"english_25k",
		"spanish",
		"french",
		"german",
		"italian",
		"portuguese",
		"russian",
		"chinese_simplified",
		"japanese_romaji",
		"korean",
		"hindi",
		"polish",
		"turkish",
		"dutch",
	}
}

// CuratedCodeLanguages returns programming language vocabularies.
func CuratedCodeLanguages() []string {
	return []string{
		"code_go",
		"code_python",
		"code_rust",
		"code_javascript",
		"code_typescript",
		"code_c",
		"code_c++",
		"code_csharp",
		"code_java",
		"code_bash",
		"code_html",
		"code_css",
		"code_sql",
		"code_lua",
		"code_php",
		"code_ruby",
		"code_swift",
		"code_kotlin",
		"code_json",
		"code_yaml",
	}
}

// FormatLanguageName returns a user-friendly display name for language codes.
func FormatLanguageName(key string) string {
	if strings.HasPrefix(key, "code_") {
		sub := strings.TrimPrefix(key, "code_")
		sub = strings.ReplaceAll(sub, "_", " ")
		return fmt.Sprintf("Code: %s", strings.Title(sub))
	}

	formatted := strings.ReplaceAll(key, "_", " ")
	return strings.Title(formatted)
}
