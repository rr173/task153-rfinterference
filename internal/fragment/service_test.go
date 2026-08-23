package fragment

import (
	"context"
	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/station"
	"task153-rfinterference/internal/store"
	"path/filepath"
	"testing"
	"time"
)

func TestPrepareDetectsDuplicate(t *testing.T) {
	db, _ := store.Open(filepath.Join(t.TempDir(), "x.db"))
	defer db.Close()
	st := station.New(db)
	ctx := context.Background()
	_, _ = st.Register(ctx, model.RegisterStationRequest{ID: "s", Name: "s"})
	_, _ = st.CreateCalibration(ctx, "s", model.CreateCalibrationRequest{TrustedUntil: time.Now().Add(time.Hour), Activate: true})
	svc := New(db, st)
	in := model.FragmentInput{StationID: "s", Sequence: "1", ObservedAt: time.Now(), CenterHz: 1000000, BandwidthHz: 1000, StrengthDBm: -60, DirectionDeg: 20}
	f, dup, err := svc.Prepare(ctx, in)
	if err != nil || dup {
		t.Fatal(err)
	}
	if err := svc.Persist(ctx, f); err != nil {
		t.Fatal(err)
	}
	_, dup, err = svc.Prepare(ctx, in)
	if err != nil || !dup {
		t.Fatalf("duplicate=%v err=%v", dup, err)
	}
}

func TestPrepareRejectsFarFutureObservation(t *testing.T) {
	db, _ := store.Open(filepath.Join(t.TempDir(), "x.db"))
	defer db.Close()
	st := station.New(db)
	ctx := context.Background()
	_, _ = st.Register(ctx, model.RegisterStationRequest{ID: "s", Name: "s"})
	_, _ = st.CreateCalibration(ctx, "s", model.CreateCalibrationRequest{TrustedUntil: time.Now().Add(time.Hour), Activate: true})
	now := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)
	svc := New(db, st)
	svc.SetClock(func() time.Time { return now })
	in := model.FragmentInput{StationID: "s", Sequence: "1", ObservedAt: now.Add(model.FutureObservationWindow + time.Second), CenterHz: 1000000, BandwidthHz: 1000, StrengthDBm: -60, DirectionDeg: 20}
	_, dup, err := svc.Prepare(ctx, in)
	if err == nil || dup {
		t.Fatalf("expected future observation rejection, got err=%v dup=%v", err, dup)
	}
	if model.CodeOf(err) != model.CodeFutureObservation {
		t.Fatalf("expected CodeFutureObservation, got %v", err)
	}
}
