package tests

import (
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

// TestDictionaryFixes covers dictionary entries that previously mapped to
// misspellings, wrong inflections, or archaic forms, plus capitalised keys
// that could never match because lookup lowercases the input word.
func TestDictionaryFixes(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "edema converts to oedema not edoema", input: "The patient has edema", expected: "The patient has oedema"},
		{name: "pummeled keeps past tense", input: "He pummeled the dough", expected: "He pummelled the dough"},
		{name: "yogurt converts to modern yoghurt", input: "I like yogurt", expected: "I like yoghurt"},
		{name: "yogurts converts to modern yoghurts", input: "Two yogurts", expected: "Two yoghurts"},
		{name: "colorize converts ize to ise", input: "colorize the image", expected: "colourise the image"},
		{name: "colorized converts ize to ise", input: "a colorized photo", expected: "a colourised photo"},
		{name: "colorizes converts ize to ise", input: "it colorizes film", expected: "it colourises film"},
		{name: "colorizing converts ize to ise", input: "colorizing old footage", expected: "colourising old footage"},
		{name: "diarization converts rather than self-mapping", input: "speaker diarization", expected: "speaker diarisation"},
		{name: "capitalised Americanization matches lowercase key", input: "The Americanization of culture", expected: "The Americanisation of culture"},
		{name: "capitalised Americanize matches lowercase key", input: "They Americanize everything", expected: "They Americanise everything"},
		{name: "capitalised Africanization matches lowercase key", input: "Africanization movements", expected: "Africanisation movements"},
		{name: "capitalised Finlandization matches lowercase key", input: "Finlandization policy", expected: "Finlandisation policy"},
		{name: "licensing is correct British English and stays unchanged", input: "The licensing agreement", expected: "The licensing agreement"},
		{name: "bussing is valid British English and stays unchanged", input: "bussing children to school", expected: "bussing children to school"},
		{name: "reflection is standard everywhere and stays unchanged", input: "uses reflection at runtime", expected: "uses reflection at runtime"},
		{name: "gram is modern international English and stays unchanged", input: "add 50 gram portions", expected: "add 50 gram portions"},
		{name: "kilograms is modern international English and stays unchanged", input: "weighs 5 kilograms", expected: "weighs 5 kilograms"},
		{name: "jailed is modern international English and stays unchanged", input: "he was jailed for fraud", expected: "he was jailed for fraud"},
		{name: "siphon is the standard scientific spelling and stays unchanged", input: "siphon the fuel", expected: "siphon the fuel"},
		{name: "lathe is the only spelling of the machine tool", input: "turned on a lathe", expected: "turned on a lathe"},
		{name: "ankle never becomes the obsolete ancle", input: "a sprained ankle", expected: "a sprained ankle"},
		{name: "stoichiometry is the standard chemistry spelling", input: "reaction stoichiometry", expected: "reaction stoichiometry"},
		{name: "binging still converts to bingeing", input: "binging on TV", expected: "bingeing on TV"},
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
