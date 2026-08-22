package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/store"
)

func TestBug02_ReceiverIdentifiersRemainCanonicalAcrossIngest(t *testing.T) {
	if got := model.CanonicalIdentifier(" North-01 "); got != "north-01" {
		t.Fatalf("canonical identifier = %q", got)
	}
	db, err := store.Open(filepath.Join(t.TempDir(), "identity.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC()
	svc := New(db)
	svc.SetClock(func() time.Time { return now })
	if _, err := svc.RegisterStation(context.Background(), model.RegisterStationRequest{ID: "North-01", Name: "North"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateCalibration(context.Background(), "north-01", model.CreateCalibrationRequest{TrustedUntil: now.Add(time.Hour), Activate: true}); err != nil {
		t.Fatal(err)
	}
	input := model.FragmentInput{StationID: " NORTH-01 ", Sequence: "Scan-A", ObservedAt: now, CenterHz: 1000000, BandwidthHz: 1000, StrengthDBm: -50, DirectionDeg: 20}
	first, err := svc.IngestBatch(context.Background(), model.BatchFragmentsRequest{Fragments: []model.FragmentInput{input}})
	if err != nil || len(first) != 1 || first[0].Status != model.FragmentAccepted {
		t.Fatalf("first ingest failed: %+v %v", first, err)
	}
	input.StationID = "north-01"
	input.Sequence = "scan-a"
	second, err := svc.IngestBatch(context.Background(), model.BatchFragmentsRequest{Fragments: []model.FragmentInput{input}})
	if err != nil || len(second) != 1 || second[0].Status != model.FragmentDuplicate || second[0].FragmentID != first[0].FragmentID {
		t.Fatalf("case-only retry was not idempotent: first=%+v second=%+v err=%v", first, second, err)
	}
}
