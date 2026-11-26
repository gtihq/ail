// Package stats provides paired statistical analysis for benchmark results.
// It uses only the Go standard library so the experiment stays reproducible.
package stats

import (
	"math"
	"sort"
)

// PairedSample is one matched pair of measurements, A and B.
type PairedSample struct {
	ID string
	A  float64
	B  float64
}

// PairedResult holds descriptive and inferential statistics for paired samples.
type PairedResult struct {
	N              int
	MeanA          float64
	MeanB          float64
	MeanDiff       float64
	MedianA        float64
	MedianB        float64
	MedianDiff     float64
	SDA            float64
	SDB            float64
	SDDiff         float64
	AbsoluteDiff   float64
	RelativeDiff   float64
	CILow          float64
	CIHigh         float64
	TStatistic     float64
	PApprox        float64
	CohensD        float64
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64{}, values...)
	sort.Float64s(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func standardDeviation(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	average := mean(values)
	sum := 0.0
	for _, value := range values {
		sum += (value - average) * (value - average)
	}
	return math.Sqrt(sum / float64(len(values)-1))
}

var tCriticalTable = []struct {
	df    float64
	value float64
}{
	{1, 12.706}, {2, 4.303}, {3, 3.182}, {4, 2.776}, {5, 2.571},
	{10, 2.228}, {15, 2.131}, {20, 2.086}, {25, 2.060}, {30, 2.042},
	{40, 2.021}, {50, 2.009}, {60, 2.000}, {120, 1.980}, {1000000, 1.960},
}

// tCritical returns the two-sided 95 percent critical value for a degrees-of-freedom count.
func tCritical(df float64) float64 {
	for _, entry := range tCriticalTable {
		if df <= entry.df {
			return entry.value
		}
	}
	return 1.960
}

// Paired computes paired statistics. B minus A is the difference direction.
func Paired(samples []PairedSample) PairedResult {
	result := PairedResult{N: len(samples)}
	if len(samples) == 0 {
		return result
	}
	valuesA := make([]float64, 0, len(samples))
	valuesB := make([]float64, 0, len(samples))
	diffs := make([]float64, 0, len(samples))
	for _, sample := range samples {
		valuesA = append(valuesA, sample.A)
		valuesB = append(valuesB, sample.B)
		diffs = append(diffs, sample.B-sample.A)
	}
	result.MeanA = mean(valuesA)
	result.MeanB = mean(valuesB)
	result.MeanDiff = mean(diffs)
	result.MedianA = median(valuesA)
	result.MedianB = median(valuesB)
	result.MedianDiff = median(diffs)
	result.SDA = standardDeviation(valuesA)
	result.SDB = standardDeviation(valuesB)
	result.SDDiff = standardDeviation(diffs)
	result.AbsoluteDiff = result.MeanB - result.MeanA
	if result.MeanA != 0 {
		result.RelativeDiff = result.AbsoluteDiff / math.Abs(result.MeanA) * 100
	}

	critical := tCritical(float64(len(samples) - 1))
	standardError := 0.0
	if len(samples) > 0 {
		standardError = result.SDDiff / math.Sqrt(float64(len(samples)))
	}
	margin := critical * standardError
	result.CILow = result.MeanDiff - margin
	result.CIHigh = result.MeanDiff + margin

	if standardError != 0 {
		result.TStatistic = result.MeanDiff / standardError
	}
	result.PApprox = math.Erfc(math.Abs(result.TStatistic) / math.Sqrt(2))
	if result.SDDiff != 0 {
		result.CohensD = result.MeanDiff / result.SDDiff
	}
	return result
}

// McNemarResult holds the result of McNemar's test for paired binary outcomes.
type McNemarResult struct {
	BothPositive  int
	BOnly         int
	AOnly         int
	BothNegative  int
	ChiSquare     float64
	PApprox       float64
}

// McNemar performs McNemar's test with a continuity correction.
func McNemar(bothPositive, bOnly, aOnly, bothNegative int) McNemarResult {
	result := McNemarResult{
		BothPositive: bothPositive,
		BOnly:        bOnly,
		AOnly:        aOnly,
		BothNegative: bothNegative,
	}
	discordant := float64(bOnly + aOnly)
	if discordant > 0 {
		numerator := math.Abs(float64(bOnly)-float64(aOnly)) - 1
		if numerator < 0 {
			numerator = 0
		}
		result.ChiSquare = numerator * numerator / discordant
		result.PApprox = math.Erfc(math.Sqrt(result.ChiSquare / 2))
	}
	return result
}

// Holm applies the Holm-Bonferroni correction to a slice of p-values.
func Holm(pValues []float64) []float64 {
	indexed := make([]struct {
		index int
		value float64
	}, len(pValues))
	for index, value := range pValues {
		indexed[index] = struct {
			index int
			value float64
		}{index: index, value: value}
	}
	sort.Slice(indexed, func(i, j int) bool { return indexed[i].value < indexed[j].value })
	adjusted := make([]float64, len(pValues))
	running := 0.0
	for rank := len(indexed) - 1; rank >= 0; rank-- {
		multiplier := float64(len(indexed) - rank)
		candidate := indexed[rank].value * multiplier
		if candidate > 1 {
			candidate = 1
		}
		if candidate < running {
			candidate = running
		}
		running = candidate
		adjusted[indexed[rank].index] = candidate
	}
	return adjusted
}

// Significant reports whether an approximate p-value is below alpha.
func Significant(pValue, alpha float64) bool {
	return pValue < alpha
}
