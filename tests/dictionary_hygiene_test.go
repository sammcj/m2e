package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const dictionaryPath = "../pkg/converter/data/american_spellings.json"

// keyPattern matches lowercase single tokens: letters with optional internal
// hyphens/apostrophes. Keys that don't match can never match at runtime because
// the converter lowercases and whitespace-tokenises input before lookup.
var keyPattern = regexp.MustCompile(`^[a-z]+(?:[-'][a-z]+)*$`)

func loadDictionary(t *testing.T) map[string]string {
	t.Helper()
	data, err := os.ReadFile(dictionaryPath)
	if err != nil {
		t.Fatalf("Failed to read dictionary %s: %v", dictionaryPath, err)
	}
	var dict map[string]string
	if err := json.Unmarshal(data, &dict); err != nil {
		t.Fatalf("Failed to parse dictionary %s: %v", dictionaryPath, err)
	}
	if len(dict) == 0 {
		t.Fatalf("Dictionary %s is empty", dictionaryPath)
	}
	return dict
}

// reportOffenders fails the test listing up to 20 offenders followed by a
// summary of the remainder, keeping large failures readable.
func reportOffenders(t *testing.T, invariant string, offenders []string) {
	t.Helper()
	if len(offenders) == 0 {
		return
	}
	sort.Strings(offenders)
	const cap = 20
	shown := offenders
	if len(shown) > cap {
		shown = shown[:cap]
	}
	msg := fmt.Sprintf("%s: %d offending entr(y/ies):\n  %s", invariant,
		len(offenders), strings.Join(shown, "\n  "))
	if len(offenders) > cap {
		msg += fmt.Sprintf("\n  ...and %d more", len(offenders)-cap)
	}
	t.Error(msg)
}

func TestDictionaryHygiene(t *testing.T) {
	dict := loadDictionary(t)

	t.Run("KeysAreLowercaseSingleTokens", func(t *testing.T) {
		var offenders []string
		for key := range dict {
			if !keyPattern.MatchString(key) {
				offenders = append(offenders, fmt.Sprintf("%q", key))
			}
		}
		reportOffenders(t, "keys must be lowercase single tokens", offenders)
	})

	t.Run("ValuesAreCleanLowercase", func(t *testing.T) {
		var offenders []string
		for key, value := range dict {
			if value == "" {
				offenders = append(offenders, fmt.Sprintf("%q -> (empty)", key))
				continue
			}
			if value != strings.TrimSpace(value) {
				offenders = append(offenders, fmt.Sprintf("%q -> %q (whitespace)", key, value))
			}
			if value != strings.ToLower(value) {
				offenders = append(offenders, fmt.Sprintf("%q -> %q (not lowercase)", key, value))
			}
		}
		reportOffenders(t, "values must be non-empty, trimmed, lowercase", offenders)
	})

	t.Run("NoSelfMapping", func(t *testing.T) {
		var offenders []string
		for key, value := range dict {
			if key == value {
				offenders = append(offenders, fmt.Sprintf("%q", key))
			}
		}
		reportOffenders(t, "keys must not map to themselves", offenders)
	})

	t.Run("NoTargetIsAlsoSource", func(t *testing.T) {
		values := make(map[string]bool, len(dict))
		for _, value := range dict {
			values[value] = true
		}
		var offenders []string
		for key := range dict {
			if values[key] {
				offenders = append(offenders, fmt.Sprintf("%q", key))
			}
		}
		reportOffenders(t, "conversion targets must not also be sources", offenders)
	})
}
