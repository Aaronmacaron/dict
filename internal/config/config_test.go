package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Dict != "dictcc-en-de.db" {
		t.Errorf("DefaultConfig().Dict = %q, want %q", cfg.Dict, "dictcc-en-de.db")
	}
}

func TestDictsDir(t *testing.T) {
	tests := []struct {
		configDir string
		want      string
	}{
		{"/home/user/.config/dict", "/home/user/.config/dict/dicts"},
		{"/tmp/test", "/tmp/test/dicts"},
		{"relative/path", "relative/path/dicts"},
	}

	for _, tt := range tests {
		t.Run(tt.configDir, func(t *testing.T) {
			got := DictsDir(tt.configDir)
			if got != tt.want {
				t.Errorf("DictsDir(%q) = %q, want %q", tt.configDir, got, tt.want)
			}
		})
	}
}

func TestPath(t *testing.T) {
	tests := []struct {
		configDir string
		want      string
	}{
		{"/home/user/.config/dict", "/home/user/.config/dict/config.toml"},
		{"/tmp/test", "/tmp/test/config.toml"},
	}

	for _, tt := range tests {
		t.Run(tt.configDir, func(t *testing.T) {
			got := Path(tt.configDir)
			if got != tt.want {
				t.Errorf("Path(%q) = %q, want %q", tt.configDir, got, tt.want)
			}
		})
	}
}

func TestConfig_DictPath(t *testing.T) {
	cfg := Config{Dict: "mydict.db"}
	configDir := "/home/user/.config/dict"

	got := cfg.DictPath(configDir)
	want := "/home/user/.config/dict/dicts/mydict.db"

	if got != want {
		t.Errorf("DictPath() = %q, want %q", got, want)
	}
}

func TestDirWithOverride(t *testing.T) {
	t.Run("with override returns override", func(t *testing.T) {
		got, err := DirWithOverride("/custom/path")
		if err != nil {
			t.Fatalf("DirWithOverride() error = %v", err)
		}
		if got != "/custom/path" {
			t.Errorf("DirWithOverride() = %q, want %q", got, "/custom/path")
		}
	})

	t.Run("empty override returns default", func(t *testing.T) {
		got, err := DirWithOverride("")
		if err != nil {
			t.Fatalf("DirWithOverride() error = %v", err)
		}
		// Should return something (the actual system config dir)
		if got == "" {
			t.Error("DirWithOverride() returned empty string")
		}
	})
}

func TestLoadFrom(t *testing.T) {
	t.Run("missing config returns defaults", func(t *testing.T) {
		tmpDir := t.TempDir()

		cfg, err := LoadFrom(tmpDir)
		if err != nil {
			t.Fatalf("LoadFrom() error = %v", err)
		}

		if cfg != DefaultConfig() {
			t.Errorf("LoadFrom() = %+v, want %+v", cfg, DefaultConfig())
		}
	})

	t.Run("loads existing config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")

		err := os.WriteFile(configPath, []byte(`dict = "custom.db"`), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		cfg, err := LoadFrom(tmpDir)
		if err != nil {
			t.Fatalf("LoadFrom() error = %v", err)
		}

		if cfg.Dict != "custom.db" {
			t.Errorf("LoadFrom().Dict = %q, want %q", cfg.Dict, "custom.db")
		}
	})

	t.Run("invalid TOML returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")

		err := os.WriteFile(configPath, []byte(`invalid toml {{{{`), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		_, err = LoadFrom(tmpDir)
		if err == nil {
			t.Error("LoadFrom() expected error for invalid TOML")
		}
	})
}

func TestConfig_Save(t *testing.T) {
	t.Run("creates config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := Config{Dict: "test.db"}

		err := cfg.Save(tmpDir)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Verify file exists and contains expected content
		data, err := os.ReadFile(filepath.Join(tmpDir, "config.toml"))
		if err != nil {
			t.Fatalf("Failed to read saved config: %v", err)
		}

		if string(data) == "" {
			t.Error("Save() created empty file")
		}
	})

	t.Run("creates directory if missing", func(t *testing.T) {
		tmpDir := filepath.Join(t.TempDir(), "nested", "dir")
		cfg := Config{Dict: "test.db"}

		err := cfg.Save(tmpDir)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		configPath := filepath.Join(tmpDir, "config.toml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Save() did not create config file")
		}
	})

	t.Run("round trip preserves data", func(t *testing.T) {
		tmpDir := t.TempDir()
		original := Config{Dict: "roundtrip.db"}

		err := original.Save(tmpDir)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		loaded, err := LoadFrom(tmpDir)
		if err != nil {
			t.Fatalf("LoadFrom() error = %v", err)
		}

		if loaded.Dict != original.Dict {
			t.Errorf("Round trip: got Dict = %q, want %q", loaded.Dict, original.Dict)
		}
	})
}
