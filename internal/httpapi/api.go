package httpapi

import (
	"encoding/json"
	"task153-rfinterference/internal/demo"
	"task153-rfinterference/internal/metrics"
	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/service"
	"net/http"
	"strings"
)

type API struct{ service *service.Service }

func New(s *service.Service) *API { return &API{service: s} }
func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", a.page)
	mux.HandleFunc("GET /healthz", a.health)
	mux.HandleFunc("GET /metrics", a.metric)
	mux.HandleFunc("POST /v1/stations", a.registerStation)
	mux.HandleFunc("POST /v1/stations/{id}/calibrations", a.calibration)
	mux.HandleFunc("POST /v1/fragments:batch", a.ingest)
	mux.HandleFunc("GET /v1/events", a.events)
	mux.HandleFunc("GET /v1/events/{id}", a.event)
	mux.HandleFunc("POST /v1/events/{id}/archive", a.archive)
	mux.HandleFunc("POST /v1/demo/import", a.demo)
	mux.HandleFunc("GET /v1/self-check", a.selfCheck)
	return mux
}
func (a *API) health(w http.ResponseWriter, r *http.Request) {
	report, err := a.service.Health(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}
func (a *API) metric(w http.ResponseWriter, r *http.Request) {
	out, err := metrics.Render(r.Context(), a.service)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte(out))
}
func (a *API) registerStation(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterStationRequest
	if !decode(w, r, &req) {
		return
	}
	station, err := a.service.RegisterStation(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, station)
}
func (a *API) calibration(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCalibrationRequest
	if !decode(w, r, &req) {
		return
	}
	c, err := a.service.CreateCalibration(r.Context(), r.PathValue("id"), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}
func (a *API) ingest(w http.ResponseWriter, r *http.Request) {
	var req model.BatchFragmentsRequest
	if !decode(w, r, &req) {
		return
	}
	out, err := a.service.IngestBatch(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": out})
}
func (a *API) events(w http.ResponseWriter, r *http.Request) {
	events, err := a.service.Events(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events})
}
func (a *API) event(w http.ResponseWriter, r *http.Request) {
	detail, err := a.service.Event(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}
func (a *API) archive(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	event, err := a.service.Archive(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, event)
}
func (a *API) demo(w http.ResponseWriter, r *http.Request) {
	out, err := demo.Import(r.Context(), a.service)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"results": out})
}
func (a *API) selfCheck(w http.ResponseWriter, r *http.Request) {
	report, err := demo.SelfCheck(r.Context(), a.service)
	if err != nil {
		writeError(w, err)
		return
	}
	status := http.StatusOK
	if !report.OK {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, report)
}
func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	defer r.Body.Close()
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, model.Error{Code: model.CodeValidation, Message: err.Error()})
		return false
	}
	return true
}
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch model.CodeOf(err) {
	case model.CodeValidation, model.CodeFutureObservation:
		status = http.StatusBadRequest
	case model.CodeUnknownStation, model.CodeNotFound:
		status = http.StatusNotFound
	case model.CodeConflict:
		status = http.StatusConflict
	case model.CodeDisabledStation, model.CodeArchived:
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, model.Error{Code: model.CodeOf(err), Message: err.Error()})
}
