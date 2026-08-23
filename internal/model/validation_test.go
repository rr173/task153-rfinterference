package model

import (
	"testing"
	"time"
)

func validInput(now time.Time) FragmentInput {
	return FragmentInput{
		StationID:    "s",
		Sequence:     "1",
		ObservedAt:   now.Add(-time.Minute),
		CenterHz:     1000000,
		BandwidthHz:  1000,
		StrengthDBm:  -60,
		DirectionDeg: 20,
	}
}

func TestValidateFragmentRejectsFarFutureObservation(t *testing.T) {
	now := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)
	in := validInput(now)
	in.ObservedAt = now.Add(FutureObservationWindow + time.Second)
	if err := ValidateFragment(in, now); err == nil || CodeOf(err) != CodeFutureObservation {
		t.Fatalf("expected future observation error, got %v", err)
	}
}

func TestValidateFragmentAcceptsObservationWithinFutureWindow(t *testing.T) {
	now := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)
	for _, delta := range []time.Duration{
		-time.Hour,            // clearly in the past
		-1 * time.Second,      // just behind now
		0,                     // exactly now
		FutureObservationWindow, // at the edge of the allowed window
	} {
		in := validInput(now)
		in.ObservedAt = now.Add(delta)
		if err := ValidateFragment(in, now); err != nil {
			t.Fatalf("delta=%s: expected acceptance, got %v", delta, err)
		}
	}
}
