package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"github.com/muesli/termenv"

	"github.com/Aaronmacaron/dict/internal/dict"
)

// ConfigureOutput sets the color profile based on the given mode.
// Modes: "always" (force colors), "never" (no colors), "auto" (detect).
func ConfigureOutput(colorMode string) {
	var profile termenv.Profile

	switch colorMode {
	case "always":
		profile = termenv.ANSI256
	case "never":
		profile = termenv.Ascii
	default: // "auto"
		if os.Getenv("NO_COLOR") != "" {
			profile = termenv.Ascii
		} else if !term.IsTerminal(os.Stdout.Fd()) {
			profile = termenv.Ascii
		} else {
			profile = termenv.ANSI256
		}
	}

	lipgloss.SetColorProfile(profile)
}

const (
	defaultGroupCap = 5
	// Border overhead for a 2-column table: │ col1 │ col2 │
	tableBorderOverhead = 3
	defaultTermWidth    = 80
)

// getTermWidth returns the terminal width, or a default if detection fails.
// Overridable for testing.
var getTermWidth = func() int {
	w, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil || w <= 0 {
		return defaultTermWidth
	}
	return w
}

var (
	white     = lipgloss.Color("15")
	lightBlue = lipgloss.Color("12")
	lightGray = lipgloss.Color("245")
	gray      = lipgloss.Color("240")
	orange    = lipgloss.Color("221")

	boldStyle    = lipgloss.NewStyle().Bold(true)
	genderStyle  = lipgloss.NewStyle().Foreground(lightBlue) // Blue
	contextStyle = lipgloss.NewStyle().Foreground(lightGray) // Gray
	abbrStyle    = lipgloss.NewStyle().Foreground(orange)    // Orange
	subjectStyle = lipgloss.NewStyle().
			Foreground(white). // White
			Background(gray)   // Gray background

	wordTypeLabels = map[string]string{
		"noun":   "Nouns",
		"verb":   "Verbs",
		"adj":    "Adjectives",
		"adv":    "Adverbs",
		"prep":   "Prepositions",
		"conj":   "Conjunctions",
		"pron":   "Pronouns",
		"prefix": "Prefixes",
		"suffix": "Suffixes",
		"past-p": "Past Participles",
		"pres-p": "Present Participles",
	}

	wordTypeColors = map[string]lipgloss.Color{
		"noun":   lipgloss.Color("75"),  // steel blue
		"verb":   lipgloss.Color("114"), // green
		"adj":    lipgloss.Color("180"), // gold
		"adv":    lipgloss.Color("139"), // mauve
		"prep":   lipgloss.Color("109"), // teal
		"conj":   lipgloss.Color("174"), // pink
		"pron":   lipgloss.Color("146"), // lavender
		"prefix": lipgloss.Color("137"), // tan
		"suffix": lipgloss.Color("137"), // tan
		"past-p": lipgloss.Color("114"), // green (same as verb)
		"pres-p": lipgloss.Color("114"), // green (same as verb)
	}

	defaultWordTypeColor = lipgloss.Color("245") // gray
)

type wordTypeGroup struct {
	wordType     string
	label        string
	color        lipgloss.Color
	translations []dict.ScoredTranslation
	bestScore    float64
}

// formattedRow holds pre-formatted cell strings for a single translation.
type formattedRow struct {
	lang1 string
	lang2 string
}

// formatTerm formats a Term with styling.
// - Search word is bold
// - Metadata (gender, context, abbr, subjects) has colored styling
func formatTerm(term dict.Term, searchWord string) string {
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
	if len(term.Subjects) > 0 {
		parts = append(parts, formatSubjects(term.Subjects))
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

// wordTypeLabel returns a human-readable label for a word type.
func wordTypeLabel(wordType string) string {
	if wordType == "" {
		return "Other"
	}
	if label, ok := wordTypeLabels[wordType]; ok {
		return label
	}
	// Title-case fallback for unknown/combo types
	words := strings.Fields(wordType)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// wordTypeColor returns the color for a word type.
func wordTypeColor(wordType string) lipgloss.Color {
	if c, ok := wordTypeColors[wordType]; ok {
		return c
	}
	return defaultWordTypeColor
}

// groupByWordType partitions translations into groups by word type,
// sorted by best score within each group, with "Other" (empty type) last.
func groupByWordType(translations []dict.ScoredTranslation) []wordTypeGroup {
	groupMap := make(map[string]*wordTypeGroup)
	var order []string

	for _, st := range translations {
		wt := st.Translation.WordType
		g, exists := groupMap[wt]
		if !exists {
			g = &wordTypeGroup{
				wordType: wt,
				label:    wordTypeLabel(wt),
				color:    wordTypeColor(wt),
			}
			groupMap[wt] = g
			order = append(order, wt)
		}
		g.translations = append(g.translations, st)
		if st.Score > g.bestScore {
			g.bestScore = st.Score
		}
	}

	// Build slice from map
	groups := make([]wordTypeGroup, 0, len(groupMap))
	for _, wt := range order {
		groups = append(groups, *groupMap[wt])
	}

	// Sort by best score descending, "Other" (empty type) always last
	sort.SliceStable(groups, func(i, j int) bool {
		iEmpty := groups[i].wordType == ""
		jEmpty := groups[j].wordType == ""
		if iEmpty != jEmpty {
			return !iEmpty
		}
		return groups[i].bestScore > groups[j].bestScore
	})

	return groups
}

// clampColumnWidths shrinks column widths proportionally so the total table
// width (columns + borders) fits within maxWidth.
func clampColumnWidths(colWidths [2]int, maxWidth int) [2]int {
	total := colWidths[0] + colWidths[1] + tableBorderOverhead
	if total <= maxWidth {
		return colWidths
	}
	available := maxWidth - tableBorderOverhead
	if available < 2 {
		available = 2
	}
	ratio := float64(colWidths[0]) / float64(colWidths[0]+colWidths[1])
	colWidths[0] = int(float64(available) * ratio)
	colWidths[1] = available - colWidths[0]
	return colWidths
}

// truncateCell truncates a styled string to fit within maxWidth, appending "…"
// if truncation occurs. Handles ANSI escape codes correctly.
func truncateCell(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	return ansi.Truncate(s, maxWidth, "…")
}

// renderGroupHeader renders a styled section header like " ── Nouns ──".
func renderGroupHeader(label string, color lipgloss.Color) string {
	style := lipgloss.NewStyle().Foreground(color).Bold(true)
	dash := lipgloss.NewStyle().Foreground(color).Render("──")
	return " " + dash + " " + style.Render(label) + " " + dash
}

// renderGroupTable renders a headerless table for a group of translations
// with fixed column widths for cross-group alignment.
func renderGroupTable(rows []formattedRow, colWidths [2]int) string {
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lightGray)).
		StyleFunc(func(row, col int) lipgloss.Style {
			return lipgloss.NewStyle().Width(colWidths[col])
		})

	for _, r := range rows {
		t.Row(r.lang1, r.lang2)
	}

	return t.Render()
}

// renderLanguageHeader renders the global header of the output that shows what the two languages are
func renderLanguageHeader(lang1, lang2 string, colWidths [2]int) string {
	t := table.New().Border(lipgloss.HiddenBorder()).StyleFunc(func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Bold(true)
		if col == 0 {
			width := max(colWidths[0]-2, len(lang1))
			return style.Width(width).Align(lipgloss.Right)
		} else if col == 2 {
			width := max(colWidths[1]-2, len(lang2))
			return style.Width(width)
		}

		return lipgloss.NewStyle().Width(3)
	})

	t.Row(lang1, "<=>", lang2)

	return t.Render() + "\n"
}

// FormatResults formats results as grouped styled tables by word type.
func FormatResults(result *dict.LookupResult, showAll bool) string {
	if len(result.Translations) == 0 {
		return ""
	}

	groups := groupByWordType(result.Translations)

	// First pass: format all visible cells and compute column widths.
	type groupDisplay struct {
		group     wordTypeGroup
		rows      []formattedRow
		truncated int
	}

	var displays []groupDisplay
	anyTruncated := false

	for _, g := range groups {
		showing := g.translations
		truncated := 0
		if !showAll && len(showing) > defaultGroupCap {
			truncated = len(showing) - defaultGroupCap
			showing = showing[:defaultGroupCap]
			anyTruncated = true
		}

		rows := make([]formattedRow, len(showing))
		for i, st := range showing {
			rows[i] = formattedRow{
				lang1: formatTerm(st.Translation.Lang1, result.Query),
				lang2: formatTerm(st.Translation.Lang2, result.Query),
			}
		}

		displays = append(displays, groupDisplay{
			group:     g,
			rows:      rows,
			truncated: truncated,
		})
	}

	// Compute max visual width per column across all groups.
	var colWidths [2]int
	for _, d := range displays {
		for _, r := range d.rows {
			if w := lipgloss.Width(r.lang1); w > colWidths[0] {
				colWidths[0] = w
			}
			if w := lipgloss.Width(r.lang2); w > colWidths[1] {
				colWidths[1] = w
			}
		}
	}

	// Clamp column widths to fit the terminal.
	termWidth := getTermWidth()
	colWidths = clampColumnWidths(colWidths, termWidth)

	// Truncate cells that exceed their column width.
	for i := range displays {
		for j := range displays[i].rows {
			displays[i].rows[j].lang1 = truncateCell(displays[i].rows[j].lang1, colWidths[0])
			displays[i].rows[j].lang2 = truncateCell(displays[i].rows[j].lang2, colWidths[1])
		}
	}

	// Second pass: render.
	var sb strings.Builder

	// Header indicating languages
	sb.WriteString(renderLanguageHeader(result.Lang1Name, result.Lang2Name, colWidths))

	for i, d := range displays {
		if i > 0 {
			sb.WriteString("\n")
		}

		// Section header
		sb.WriteString(renderGroupHeader(d.group.label, d.group.color))
		sb.WriteString("\n")

		// Table
		sb.WriteString(renderGroupTable(d.rows, colWidths))

		// Truncation hint
		if d.truncated > 0 {
			hint := fmt.Sprintf("  + %d more", d.truncated)
			sb.WriteString("\n")
			sb.WriteString(lipgloss.NewStyle().Foreground(lightGray).Render(hint))
		}

		sb.WriteString("\n")
	}

	// Footer hint
	if anyTruncated {
		sb.WriteString("\n")
		sb.WriteString(lipgloss.NewStyle().Foreground(lightGray).Render("Use -a to show all results."))
	}

	return sb.String()
}
