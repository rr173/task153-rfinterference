package attribution

import (
	"testing"
	"time"

	"task153-rfinterference/internal/association"
	"task153-rfinterference/internal/model"
)

func TestBug09_DirectionEvidenceWrapsAcrossNorth(t *testing.T) {
	if got := model.DirectionDeviation(359, 1); got != 2 {
		t.Fatalf("north-boundary deviation=%v, want 2", got)
	}
	if !association.DirectionCompatible([]model.Fragment{{DirectionDeg: 359}}, model.Fragment{DirectionDeg: 1}) {
		t.Fatal("near-north direction evidence was not considered compatible")
	}
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	computer := New()
	computer.SetClock(func() time.Time { return now })
	fragments := []model.Fragment{{StationID: "a", ObservedAt: now, DirectionDeg: 359}, {StationID: "b", ObservedAt: now, DirectionDeg: 1}, {StationID: "c", ObservedAt: now, DirectionDeg: 4}}
	calibrations := map[string]model.Calibration{"a": {TrustedUntil: now.Add(time.Hour)}, "b": {TrustedUntil: now.Add(time.Hour)}, "c": {TrustedUntil: now.Add(time.Hour)}}
	snapshot := computer.Compute(model.Event{ID: "north", Revision: 1}, fragments, calibrations)
	if snapshot.Verdict != model.VerdictLocatable || snapshot.DirectionMin != 359 || snapshot.DirectionMax != 4 {
		t.Fatalf("north-wrapping evidence produced %+v", snapshot)
	}
}
