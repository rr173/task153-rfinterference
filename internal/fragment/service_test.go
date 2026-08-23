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

// TestPrepareRetransmitIsIdempotent ensures a receiver that retransmits a scan
// fragment using a different casing or surrounding whitespace resolves to the
// same stored evidence instead of creating a new record.
func TestPrepareRetransmitIsIdempotent(t *testing.T) {
	db, _ := store.Open(filepath.Join(t.TempDir(), "x.db"))
	defer db.Close()
	st := station.New(db)
	ctx := context.Background()
	if _, err := st.Register(ctx, model.RegisterStationRequest{ID: "north", Name: "North"}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateCalibration(ctx, "north", model.CreateCalibrationRequest{TrustedUntil: time.Now().Add(time.Hour), Activate: true}); err != nil {
		t.Fatal(err)
	}
	svc := New(db, st)
	original := model.FragmentInput{StationID: "north", Sequence: "seq-1", ObservedAt: time.Now(), CenterHz: 1000000, BandwidthHz: 1000, StrengthDBm: -60, DirectionDeg: 20}
	f, dup, err := svc.Prepare(ctx, original)
	if err != nil || dup {
		t.Fatalf("prepare original: dup=%v err=%v", dup, err)
	}
	if err := svc.Persist(ctx, f); err != nil {
		t.Fatal(err)
	}

	for _, variant := range []model.FragmentInput{
		{StationID: "North", Sequence: "seq-1", ObservedAt: original.ObservedAt, CenterHz: original.CenterHz, BandwidthHz: original.BandwidthHz, StrengthDBm: original.StrengthDBm, DirectionDeg: original.DirectionDeg},
		{StationID: " NORTH ", Sequence: "  seq-1  ", ObservedAt: original.ObservedAt, CenterHz: original.CenterHz, BandwidthHz: original.BandwidthHz, StrengthDBm: original.StrengthDBm, DirectionDeg: original.DirectionDeg},
		{StationID: "north", Sequence: "SEQ-1", ObservedAt: original.ObservedAt, CenterHz: original.CenterHz, BandwidthHz: original.BandwidthHz, StrengthDBm: original.StrengthDBm, DirectionDeg: original.DirectionDeg},
	} {
		got, dup, err := svc.Prepare(ctx, variant)
		if err != nil {
			t.Fatalf("retransmit %+v: err=%v", variant.StationID, err)
		}
		if !dup {
			t.Fatalf("retransmit %+v was not detected as duplicate", variant.StationID)
		}
		if got.ID != f.ID {
			t.Fatalf("retransmit %+v returned fragment %q, want original %q", variant.StationID, got.ID, f.ID)
		}
	}

	count, err := db.CountFragments(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected a single stored fragment, got %d", count)
	}
}
