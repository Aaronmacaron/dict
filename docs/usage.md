# Dict User Guide

Dict is an offline dictionary program for looking up word translations. It uses SQLite databases from dict.cc to provide fast, local lookups without requiring an internet connection.

## Quick Start

1. Register a dictionary:
   ```sh
   dict m register path/to/dictcc-en-de.db
   ```

2. Look up a word:
   ```sh
   dict moon
   ```

## Commands

### Word Lookup (Default)

Look up translations for a word:

```sh
dict <word>
```

**Examples:**
```sh
dict moon          # Look up "moon"
dict Mond          # Look up "Mond"
dict "full moon"   # Look up a phrase
```

### Dictionary Management

Management commands are accessed via `dict manage` or the shorthand `dict m`.

#### List Dictionaries

Show all registered dictionaries with their language pairs:

```sh
dict m list
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
dict m register <path>
```

**Example:**
```sh
dict m register ~/Downloads/dictcc-fr-en.db
```

#### Set Default Dictionary

Change the default dictionary:

```sh
dict m default <name>
```

**Example:**
```sh
dict m default dictcc-fr-en.db
```

## Global Flags

These flags can be used with any command:

| Flag | Short | Description |
|------|-------|-------------|
| `--dict` | `-d` | Use a specific dictionary (by name) |
| `--config-dir` | `-c` | Use a custom config directory |

**Examples:**
```sh
dict -d dictcc-fr-en.db bonjour     # Use French-English dictionary
dict -c /custom/path m list         # Use custom config directory
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

Results are displayed in a table with German and English columns:

- **Bold text**: Matches your search term
- **{m} {f} {n} {pl}**: Grammatical gender (masculine, feminine, neuter, plural)
- **[context]**: Usage context or notes
- **<abbr>**: Abbreviations or special markers
- **Gray tags**: Subject areas (e.g., astron., law, med.)
