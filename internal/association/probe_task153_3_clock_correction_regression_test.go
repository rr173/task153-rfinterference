package association

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"task153-rfinterference/internal/fragment"
	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/station"
	"task153-rfinterference/internal/store"
)

type fixedCorrector struct {
	corrected time.Time
	calibration model.Calibration
}

func (f fixedCorrector) CorrectTime(context.Context, string, time.Time) (time.Time, model.Calibration, error) {
	return f.corrected, f.calibration, nil
}

func TestBug03_CalibratedTimeIsUsedThroughoutAssociation(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "clock.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC()
	stations := station.New(db)
	stations.SetClock(func() time.Time { return now })
	if _, err := stations.Register(context.Background(), model.RegisterStationRequest{ID: "north", Name: "North"}); err != nil {
		t.Fatal(err)
	}
	if _, err := stations.CreateCalibration(context.Background(), "north", model.CreateCalibrationRequest{ClockOffsetMillis: 1200, TrustedUntil: now.Add(time.Hour), Activate: true}); err != nil {
		t.Fatal(err)
	}
	corrected, _, err := stations.CorrectTime(context.Background(), "north", now)
	if err != nil || !corrected.Equal(now.Add(-1200*time.Millisecond)) {
		t.Fatalf("station correction = %s, err=%v", corrected, err)
	}
	fragmentService := fragment.New(db, fixedCorrector{corrected: corrected, calibration: model.Calibration{TrustedUntil: now.Add(time.Hour)}})
	fragmentService.SetClock(func() time.Time { return now })
	prepared, _, err := fragmentService.Prepare(context.Background(), model.FragmentInput{StationID: "north", Sequence: "s1", ObservedAt: now, CenterHz: 1000000, BandwidthHz: 1000, StrengthDBm: -50, DirectionDeg: 20})
	if err != nil || !prepared.CorrectedAt.Equal(corrected) {
		t.Fatalf("prepared correction = %s, err=%v", prepared.CorrectedAt, err)
	}
	if !TimeCompatible(model.Event{StartAt: corrected, EndAt: corrected}, model.Fragment{ObservedAt: corrected.Add(4 * time.Minute), CorrectedAt: corrected}) {
		t.Fatal("association ignored the calibrated event time")
	}
}
