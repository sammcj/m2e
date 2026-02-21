package tests

import (
	"strings"
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

// Baseline text samples for benchmarking the full conversion pipeline.

const smallText = `The color of the center of the organization was gray. She traveled to the theater to analyze the behavior of the neighboring civilization.`

const mediumText = `The color of the center of the organization was gray. She traveled to the theater to analyze the behavior of the neighboring civilization.
He realized that the labor union had authorized a new program to standardize the utilization of defense resources.
The catalog of jewelry was organized by the modeling agency, which specialized in the customization of accessories.
Meanwhile, the counselor analyzed the favorable rumors about the harbor's modernization and the cancellation of unauthorized maneuvers.
The neighbor apologized for the offense and acknowledged that his behavior toward the organization was not favorable.
She emphasized that the theater's catalog had been digitized and reorganized to maximize utilization of the new program.
The traveler recognized the humor in the situation and signaled to the counselor that the favorable outcome had materialized.
His defense of the organization's labor practices was characterized by a rigorous analysis of the authorized memorandum.
The center of the harbor was being modernized, and the neighboring civilization had recognized the favorable rumors.
The modeling agency specialized in customization and had organized a catalog that emphasized both color and humor.
`

func makeLargeText(repetitions int) string {
	return strings.Repeat(mediumText, repetitions)
}

// BenchmarkConvertToBritish_Small benchmarks conversion of a short sentence.
func BenchmarkConvertToBritish_Small(b *testing.B) {
	conv, err := converter.NewConverter()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv.ConvertToBritish(smallText, false)
	}
}

// BenchmarkConvertToBritish_Medium benchmarks conversion of a paragraph.
func BenchmarkConvertToBritish_Medium(b *testing.B) {
	conv, err := converter.NewConverter()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv.ConvertToBritish(mediumText, false)
	}
}

// BenchmarkConvertToBritish_Large benchmarks conversion of a large document (~100 paragraphs).
func BenchmarkConvertToBritish_Large(b *testing.B) {
	conv, err := converter.NewConverter()
	if err != nil {
		b.Fatal(err)
	}
	largeText := makeLargeText(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv.ConvertToBritish(largeText, false)
	}
}

// BenchmarkConvertToBritish_VeryLarge benchmarks conversion of a very large document (~1000 paragraphs).
func BenchmarkConvertToBritish_VeryLarge(b *testing.B) {
	conv, err := converter.NewConverter()
	if err != nil {
		b.Fatal(err)
	}
	veryLargeText := makeLargeText(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv.ConvertToBritish(veryLargeText, false)
	}
}

// BenchmarkConvertToBritishSimple_Medium benchmarks the simple conversion path (no code-awareness).
func BenchmarkConvertToBritishSimple_Medium(b *testing.B) {
	conv, err := converter.NewConverter()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv.ConvertToBritishSimple(mediumText, false)
	}
}

// BenchmarkConvertToBritishSimple_Large benchmarks the simple conversion path with large text.
func BenchmarkConvertToBritishSimple_Large(b *testing.B) {
	conv, err := converter.NewConverter()
	if err != nil {
		b.Fatal(err)
	}
	largeText := makeLargeText(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv.ConvertToBritishSimple(largeText, false)
	}
}

// BenchmarkConvertWithSmartQuotes benchmarks conversion with smart quote normalisation enabled.
func BenchmarkConvertWithSmartQuotes(b *testing.B) {
	conv, err := converter.NewConverter()
	if err != nil {
		b.Fatal(err)
	}
	// Text with smart quotes mixed in
	text := strings.ReplaceAll(mediumText, "\"", "\u201C") // Replace with smart quotes
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv.ConvertToBritish(text, true)
	}
}

// BenchmarkConvertNoChanges benchmarks text that has no American spellings (worst case for lookups).
func BenchmarkConvertNoChanges(b *testing.B) {
	conv, err := converter.NewConverter()
	if err != nil {
		b.Fatal(err)
	}
	// British English text - nothing to convert, exercises all fallback paths
	britishText := strings.Repeat("The colour of the centre of the organisation was grey. She travelled to the theatre to analyse the behaviour of the neighbouring civilisation.\n", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv.ConvertToBritish(britishText, false)
	}
}
