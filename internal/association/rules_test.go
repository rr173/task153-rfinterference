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
	if !DirectionCompatible([]model.Fragment{{DirectionDeg: 38}}, f) {
		t.Fatal("expected close directions to be compatible")
	}
	if DirectionCompatible([]model.Fragment{{DirectionDeg: 200}}, f) {
		t.Fatal("unexpected direction compatibility")
	}
}

// TestDirectionCompatibleGroupsColocatedEvidence guards against the field
// report where receivers sighting the same source from close arrival directions
// had their evidence split across separate events instead of consolidating.
func TestDirectionCompatibleGroupsColocatedEvidence(t *testing.T) {
	now := time.Now()
	event := model.Event{StartAt: now.Add(-time.Minute), EndAt: now, MinHz: 990000, MaxHz: 1010000}
	// Existing evidence for the same source from one receiver direction.
	existing := []model.Fragment{{DirectionDeg: 20}}
	// A second receiver sees the same band and time from a nearby direction;
	// this must stay compatible so it consolidates into the active event.
	for _, dir := range []float64{28, 31, 45, 64} {
		f := model.Fragment{CorrectedAt: now, CenterHz: 1000000, BandwidthHz: 2000, DirectionDeg: dir}
		if !FrequencyCompatible(event, f) || !TimeCompatible(event, f) {
			t.Fatalf("expected frequency/time compatibility for direction %.0f", dir)
		}
		if !DirectionCompatible(existing, f) {
			t.Fatalf("direction %.0f should be compatible with existing 20 degree evidence", dir)
		}
	}
}
