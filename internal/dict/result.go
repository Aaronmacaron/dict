package dict

// Translation represents a bidirectional translation entry (stable, query-independent).
type Translation struct {
	Lang1      Term
	Lang2      Term
	WordType   string  // Word class: noun, verb, adj, adv, etc.
	Popularity float64 // Normalized usage frequency (0.0-1.0, higher = more common)
}

// ScoredTranslation pairs a translation with its query-specific relevance score.
type ScoredTranslation struct {
	Translation Translation
	Score       float64 // hybrid score 0.0–1.0
}

// LookupResult contains the results of a dictionary lookup.
type LookupResult struct {
	Query        string
	Lang1Name    string // Human-readable name of the term1 language
	Lang2Name    string // Human-readable name of the term2 language
	Translations []ScoredTranslation
}
