package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/store"
)

func TestBug04_LateEvidenceStaysWithArchivedEvent(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "archive.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	clock := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)
	svc := New(db)
	svc.SetClock(func() time.Time { return clock })
	if _, err := svc.RegisterStation(ctx, model.RegisterStationRequest{ID: "north", Name: "North"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateCalibration(ctx, "north", model.CreateCalibrationRequest{TrustedUntil: clock.Add(2 * time.Hour), Activate: true}); err != nil {
		t.Fatal(err)
	}
	old := clock.Add(-30 * time.Minute)
	first, err := svc.IngestBatch(ctx, model.BatchFragmentsRequest{Fragments: []model.FragmentInput{{StationID: "north", Sequence: "old-1", ObservedAt: old, CenterHz: 1000000, BandwidthHz: 1000, StrengthDBm: -50, DirectionDeg: 20}}})
	if err != nil || len(first) != 1 || first[0].EventID == "" {
		t.Fatalf("initial ingest failed: %+v %v", first, err)
	}
	if _, err := svc.Archive(ctx, first[0].EventID); err != nil {
		t.Fatal(err)
	}
	clock = clock.Add(30 * time.Minute)
	late, err := svc.IngestBatch(ctx, model.BatchFragmentsRequest{Fragments: []model.FragmentInput{{StationID: "north", Sequence: "old-2", ObservedAt: old.Add(time.Minute), CenterHz: 1000000, BandwidthHz: 1000, StrengthDBm: -51, DirectionDeg: 21}}})
	if err != nil || len(late) != 1 || late[0].Status != model.FragmentExcluded || late[0].EventID != first[0].EventID {
		t.Fatalf("late evidence was not retained as a supplement: first=%+v late=%+v err=%v", first, late, err)
	}
	events, err := svc.Events(ctx)
	if err != nil || len(events) != 1 || !events[0].Frozen {
		t.Fatalf("archived report changed: %+v %v", events, err)
	}
}
