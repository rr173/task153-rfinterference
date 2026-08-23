package httpapi

import (
	"bytes"
	"context"
	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/service"
	"task153-rfinterference/internal/store"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestStationEndpoint(t *testing.T) {
	db, _ := store.Open(filepath.Join(t.TempDir(), "api.db"))
	defer db.Close()
	svc := service.New(db)
	_ = svc.Recover(context.Background())
	h := New(svc).Handler()
	r := httptest.NewRequest(http.MethodPost, "/v1/stations", bytes.NewBufferString(`{"id":"test","name":"Test","latitude":1,"longitude":2}`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
}

func TestOperatorPage(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "page.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := New(service.New(db)).Handler()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "text/html; charset=utf-8" || !bytes.Contains(w.Body.Bytes(), []byte("无线干扰事件归因")) {
		t.Fatalf("unexpected operator page: %d %s %s", w.Code, w.Header().Get("Content-Type"), w.Body.String())
	}
}

func TestIngestRejectsFarFutureObservation(t *testing.T) {
	db, _ := store.Open(filepath.Join(t.TempDir(), "future.db"))
	defer db.Close()
	svc := service.New(db)
	_ = svc.Recover(context.Background())
	now := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)
	svc.SetClock(func() time.Time { return now })
	ctx := context.Background()
	if _, err := svc.RegisterStation(ctx, model.RegisterStationRequest{ID: "s", Name: "s"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateCalibration(ctx, "s", model.CreateCalibrationRequest{TrustedUntil: now.Add(time.Hour), Activate: true}); err != nil {
		t.Fatal(err)
	}
	h := New(svc).Handler()
	body := `{"fragments":[{"station_id":"s","sequence":"1","observed_at":"2026-08-22T10:05:01Z","center_hz":1000000,"bandwidth_hz":1000,"strength_dbm":-60,"direction_deg":20}]}`
	r := httptest.NewRequest(http.MethodPost, "/v1/fragments:batch", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with per-fragment result, got %d %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(model.CodeFutureObservation)) {
		t.Fatalf("expected %s in result, got %s", model.CodeFutureObservation, w.Body.String())
	}
}

func TestWriteErrorMapsFutureObservationToBadRequest(t *testing.T) {
	// writeError is what every endpoint uses to translate a model.Error to an HTTP
	// status. Future observations must surface as a client error (400), never the
	// 500 INTERNAL fallback, so realtime-event protection is stable across boundaries.
	w := httptest.NewRecorder()
	writeError(w, model.NewError(model.CodeFutureObservation, "observation is too far in the future"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d for future observation, got %d %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(model.CodeFutureObservation)) {
		t.Fatalf("expected %s in body, got %s", model.CodeFutureObservation, w.Body.String())
	}
}
