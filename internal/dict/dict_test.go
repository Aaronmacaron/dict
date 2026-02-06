package dict

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestScoreMatch(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		word      string
		wantScore int
	}{
		// Exact matches
		{"exact match", "moon", "moon", 1000},
		{"exact match case insensitive", "Moon", "moon", 1000},
		{"exact match uppercase query", "moon", "MOON", 1000},

		// First word exact match
		{"first word exact", "moon landing", "moon", 900},
		{"first word exact case insensitive", "Moon landing", "moon", 900},

		// Later word exact match
		{"later word exact", "full moon", "moon", 800},
		{"middle word exact", "the full moon rises", "moon", 800},

		// Prefix of whole term
		{"prefix of whole", "moonlight", "moon", 600},
		{"prefix of whole case insensitive", "Moonlight", "moon", 600},

		// Prefix of any word
		{"prefix of later word", "full moonlight", "moon", 400},

		// Substring match
		{"substring", "Halbmond", "mond", 200},
		{"substring in middle", "Vollmondnacht", "mond", 200},

		// No match
		{"no match", "sun", "moon", 0},
		{"empty text", "", "moon", 0},
		// Note: empty word returns 600 because HasPrefix(text, "") is always true
		{"empty word matches prefix", "moon", "", 600},

		// Hyphenated and slashed terms
		{"hyphenated match", "half-moon", "moon", 800},
		{"slash separated", "day/night", "night", 800},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScoreMatch(tt.text, tt.word)
			if got != tt.wantScore {
				t.Errorf("ScoreMatch(%q, %q) = %d, want %d", tt.text, tt.word, got, tt.wantScore)
			}
		})
	}
}

func TestLanguageName(t *testing.T) {
	tests := []struct {
		langID int
		want   string
	}{
		{1, "English"},
		{2, "German"},
		{36, "French"},
		{55, "Italian"},
		{108, "Spanish"},
		{999, "Unknown(999)"},
		{0, "Unknown(0)"},
		{-1, "Unknown(-1)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := LanguageName(tt.langID)
			if got != tt.want {
				t.Errorf("LanguageName(%d) = %q, want %q", tt.langID, got, tt.want)
			}
		})
	}
}

func TestDetectLanguages(t *testing.T) {
	t.Run("detects languages from fixture", func(t *testing.T) {
		dbPath := setupTestDB(t)

		langs, err := DetectLanguages(dbPath)
		if err != nil {
			t.Fatalf("DetectLanguages() error = %v", err)
		}

		if len(langs) != 2 {
			t.Errorf("DetectLanguages() returned %d languages, want 2", len(langs))
		}

		// Should be sorted: 1 (English), 2 (German)
		if len(langs) >= 2 && (langs[0] != 1 || langs[1] != 2) {
			t.Errorf("DetectLanguages() = %v, want [1, 2]", langs)
		}
	})

	t.Run("invalid database returns error", func(t *testing.T) {
		_, err := DetectLanguages("/nonexistent/path.db")
		if err == nil {
			t.Error("DetectLanguages() expected error for nonexistent path")
		}
	})
}

func TestOpen(t *testing.T) {
	t.Run("opens valid database", func(t *testing.T) {
		dbPath := setupTestDB(t)

		d, err := Open(dbPath)
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		defer d.Close()

		if d.db == nil {
			t.Error("Open() returned SQLiteDict with nil db")
		}
		if d.subjects == nil {
			t.Error("Open() returned SQLiteDict with nil subjects")
		}
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		_, err := Open("/nonexistent/path.db")
		if err == nil {
			t.Error("Open() expected error for nonexistent path")
		}
	})
}

func TestLookup(t *testing.T) {
	dbPath := setupTestDB(t)
	d, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer d.Close()

	t.Run("finds exact match", func(t *testing.T) {
		results, err := d.Lookup("moon")
		if err != nil {
			t.Fatalf("Lookup() error = %v", err)
		}

		if len(results) == 0 {
			t.Fatal("Lookup() returned no results")
		}

		// First result should be exact match
		if results[0].English.Text != "moon" {
			t.Errorf("First result English.Text = %q, want %q", results[0].English.Text, "moon")
		}
	})

	t.Run("scores exact matches higher than prefix", func(t *testing.T) {
		results, err := d.Lookup("moon")
		if err != nil {
			t.Fatalf("Lookup() error = %v", err)
		}

		if len(results) < 2 {
			t.Skip("Need at least 2 results to test ordering")
		}

		// Exact match "moon" should come before "moonlight"
		foundExact := false
		foundPrefix := false
		exactPos := -1
		prefixPos := -1

		for i, r := range results {
			if r.English.Text == "moon" {
				foundExact = true
				exactPos = i
			}
			if r.English.Text == "moonlight" {
				foundPrefix = true
				prefixPos = i
			}
		}

		if foundExact && foundPrefix && exactPos > prefixPos {
			t.Errorf("Exact match 'moon' (pos %d) should come before 'moonlight' (pos %d)", exactPos, prefixPos)
		}
	})

	t.Run("resolves subjects", func(t *testing.T) {
		results, err := d.Lookup("moon")
		if err != nil {
			t.Fatalf("Lookup() error = %v", err)
		}

		if len(results) == 0 {
			t.Fatal("Lookup() returned no results")
		}

		// First result should have astronomy subject
		if len(results[0].SubjectsEN) == 0 {
			t.Error("Expected SubjectsEN to be populated")
		} else if results[0].SubjectsEN[0] != "astron." {
			t.Errorf("SubjectsEN[0] = %q, want %q", results[0].SubjectsEN[0], "astron.")
		}
	})

	t.Run("no results for unknown word", func(t *testing.T) {
		results, err := d.Lookup("xyznonexistent")
		if err != nil {
			t.Fatalf("Lookup() error = %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Lookup() returned %d results, want 0", len(results))
		}
	})
}

// setupTestDB copies the fixture database to a temp directory for test isolation
func setupTestDB(t *testing.T) string {
	t.Helper()

	src := "../../testdata/fixtures/minimal.db"
	dst := filepath.Join(t.TempDir(), "test.db")

	srcFile, err := os.Open(src)
	if err != nil {
		t.Fatalf("Failed to open fixture: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		t.Fatalf("Failed to create temp db: %v", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		t.Fatalf("Failed to copy fixture: %v", err)
	}

	return dst
}
