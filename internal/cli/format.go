package cli

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/muesli/termenv"

	"example.com/dict/internal/dict"
)

func init() {
	// Force ANSI color output even when not in a TTY
	lipgloss.SetColorProfile(termenv.ANSI256)
}

var (
	white     = lipgloss.Color("15")
	lightBlue = lipgloss.Color("12")
	lightGray = lipgloss.Color("245")
	gray      = lipgloss.Color("240")
	orange    = lipgloss.Color("221")

	headerStyle  = lipgloss.NewStyle().Bold(true)
	boldStyle    = lipgloss.NewStyle().Bold(true)
	genderStyle  = lipgloss.NewStyle().Foreground(lightBlue) // Blue
	contextStyle = lipgloss.NewStyle().Foreground(lightGray) // Gray
	abbrStyle    = lipgloss.NewStyle().Foreground(orange)    // Orange
	subjectStyle = lipgloss.NewStyle().
			Foreground(white). // White
			Background(gray)   // Gray background
)

// formatTerm formats a ParsedTerm with styling.
// - Search word is bold
// - Metadata (gender, context, abbr) has colored styling
func formatTerm(term dict.ParsedTerm, searchWord string) string {
	// Highlight search word in the clean text
	text := highlightWord(term.Text, searchWord)

	// Add colored metadata
	var parts []string
	parts = append(parts, text)

	if term.Gender != "" {
		parts = append(parts, genderStyle.Render("{"+term.Gender+"}"))
	}
	if term.Abbreviation != "" {
		parts = append(parts, abbrStyle.Render("<"+term.Abbreviation+">"))
	}
	for _, ctx := range term.Context {
		parts = append(parts, contextStyle.Render("["+ctx+"]"))
	}

	return strings.Join(parts, " ")
}

// formatSubjects formats subject tags with styling.
func formatSubjects(subjects []string) string {
	var styled []string
	for _, s := range subjects {
		styled = append(styled, subjectStyle.Render(" "+s+" "))
	}
	return strings.Join(styled, " ")
}

// highlightWord makes the search word bold in the text (case-insensitive).
func highlightWord(text, word string) string {
	if word == "" {
		return text
	}

	lower := strings.ToLower(text)
	wordLower := strings.ToLower(word)

	idx := strings.Index(lower, wordLower)
	if idx == -1 {
		return text
	}

	// Preserve original case of matched text
	matched := text[idx : idx+len(word)]
	return text[:idx] + boldStyle.Render(matched) + text[idx+len(word):]
}

// FormatResults formats results as a styled table.
func FormatResults(results []dict.Result, searchWord string) string {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lightGray)).
		Headers("GERMAN 🇩🇪", "ENGLISH 🇬🇧").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return lipgloss.NewStyle()
		})

	for _, r := range results {
		german := formatTerm(r.German, searchWord)
		english := formatTerm(r.English, searchWord)

		// Append subjects if present (in respective language)
		if len(r.SubjectsDE) > 0 {
			german = german + " " + formatSubjects(r.SubjectsDE)
		}
		if len(r.SubjectsEN) > 0 {
			english = english + " " + formatSubjects(r.SubjectsEN)
		}

		t.Row(german, english)
	}

	return t.Render()
}
