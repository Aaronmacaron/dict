package cli

import (
	"testing"
)

func TestResolvePager_EnvVar(t *testing.T) {
	pagerCmd, args := resolvePager(func(key string) string {
		if key == "PAGER" {
			return "cat"
		}
		return ""
	})

	if pagerCmd == "" {
		t.Skip("cat not found in PATH")
	}

	if len(args) != 0 {
		t.Errorf("expected no args for cat, got %v", args)
	}
}

func TestResolvePager_EnvVarWithArgs(t *testing.T) {
	pagerCmd, args := resolvePager(func(key string) string {
		if key == "PAGER" {
			return "less -X"
		}
		return ""
	})

	if pagerCmd == "" {
		t.Skip("less not found in PATH")
	}

	if len(args) != 1 || args[0] != "-X" {
		t.Errorf("expected args [-X], got %v", args)
	}
}

func TestResolvePager_InvalidPager(t *testing.T) {
	pagerCmd, _ := resolvePager(func(key string) string {
		if key == "PAGER" {
			return "nonexistent_pager_binary_xyz"
		}
		return ""
	})

	if pagerCmd != "" {
		t.Errorf("expected empty pager for nonexistent binary, got %q", pagerCmd)
	}
}

func TestResolvePager_EmptyEnvFallback(t *testing.T) {
	pagerCmd, _ := resolvePager(func(key string) string {
		return ""
	})

	// Should fall back to less or more (both likely available)
	if pagerCmd == "" {
		t.Skip("neither less nor more found in PATH")
	}
}
