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

func TestOverlapsFrequencySharesWindowRule(t *testing.T) {
	narrow := model.Fragment{CenterHz: 1_000_000, BandwidthHz: 2_000}
	wide := model.Fragment{CenterHz: 1_000_000, BandwidthHz: 6_000}
	if !OverlapsFrequency(narrow, wide, 0) {
		t.Fatal("edge-overlapping scans of different bandwidth were split")
	}
	disjoint := model.Fragment{CenterHz: 2_000_000, BandwidthHz: 4_000}
	if OverlapsFrequency(narrow, disjoint, FrequencyTolerance(narrow, disjoint)) {
		t.Fatal("disjoint band was merged")
	}
}

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
