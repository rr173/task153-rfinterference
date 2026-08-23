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

func TestComputerFlagsExpiredCalibrationAsUntrusted(t *testing.T) {
	now := time.Now()
	c := New()
	event := model.Event{ID: "e", Revision: 1}
	fs := []model.Fragment{{StationID: "a", CorrectedAt: now, DirectionDeg: 20}, {StationID: "b", CorrectedAt: now, DirectionDeg: 28}, {StationID: "c", CorrectedAt: now, DirectionDeg: 35}}
	// Every calibration's trust window ended before the evidence was observed, so the
	// observations must not be treated as trusted localization evidence.
	cal := map[string]model.Calibration{"a": {TrustedUntil: now.Add(-time.Hour)}, "b": {TrustedUntil: now.Add(-time.Hour)}, "c": {TrustedUntil: now.Add(-time.Hour)}}
	a := c.Compute(event, fs, cal)
	if a.Verdict != model.VerdictTimeUntrusted {
		t.Fatalf("expected time_untrusted for expired calibration, got %+v", a)
	}
	if a.Confidence > 0.3 {
		t.Fatalf("confidence should be low for expired calibration, got %v", a.Confidence)
	}
}
