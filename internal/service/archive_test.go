package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/store"
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

// TestArchivedEventStaysImmutableAcrossLifecycle is the regression guard for the
// archive boundary: after an event is archived, lifecycle processing (Recover,
// which re-applies ApplyLifecycle to active events) and any later refresh must
// not reactivate it, bump its revision, or write a fresh attribution snapshot.
// Historical conclusions have to stay frozen.
func TestArchivedEventStaysImmutableAcrossLifecycle(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "immutable.db"))
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

	now := clock
	first, err := svc.IngestBatch(ctx, model.BatchFragmentsRequest{Fragments: []model.FragmentInput{{
		StationID: "north", Sequence: "s1", ObservedAt: now, CenterHz: 1000000, BandwidthHz: 1000, StrengthDBm: -50, DirectionDeg: 20,
	}}})
	if err != nil || len(first) != 1 || first[0].EventID == "" {
		t.Fatalf("initial ingest failed: %+v %v", first, err)
	}
	eventID := first[0].EventID

	archived, err := svc.Archive(ctx, eventID)
	if err != nil {
		t.Fatal(err)
	}
	if !archived.Frozen || archived.Status != model.EventArchived {
		t.Fatalf("archive did not freeze: %+v", archived)
	}
	frozenRevision := archived.Revision
	frozenSnapshots := 1 // Archive no longer writes a snapshot; the single refresh during ingest did.

	// Recover re-runs ApplyLifecycle over active events; it must skip the
	// archived one entirely rather than reactivate it.
	if err := svc.Recover(ctx); err != nil {
		t.Fatal(err)
	}

	// Advance the clock well past the ageing threshold so lifecycle would mark
	// the event insufficient_evidence if it were still active.
	clock = clock.Add(2 * time.Hour)

	events, err := svc.Events(ctx)
	if err != nil || len(events) != 1 {
		t.Fatalf("expected a single event: %+v %v", events, err)
	}
	if events[0].Status != model.EventArchived || !events[0].Frozen || events[0].Revision != frozenRevision {
		t.Fatalf("archived event was mutated by lifecycle: %+v", events[0])
	}

	detail, err := svc.Event(ctx, eventID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Event.Status != model.EventArchived || detail.Event.Revision != frozenRevision {
		t.Fatalf("archived event mutated on detail read: %+v", detail.Event)
	}
	if len(detail.Attribution) != frozenSnapshots {
		t.Fatalf("attribution snapshots for archived event drifted: got %d want %d", len(detail.Attribution), frozenSnapshots)
	}

	// Re-archiving an already-archived event must be rejected: there is no
	// permitted transition away from the terminal archived state.
	if _, err := svc.Archive(ctx, eventID); err == nil {
		t.Fatalf("re-archiving an archived event should fail")
	}

	// The archived event is terminal and cannot transition back to observing.
	if model.EventTransitionAllowed(model.EventArchived, model.EventObserving) {
		t.Fatalf("archived -> observing must not be allowed")
	}
}
