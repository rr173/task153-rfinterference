package httpapi

import (
	"bytes"
	"context"
	"task153-rfinterference/internal/service"
	"task153-rfinterference/internal/store"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
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
