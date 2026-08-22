package station

import (
	"time"

	"task153-rfinterference/internal/model"
)

type TrustAssessment struct {
	Trusted      bool          `json:"trusted"`
	Age          time.Duration `json:"age"`
	ClockPenalty float64       `json:"clock_penalty"`
	Reason       string        `json:"reason,omitempty"`
}

func AssessTrust(calibration model.Calibration, observedAt, now time.Time) TrustAssessment {
	assessment := TrustAssessment{Trusted: true}
	if calibration.TrustedUntil.IsZero() {
		assessment.Trusted = false
		assessment.Reason = "station has no active calibration"
		return assessment
	}
	if observedAt.After(calibration.TrustedUntil) {
		assessment.Trusted = false
		assessment.Reason = "observation is newer than calibration trust window"
		return assessment
	}
	assessment.Age = now.Sub(calibration.CreatedAt)
	offset := calibration.ClockOffsetMillis
	if offset < 0 {
		offset = -offset
	}
	assessment.ClockPenalty = float64(offset) / 120000
	if assessment.ClockPenalty > 1 {
		assessment.ClockPenalty = 1
	}
	return assessment
}

func EffectiveDirectionMargin(calibration model.Calibration) float64 {
	margin := calibration.DirectionErrorDeg + float64(abs64(calibration.ClockOffsetMillis))/1000*0.2
	if margin < 3 {
		return 3
	}
	if margin > 45 {
		return 45
	}
	return margin
}

func abs64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
