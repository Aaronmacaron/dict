package cli

import (
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestCopyFile(t *testing.T) {
	t.Run("copies file content", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := t.TempDir()

		srcPath := filepath.Join(srcDir, "source.txt")
		dstPath := filepath.Join(dstDir, "dest.txt")
		content := []byte("test content for copy")

		err := os.WriteFile(srcPath, content, 0644)
		if err != nil {
			t.Fatalf("Failed to write source file: %v", err)
		}

		err = CopyFile(srcPath, dstPath)
		if err != nil {
			t.Fatalf("CopyFile() error = %v", err)
		}

		got, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("Failed to read dest file: %v", err)
		}

		if string(got) != string(content) {
			t.Errorf("CopyFile() content = %q, want %q", string(got), string(content))
		}
	})

	t.Run("missing source fails", func(t *testing.T) {
		err := CopyFile("/nonexistent/source", filepath.Join(t.TempDir(), "dest"))
		if err == nil {
			t.Error("CopyFile() expected error for missing source")
		}
	})
}

func TestValidateDictFile(t *testing.T) {
	t.Run("valid dictionary passes", func(t *testing.T) {
		dbPath := createValidTestDB(t)

		err := ValidateDictFile(dbPath)
		if err != nil {
			t.Errorf("ValidateDictFile() error = %v", err)
		}
	})

	t.Run("missing subjects table fails", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "empty.db")

		db, err := sql.Open("sqlite3", tmpFile)
		if err != nil {
			t.Fatalf("Failed to create test db: %v", err)
		}
		_, err = db.Exec("CREATE TABLE other (id INTEGER)")
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
		db.Close()

		err = ValidateDictFile(tmpFile)
		if err == nil {
			t.Error("ValidateDictFile() expected error for missing subjects table")
		}
	})

	t.Run("non-sqlite file fails", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "notdb.txt")
		err := os.WriteFile(tmpFile, []byte("not a database"), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		err = ValidateDictFile(tmpFile)
		if err == nil {
			t.Error("ValidateDictFile() expected error for non-sqlite file")
		}
	})
}

func TestDetectLanguagePair(t *testing.T) {
	t.Run("detects language pair", func(t *testing.T) {
		dbPath := setupTestDB(t)

		pair, err := DetectLanguagePair(dbPath)
		if err != nil {
			t.Fatalf("DetectLanguagePair() error = %v", err)
		}

		if pair != "English-German" {
			t.Errorf("DetectLanguagePair() = %q, want %q", pair, "English-German")
		}
	})

	t.Run("invalid database returns error", func(t *testing.T) {
		_, err := DetectLanguagePair("/nonexistent/path.db")
		if err == nil {
			t.Error("DetectLanguagePair() expected error for invalid path")
		}
	})
}

// createValidTestDB creates a minimal valid dictionary database
func createValidTestDB(t *testing.T) string {
	t.Helper()

	tmpFile := filepath.Join(t.TempDir(), "valid.db")

	db, err := sql.Open("sqlite3", tmpFile)
	if err != nil {
		t.Fatalf("Failed to create test db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE subjects (
			subj_id INTEGER,
			lang_id INTEGER,
			abbr VARCHAR
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create subjects table: %v", err)
	}

	_, err = db.Exec(`INSERT INTO subjects VALUES (1, 1, 'test.')`)
	if err != nil {
		t.Fatalf("Failed to insert subject: %v", err)
	}

	return tmpFile
}

// setupTestDB copies the fixture database to a temp directory
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
