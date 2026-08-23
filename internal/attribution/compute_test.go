package attribution

import (
	"task153-rfinterference/internal/model"
	"testing"
	"time"
)

func TestComputerReportsLocatableEvidence(t *testing.T) {
	now := time.Now()
	c := New()
	event := model.Event{ID: "e", Revision: 1}
	fs := []model.Fragment{{StationID: "a", ObservedAt: now, DirectionDeg: 20}, {StationID: "b", ObservedAt: now, DirectionDeg: 28}, {StationID: "c", ObservedAt: now, DirectionDeg: 35}}
	cal := map[string]model.Calibration{"a": {TrustedUntil: now.Add(time.Hour)}, "b": {TrustedUntil: now.Add(time.Hour)}, "c": {TrustedUntil: now.Add(time.Hour)}}
	a := c.Compute(event, fs, cal)
	if a.Verdict != model.VerdictLocatable {
		t.Fatalf("%+v", a)
	}
}

func TestComputerLocatesAcrossNorth(t *testing.T) {
	now := time.Now()
	c := New()
	event := model.Event{ID: "e", Revision: 1}
	fs := []model.Fragment{{StationID: "a", ObservedAt: now, DirectionDeg: 359}, {StationID: "b", ObservedAt: now, DirectionDeg: 1}}
	cal := map[string]model.Calibration{"a": {TrustedUntil: now.Add(time.Hour)}, "b": {TrustedUntil: now.Add(time.Hour)}}
	a := c.Compute(event, fs, cal)
	if a.Verdict != model.VerdictLocatable {
		t.Fatalf("expected locatable across north, got %+v", a)
	}
	if a.DirectionMin != 359 || a.DirectionMax != 1 {
		t.Fatalf("direction range should straddle north as 359..1, got %.1f..%.1f", a.DirectionMin, a.DirectionMax)
	}
}
