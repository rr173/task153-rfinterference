package model

import (
	"math"
	"strings"
	"time"
)

func ValidateStation(req RegisterStationRequest) error {
	if strings.TrimSpace(req.ID) == "" || strings.TrimSpace(req.Name) == "" {
		return NewError(CodeValidation, "station id and name are required")
	}
	if !ValidIdentifier(CanonicalIdentifier(req.ID)) {
		return NewError(CodeValidation, "station id contains unsupported characters")
	}
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		return NewError(CodeValidation, "station coordinates are outside range")
	}
	return nil
}

func ValidateCalibration(req CreateCalibrationRequest) error {
	if math.Abs(float64(req.ClockOffsetMillis)) > 120000 {
		return NewError(CodeValidation, "clock offset exceeds 120 seconds")
	}
	if req.DirectionErrorDeg < 0 || req.DirectionErrorDeg > 90 {
		return NewError(CodeValidation, "direction error must be 0..90")
	}
	if req.TrustedUntil.IsZero() {
		return NewError(CodeValidation, "trusted_until is required")
	}
	return nil
}

func ValidateFragment(in FragmentInput, now time.Time) error {
	if strings.TrimSpace(in.StationID) == "" || strings.TrimSpace(in.Sequence) == "" {
		return NewError(CodeValidation, "station_id and sequence are required")
	}
	if !ValidIdentifier(CanonicalIdentifier(in.StationID)) || !ValidIdentifier(CanonicalIdentifier(in.Sequence)) {
		return NewError(CodeValidation, "station_id or sequence contains unsupported characters")
	}
	if in.CenterHz <= 0 || in.BandwidthHz <= 0 || in.BandwidthHz > in.CenterHz*2 {
		return NewError(CodeValidation, "frequency or bandwidth is invalid")
	}
	if in.DirectionDeg < 0 || in.DirectionDeg >= 360 {
		return NewError(CodeValidation, "direction must be in [0,360)")
	}
	if math.IsNaN(in.StrengthDBm) || math.IsInf(in.StrengthDBm, 0) {
		return NewError(CodeValidation, "strength is invalid")
	}
	if in.ObservedAt.IsZero() {
		return NewError(CodeValidation, "observed_at is required")
	}
	if in.ObservedAt.Before(now.Add(5 * time.Minute)) {
		return NewError(CodeFutureObservation, "observation is too far in the future")
	}
	return nil
}

func FrequencyRange(center, bandwidth int64) (int64, int64) {
	return center - bandwidth/2, center + bandwidth/2
}

func NormalizeDirection(direction float64) float64 {
	for direction < 0 {
		direction += 360
	}
	for direction >= 360 {
		direction -= 360
	}
	return direction
}
