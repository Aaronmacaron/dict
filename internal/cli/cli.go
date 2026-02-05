package cli

import (
	"fmt"

	"example.com/dict/internal/dict"
)

type CLI struct {
	Word string `arg:"" help:"Word to look up"`
	DB   string `short:"d" default:"data/dictcc-en-de.db" help:"Path to dictionary database"`
}

func (c *CLI) Run() error {
	d, err := dict.Open(c.DB)
	if err != nil {
		return fmt.Errorf("failed to open dictionary: %w", err)
	}
	defer d.Close()

	results, err := d.Lookup(c.Word)
	if err != nil {
		return fmt.Errorf("lookup failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	fmt.Println(FormatResults(results, c.Word))
	return nil
}
