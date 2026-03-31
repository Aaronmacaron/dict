package cli

import (
	"path/filepath"
	"testing"

	"github.com/Aaronmacaron/dict/internal/config"
)

func TestContext_ResolveDictPath(t *testing.T) {
	t.Run("uses DictName override when set", func(t *testing.T) {
		ctx := &Context{
			ConfigDir: "/test/config",
			DictName:  "override.db",
			Config:    config.Config{Dict: "default.db"},
		}

		got, err := ctx.ResolveDictPath()
		if err != nil {
			t.Fatalf("ResolveDictPath() error = %v", err)
		}
		want := filepath.Join("/test/config", "dicts", "override.db")

		if got != want {
			t.Errorf("ResolveDictPath() = %q, want %q", got, want)
		}
	})

	t.Run("uses Config.Dict when DictName is empty", func(t *testing.T) {
		ctx := &Context{
			ConfigDir: "/test/config",
			DictName:  "",
			Config:    config.Config{Dict: "configured.db"},
		}

		got, err := ctx.ResolveDictPath()
		if err != nil {
			t.Fatalf("ResolveDictPath() error = %v", err)
		}
		want := filepath.Join("/test/config", "dicts", "configured.db")

		if got != want {
			t.Errorf("ResolveDictPath() = %q, want %q", got, want)
		}
	})

	t.Run("returns error when no dictionary configured", func(t *testing.T) {
		ctx := &Context{
			ConfigDir: "/test/config",
			DictName:  "",
			Config:    config.Config{Dict: ""},
		}

		_, err := ctx.ResolveDictPath()
		if err == nil {
			t.Error("ResolveDictPath() expected error when no dictionary configured")
		}
	})

	t.Run("constructs correct path structure", func(t *testing.T) {
		ctx := &Context{
			ConfigDir: "/home/user/.config/dict",
			Config:    config.Config{Dict: "dictcc-en-de.db"},
		}

		got, err := ctx.ResolveDictPath()
		if err != nil {
			t.Fatalf("ResolveDictPath() error = %v", err)
		}

		if !filepath.IsAbs(got) {
			t.Error("ResolveDictPath() should return absolute path when configDir is absolute")
		}

		if filepath.Base(got) != "dictcc-en-de.db" {
			t.Errorf("ResolveDictPath() filename = %q, want %q", filepath.Base(got), "dictcc-en-de.db")
		}

		if filepath.Base(filepath.Dir(got)) != "dicts" {
			t.Errorf("ResolveDictPath() parent dir = %q, want %q", filepath.Base(filepath.Dir(got)), "dicts")
		}
	})
}
