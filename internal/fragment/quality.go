package fragment

import (
	"math"

	"task153-rfinterference/internal/model"
)

type Quality struct {
	Usable bool
	Reason string
	Score  float64
}

// Assess rejects physically implausible scan values before they become event
// evidence and gives association a stable measure of evidence quality.
func Assess(input model.FragmentInput) Quality {
	if input.StrengthDBm < -180 || input.StrengthDBm > 30 {
		return Quality{Reason: "signal strength is outside receiver operating range"}
	}
	if input.BandwidthHz < 100 {
		return Quality{Reason: "bandwidth is too narrow to form an interference observation"}
	}
	if input.BandwidthHz > 100000000 {
		return Quality{Reason: "bandwidth exceeds configured receiver span"}
	}
	strength := math.Min(1, math.Max(0, (input.StrengthDBm+140)/110))
	bandwidth := math.Min(1, math.Log10(float64(input.BandwidthHz))/8)
	return Quality{Usable: true, Score: 0.55*strength + 0.45*bandwidth}
}

func Comparable(a, b model.Fragment) bool {
	if math.Abs(a.StrengthDBm-b.StrengthDBm) > 55 {
		return false
	}
	return OverlapsFrequency(a, b, FrequencyTolerance(a, b))
}

func FrequencyTolerance(a, b model.Fragment) int64 {
	base := (a.BandwidthHz + b.BandwidthHz) / 4
	if base < 10000 {
		return 10000
	}
	if base > 100000 {
		return 100000
	}
	return base
}
