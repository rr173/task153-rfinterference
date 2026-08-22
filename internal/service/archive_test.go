package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/rr173/task153-rfinterference/internal/model"
	"github.com/rr173/task153-rfinterference/internal/store"
)

func TestLateFragmentIsRetainedOnArchivedEvent(t *testing.T) {
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
	first, err := svc.IngestBatch(ctx, model.BatchFragmentsRequest{Fragments: []model.FragmentInput{{
		StationID: "north", Sequence: "old-1", ObservedAt: old, CenterHz: 1000000, BandwidthHz: 1000, StrengthDBm: -50, DirectionDeg: 20,
	}}})
	if err != nil || len(first) != 1 || first[0].EventID == "" {
		t.Fatalf("initial ingest failed: %+v %v", first, err)
	}
	eventID := first[0].EventID
	if _, err := svc.Archive(ctx, eventID); err != nil {
		t.Fatal(err)
	}

	clock = clock.Add(30 * time.Minute)
	late, err := svc.IngestBatch(ctx, model.BatchFragmentsRequest{Fragments: []model.FragmentInput{{
		StationID: "north", Sequence: "old-2", ObservedAt: old.Add(time.Minute), CenterHz: 1000000, BandwidthHz: 1000, StrengthDBm: -51, DirectionDeg: 21,
	}}})
	if err != nil || len(late) != 1 {
		t.Fatalf("late ingest failed: %+v %v", late, err)
	}
	if late[0].Status != model.FragmentExcluded || late[0].EventID != eventID {
		t.Fatalf("late fragment was not retained on archived event: %+v", late[0])
	}

	events, err := svc.Events(ctx)
	if err != nil || len(events) != 1 || events[0].ID != eventID || !events[0].Frozen {
		t.Fatalf("archived event changed or a new event was created: %+v %v", events, err)
	}
	detail, err := svc.Event(ctx, eventID)
	if err != nil || len(detail.Exclusions) != 1 || detail.Exclusions[0].FragmentID != late[0].FragmentID {
		t.Fatalf("supplementary exclusion was not persisted: %+v %v", detail, err)
	}
}
