package dict

import (
	"regexp"
	"strings"
)

var (
	genderRe   = regexp.MustCompile(`\{([^}]+)\}`)
	abbrRe     = regexp.MustCompile(`<([^>]+)>`)
	contextRe  = regexp.MustCompile(`\[([^\]]+)\]`)
	optionalRe = regexp.MustCompile(`\(([^)]+)\)`)
)

// Term represents a dictionary term with its metadata separated.
type Term struct {
	Full         string   // Original full term: "Mond {m} [poet.]"
	Text         string   // Clean term: "Mond"
	Gender       string   // "m", "f", "n", "pl", "" (languages with grammatical gender)
	Abbreviation string   // Text in <...>
	Context      []string // Text in [...] - can have multiple
	Optional     []string // Text in (...) - optional parts
	Subjects     []string // Subject areas, e.g., ["astron."]
}

// ParseTerm parses a raw term string into a Term.
func ParseTerm(raw string) Term {
	result := Term{Full: raw}

	text := raw

	// Extract gender {...}
	if matches := genderRe.FindAllStringSubmatch(text, -1); len(matches) > 0 {
		result.Gender = matches[0][1] // Use first match as primary gender
	}
	text = genderRe.ReplaceAllString(text, "")

	// Extract abbreviation <...>
	if matches := abbrRe.FindStringSubmatch(text); len(matches) > 1 {
		result.Abbreviation = matches[1]
	}
	text = abbrRe.ReplaceAllString(text, "")

	// Extract context [...] - can have multiple
	if matches := contextRe.FindAllStringSubmatch(text, -1); len(matches) > 0 {
		for _, m := range matches {
			result.Context = append(result.Context, m[1])
		}
	}
	text = contextRe.ReplaceAllString(text, "")

	// Extract optional parts (...) - can have multiple
	if matches := optionalRe.FindAllStringSubmatch(text, -1); len(matches) > 0 {
		for _, m := range matches {
			result.Optional = append(result.Optional, m[1])
		}
	}
	text = optionalRe.ReplaceAllString(text, "")

	// Clean up whitespace
	result.Text = strings.TrimSpace(strings.Join(strings.Fields(text), " "))

	return result
}
