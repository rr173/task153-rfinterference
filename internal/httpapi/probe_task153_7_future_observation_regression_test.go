package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"task153-rfinterference/internal/fragment"
	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/station"
	"task153-rfinterference/internal/store"
)

func TestBug07_FutureObservationsAreRejectedAtEveryBoundary(t *testing.T) {
	now := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)
	in := model.FragmentInput{StationID: "site-a", Sequence: "future-1", ObservedAt: now.Add(6 * time.Minute), CenterHz: 900_000_000, BandwidthHz: 20_000, StrengthDBm: -62, DirectionDeg: 40}
	if err := model.ValidateFragment(in, now); model.CodeOf(err) != model.CodeFutureObservation {
		t.Fatalf("model validation code=%q err=%v", model.CodeOf(err), err)
	}

	db, err := store.Open(filepath.Join(t.TempDir(), "future.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	stations := station.New(db)
	if _, err := stations.Register(context.Background(), model.RegisterStationRequest{ID: "site-a", Name: "Site A"}); err != nil {
		t.Fatal(err)
	}
	if _, err := stations.CreateCalibration(context.Background(), "site-a", model.CreateCalibrationRequest{TrustedUntil: now.Add(time.Hour), Activate: true}); err != nil {
		t.Fatal(err)
	}
	fragments := fragment.New(db, stations)
	fragments.SetClock(func() time.Time { return now })
	if _, _, err := fragments.Prepare(context.Background(), in); model.CodeOf(err) != model.CodeFutureObservation {
		t.Fatalf("fragment preparation code=%q err=%v", model.CodeOf(err), err)
	}

	w := httptest.NewRecorder()
	writeError(w, model.NewError(model.CodeFutureObservation, "observation is too far in the future"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("future-observation response status=%d, want %d", w.Code, http.StatusBadRequest)
	}
}
