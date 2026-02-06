package cli

import (
	"regexp"
	"strings"
	"testing"

	"example.com/dict/internal/dict"
)

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
		got := FormatResults([]dict.Result{}, "moon")
		stripped := stripANSI(got)

		// Should still have headers
		if !strings.Contains(stripped, "GERMAN") || !strings.Contains(stripped, "ENGLISH") {
			t.Error("FormatResults() should contain column headers")
		}
	})

	t.Run("formats results with content", func(t *testing.T) {
		results := []dict.Result{
			{
				German:  dict.ParseTerm("Mond {m}"),
				English: dict.ParseTerm("moon"),
			},
		}

		got := FormatResults(results, "moon")
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
		results := []dict.Result{
			{
				German:     dict.ParseTerm("Mond {m}"),
				English:    dict.ParseTerm("moon"),
				SubjectsDE: []string{"astron."},
				SubjectsEN: []string{"astron."},
			},
		}

		got := FormatResults(results, "moon")
		stripped := stripANSI(got)

		// Should contain subject tags
		if !strings.Contains(stripped, "astron.") {
			t.Error("FormatResults() should contain subject tags")
		}
	})

	t.Run("highlights search word", func(t *testing.T) {
		results := []dict.Result{
			{
				German:  dict.ParseTerm("Mond {m}"),
				English: dict.ParseTerm("moon"),
			},
		}

		got := FormatResults(results, "moon")

		// Should contain ANSI codes for highlighting
		if !strings.Contains(got, "\x1b[") {
			t.Error("FormatResults() should contain ANSI styling codes")
		}
	})
}
