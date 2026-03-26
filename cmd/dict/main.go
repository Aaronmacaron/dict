package main

import (
	"fmt"
	"os"

	"example.com/dict/internal/cli"
	"example.com/dict/internal/config"
	"github.com/alecthomas/kong"
)

func main() {
	var c cli.CLI
	parser, err := kong.New(&c,
		kong.Name("dict"),
		kong.Description("Offline dictionary for looking up word translations."),
		kong.Help(cli.HelpPrinter),
		kong.ConfigureHelp(kong.HelpOptions{
			NoAppSummary: true,
		}),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dict: %v\n", err)
		os.Exit(1)
	}

	_, err = parser.Parse(os.Args[1:])
	if err != nil {
		cli.ConfigureOutput(c.Color)
		cli.PrintHelp()
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Configure color output
	cli.ConfigureOutput(c.Color)

	// Resolve config directory
	configDir, err := config.DirWithOverride(c.ConfigDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dict: %v\n", err)
		os.Exit(1)
	}

	// Load config
	cfg, err := config.LoadFrom(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dict: %v\n", err)
		os.Exit(1)
	}

	// Create context for commands
	runCtx := &cli.Context{
		ConfigDir: configDir,
		DictName:  c.Dict,
		Config:    cfg,
	}

	err = c.Run(runCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dict: %v\n", err)
		os.Exit(1)
	}
}
