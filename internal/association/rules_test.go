package association

import (
	"task153-rfinterference/internal/model"
	"testing"
	"time"
)

func TestCompatibilityRules(t *testing.T) {
	now := time.Now()
	event := model.Event{StartAt: now.Add(-time.Minute), EndAt: now, MinHz: 990000, MaxHz: 1010000}
	f := model.Fragment{CorrectedAt: now, CenterHz: 1000000, BandwidthHz: 2000, DirectionDeg: 10}
	if !FrequencyCompatible(event, f) || !TimeCompatible(event, f) {
		t.Fatal("expected compatibility")
	}
	if DirectionCompatible([]model.Fragment{{DirectionDeg: 200}}, f) {
		t.Fatal("unexpected direction compatibility")
	}
	// Closely agreeing directions must associate even when they straddle north.
	if !DirectionCompatible([]model.Fragment{{DirectionDeg: 359}}, model.Fragment{DirectionDeg: 1}) {
		t.Fatal("north-straddling directions should remain compatible")
	}
	if !DirectionCompatible([]model.Fragment{{DirectionDeg: 20}}, model.Fragment{DirectionDeg: 27}) {
		t.Fatal("nearby directions should remain compatible")
	}
}
