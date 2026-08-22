package model

import (
	"math"
	"sort"
)

type FrequencySpan struct {
	MinHz int64 `json:"min_hz"`
	MaxHz int64 `json:"max_hz"`
}

func NewFrequencySpan(centerHz, bandwidthHz int64) FrequencySpan {
	min, max := FrequencyRange(centerHz, bandwidthHz)
	return FrequencySpan{MinHz: min, MaxHz: max}
}

func (s FrequencySpan) CenterHz() int64 {
	return (s.MinHz + s.MaxHz) / 2
}

func (s FrequencySpan) WidthHz() int64 {
	return s.MaxHz - s.MinHz
}

func (s FrequencySpan) Overlaps(other FrequencySpan, toleranceHz int64) bool {
	return s.MinHz <= other.MaxHz+toleranceHz && other.MinHz <= s.MaxHz+toleranceHz
}

func MergeFrequencySpans(spans []FrequencySpan) FrequencySpan {
	if len(spans) == 0 {
		return FrequencySpan{}
	}
	merged := spans[0]
	for _, span := range spans[1:] {
		if span.MinHz < merged.MinHz {
			merged.MinHz = span.MinHz
		}
		if span.MaxHz > merged.MaxHz {
			merged.MaxHz = span.MaxHz
		}
	}
	return merged
}

func MedianStrength(fragments []Fragment) float64 {
	if len(fragments) == 0 {
		return 0
	}
	values := make([]float64, 0, len(fragments))
	for _, fragment := range fragments {
		values = append(values, fragment.StrengthDBm)
	}
	sort.Float64s(values)
	middle := len(values) / 2
	if len(values)%2 == 1 {
		return values[middle]
	}
	return (values[middle-1] + values[middle]) / 2
}

func DirectionDeviation(a, b float64) float64 {
	delta := math.Abs(NormalizeDirection(a) - NormalizeDirection(b))
	if delta > 180 {
		return 360 - delta
	}
	return delta
}
