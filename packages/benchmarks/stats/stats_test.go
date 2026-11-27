package stats_test

import (
	"math"
	"testing"

	"github.com/gtihq/ail/packages/benchmarks/stats"
)

func closeEnough(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func TestPairedKnownValues(t *testing.T) {
	samples := []stats.PairedSample{
		{ID: "1", A: 1, B: 2},
		{ID: "2", A: 2, B: 4},
		{ID: "3", A: 3, B: 6},
		{ID: "4", A: 4, B: 8},
	}
	result := stats.Paired(samples)
	if result.N != 4 {
		t.Fatalf("N = %d", result.N)
	}
	if !closeEnough(result.MeanA, 2.5, 1e-9) || !closeEnough(result.MeanB, 5, 1e-9) {
		t.Fatalf("means = %f / %f", result.MeanA, result.MeanB)
	}
	if !closeEnough(result.MeanDiff, 2.5, 1e-9) {
		t.Fatalf("mean diff = %f", result.MeanDiff)
	}
	if !closeEnough(result.MedianDiff, 2.5, 1e-9) {
		t.Fatalf("median diff = %f", result.MedianDiff)
	}
	if !closeEnough(result.RelativeDiff, 100, 1e-9) {
		t.Fatalf("relative diff = %f", result.RelativeDiff)
	}
	if result.CILow >= result.MeanDiff || result.CIHigh <= result.MeanDiff {
		t.Fatalf("CI does not contain mean: %f to %f", result.CILow, result.CIHigh)
	}
}

func TestPairedIdenticalSamples(t *testing.T) {
	samples := []stats.PairedSample{
		{ID: "1", A: 10, B: 10},
		{ID: "2", A: 20, B: 20},
		{ID: "3", A: 30, B: 30},
	}
	result := stats.Paired(samples)
	if result.MeanDiff != 0 || result.SDDiff != 0 || result.TStatistic != 0 {
		t.Fatalf("identical samples should have zero effect: %+v", result)
	}
	if !closeEnough(result.PApprox, 1, 1e-9) {
		t.Fatalf("p = %f, want 1", result.PApprox)
	}
}

func TestPairedEmpty(t *testing.T) {
	result := stats.Paired(nil)
	if result.N != 0 || result.MeanDiff != 0 {
		t.Fatalf("empty result = %+v", result)
	}
}

func TestMcNemar(t *testing.T) {
	result := stats.McNemar(5, 10, 0, 5)
	expected := 81.0 / 10.0
	if !closeEnough(result.ChiSquare, expected, 1e-9) {
		t.Fatalf("chi-square = %f, want %f", result.ChiSquare, expected)
	}
	if result.PApprox >= 0.05 {
		t.Fatalf("p = %f, want significant", result.PApprox)
	}
	identical := stats.McNemar(5, 3, 3, 5)
	if identical.ChiSquare != 0 || !closeEnough(identical.PApprox, 1, 1e-9) {
		t.Fatalf("symmetric discordance = %+v", identical)
	}
}

func TestHolmCorrection(t *testing.T) {
	adjusted := stats.Holm([]float64{0.01, 0.02, 0.03, 0.04})
	if len(adjusted) != 4 {
		t.Fatalf("length = %d", len(adjusted))
	}
	if !closeEnough(adjusted[0], 0.06, 1e-9) || !closeEnough(adjusted[1], 0.06, 1e-9) ||
		!closeEnough(adjusted[2], 0.06, 1e-9) || !closeEnough(adjusted[3], 0.04, 1e-9) {
		t.Fatalf("unexpected adjustment: %+v", adjusted)
	}
}

func TestSignificance(t *testing.T) {
	if !stats.Significant(0.01, 0.05) {
		t.Fatal("0.01 should be significant at 0.05")
	}
	if stats.Significant(0.10, 0.05) {
		t.Fatal("0.10 should not be significant at 0.05")
	}
}
