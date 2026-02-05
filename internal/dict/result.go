package dict

// Result represents a single dictionary translation entry.
type Result struct {
	// Core fields (always displayed)
	German   ParsedTerm // German term with parsed metadata
	English  ParsedTerm // English term with parsed metadata
	WordType string     // Word class: noun, verb, adj, adv, etc.

	// Extended fields (exposed in API, not displayed by default)
	Subjects   []string // Subject areas, e.g., ["astron."]
	Popularity int      // Usage frequency (10-53, higher = more common)
	SortScore  int      // Relevance score (lower = simpler/more relevant)
}
