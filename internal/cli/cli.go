package cli

import (
	"fmt"

	"example.com/dict/internal/dict"
)

// CLI is the root command structure.
type CLI struct {
	// Global flags
	Dict      string `short:"d" help:"Use a specific dictionary by name"`
	ConfigDir string `short:"c" help:"Use a custom config directory"`
	Color     string `default:"auto" enum:"auto,always,never" help:"Color output mode"`

	// Lookup flags
	All bool `short:"a" help:"Show all results (no per-group limit)"`

	// Management flags
	List     bool   `short:"l" help:"List registered dictionaries"`
	Register string `help:"Register a dictionary file at the given path" type:"existingfile" placeholder:"PATH"`
	Default  string `help:"Set the default dictionary" placeholder:"NAME"`

	// Positional argument for lookup
	Word string `arg:"" optional:"" help:"Word to look up"`
}

// Run executes the appropriate action based on flags.
func (c *CLI) Run(ctx *Context) error {
	// Management flags take priority
	switch {
	case c.List:
		return runList(ctx)
	case c.Register != "":
		return runRegister(ctx, c.Register)
	case c.Default != "":
		return runDefault(ctx, c.Default)
	}

	// Default: word lookup
	if c.Word == "" {
		return fmt.Errorf("no word specified\n\nRun 'dict --help' for usage information.")
	}

	dictPath, err := ctx.ResolveDictPath()
	if err != nil {
		return err
	}
	d, err := dict.Open(dictPath)
	if err != nil {
		return fmt.Errorf("failed to open dictionary: %w", err)
	}
	defer d.Close()

	result, err := d.Lookup(c.Word)
	if err != nil {
		return fmt.Errorf("lookup failed: %w", err)
	}

	if len(result.Translations) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	printWithPager(FormatResults(result, c.All))
	return nil
}
