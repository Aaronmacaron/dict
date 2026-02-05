package main

import (
	"example.com/dict/internal/cli"
	"github.com/alecthomas/kong"
)

func main() {
	var c cli.CLI
	ctx := kong.Parse(&c,
		kong.Name("dict"),
		kong.Description("Look up words in the dictionary"),
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
