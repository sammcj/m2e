package tests

import (
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

// TestImportedDictionaryEntries samples the entries imported from
// tmgldn/en-mappings (see scripts/import-en-mappings), one or more per
// transformation pattern group.
func TestImportedDictionaryEntries(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "ense to ence", input: "cyberdefense systems", expected: "cyberdefence systems"},
		{name: "er to re", input: "accoutered in finery", expected: "accoutred in finery"},
		{name: "og to ogue", input: "gene homologs", expected: "gene homologues"},
		{name: "or to our", input: "ardors and armors", expected: "ardours and armours"},
		{name: "consonant doubling", input: "barreled downhill", expected: "barrelled downhill"},
		{name: "e insertion", input: "the acknowledgments page", expected: "the acknowledgements page"},
		{name: "adrenaline gains final e", input: "pure adrenalin", expected: "pure adrenaline"},
		{name: "ambiance to ambience", input: "great ambiance", expected: "great ambience"},
		{name: "airfoil to aerofoil", input: "wing airfoil design", expected: "wing aerofoil design"},
		{name: "thru to through", input: "drive thru here", expected: "drive through here"},
		{name: "mom to mum", input: "my mom said", expected: "my mum said"},
		{name: "case preserved on imported words", input: "Ambiance matters", expected: "Ambience matters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := conv.ConvertToBritish(tt.input, false)
			if result != tt.expected {
				t.Errorf("ConvertToBritish(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestBlocklistedEntriesNotImported confirms that entries curated out of the
// import (vocabulary swaps, technical collisions, contested spellings - see
// scripts/import-en-mappings/blocklist.json) do not convert.
func TestBlocklistedEntriesNotImported(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	inputs := []string{
		"math homework",
		"npm lists dependents here",
		"curb your enthusiasm",
		"we had gotten there",
		"a network adapter",
		"the neuron fires",
		"centennial celebrations",
		"guerrilla tactics",
		"all for naught",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if result := conv.ConvertToBritish(input, false); result != input {
				t.Errorf("ConvertToBritish(%q) = %q, expected unchanged", input, result)
			}
		})
	}
}

// TestBritishEnglishUnchanged runs realistic British/international English
// prose through the converter and asserts nothing changes: valid British text
// must never be "converted".
func TestBritishEnglishUnchanged(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	paragraphs := []string{
		"The colour of the aluminium organiser was analysed at the centre of the programme.",
		"Reflection in Go uses the reflect package to analyse behaviour at runtime.",
		"The licensing agreement covers travelling employees and their favourite labelled organisers.",
		"She weighed fifty grams of flour, then a few kilograms of sugar, using a calibrated gauge.",
		"He was jailed after the burglary; the jailer siphoned fuel from the lathe by the kerb.",
		"Reaction stoichiometry requires precise measurement and rigorous defence of methodology.",
	}

	for _, p := range paragraphs {
		t.Run(p[:24], func(t *testing.T) {
			if result := conv.ConvertToBritish(p, false); result != p {
				t.Errorf("British text changed:\n  in:  %q\n  out: %q", p, result)
			}
		})
	}
}

// TestConversionIdempotent asserts converting already-converted text is a
// no-op. Combined with the hygiene test's no-key-is-a-value invariant this
// guards against double-conversion chains.
func TestConversionIdempotent(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	inputs := []string{
		"I love the color and flavor of this yogurt.",
		"The organization utilized specialized colorized analyzers.",
		"He apologized for the misdemeanor near the harbor's center.",
		"My mom drove thru the neighborhood with great ambiance.",
	}

	for _, input := range inputs {
		t.Run(input[:20], func(t *testing.T) {
			once := conv.ConvertToBritish(input, false)
			twice := conv.ConvertToBritish(once, false)
			if once != twice {
				t.Errorf("conversion not idempotent:\n  once:  %q\n  twice: %q", once, twice)
			}
		})
	}
}
