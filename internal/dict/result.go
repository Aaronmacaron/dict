package dict

// Result represents a single dictionary translation entry.
type Result struct {
	// Core fields (always displayed)
	German   ParsedTerm // German term with parsed metadata
	English  ParsedTerm // English term with parsed metadata
	WordType string     // Word class: noun, verb, adj, adv, etc.

	// Extended fields (exposed in API, not displayed by default)
	SubjectsDE []string // Subject areas in German, e.g., ["astron."]
	SubjectsEN []string // Subject areas in English, e.g., ["astron."]
	Popularity int      // Usage frequency (10-53, higher = more common)
	SortScore  int      // Relevance score (lower = simpler/more relevant)
}
