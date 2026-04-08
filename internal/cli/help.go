package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss"
)

var (
	helpTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))
	helpDesc  = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	helpSectionHeader = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("114"))

	helpFlag      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("180"))
	helpFlagDesc  = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	helpFlagMeta  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	helpExample   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	helpExampleEx = lipgloss.NewStyle().Foreground(lipgloss.Color("109"))
)

// HelpPrinter is a custom Kong help printer that uses lipgloss for styling.
func HelpPrinter(options kong.HelpOptions, ctx *kong.Context) error {
	printHelp(ctx.Stdout)
	return nil
}

// PrintHelp prints the help page to stderr. Used for error output.
func PrintHelp() {
	printHelp(os.Stderr)
}

func printHelp(w io.Writer) {
	// Title and description
	fmt.Fprintln(w, helpTitle.Render("dict")+" "+helpFlagMeta.Render("- offline dictionary"))
	fmt.Fprintln(w, helpDesc.Render("Look up word translations using local dict.cc databases."))

	// Usage
	fmt.Fprintln(w)
	fmt.Fprintln(w, helpSectionHeader.Render("Usage"))
	fmt.Fprintln(w, "  "+helpExampleEx.Render("dict <word>")+"              "+helpFlagDesc.Render("Look up a word"))
	fmt.Fprintln(w, "  "+helpExampleEx.Render("dict [flags]")+"             "+helpFlagDesc.Render("Run a management action"))

	// Lookup section
	printSection(w, "Lookup", []flagEntry{
		{flags: "<word>", desc: "Word or phrase to look up", example: "dict moon"},
		{flags: "-a, --all", desc: "Show all results (no per-group limit)", example: "dict -a moon"},
	})

	// Dictionary management section
	printSection(w, "Dictionary Management", []flagEntry{
		{flags: "-l, --list", desc: "List registered dictionaries", example: "dict --list"},
		{flags: "    --register PATH", desc: "Register a dictionary file", example: "dict --register ~/Downloads/dictcc-fr-en.db"},
		{flags: "    --default NAME", desc: "Set the default dictionary", example: "dict --default dictcc-fr-en.db"},
	})

	// Global flags section
	printSection(w, "Global Flags", []flagEntry{
		{flags: "-d, --dict NAME", desc: "Use a specific dictionary for this lookup", example: "dict -d dictcc-fr-en.db bonjour"},
		{flags: "-c, --config-dir PATH", desc: "Use a custom config directory", example: "dict -c /custom/path --list"},
		{flags: "    --color MODE", desc: "Color output: auto, always, never (default: auto)"},
		{flags: "-v, --version", desc: "Show version"},
	})

	// Examples
	fmt.Fprintln(w)
	fmt.Fprintln(w, helpSectionHeader.Render("Examples"))
	printExamples(w, []exampleEntry{
		{cmd: "dict moon", desc: "Look up \"moon\" in the default dictionary"},
		{cmd: "dict \"full moon\"", desc: "Look up a phrase"},
		{cmd: "dict -a Mond", desc: "Show all results for \"Mond\""},
		{cmd: "dict --register ./dictcc-en-de.db", desc: "Register a new dictionary"},
		{cmd: "dict --list", desc: "Show all registered dictionaries"},
		{cmd: "dict -d dictcc-fr-en.db bonjour", desc: "Look up in a specific dictionary"},
	})

	fmt.Fprintln(w)
}

type flagEntry struct {
	flags   string
	desc    string
	example string
}

type exampleEntry struct {
	cmd  string
	desc string
}

func printSection(w io.Writer, title string, entries []flagEntry) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, helpSectionHeader.Render(title))

	for _, e := range entries {
		// Flag name (padded to align descriptions)
		flagStr := helpFlag.Render(e.flags)
		flagWidth := lipgloss.Width(flagStr)

		padding := 28 - flagWidth
		if padding < 2 {
			padding = 2
		}

		fmt.Fprintf(w, "  %s%s%s\n", flagStr, strings.Repeat(" ", padding), helpFlagDesc.Render(e.desc))
		if e.example != "" {
			fmt.Fprintf(w, "  %s%s\n", strings.Repeat(" ", 28), helpExample.Render("$ "+e.example))
		}
	}
}

func printExamples(w io.Writer, examples []exampleEntry) {
	for _, e := range examples {
		fmt.Fprintf(w, "  %s\n", helpExampleEx.Render("$ "+e.cmd))
		fmt.Fprintf(w, "    %s\n", helpExample.Render(e.desc))
	}
}
