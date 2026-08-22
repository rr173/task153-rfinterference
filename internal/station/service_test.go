package station

import (
	"context"
	"github.com/rr173/task153-rfinterference/internal/model"
	"github.com/rr173/task153-rfinterference/internal/store"
	"path/filepath"
	"testing"
	"time"
)

func TestCalibrationCorrectsClock(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "rf.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := New(db)
	ctx := context.Background()
	if _, err := svc.Register(ctx, model.RegisterStationRequest{ID: "s1", Name: "one"}); err != nil {
		t.Fatal(err)
	}
	until := time.Now().Add(time.Hour)
	if _, err := svc.CreateCalibration(ctx, "s1", model.CreateCalibrationRequest{ClockOffsetMillis: 120, DirectionErrorDeg: 3, TrustedUntil: until, Activate: true}); err != nil {
		t.Fatal(err)
	}
	at := time.Now().UTC()
	corrected, _, err := svc.CorrectTime(ctx, "s1", at)
	if err != nil {
		t.Fatal(err)
	}
	if corrected.Sub(at) != -120*time.Millisecond {
		t.Fatalf("got %s", corrected.Sub(at))
	}
}
