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
