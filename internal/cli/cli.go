package cli

import (
	"fmt"

	"example.com/dict/internal/dict"
)

// CLI is the root command structure.
type CLI struct {
	// Global flags
	Dict      string `short:"d" help:"Dictionary name (overrides config)"`
	ConfigDir string `short:"c" help:"Config directory (overrides default)"`
	Color     string `default:"auto" enum:"auto,always,never" help:"Color output mode (auto, always, never)"`

	// Commands
	Lookup LookupCmd `cmd:"" default:"withargs" help:"Look up a word in the dictionary"`
	Manage ManageCmd `cmd:"" name:"manage" aliases:"m" help:"Manage dictionaries"`
}

// LookupCmd handles word lookup (the default command).
type LookupCmd struct {
	Word string `arg:"" help:"Word to look up"`
	All  bool   `short:"a" help:"Show all results (no per-group limit)"`
}

// Run executes the lookup command.
func (cmd *LookupCmd) Run(ctx *Context) error {
	if cmd.Word == "" {
		return fmt.Errorf("no word specified")
	}

	dictPath := ctx.ResolveDictPath()
	d, err := dict.Open(dictPath)
	if err != nil {
		return fmt.Errorf("failed to open dictionary: %w", err)
	}
	defer d.Close()

	result, err := d.Lookup(cmd.Word)
	if err != nil {
		return fmt.Errorf("lookup failed: %w", err)
	}

	if len(result.Translations) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	printWithPager(FormatResults(result, cmd.All))
	return nil
}

// ManageCmd contains dictionary management subcommands.
type ManageCmd struct {
	Register RegisterCmd `cmd:"" help:"Register a dictionary file"`
	Default  DefaultCmd  `cmd:"" name:"default" help:"Set the default dictionary"`
	List     ListCmd     `cmd:"" help:"List available dictionaries"`
}

// RegisterCmd copies a dictionary file to config/dicts/.
type RegisterCmd struct {
	Path string `arg:"" type:"existingfile" help:"Path to dictionary file"`
}

// DefaultCmd sets the default dictionary in config.toml.
type DefaultCmd struct {
	Name string `arg:"" help:"Dictionary name to set as default"`
}

// ListCmd lists all dictionaries in config/dicts/.
type ListCmd struct{}
