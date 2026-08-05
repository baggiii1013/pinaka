package engine

import (
	"testing"
)

func TestAssetLoaderLanguages(t *testing.T) {
	al := GetAssetLoader()
	langs := al.ListLanguages()
	if len(langs) == 0 {
		t.Fatalf("expected embedded languages, found 0")
	}

	// Verify core languages exist
	expected := []string{"english", "english_1k", "code_go", "code_python"}
	langMap := make(map[string]bool)
	for _, l := range langs {
		langMap[l] = true
	}

	for _, exp := range expected {
		if !langMap[exp] {
			t.Errorf("expected language '%s' to be available", exp)
		}
	}
}

func TestAssetLoaderWordLoadingAndCaching(t *testing.T) {
	al := GetAssetLoader()

	// Load english
	words := al.GetLanguageWords("english")
	if len(words) == 0 {
		t.Fatalf("expected words for 'english', got empty")
	}

	// Load code_go
	goWords := al.GetLanguageWords("code_go")
	if len(goWords) == 0 {
		t.Fatalf("expected words for 'code_go', got empty")
	}

	// Verify caching doesn't crash or corrupt
	cachedWords := al.GetLanguageWords("code_go")
	if len(cachedWords) != len(goWords) {
		t.Errorf("expected cached words count %d, got %d", len(goWords), len(cachedWords))
	}
}

func TestAssetLoaderInvalidLanguageFallback(t *testing.T) {
	al := GetAssetLoader()
	words := al.GetLanguageWords("non_existent_language_12345")
	if len(words) == 0 {
		t.Fatalf("expected fallback words for non-existent language, got empty")
	}
}

func TestAssetLoaderQuotes(t *testing.T) {
	al := GetAssetLoader()
	quotes := al.GetQuotes("english")
	if len(quotes) == 0 {
		t.Fatalf("expected quotes for 'english', got empty")
	}

	q := GetRandomQuoteForLanguage("english")
	if q.Text == "" {
		t.Errorf("expected non-empty random quote text")
	}
}

func TestGenerateWordsForLanguage(t *testing.T) {
	words := GenerateWordsForLanguage("code_rust", 25)
	if len(words) != 25 {
		t.Fatalf("expected 25 words generated, got %d", len(words))
	}

	optWords := GenerateWordsWithOptionsForLanguage("english", 50, true, true)
	if len(optWords) != 50 {
		t.Fatalf("expected 50 words generated with options, got %d", len(optWords))
	}
}

func TestFormatLanguageName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"english", "English"},
		{"english_1k", "English 1k"},
		{"code_go", "Code: Go"},
		{"code_c++", "Code: C++"},
	}

	for _, tt := range tests {
		actual := FormatLanguageName(tt.input)
		if actual != tt.expected {
			t.Errorf("FormatLanguageName(%q) = %q, expected %q", tt.input, actual, tt.expected)
		}
	}
}
