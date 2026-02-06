package dict

import (
	"testing"
)

func TestParseTerm(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantText   string
		wantGender string
		wantAbbr   string
		wantCtx    []string
		wantOpt    []string
	}{
		{
			name:     "simple term",
			input:    "moon",
			wantText: "moon",
		},
		{
			name:       "german noun with masculine gender",
			input:      "Rentner {m}",
			wantText:   "Rentner",
			wantGender: "m",
		},
		{
			name:       "german noun with feminine gender",
			input:      "Schwester {f}",
			wantText:   "Schwester",
			wantGender: "f",
		},
		{
			name:       "german noun with neuter gender",
			input:      "Kind {n}",
			wantText:   "Kind",
			wantGender: "n",
		},
		{
			name:       "german noun with plural",
			input:      "Schwestern {pl}",
			wantText:   "Schwestern",
			wantGender: "pl",
		},
		{
			name:     "term with abbreviation",
			input:    "old-age pensioner <OAP>",
			wantText: "old-age pensioner",
			wantAbbr: "OAP",
		},
		{
			name:     "term with single context",
			input:    "colour [Br.]",
			wantText: "colour",
			wantCtx:  []string{"Br."},
		},
		{
			name:     "term with multiple contexts",
			input:    "bloke [Br.] [coll.]",
			wantText: "bloke",
			wantCtx:  []string{"Br.", "coll."},
		},
		{
			name:     "term with optional parts",
			input:    "to go (somewhere)",
			wantText: "to go",
			wantOpt:  []string{"somewhere"},
		},
		{
			name:       "term with all metadata",
			input:      "Arzt {m} <Dr.> [med.] (fachsprachlich)",
			wantText:   "Arzt",
			wantGender: "m",
			wantAbbr:   "Dr.",
			wantCtx:    []string{"med."},
			wantOpt:    []string{"fachsprachlich"},
		},
		{
			name:     "complex english term",
			input:    "old-age pensioner <OAP> [Br.] [coll.]",
			wantText: "old-age pensioner",
			wantAbbr: "OAP",
			wantCtx:  []string{"Br.", "coll."},
		},
		{
			name:     "empty string",
			input:    "",
			wantText: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			wantText: "",
		},
		{
			name:       "term with extra whitespace",
			input:      "  Mond   {m}  ",
			wantText:   "Mond",
			wantGender: "m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTerm(tt.input)

			if got.Full != tt.input {
				t.Errorf("Full = %q, want %q", got.Full, tt.input)
			}
			if got.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", got.Text, tt.wantText)
			}
			if got.Gender != tt.wantGender {
				t.Errorf("Gender = %q, want %q", got.Gender, tt.wantGender)
			}
			if got.Abbreviation != tt.wantAbbr {
				t.Errorf("Abbreviation = %q, want %q", got.Abbreviation, tt.wantAbbr)
			}
			if !slicesEqual(got.Context, tt.wantCtx) {
				t.Errorf("Context = %v, want %v", got.Context, tt.wantCtx)
			}
			if !slicesEqual(got.Optional, tt.wantOpt) {
				t.Errorf("Optional = %v, want %v", got.Optional, tt.wantOpt)
			}
		})
	}
}

// slicesEqual compares two string slices for equality
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
