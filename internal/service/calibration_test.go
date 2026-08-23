package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/store"
)

// TestCalibrationAlignsEvidenceIntoOneEvent guards the behaviour described in
// the field report: when a receiver has a clock calibration configured, scan
// fragments that are close in time must still be grouped into a single event
// whose time window and association result reflect the calibrated time.
//
// Two stations report the same physical instant with clocks that disagree by
// a few hundred milliseconds. After calibration both CorrectedAt values fall on
// the same moment, so the second fragment must extend the first event rather
// than spawn a new one.
func TestCalibrationAlignsEvidenceIntoOneEvent(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "calib.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	clock := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)
	svc := New(db)
	svc.SetClock(func() time.Time { return clock })

	// Station "a" reports 250ms late (clock slow), station "b" reports 300ms
	// early (clock fast). Each calibration cancels the station clock error so
	// both observations correct to exactly `clock`: corrected = observed - offset.
	for _, st := range []struct {
		id     string
		offset int64
	}{
		// Each station clock runs fast by `offset`, so observing `clock+offset`
		// corrects back to `clock` (corrected = observed - offset).
		{"a", 250},
		{"b", 300},
	} {
		if _, err := svc.RegisterStation(ctx, model.RegisterStationRequest{ID: st.id, Name: st.id}); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.CreateCalibration(ctx, st.id, model.CreateCalibrationRequest{
			ClockOffsetMillis: st.offset,
			DirectionErrorDeg: 3,
			TrustedUntil:      clock.Add(time.Hour),
			Activate:          true,
		}); err != nil {
			t.Fatal(err)
		}
	}

	out, err := svc.IngestBatch(ctx, model.BatchFragmentsRequest{Fragments: []model.FragmentInput{
		{StationID: "a", Sequence: "1", ObservedAt: clock.Add(250 * time.Millisecond), CenterHz: 433920000, BandwidthHz: 12000, StrengthDBm: -54, DirectionDeg: 20},
		{StationID: "b", Sequence: "1", ObservedAt: clock.Add(300 * time.Millisecond), CenterHz: 433920000, BandwidthHz: 10000, StrengthDBm: -56, DirectionDeg: 24},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if out[0].EventID != out[1].EventID {
		t.Fatalf("calibration should keep close evidence in one event, got %v", out)
	}

	detail, err := svc.Event(ctx, out[0].EventID)
	if err != nil {
		t.Fatal(err)
	}
	if !detail.Event.StartAt.Equal(clock) || !detail.Event.EndAt.Equal(clock) {
		t.Fatalf("event window should span the calibrated instant, got %s..%s", detail.Event.StartAt, detail.Event.EndAt)
	}
	for _, f := range detail.Fragments {
		if !f.CorrectedAt.Equal(clock) {
			t.Fatalf("fragment %s corrected time %s should equal calibrated instant %s", f.ID, f.CorrectedAt, clock)
		}
	}
}
