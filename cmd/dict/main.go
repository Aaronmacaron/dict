package main

import (
	"example.com/dict/internal/cli"
	"example.com/dict/internal/config"
	"github.com/alecthomas/kong"
)

func main() {
	var c cli.CLI
	ctx := kong.Parse(&c,
		kong.Name("dict"),
		kong.Description("Look up words in the dictionary"),
		kong.UsageOnError(),
	)

	// Configure color output
	cli.ConfigureOutput(c.Color)

	// Resolve config directory
	configDir, err := config.DirWithOverride(c.ConfigDir)
	ctx.FatalIfErrorf(err)

	// Load config
	cfg, err := config.LoadFrom(configDir)
	ctx.FatalIfErrorf(err)

	// Create context for commands
	runCtx := &cli.Context{
		ConfigDir: configDir,
		DictName:  c.Dict,
		Config:    cfg,
	}

	err = ctx.Run(runCtx)
	ctx.FatalIfErrorf(err)
}
