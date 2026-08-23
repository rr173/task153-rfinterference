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

func TestPrepareRecordsCalibrationExpiryHint(t *testing.T) {
	db, _ := store.Open(filepath.Join(t.TempDir(), "x.db"))
	defer db.Close()
	st := station.New(db)
	ctx := context.Background()
	_, _ = st.Register(ctx, model.RegisterStationRequest{ID: "s", Name: "s"})

	// Calibrated, then the trust window has already elapsed.
	expired := time.Now().Add(-time.Hour)
	if _, err := st.CreateCalibration(ctx, "s", model.CreateCalibrationRequest{TrustedUntil: expired, Activate: true}); err != nil {
		t.Fatal(err)
	}
	late := time.Now()
	svc := New(db, st)

	// Evidence observed after the calibration trust window must carry the expiry hint.
	out, _, err := svc.Prepare(ctx, model.FragmentInput{StationID: "s", Sequence: "late", ObservedAt: late, CenterHz: 1000000, BandwidthHz: 1000, StrengthDBm: -60, DirectionDeg: 20})
	if err != nil {
		t.Fatal(err)
	}
	if out.ExclusionReason != "calibration trust window elapsed" {
		t.Fatalf("expected expiry hint, got %q", out.ExclusionReason)
	}

	// A fresh calibration whose window still covers the observation must not be flagged.
	if _, err := st.CreateCalibration(ctx, "s", model.CreateCalibrationRequest{TrustedUntil: time.Now().Add(time.Hour), Activate: true}); err != nil {
		t.Fatal(err)
	}
	fresh, _, err := svc.Prepare(ctx, model.FragmentInput{StationID: "s", Sequence: "fresh", ObservedAt: time.Now(), CenterHz: 1000000, BandwidthHz: 1000, StrengthDBm: -60, DirectionDeg: 20})
	if err != nil {
		t.Fatal(err)
	}
	if fresh.ExclusionReason != "" {
		t.Fatalf("expected no expiry hint within trust window, got %q", fresh.ExclusionReason)
	}
}
