package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/x/term"
)

// getTermHeight returns the terminal height, or 0 if detection fails.
var getTermHeight = func() int {
	_, h, err := term.GetSize(os.Stdout.Fd())
	if err != nil || h <= 0 {
		return 0
	}
	return h
}

// printWithPager prints output directly if it fits in the terminal,
// otherwise pipes it through a pager.
func printWithPager(output string) {
	if !term.IsTerminal(os.Stdout.Fd()) {
		fmt.Println(output)
		return
	}

	height := getTermHeight()
	lines := strings.Count(output, "\n") + 1

	if height == 0 || lines <= height {
		fmt.Println(output)
		return
	}

	pagerCmd, args := resolvePager()
	if pagerCmd == "" {
		fmt.Println(output)
		return
	}

	cmd := exec.Command(pagerCmd, args...)
	cmd.Stdin = strings.NewReader(output + "\n")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println(output)
	}
}

// resolvePager returns the pager command and arguments to use.
// It checks $PAGER, then falls back to "less", then "more".
// Returns empty string if no pager is available.
func resolvePager(pagerEnvLookup ...func(string) string) (string, []string) {
	lookupEnv := os.Getenv
	if len(pagerEnvLookup) > 0 {
		lookupEnv = pagerEnvLookup[0]
	}

	if pager := lookupEnv("PAGER"); pager != "" {
		parts := strings.Fields(pager)
		if len(parts) == 0 {
			return "", nil
		}
		if _, err := exec.LookPath(parts[0]); err == nil {
			return parts[0], parts[1:]
		}
		return "", nil
	}

	if path, err := exec.LookPath("less"); err == nil {
		return path, []string{"-R"}
	}

	if path, err := exec.LookPath("more"); err == nil {
		return path, nil
	}

	return "", nil
}
