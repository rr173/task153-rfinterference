package service

import (
	"context"
	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/store"
	"path/filepath"
	"testing"
	"time"
)

func TestIngestGroupsThreeStations(t *testing.T) {
	db, _ := store.Open(filepath.Join(t.TempDir(), "rf.db"))
	defer db.Close()
	svc := New(db)
	ctx := context.Background()
	now := time.Now().UTC()
	svc.SetClock(func() time.Time { return now })
	for _, id := range []string{"a", "b", "c"} {
		if _, err := svc.RegisterStation(ctx, model.RegisterStationRequest{ID: id, Name: id}); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.CreateCalibration(ctx, id, model.CreateCalibrationRequest{TrustedUntil: now.Add(time.Hour), DirectionErrorDeg: 3, Activate: true}); err != nil {
			t.Fatal(err)
		}
	}
	out, err := svc.IngestBatch(ctx, model.BatchFragmentsRequest{Fragments: []model.FragmentInput{{StationID: "a", Sequence: "1", ObservedAt: now, CenterHz: 1000000, BandwidthHz: 1000, StrengthDBm: -50, DirectionDeg: 20}, {StationID: "b", Sequence: "1", ObservedAt: now.Add(time.Second), CenterHz: 1001000, BandwidthHz: 1000, StrengthDBm: -52, DirectionDeg: 27}, {StationID: "c", Sequence: "1", ObservedAt: now.Add(2 * time.Second), CenterHz: 1000500, BandwidthHz: 1000, StrengthDBm: -53, DirectionDeg: 31}}})
	if err != nil {
		t.Fatal(err)
	}
	if out[0].EventID != out[2].EventID {
		t.Fatalf("events differ: %+v", out)
	}
	detail, err := svc.Event(ctx, out[0].EventID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Event.Status != model.EventConfirmed {
		t.Fatalf("got %s", detail.Event.Status)
	}
}
