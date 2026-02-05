package dict

import (
	"database/sql"
	"sort"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Dictionary provides word lookup functionality.
type Dictionary interface {
	Lookup(word string) ([]Result, error)
	Close() error
}

// SQLiteDict implements Dictionary using a SQLite database.
type SQLiteDict struct {
	db       *sql.DB
	subjects map[int]map[int]string // subj_id -> lang_id -> abbr
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

	return &SQLiteDict{db: db, subjects: subjects}, nil
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

// scoredResult pairs a result with its match score for sorting.
type scoredResult struct {
	result Result
	score  int
}

// Lookup searches for translations of the given word.
func (d *SQLiteDict) Lookup(word string) ([]Result, error) {
	// Fetch more results than needed since we'll filter some out
	query := `
		SELECT
			m.term1,
			m.term2,
			m.entry_type,
			m.subj_ids,
			m.vt_usage,
			m.sort2
		FROM main_ft m
		WHERE main_ft MATCH ?
		ORDER BY m.vt_usage DESC, m.sort2 ASC
		LIMIT 100
	`

	rows, err := d.db.Query(query, word)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scored []scoredResult
	for rows.Next() {
		var germanRaw, englishRaw string
		var r Result
		var subjIDs string
		var sort2 int

		err := rows.Scan(
			&germanRaw,
			&englishRaw,
			&r.WordType,
			&subjIDs,
			&r.Popularity,
			&sort2,
		)
		if err != nil {
			return nil, err
		}

		// Parse terms into structured data
		r.German = ParseTerm(germanRaw)
		r.English = ParseTerm(englishRaw)
		r.SortScore = sort2
		// Resolve subjects for both languages
		r.SubjectsEN = d.resolveSubjects(subjIDs, 1) // lang_id 1 = English
		r.SubjectsDE = d.resolveSubjects(subjIDs, 2) // lang_id 2 = German

		// Calculate match score (best of German or English)
		germanScore := scoreMatch(r.German.Text, word)
		englishScore := scoreMatch(r.English.Text, word)
		score := germanScore
		if englishScore > score {
			score = englishScore
		}

		// Only include if there's a match
		if score > 0 {
			scored = append(scored, scoredResult{r, score})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Sort by score (desc), then popularity (desc), then sort2 (asc)
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		if scored[i].result.Popularity != scored[j].result.Popularity {
			return scored[i].result.Popularity > scored[j].result.Popularity
		}
		return scored[i].result.SortScore < scored[j].result.SortScore
	})

	// Extract results (limit to 20)
	results := make([]Result, 0, 20)
	for i := 0; i < len(scored) && i < 20; i++ {
		results = append(results, scored[i].result)
	}

	return results, nil
}

// scoreMatch calculates how well the term matches the search word.
// Higher scores indicate better matches.
// Exact word matches rank higher than prefix matches.
func scoreMatch(text, word string) int {
	text = strings.ToLower(text)
	word = strings.ToLower(word)

	// Exact match (whole term)
	if text == word {
		return 1000
	}

	// Split into words/parts
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '-' || r == '/'
	})

	// First word/part is exact match (e.g., "moon landing")
	if len(parts) > 0 && parts[0] == word {
		return 900
	}

	// Query appears as exact word somewhere (e.g., "full moon")
	for i := 1; i < len(parts); i++ {
		if parts[i] == word {
			return 800
		}
	}

	// Term starts with the search word as prefix (e.g., "moonlight")
	if strings.HasPrefix(text, word) {
		return 600
	}

	// First word/part starts with query (e.g., "Mondschein" for "Mond")
	if len(parts) > 0 && strings.HasPrefix(parts[0], word) {
		return 500
	}

	// Query is prefix of any word/part
	for _, part := range parts {
		if strings.HasPrefix(part, word) {
			return 400
		}
	}

	// Query appears somewhere in the term
	if strings.Contains(text, word) {
		return 200
	}

	return 0 // No match
}

// resolveSubjects converts comma-separated subject IDs to subject abbreviations.
// langID: 1 = English, 2 = German
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
