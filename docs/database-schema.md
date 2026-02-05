# Database Schema

The dictionary database (`dictcc-en-de.db`) contains English-German translations from dict.cc.

## Main Tables

### `main_ft` (FTS3 Virtual Table)

The primary table for dictionary lookups. Uses SQLite Full-Text Search (FTS3).

| Column       | Type    | Description                                      |
|--------------|---------|--------------------------------------------------|
| `id`         | INTEGER | Primary key                                      |
| `term1`      | VARCHAR | German term                                      |
| `term2`      | VARCHAR | English term                                     |
| `sort1`      | INTEGER | Sort priority for term1                          |
| `sort2`      | INTEGER | Sort priority for term2                          |
| `subj_ids`   | VARCHAR | Comma-separated subject IDs (e.g., `,117,`)      |
| `entry_type` | VARCHAR | Word class (noun, verb, adj, adv, etc.)          |
| `vt_usage`   | INTEGER | Usage/frequency indicator                        |

**Entry Types:** `noun`, `verb`, `adj`, `adv`, `prep`, `conj`, `pron`, `prefix`, `suffix`, `past-p`, `pres-p`, and combinations like `adj past-p`

**Example row:**
```
id: 1
term1: Rentner {m}
term2: old-age pensioner <OAP> [Br.]
entry_type: noun
```

### `subjects`

Subject/category classifications for specialized terminology.

| Column        | Type    | Description                          |
|---------------|---------|--------------------------------------|
| `subj_id`     | INTEGER | Subject ID (referenced in main_ft)   |
| `lang_id`     | INTEGER | Language (1=English, 2=German)       |
| `abbr`        | VARCHAR | Abbreviation (e.g., "med.", "jur.")  |
| `description` | VARCHAR | Full description                     |

**Example subjects:** `med.` (Medicine), `jur.` (Law), `chem.` (Chemistry)

### `singlewords` and `singlewords_[a-z]`

**Autocomplete/search suggestion index** - not used for actual translations.

These tables extract and deduplicate searchable terms from `main_ft` phrases. They are partitioned alphabetically (`singlewords_a` through `singlewords_z`, plus base `singlewords` for symbols/numbers) to enable fast prefix lookups.

| Column        | Type    | Description                                    |
|---------------|---------|------------------------------------------------|
| `id`          | INTEGER | Primary key (not related to main_ft.id)        |
| `colnum`      | INTEGER | Language: 1 = German, 2 = English              |
| `term`        | VARCHAR | Original term with casing preserved            |
| `term4search` | VARCHAR | Normalized lowercase for case-insensitive search |

**Example:**
```
term: Moon Boot
term4search: moon boot
colnum: 2 (English)
```

**Typical GUI workflow:**
1. User types "moo" → query `singlewords_m` for autocomplete suggestions
2. User selects "moon" → query `main_ft` for full translations

**For a CLI tool:** These tables can be skipped entirely. Use `main_ft` with FTS3 MATCH queries directly since there's no autocomplete UI.

**Statistics:** ~1.7 million entries total across all singlewords tables

## Querying

### Basic search using FTS3

```sql
SELECT term1, term2, entry_type
FROM main_ft
WHERE main_ft MATCH 'moon';
```

### Search with LIKE (for partial matches)

```sql
SELECT term1, term2, entry_type
FROM main_ft
WHERE term1 LIKE '%moon%' OR term2 LIKE '%moon%'
LIMIT 20;
```

### Join with subjects for category info

```sql
SELECT m.term1, m.term2, m.entry_type, s.abbr, s.description
FROM main_ft m
LEFT JOIN subjects s ON m.subj_ids LIKE '%,' || s.subj_id || ',%'
WHERE m.term2 LIKE '%chemistry%';
```

## Statistics

- Total entries: ~1.3 million translations
- Database size: ~287 MB
