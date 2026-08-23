package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"task153-rfinterference/internal/association"
	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/store"
)

func TestBug05_ArchivedReportsRemainImmutable(t *testing.T) {
	if model.EventTransitionAllowed(model.EventArchived, model.EventObserving) {
		t.Fatal("archived event was allowed to resume observation")
	}
	db, err := store.Open(filepath.Join(t.TempDir(), "archive.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC()
	engine := association.New(db)
	frozen := model.Event{ID: "frozen", Status: model.EventArchived, Frozen: true, EndAt: now}
	engine.ApplyLifecycle(&frozen, nil, now.Add(time.Hour))
	if frozen.Status != model.EventArchived {
		t.Fatalf("lifecycle reopened archived report: %+v", frozen)
	}
	event := model.Event{ID: "report", Status: model.EventObserving, StartAt: now, EndAt: now, MinHz: 1, MaxHz: 2, Revision: 1, CreatedAt: now, UpdatedAt: now}
	if err := db.SaveEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	svc := New(db)
	svc.SetClock(func() time.Time { return now })
	archived, err := svc.Archive(context.Background(), event.ID)
	if err != nil || !archived.Frozen || archived.Status != model.EventArchived {
		t.Fatalf("archive changed report mutability: %+v %v", archived, err)
	}
}
