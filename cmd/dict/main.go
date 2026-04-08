package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/Aaronmacaron/dict/internal/cli"
	"github.com/Aaronmacaron/dict/internal/config"
	"github.com/alecthomas/kong"
)

// Version is set at build time via ldflags.
var Version = "dev"

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

	// Set version string (fall back to module info for go install users)
	c.VersionStr = Version
	if c.VersionStr == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			c.VersionStr = info.Main.Version
		}
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
