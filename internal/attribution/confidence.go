package attribution

import (
	"math"

	"task153-rfinterference/internal/model"
)

// ConfidenceFactors exposes the independently meaningful parts of a location
// confidence score so downstream applications can explain score changes.
type ConfidenceFactors struct {
	StationFactor   float64
	EvidenceFactor  float64
	DirectionFactor float64
	StrengthFactor  float64
}

func BuildConfidence(fragments []model.Fragment, stationCount int, directionSpan float64) ConfidenceFactors {
	factors := ConfidenceFactors{}
	factors.StationFactor = math.Min(1, float64(stationCount)/3)
	factors.EvidenceFactor = math.Min(1, float64(len(fragments))/6)
	factors.DirectionFactor = 1 - math.Min(1, directionSpan/90)
	median := model.MedianStrength(fragments)
	factors.StrengthFactor = math.Min(1, math.Max(0, (median+130)/90))
	return factors
}

func (f ConfidenceFactors) Score() float64 {
	return 0.12 + f.StationFactor*0.38 + f.EvidenceFactor*0.20 + f.DirectionFactor*0.20 + f.StrengthFactor*0.10
}

func (f ConfidenceFactors) Clamp() float64 {
	return math.Min(0.95, math.Max(0.05, f.Score()))
}
