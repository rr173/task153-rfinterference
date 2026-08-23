package model

import (
	"math"
	"testing"
)

func TestDirectionDeviationWrapsAcrossNorth(t *testing.T) {
	cases := []struct {
		name          string
		a, b, wantDeg float64
	}{
		{"straddles north", 359, 1, 2},
		{"straddles north reversed", 1, 359, 2},
		{"wide north arc", 350, 10, 20},
		{"southern close", 180, 170, 10},
		{"identical", 42, 42, 0},
		{"antipodal", 0, 180, 180},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DirectionDeviation(tc.a, tc.b)
			if math.Abs(got-tc.wantDeg) > 1e-9 {
				t.Fatalf("DirectionDeviation(%g,%g) = %g, want %g", tc.a, tc.b, got, tc.wantDeg)
			}
		})
	}
}
