package cli

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"example.com/dict/internal/config"
	"example.com/dict/internal/dict"
)

// runRegister copies a dictionary file to config/dicts/.
func runRegister(ctx *Context, path string) error {
	dictsDir := config.DictsDir(ctx.ConfigDir)

	// Ensure dicts directory exists
	if err := os.MkdirAll(dictsDir, 0755); err != nil {
		return fmt.Errorf("creating dicts directory: %w", err)
	}

	// Validate the file is a valid dictionary
	if err := validateDictFile(path); err != nil {
		return fmt.Errorf("invalid dictionary file: %w", err)
	}

	// Copy file to dicts directory
	destName := filepath.Base(path)
	destPath := filepath.Join(dictsDir, destName)

	if err := copyFile(path, destPath); err != nil {
		return fmt.Errorf("copying dictionary: %w", err)
	}

	fmt.Printf("Registered dictionary: %s\n", destName)

	// Auto-set as default if no default is configured
	if ctx.Config.Dict == "" {
		ctx.Config.Dict = destName
		if err := ctx.Config.Save(ctx.ConfigDir); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("Default dictionary set to: %s\n", destName)
	}

	return nil
}

// runDefault sets the default dictionary in config.toml.
func runDefault(ctx *Context, name string) error {
	// Verify the dictionary exists
	dictPath := filepath.Join(config.DictsDir(ctx.ConfigDir), name)
	if _, err := os.Stat(dictPath); os.IsNotExist(err) {
		return fmt.Errorf("dictionary not found: %s", name)
	}

	// Update config
	ctx.Config.Dict = name
	if err := ctx.Config.Save(ctx.ConfigDir); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Default dictionary set to: %s\n", name)
	return nil
}

// runList lists all dictionaries in config/dicts/.
func runList(ctx *Context) error {
	dictsDir := config.DictsDir(ctx.ConfigDir)

	entries, err := os.ReadDir(dictsDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No dictionaries registered.")
			fmt.Println("Use 'dict --register <path>' to add a dictionary.")
			return nil
		}
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No dictionaries registered.")
		fmt.Println("Use 'dict --register <path>' to add a dictionary.")
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		dictPath := filepath.Join(dictsDir, name)

		langPair, err := detectLanguagePair(dictPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s - %v\n", name, err)
			continue
		}

		// Mark default with asterisk
		marker := "  "
		if name == ctx.Config.Dict {
			marker = "* "
		}

		fmt.Printf("%s%s (%s)\n", marker, name, langPair)
	}

	return nil
}

// validateDictFile checks if a file is a valid dictionary database.
func validateDictFile(path string) error {
	db, err := sql.Open("sqlite3", path+"?mode=ro")
	if err != nil {
		return fmt.Errorf("cannot open as SQLite: %w", err)
	}
	defer db.Close()

	// Check for required tables
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM subjects LIMIT 1").Scan(&count)
	if err != nil {
		return fmt.Errorf("missing subjects table: %w", err)
	}

	return nil
}

// detectLanguagePair detects the language pair in a dictionary.
func detectLanguagePair(dbPath string) (string, error) {
	d, err := dict.Open(dbPath)
	if err != nil {
		return "", fmt.Errorf("not a valid dictionary: %w", err)
	}
	defer d.Close()

	lang1, lang2 := d.LangNames()
	return fmt.Sprintf("%s-%s", lang1, lang2), nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
