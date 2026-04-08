# Dict User Guide

Dict is an offline dictionary program for looking up word translations. It uses SQLite databases from dict.cc to provide fast, local lookups without requiring an internet connection.

## Quick Start

1. Register a dictionary:
   ```sh
   dict --register path/to/dictcc-en-de.db
   ```

2. Look up a word:
   ```sh
   dict moon
   ```

## Word Lookup

Look up translations for a word:

```sh
dict <word>
```

**Examples:**
```sh
dict moon          # Look up "moon"
dict Mond          # Look up "Mond"
dict "full moon"   # Look up a phrase
dict -a moon       # Show all results (no per-group limit)
```

Results are grouped by word type (nouns, verbs, adjectives, etc.) with up to 5 results per group. Use `--all` / `-a` to show all results.

## Dictionary Management

#### List Dictionaries

Show all registered dictionaries with their language pairs:

```sh
dict --list
dict -l
```

**Output:**
```
* dictcc-en-de.db (English-German)
  dictcc-es-en.db (Spanish-English)
```

The `*` indicates the default dictionary.

#### Register a Dictionary

Copy a dictionary file to the config directory:

```sh
dict --register <path>
```

**Example:**
```sh
dict --register ~/Downloads/dictcc-fr-en.db
```

#### Set Default Dictionary

Change the default dictionary:

```sh
dict --default <name>
```

**Example:**
```sh
dict --default dictcc-fr-en.db
```

## Flags

### Lookup Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--all` | `-a` | Show all results (no per-group limit) |

### Management Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--list` | `-l` | List registered dictionaries |
| `--register PATH` | | Register a dictionary file |
| `--default NAME` | | Set the default dictionary |

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--dict` | `-d` | Use a specific dictionary (by name) |
| `--config-dir` | `-c` | Use a custom config directory |
| `--color` | | Color output mode (auto, always, never) |
| `--version` | `-v` | Show version |

**Examples:**
```sh
dict -d dictcc-fr-en.db bonjour     # Use French-English dictionary
dict -c /custom/path --list         # Use custom config directory
```

## Configuration

### Config Location

Configuration is stored in platform-specific directories:

| Platform | Location |
|----------|----------|
| Linux | `~/.config/dict/` |
| macOS | `~/Library/Application Support/dict/` |
| Windows | `%AppData%\dict\` |

### Directory Structure

```
dict/
├── config.toml     # Configuration file
└── dicts/          # Dictionary databases
    ├── dictcc-en-de.db
    └── dictcc-fr-en.db
```

### config.toml

```toml
# Default dictionary (filename in dicts/ folder)
dict = "dictcc-en-de.db"
```

## Output Format

Results are grouped by word type (e.g., Nouns, Verbs, Adjectives) with a colored section header for each group. Each group shows up to 5 results by default, with a `+ N more` hint when truncated.

Within each group, translations are displayed in a table with German and English columns:

- **Bold text**: Matches your search term
- **{m} {f} {n} {pl}**: Grammatical gender (masculine, feminine, neuter, plural)
- **[context]**: Usage context or notes
- **<abbr>**: Abbreviations or special markers
- **Gray tags**: Subject areas (e.g., astron., law, med.)
