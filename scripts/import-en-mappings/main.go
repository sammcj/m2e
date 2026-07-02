// Command import-en-mappings merges vetted entries from tmgldn/en-mappings
// into m2e's built-in dictionary (pkg/converter/data/american_spellings.json).
//
// Source dataset: https://github.com/tmgldn/en-mappings (_spellings.ts),
// vendored here as spellings_snapshot.ts at upstream commit
// b0bab798cf62f186d091f014453eafaee0672667. The author offered the dataset to
// m2e and gave permission in https://github.com/sammcj/m2e/issues/29.
//
// The snapshot rows are tuples: [from, to, locale, confidence, caseFlag?, wildcardFlag?].
// Only plain string targets at confidence 2 are considered; suffix-function
// rules, style-only entries (hyphenation, multi-word targets) and anything in
// blocklist.json are dropped. Post-merge dictionary invariants are enforced
// before anything is written (mirroring tests/dictionary_hygiene_test.go).
//
// Usage (from repo root):
//
//	go run ./scripts/import-en-mappings -report /tmp/import-report.txt          # dry run
//	go run ./scripts/import-en-mappings -write                                  # apply merge
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"os"
	"regexp"
	"sort"
	"strings"
)

// mapping is one upstream tuple row. The tuple's locale field is deliberately
// not carried: locale-specific risks are handled by the curated blocklist.
// Flagged marks rows the upstream author annotated with a trailing comment
// ("newly added, should check locales") - unvetted, so never imported.
type mapping struct {
	From    string
	To      string
	Flagged bool
}

type stats map[string]int

var keyRe = regexp.MustCompile(`^[a-z]+(?:[-'][a-z]+)*$`)

func main() {
	snapshot := flag.String("snapshot", "scripts/import-en-mappings/spellings_snapshot.ts", "path to vendored _spellings.ts")
	dictPath := flag.String("dict", "pkg/converter/data/american_spellings.json", "path to m2e dictionary")
	blockPath := flag.String("blocklist", "scripts/import-en-mappings/blocklist.json", "path to curated blocklist")
	reportPath := flag.String("report", "", "write full report to this path (stdout summary always printed)")
	write := flag.Bool("write", false, "write merged dictionary (default: dry run)")
	flag.Parse()

	if err := run(*snapshot, *dictPath, *blockPath, *reportPath, *write); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(snapshot, dictPath, blockPath, reportPath string, write bool) error {
	rows, err := parseSnapshot(snapshot)
	if err != nil {
		return err
	}
	dict, err := loadDict(dictPath)
	if err != nil {
		return err
	}
	blocklist, err := loadBlocklist(blockPath)
	if err != nil {
		return err
	}

	added, dropStats, conflicts := filterCandidates(rows, dict, blocklist)

	merged := make(map[string]string, len(dict)+len(added))
	maps.Copy(merged, dict)
	for _, m := range added {
		merged[m.From] = m.To
	}
	if violations := checkInvariants(merged); len(violations) > 0 {
		return fmt.Errorf("merged dictionary violates invariants, not writing:\n  %s",
			strings.Join(violations, "\n  "))
	}

	report := buildReport(len(dict), added, dropStats, conflicts)
	fmt.Print(summarise(len(dict), len(merged), added, dropStats, conflicts))
	if reportPath != "" {
		if err := os.WriteFile(reportPath, []byte(report), 0o644); err != nil {
			return err
		}
		fmt.Printf("full report: %s\n", reportPath)
	}

	if !write {
		fmt.Println("dry run only - pass -write to update the dictionary")
		return nil
	}
	if err := writeDict(dictPath, merged); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d entries)\n", dictPath, len(merged))
	return nil
}

// parseSnapshot extracts [from, to, locale, confidence, ...] tuple rows from
// the vendored TypeScript source.
func parseSnapshot(path string) ([]mapping, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading snapshot: %w", err)
	}
	rowRe := regexp.MustCompile(`(?m)^\s*\[(.+?)\],?\s*(//.*)?$`)
	partRe := regexp.MustCompile(`"(?:[^"\\]|\\.)*"|[^,\s][^,]*`)

	var rows []mapping
	for _, row := range rowRe.FindAllStringSubmatch(string(data), -1) {
		parts := partRe.FindAllString(row[1], -1)
		if len(parts) < 4 {
			continue
		}
		for i, p := range parts {
			parts[i] = strings.Trim(strings.TrimSpace(p), `"`)
		}
		// Function targets (processIz/processLyz suffix rules) and low-confidence
		// entries are out of scope: m2e enumerates exact inflections only.
		if parts[1] == "processIz" || parts[1] == "processLyz" || parts[3] != "2" {
			continue
		}
		var locale int
		if _, err := fmt.Sscanf(parts[2], "%d", &locale); err != nil {
			continue
		}
		rows = append(rows, mapping{From: parts[0], To: parts[1], Flagged: row[2] != ""})
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no usable rows parsed from %s", path)
	}
	return rows, nil
}

func loadDict(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading dictionary: %w", err)
	}
	var dict map[string]string
	if err := json.Unmarshal(data, &dict); err != nil {
		return nil, fmt.Errorf("parsing dictionary: %w", err)
	}
	return dict, nil
}

func loadBlocklist(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading blocklist: %w", err)
	}
	var wrapper struct {
		Blocked map[string]string `json:"blocked"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("parsing blocklist: %w", err)
	}
	return wrapper.Blocked, nil
}

// filterCandidates applies the mechanical and curated filters, returning the
// entries to add, drop counts per reason, and conflicts with existing entries.
func filterCandidates(rows []mapping, dict, blocklist map[string]string) ([]mapping, stats, []string) {
	var added []mapping
	drops := stats{}
	var conflicts []string
	seen := make(map[string]string)
	for _, m := range rows {
		switch {
		case m.Flagged:
			drops["upstream comment-flagged (unvetted)"]++
		case seen[m.From] == m.To:
			drops["duplicate upstream row"]++
		case seen[m.From] != "":
			conflicts = append(conflicts, fmt.Sprintf("%s: upstream rows disagree, %q vs %q (keeping first)", m.From, seen[m.From], m.To))
		case !keyRe.MatchString(m.From):
			drops["non-lowercase or multi-word key"]++
		case strings.Contains(m.To, " "):
			drops["multi-word target (style, not spelling)"]++
		case strings.ReplaceAll(m.To, "-", "") == m.From:
			drops["hyphenation style"]++
		case !keyRe.MatchString(m.To):
			drops["target fails charset check"]++
		case m.From == m.To:
			drops["identity mapping"]++
		case blocklist[m.From] != "":
			drops["blocklisted: "+blocklist[m.From]]++
		case dict[m.From] == m.To:
			drops["already in dictionary"]++
		case dict[m.From] != "":
			conflicts = append(conflicts, fmt.Sprintf("%s: m2e=%q theirs=%q (keeping m2e)", m.From, dict[m.From], m.To))
		default:
			seen[m.From] = m.To
			added = append(added, m)
		}
	}
	return added, drops, conflicts
}

// checkInvariants mirrors tests/dictionary_hygiene_test.go: it must hold for
// the merged dictionary before it can be written.
func checkInvariants(dict map[string]string) []string {
	values := make(map[string]string, len(dict))
	for k, v := range dict {
		values[v] = k
	}
	var violations []string
	for k, v := range dict {
		switch {
		case !keyRe.MatchString(k):
			violations = append(violations, fmt.Sprintf("key %q fails charset/case check", k))
		case v == "" || v != strings.ToLower(v) || strings.TrimSpace(v) != v:
			violations = append(violations, fmt.Sprintf("value %q for key %q fails case/whitespace check", v, k))
		case k == v:
			violations = append(violations, fmt.Sprintf("self-mapping %q", k))
		case values[k] != "":
			violations = append(violations, fmt.Sprintf("key %q is also the target of %q -> chain/valid-British risk", k, values[k]))
		}
	}
	sort.Strings(violations)
	return violations
}

// classify buckets an added mapping by transformation pattern so the review
// report groups mechanical families together and isolates the rest.
func classify(from, to string) string {
	type rule struct {
		name string
		ok   func(string, string) bool
	}
	swap := func(old, new string) func(string, string) bool {
		return func(f, t string) bool { return strings.ReplaceAll(f, old, new) == t }
	}
	rules := []rule{
		{"-iz- to -is-", swap("iz", "is")},
		{"-yz- to -ys-", swap("yz", "ys")},
		{"-or to -our", swap("or", "our")},
		{"-iz-/-or combined", func(f, t string) bool {
			return strings.ReplaceAll(strings.ReplaceAll(f, "iz", "is"), "or", "our") == t
		}},
		{"-og to -ogue", swap("og", "ogue")},
		{"-ense to -ence", swap("ense", "ence")},
		{"-er to -re", suffixSwap(map[string]string{"er": "re", "ers": "res", "ered": "red", "ering": "ring"})},
		{"single letter doubled", letterDoubled},
		{"single letter inserted (digraph/e)", letterInserted},
	}
	for _, r := range rules {
		if r.ok(from, to) {
			return r.name
		}
	}
	return "other"
}

func suffixSwap(pairs map[string]string) func(string, string) bool {
	return func(f, t string) bool {
		for old, new := range pairs {
			if strings.HasSuffix(f, old) && t == f[:len(f)-len(old)]+new {
				return true
			}
		}
		return false
	}
}

// letterDoubled reports whether to is from with exactly one letter doubled in
// place (modeled -> modelled).
func letterDoubled(from, to string) bool {
	if len(to) != len(from)+1 {
		return false
	}
	for i := range from {
		if from[:i+1]+from[i:] == to {
			return true
		}
	}
	return false
}

// letterInserted reports whether to is from with one extra letter inserted
// that is not a doubling (edema -> oedema, aging -> ageing).
func letterInserted(from, to string) bool {
	if len(to) != len(from)+1 {
		return false
	}
	for i := 0; i <= len(from); i++ {
		if from[:i]+string(to[i])+from[i:] == to {
			return true
		}
	}
	return false
}

func groupAdded(added []mapping) map[string][]mapping {
	groups := make(map[string][]mapping)
	for _, m := range added {
		g := classify(m.From, m.To)
		groups[g] = append(groups[g], m)
	}
	for _, g := range groups {
		sort.Slice(g, func(i, j int) bool { return g[i].From < g[j].From })
	}
	return groups
}

func sortedGroupNames(groups map[string][]mapping) []string {
	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	// "other" last: it is the bucket needing word-by-word review.
	sort.Slice(names, func(i, j int) bool {
		if (names[i] == "other") != (names[j] == "other") {
			return names[j] == "other"
		}
		return names[i] < names[j]
	})
	return names
}

func summarise(before, after int, added []mapping, drops stats, conflicts []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "dictionary: %d -> %d entries (+%d)\n", before, after, len(added))
	fmt.Fprintf(&b, "dropped:\n")
	for _, reason := range sortedKeys(drops) {
		fmt.Fprintf(&b, "  %5d  %s\n", drops[reason], reason)
	}
	fmt.Fprintf(&b, "conflicts kept as-is: %d\n", len(conflicts))
	groups := groupAdded(added)
	fmt.Fprintf(&b, "added by pattern:\n")
	for _, name := range sortedGroupNames(groups) {
		fmt.Fprintf(&b, "  %5d  %s\n", len(groups[name]), name)
	}
	return b.String()
}

func buildReport(before int, added []mapping, drops stats, conflicts []string) string {
	var b strings.Builder
	b.WriteString(summarise(before, before+len(added), added, drops, conflicts))
	if len(conflicts) > 0 {
		b.WriteString("\nconflicts (existing m2e entry kept):\n")
		for _, c := range conflicts {
			fmt.Fprintf(&b, "  %s\n", c)
		}
	}
	groups := groupAdded(added)
	for _, name := range sortedGroupNames(groups) {
		fmt.Fprintf(&b, "\n== %s (%d) ==\n", name, len(groups[name]))
		for _, m := range groups[name] {
			fmt.Fprintf(&b, "  %s -> %s\n", m.From, m.To)
		}
	}
	return b.String()
}

func sortedKeys(m stats) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func writeDict(path string, dict map[string]string) error {
	keys := make([]string, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("{\n")
	for i, k := range keys {
		sep := ","
		if i == len(keys)-1 {
			sep = ""
		}
		fmt.Fprintf(&b, "  %q: %q%s\n", k, dict[k], sep)
	}
	b.WriteString("}\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
