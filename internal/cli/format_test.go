package cli

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/Aaronmacaron/dict/internal/dict"
)

func TestMain(m *testing.M) {
	ConfigureOutput("always")
	os.Exit(m.Run())
}

// stripANSI removes ANSI escape codes for test comparison
func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

func TestHighlightWord(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		word     string
		contains string
	}{
		{"highlights match", "moonlight", "moon", "moon"},
		{"case insensitive", "Moonlight", "moon", "Moon"},
		{"no match returns original", "sunlight", "moon", "sunlight"},
		{"empty word returns original", "test", "", "test"},
		{"empty text", "", "moon", ""},
		{"word at end", "full moon", "moon", "moon"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HighlightWord(tt.text, tt.word)
			stripped := stripANSI(got)

			if !strings.Contains(stripped, tt.contains) {
				t.Errorf("HighlightWord(%q, %q) stripped = %q, want to contain %q",
					tt.text, tt.word, stripped, tt.contains)
			}
		})
	}

	t.Run("adds bold styling", func(t *testing.T) {
		got := HighlightWord("moonlight", "moon")

		// Should contain ANSI codes (not equal to plain text)
		if got == "moonlight" {
			t.Error("HighlightWord() should add ANSI styling")
		}

		// Stripped version should still have the text
		stripped := stripANSI(got)
		if stripped != "moonlight" {
			t.Errorf("Stripped result = %q, want %q", stripped, "moonlight")
		}
	})
}

func TestFormatSubjects(t *testing.T) {
	tests := []struct {
		name     string
		subjects []string
		wantLen  int // minimum length of output
	}{
		{"empty subjects", []string{}, 0},
		{"single subject", []string{"astron."}, 7},
		{"multiple subjects", []string{"astron.", "med."}, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatSubjects(tt.subjects)
			stripped := stripANSI(got)

			if len(stripped) < tt.wantLen {
				t.Errorf("FormatSubjects() len = %d, want >= %d", len(stripped), tt.wantLen)
			}

			// Verify all subjects are present
			for _, s := range tt.subjects {
				if !strings.Contains(stripped, s) {
					t.Errorf("FormatSubjects() missing subject %q", s)
				}
			}
		})
	}
}

func TestFormatResults(t *testing.T) {
	t.Run("formats empty results", func(t *testing.T) {
		got := FormatResults(&dict.LookupResult{Query: "moon"}, false)
		if got != "" {
			t.Errorf("FormatResults() for empty results = %q, want empty string", got)
		}
	})

	t.Run("formats results with content", func(t *testing.T) {
		result := &dict.LookupResult{
			Query: "moon",
			Translations: []dict.ScoredTranslation{
				{
					Translation: dict.Translation{
						Lang1:    dict.ParseTerm("Mond {m}"),
						Lang2:    dict.ParseTerm("moon"),
						WordType: "noun",
					},
				},
			},
		}

		got := FormatResults(result, false)
		stripped := stripANSI(got)

		// Should contain the terms
		if !strings.Contains(stripped, "Mond") {
			t.Error("FormatResults() should contain German term")
		}
		if !strings.Contains(stripped, "moon") {
			t.Error("FormatResults() should contain English term")
		}
	})

	t.Run("formats results with subjects", func(t *testing.T) {
		lang1Term := dict.ParseTerm("Mond {m}")
		lang1Term.Subjects = []string{"astron."}
		lang2Term := dict.ParseTerm("moon")
		lang2Term.Subjects = []string{"astron."}

		result := &dict.LookupResult{
			Query: "moon",
			Translations: []dict.ScoredTranslation{
				{
					Translation: dict.Translation{
						Lang1:    lang1Term,
						Lang2:    lang2Term,
						WordType: "noun",
					},
				},
			},
		}

		got := FormatResults(result, false)
		stripped := stripANSI(got)

		// Should contain subject tags
		if !strings.Contains(stripped, "astron.") {
			t.Error("FormatResults() should contain subject tags")
		}
	})

	t.Run("highlights search word", func(t *testing.T) {
		result := &dict.LookupResult{
			Query: "moon",
			Translations: []dict.ScoredTranslation{
				{
					Translation: dict.Translation{
						Lang1:    dict.ParseTerm("Mond {m}"),
						Lang2:    dict.ParseTerm("moon"),
						WordType: "noun",
					},
				},
			},
		}

		got := FormatResults(result, false)

		// Should contain ANSI codes for highlighting
		if !strings.Contains(got, "\x1b[") {
			t.Error("FormatResults() should contain ANSI styling codes")
		}
	})
}

func TestGroupByWordType(t *testing.T) {
	t.Run("groups by word type", func(t *testing.T) {
		translations := []dict.ScoredTranslation{
			{Translation: dict.Translation{WordType: "noun"}, Score: 0.9},
			{Translation: dict.Translation{WordType: "verb"}, Score: 0.8},
			{Translation: dict.Translation{WordType: "noun"}, Score: 0.7},
			{Translation: dict.Translation{WordType: "adj"}, Score: 0.6},
		}

		groups := GroupByWordType(translations)

		if len(groups) != 3 {
			t.Fatalf("got %d groups, want 3", len(groups))
		}

		// First group should be noun (best score 0.9)
		if groups[0].wordType != "noun" {
			t.Errorf("first group type = %q, want %q", groups[0].wordType, "noun")
		}
		if len(groups[0].translations) != 2 {
			t.Errorf("noun group has %d translations, want 2", len(groups[0].translations))
		}

		// Second group should be verb (best score 0.8)
		if groups[1].wordType != "verb" {
			t.Errorf("second group type = %q, want %q", groups[1].wordType, "verb")
		}
	})

	t.Run("empty word type goes last", func(t *testing.T) {
		translations := []dict.ScoredTranslation{
			{Translation: dict.Translation{WordType: ""}, Score: 1.0},
			{Translation: dict.Translation{WordType: "noun"}, Score: 0.5},
		}

		groups := GroupByWordType(translations)

		if len(groups) != 2 {
			t.Fatalf("got %d groups, want 2", len(groups))
		}
		if groups[0].wordType != "noun" {
			t.Errorf("first group = %q, want %q", groups[0].wordType, "noun")
		}
		if groups[1].wordType != "" {
			t.Errorf("last group = %q, want empty", groups[1].wordType)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		groups := GroupByWordType(nil)
		if len(groups) != 0 {
			t.Errorf("got %d groups, want 0", len(groups))
		}
	})
}

func TestClampColumnWidths(t *testing.T) {
	t.Run("no change when fits", func(t *testing.T) {
		got := ClampColumnWidths([2]int{20, 20}, 80)
		if got != [2]int{20, 20} {
			t.Errorf("got %v, want [20 20]", got)
		}
	})

	t.Run("shrinks proportionally", func(t *testing.T) {
		// 40+40+3=83, needs to fit in 43 → available=40, ratio=0.5 → 20,20
		got := ClampColumnWidths([2]int{40, 40}, 43)
		if got[0]+got[1] != 40 {
			t.Errorf("total %d, want 40", got[0]+got[1])
		}
	})

	t.Run("preserves ratio", func(t *testing.T) {
		// 30+10+3=43, needs to fit in 23 → available=20, ratio=0.75 → 15,5
		got := ClampColumnWidths([2]int{30, 10}, 23)
		if got[0]+got[1] != 20 {
			t.Errorf("total %d, want 20", got[0]+got[1])
		}
		if got[0] <= got[1] {
			t.Errorf("col0 (%d) should be larger than col1 (%d)", got[0], got[1])
		}
	})
}

func TestTruncateCell(t *testing.T) {
	t.Run("no truncation needed", func(t *testing.T) {
		got := TruncateCell("hello", 10)
		if stripANSI(got) != "hello" {
			t.Errorf("got %q, want %q", stripANSI(got), "hello")
		}
	})

	t.Run("truncates with ellipsis", func(t *testing.T) {
		got := TruncateCell("hello world this is long", 10)
		stripped := stripANSI(got)
		w := lipgloss.Width(stripped)
		if w > 10 {
			t.Errorf("truncated visual width %d exceeds max 10", w)
		}
		if !strings.Contains(stripped, "…") {
			t.Error("truncated string should contain ellipsis")
		}
	})
}

func TestWordTypeLabel(t *testing.T) {
	tests := []struct {
		wordType string
		want     string
	}{
		{"noun", "Nouns"},
		{"verb", "Verbs"},
		{"adj", "Adjectives"},
		{"", "Other"},
		{"adj past-p", "Adj Past-p"},
	}

	for _, tt := range tests {
		t.Run(tt.wordType, func(t *testing.T) {
			got := WordTypeLabel(tt.wordType)
			if got != tt.want {
				t.Errorf("WordTypeLabel(%q) = %q, want %q", tt.wordType, got, tt.want)
			}
		})
	}
}

func TestFormatResultsGrouped(t *testing.T) {
	t.Run("truncation hint shown", func(t *testing.T) {
		var translations []dict.ScoredTranslation
		for i := 0; i < 8; i++ {
			translations = append(translations, dict.ScoredTranslation{
				Translation: dict.Translation{
					Lang1:    dict.ParseTerm("Wort"),
					Lang2:    dict.ParseTerm("word"),
					WordType: "noun",
				},
				Score: float64(8-i) / 10.0,
			})
		}

		got := FormatResults(&dict.LookupResult{Query: "word", Translations: translations}, false)
		stripped := stripANSI(got)

		if !strings.Contains(stripped, "+ 3 more") {
			t.Error("should show '+ 3 more' hint for 8 results with cap 5")
		}
		if !strings.Contains(stripped, "Use -a to show all results.") {
			t.Error("should show footer hint when truncated")
		}
	})

	t.Run("full mode shows all", func(t *testing.T) {
		var translations []dict.ScoredTranslation
		for i := 0; i < 8; i++ {
			translations = append(translations, dict.ScoredTranslation{
				Translation: dict.Translation{
					Lang1:    dict.ParseTerm("Wort"),
					Lang2:    dict.ParseTerm("word"),
					WordType: "noun",
				},
				Score: float64(8-i) / 10.0,
			})
		}

		got := FormatResults(&dict.LookupResult{Query: "word", Translations: translations}, true)
		stripped := stripANSI(got)

		if strings.Contains(stripped, "+ 3 more") {
			t.Error("full mode should not show truncation hint")
		}
		if strings.Contains(stripped, "Use -a to show all results.") {
			t.Error("full mode should not show footer hint")
		}
	})

	t.Run("section headers present", func(t *testing.T) {
		translations := []dict.ScoredTranslation{
			{
				Translation: dict.Translation{
					Lang1:    dict.ParseTerm("Mond {m}"),
					Lang2:    dict.ParseTerm("moon"),
					WordType: "noun",
				},
				Score: 0.9,
			},
			{
				Translation: dict.Translation{
					Lang1:    dict.ParseTerm("monden"),
					Lang2:    dict.ParseTerm("to moon"),
					WordType: "verb",
				},
				Score: 0.7,
			},
		}

		got := FormatResults(&dict.LookupResult{Query: "moon", Translations: translations}, false)
		stripped := stripANSI(got)

		if !strings.Contains(stripped, "Nouns") {
			t.Error("should contain 'Nouns' section header")
		}
		if !strings.Contains(stripped, "Verbs") {
			t.Error("should contain 'Verbs' section header")
		}
	})

	t.Run("tables aligned across groups", func(t *testing.T) {
		translations := []dict.ScoredTranslation{
			{
				Translation: dict.Translation{
					Lang1:    dict.ParseTerm("Mondfinsternis {f}"),
					Lang2:    dict.ParseTerm("lunar eclipse"),
					WordType: "noun",
				},
				Score: 0.9,
			},
			{
				Translation: dict.Translation{
					Lang1:    dict.ParseTerm("monden"),
					Lang2:    dict.ParseTerm("to moon"),
					WordType: "verb",
				},
				Score: 0.7,
			},
		}

		got := FormatResults(&dict.LookupResult{Query: "moon", Translations: translations}, false)
		stripped := stripANSI(got)
		lines := strings.Split(stripped, "\n")

		// Find border lines (start with ┌, ├, or └) — they reflect the table width.
		// All tables should produce border lines of the same length.
		var borderLengths []int
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "╭") || strings.HasPrefix(trimmed, "╰") {
				borderLengths = append(borderLengths, lipgloss.Width(trimmed))
			}
		}

		if len(borderLengths) < 2 {
			t.Fatalf("expected at least 2 border lines, got %d", len(borderLengths))
		}
		for i := 1; i < len(borderLengths); i++ {
			if borderLengths[i] != borderLengths[0] {
				t.Errorf("border line %d has length %d, want %d (same as first)",
					i, borderLengths[i], borderLengths[0])
			}
		}
	})

	t.Run("table fits narrow terminal", func(t *testing.T) {
		origGetTermWidth := *GetTermWidth
		*GetTermWidth = func() int { return 40 }
		defer func() { *GetTermWidth = origGetTermWidth }()

		translations := []dict.ScoredTranslation{
			{
				Translation: dict.Translation{
					Lang1:    dict.ParseTerm("Mondfinsternis {f}"),
					Lang2:    dict.ParseTerm("lunar eclipse [astron.]"),
					WordType: "noun",
				},
				Score: 0.9,
			},
		}

		got := FormatResults(&dict.LookupResult{Query: "moon", Translations: translations}, false)
		stripped := stripANSI(got)
		lines := strings.Split(stripped, "\n")

		for _, line := range lines {
			w := lipgloss.Width(line)
			if w > 40 {
				t.Errorf("line exceeds terminal width 40: width=%d %q", w, line)
			}
		}
	})

	t.Run("truncated cells show ellipsis", func(t *testing.T) {
		origGetTermWidth := *GetTermWidth
		*GetTermWidth = func() int { return 30 }
		defer func() { *GetTermWidth = origGetTermWidth }()

		translations := []dict.ScoredTranslation{
			{
				Translation: dict.Translation{
					Lang1:    dict.ParseTerm("Mondfinsternis {f}"),
					Lang2:    dict.ParseTerm("lunar eclipse"),
					WordType: "noun",
				},
				Score: 0.9,
			},
		}

		got := FormatResults(&dict.LookupResult{Query: "moon", Translations: translations}, false)
		stripped := stripANSI(got)

		if !strings.Contains(stripped, "…") {
			t.Error("truncated output should contain ellipsis character")
		}
	})

	t.Run("plain mode has no ANSI codes", func(t *testing.T) {
		ConfigureOutput("never")
		defer ConfigureOutput("always")

		result := &dict.LookupResult{
			Query: "moon",
			Translations: []dict.ScoredTranslation{
				{
					Translation: dict.Translation{
						Lang1:    dict.ParseTerm("Mond {m}"),
						Lang2:    dict.ParseTerm("moon"),
						WordType: "noun",
					},
				},
			},
		}

		got := FormatResults(result, false)

		if strings.Contains(got, "\x1b[") {
			t.Error("plain mode should not contain ANSI escape codes")
		}

		if !strings.Contains(got, "Mond") || !strings.Contains(got, "moon") {
			t.Error("plain mode should still contain the terms")
		}
	})

	t.Run("no truncation when under cap", func(t *testing.T) {
		translations := []dict.ScoredTranslation{
			{
				Translation: dict.Translation{
					Lang1:    dict.ParseTerm("Mond {m}"),
					Lang2:    dict.ParseTerm("moon"),
					WordType: "noun",
				},
				Score: 0.9,
			},
		}

		got := FormatResults(&dict.LookupResult{Query: "moon", Translations: translations}, false)
		stripped := stripANSI(got)

		if strings.Contains(stripped, "more") {
			t.Error("should not show truncation hint when under cap")
		}
		if strings.Contains(stripped, "Use -a") {
			t.Error("should not show footer hint when nothing truncated")
		}
	})
}
