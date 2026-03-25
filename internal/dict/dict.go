package dict

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Dictionary provides word lookup functionality.
type Dictionary interface {
	Lookup(word string) (*LookupResult, error)
	Close() error
}

// SQLiteDict implements Dictionary using a SQLite database.
type SQLiteDict struct {
	db          *sql.DB
	subjects    map[int]map[int]string // subj_id -> lang_id -> abbr
	term1LangID int                    // lang_id for term1 column
	term2LangID int                    // lang_id for term2 column
}

// Open opens a dictionary database at the given path.
func Open(dbPath string) (*SQLiteDict, error) {
	db, err := sql.Open("sqlite3", dbPath+"?mode=ro")
	if err != nil {
		return nil, err
	}

	// Verify the connection works
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// Load subjects into memory
	subjects, err := loadSubjects(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Detect language IDs and map them to term1/term2
	term1LangID, term2LangID, err := detectTermLangIDs(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteDict{
		db:          db,
		subjects:    subjects,
		term1LangID: term1LangID,
		term2LangID: term2LangID,
	}, nil
}

// loadSubjects loads all subject abbreviations for all languages into memory.
func loadSubjects(db *sql.DB) (map[int]map[int]string, error) {
	rows, err := db.Query("SELECT subj_id, lang_id, abbr FROM subjects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subjects := make(map[int]map[int]string)
	for rows.Next() {
		var subjID, langID int
		var abbr string
		if err := rows.Scan(&subjID, &langID, &abbr); err != nil {
			return nil, err
		}
		if subjects[subjID] == nil {
			subjects[subjID] = make(map[int]string)
		}
		subjects[subjID][langID] = abbr
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return subjects, nil
}

// Close closes the database connection.
func (d *SQLiteDict) Close() error {
	return d.db.Close()
}

// calculateHybridScore combines match quality and popularity into a single score.
// Both matchScore and popularity are in the range [0, 1].
// Returns: weighted hybrid score in range [0, 1].
func calculateHybridScore(matchScore, popularity float64) float64 {
	return matchScore*0.50 + popularity*0.50
}

// Lookup searches for translations of the given word.
func (d *SQLiteDict) Lookup(word string) (*LookupResult, error) {
	// Fetch more results than needed since we'll filter some out
	query := `
		SELECT
			m.term1,
			m.term2,
			m.entry_type,
			m.subj_ids,
			MIN(MAX((m.vt_usage - 10.0) / 43.0, 0.0), 1.0) as popularity
		FROM main_ft m
		WHERE main_ft MATCH ?
		ORDER BY m.vt_usage DESC
		LIMIT 100
	`

	rows, err := d.db.Query(query, word)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scored []ScoredTranslation
	for rows.Next() {
		var lang1Raw, lang2Raw string
		var wordType, subjIDs string
		var popularity float64

		err := rows.Scan(
			&lang1Raw,
			&lang2Raw,
			&wordType,
			&subjIDs,
			&popularity,
		)
		if err != nil {
			return nil, err
		}

		// Parse terms into structured data with subjects
		lang1 := ParseTerm(lang1Raw)
		lang1.Subjects = d.resolveSubjects(subjIDs, d.term1LangID)
		lang2 := ParseTerm(lang2Raw)
		lang2.Subjects = d.resolveSubjects(subjIDs, d.term2LangID)

		tr := Translation{
			Lang1:      lang1,
			Lang2:      lang2,
			WordType:   strings.ToLower(strings.TrimSpace(wordType)),
			Popularity: popularity,
		}

		// Calculate match score (best of either language)
		lang1Score := scoreMatch(lang1.Text, word)
		lang2Score := scoreMatch(lang2.Text, word)
		matchScore := lang1Score
		if lang2Score > matchScore {
			matchScore = lang2Score
		}

		// Only include if there's a match
		if matchScore > 0.0 {
			hybridScore := calculateHybridScore(matchScore, popularity)
			scored = append(scored, ScoredTranslation{Translation: tr, Score: hybridScore})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Sort by hybrid score (desc)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	// Limit to 50
	if len(scored) > 50 {
		scored = scored[:50]
	}

	lang1Name, lang2Name := d.LangNames()
	return &LookupResult{
		Query:        word,
		Lang1Name:    lang1Name,
		Lang2Name:    lang2Name,
		Translations: scored,
	}, nil
}

// scoreMatch calculates how well the term matches the search word.
// Returns a score between 0.0 and 1.0, where higher scores indicate better matches.
// Exact word matches rank higher than prefix matches.
func scoreMatch(text, word string) float64 {
	text = strings.ToLower(text)
	word = strings.ToLower(word)

	// Exact match (whole term)
	if text == word {
		return 1.0
	}

	// Split into words/parts
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '-' || r == '/'
	})

	// Query appears as exact word somewhere (e.g., "full moon")
	for i := 1; i < len(parts); i++ {
		if parts[i] == word {
			return 0.9
		}
	}

	// First word/part is exact match (e.g., "moon landing")
	if len(parts) > 0 && parts[0] == word {
		return 0.8
	}

	// Term starts with the search word as prefix (e.g., "moonlight")
	if strings.HasPrefix(text, word) {
		return 0.6
	}

	// First word/part starts with query (e.g., "Mondschein" for "Mond")
	if len(parts) > 0 && strings.HasPrefix(parts[0], word) {
		return 0.5
	}

	// Query is prefix of any word/part
	for _, part := range parts {
		if strings.HasPrefix(part, word) {
			return 0.4
		}
	}

	// Query appears somewhere in the term
	if strings.Contains(text, word) {
		return 0.2
	}

	return 0.0 // No match
}

// resolveSubjects converts comma-separated subject IDs to subject abbreviations.
func (d *SQLiteDict) resolveSubjects(subjIDs string, langID int) []string {
	if subjIDs == "" {
		return nil
	}

	// subjIDs format: ",3,117," - extract IDs between commas
	ids := strings.Split(strings.Trim(subjIDs, ","), ",")
	if len(ids) == 0 || (len(ids) == 1 && ids[0] == "") {
		return nil
	}

	var subjects []string
	for _, idStr := range ids {
		if idStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		if langMap, ok := d.subjects[id]; ok {
			if abbr, ok := langMap[langID]; ok {
				subjects = append(subjects, abbr)
			}
		}
	}

	return subjects
}

// DetectLanguages returns the distinct language IDs in a dictionary database.
func DetectLanguages(dbPath string) ([]int, error) {
	db, err := sql.Open("sqlite3", dbPath+"?mode=ro")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT DISTINCT lang_id FROM subjects ORDER BY lang_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var langIDs []int
	for rows.Next() {
		var langID int
		if err := rows.Scan(&langID); err != nil {
			return nil, err
		}
		langIDs = append(langIDs, langID)
	}

	return langIDs, rows.Err()
}

// langNames maps language IDs to their English names.
// Derived from all language pairs supported by dict.cc.
var langNames = map[int]string{
	1:   "English",
	2:   "German",
	6:   "Albanian",
	20:  "Bulgarian",
	27:  "Croatian",
	28:  "Czech",
	29:  "Danish",
	30:  "Dutch",
	31:  "Esperanto",
	35:  "Finnish",
	36:  "French",
	41:  "Greek",
	48:  "Hungarian",
	49:  "Icelandic",
	55:  "Italian",
	81:  "Norwegian",
	87:  "Polish",
	88:  "Portuguese",
	92:  "Romanian",
	93:  "Russian",
	97:  "Serbian",
	105: "Slovak",
	108: "Spanish",
	111: "Swedish",
	122: "Turkish",
	138: "Latin",
	143: "Bosnian",
}

// LanguageName maps a language ID to its name.
func LanguageName(langID int) string {
	if name, ok := langNames[langID]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", langID)
}

// langPairKey encodes two language IDs into a single uint32 for map lookup.
// The smaller ID goes in the high 16 bits, the larger in the low 16 bits.
func langPairKey(a, b int) uint32 {
	if a > b {
		a, b = b, a
	}
	return uint32(a)<<16 | uint32(b)
}

// langPairToTermOrder maps a pair of lang IDs (encoded via langPairKey) to
// (term1LangID, term2LangID). Derived from dict.cc's known language pairs.
var langPairToTermOrder = map[uint32][2]int{
	langPairKey(1, 27):  {1, 27},  // EN-HR
	langPairKey(1, 30):  {1, 30},  // EN-NL
	langPairKey(1, 31):  {1, 31},  // EN-EO
	langPairKey(1, 35):  {1, 35},  // EN-FI
	langPairKey(1, 36):  {1, 36},  // EN-FR
	langPairKey(1, 41):  {41, 1},  // EL-EN
	langPairKey(1, 48):  {1, 48},  // EN-HU
	langPairKey(1, 49):  {1, 49},  // EN-IS
	langPairKey(1, 55):  {1, 55},  // EN-IT
	langPairKey(1, 6):   {1, 6},   // EN-SQ
	langPairKey(1, 81):  {1, 81},  // EN-NO
	langPairKey(1, 87):  {1, 87},  // EN-PL
	langPairKey(1, 88):  {1, 88},  // EN-PT
	langPairKey(1, 92):  {1, 92},  // EN-RO
	langPairKey(1, 93):  {1, 93},  // EN-RU
	langPairKey(1, 97):  {1, 97},  // EN-SR
	langPairKey(1, 105): {1, 105}, // EN-SK
	langPairKey(1, 108): {1, 108}, // EN-ES
	langPairKey(1, 111): {1, 111}, // EN-SV
	langPairKey(1, 122): {1, 122}, // EN-TR
	langPairKey(1, 138): {1, 138}, // EN-LA
	langPairKey(1, 2):   {2, 1},   // DE-EN
	langPairKey(2, 6):   {2, 6},   // DE-SQ
	langPairKey(2, 27):  {2, 27},  // DE-HR
	langPairKey(2, 30):  {2, 30},  // DE-NL
	langPairKey(2, 31):  {2, 31},  // DE-EO
	langPairKey(2, 35):  {2, 35},  // DE-FI
	langPairKey(2, 36):  {2, 36},  // DE-FR
	langPairKey(2, 41):  {2, 41},  // DE-EL
	langPairKey(2, 48):  {2, 48},  // DE-HU
	langPairKey(2, 49):  {2, 49},  // DE-IS
	langPairKey(2, 55):  {2, 55},  // DE-IT
	langPairKey(2, 81):  {2, 81},  // DE-NO
	langPairKey(2, 87):  {2, 87},  // DE-PL
	langPairKey(2, 88):  {2, 88},  // DE-PT
	langPairKey(2, 92):  {2, 92},  // DE-RO
	langPairKey(2, 93):  {2, 93},  // DE-RU
	langPairKey(2, 97):  {2, 97},  // DE-SR
	langPairKey(2, 105): {2, 105}, // DE-SK
	langPairKey(2, 108): {2, 108}, // DE-ES
	langPairKey(2, 111): {2, 111}, // DE-SV
	langPairKey(2, 122): {2, 122}, // DE-TR
	langPairKey(2, 138): {2, 138}, // DE-LA
	langPairKey(20, 1):  {20, 1},  // BG-EN
	langPairKey(20, 2):  {20, 2},  // BG-DE
	langPairKey(28, 1):  {28, 1},  // CS-EN
	langPairKey(28, 2):  {28, 2},  // CS-DE
	langPairKey(29, 1):  {29, 1},  // DA-EN
	langPairKey(29, 2):  {29, 2},  // DA-DE
	langPairKey(143, 1): {143, 1}, // BS-EN
	langPairKey(143, 2): {143, 2}, // BS-DE
}

// detectTermLangIDs detects the two language IDs from the subjects table
// and determines which maps to term1 and term2 using the known dict.cc
// language pair ordering.
func detectTermLangIDs(db *sql.DB) (term1LangID, term2LangID int, err error) {
	rows, err := db.Query("SELECT DISTINCT lang_id FROM subjects ORDER BY lang_id")
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	var langIDs []int
	for rows.Next() {
		var langID int
		if err := rows.Scan(&langID); err != nil {
			return 0, 0, err
		}
		langIDs = append(langIDs, langID)
	}
	if err := rows.Err(); err != nil {
		return 0, 0, err
	}

	if len(langIDs) < 2 {
		return 0, 0, fmt.Errorf("expected at least 2 languages in subjects table, got %d", len(langIDs))
	}

	key := langPairKey(langIDs[0], langIDs[1])
	if order, ok := langPairToTermOrder[key]; ok {
		return order[0], order[1], nil
	}

	// Unknown pair: fall back to the order found in the subjects table
	return langIDs[0], langIDs[1], nil
}

// LangNames returns the language names for this dictionary.
func (d *SQLiteDict) LangNames() (lang1Name, lang2Name string) {
	return LanguageName(d.term1LangID), LanguageName(d.term2LangID)
}
