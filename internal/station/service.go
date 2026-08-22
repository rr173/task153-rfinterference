package station

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rr173/task153-rfinterference/internal/model"
	"github.com/rr173/task153-rfinterference/internal/store"
)

type Service struct {
	store *store.Store
	now   func() time.Time
}

func New(s *store.Store) *Service                  { return &Service{store: s, now: time.Now} }
func (s *Service) SetClock(clock func() time.Time) { s.now = clock }

func (s *Service) Register(ctx context.Context, req model.RegisterStationRequest) (model.Station, error) {
	if err := model.ValidateStation(req); err != nil {
		return model.Station{}, err
	}
	station := model.Station{ID: model.CanonicalIdentifier(req.ID), Name: req.Name, Latitude: req.Latitude, Longitude: req.Longitude, Status: model.StationCalibrating, CreatedAt: s.now().UTC()}
	if err := s.store.SaveStation(ctx, station); err != nil {
		return model.Station{}, err
	}
	return station, nil
}
func (s *Service) CreateCalibration(ctx context.Context, stationID string, req model.CreateCalibrationRequest) (model.Calibration, error) {
	if err := model.ValidateCalibration(req); err != nil {
		return model.Calibration{}, err
	}
	station, err := s.store.GetStation(ctx, stationID)
	if err != nil {
		return model.Calibration{}, err
	}
	if station.Status == model.StationDisabled {
		return model.Calibration{}, model.NewError(model.CodeDisabledStation, "station %s is disabled", stationID)
	}
	version, err := s.store.NextCalibrationVersion(ctx, stationID)
	if err != nil {
		return model.Calibration{}, err
	}
	status := model.CalibrationDraft
	if req.Activate {
		status = model.CalibrationActive
	}
	c := model.Calibration{ID: uuid.NewString(), StationID: stationID, Version: version, ClockOffsetMillis: req.ClockOffsetMillis, DirectionErrorDeg: req.DirectionErrorDeg, TrustedUntil: req.TrustedUntil.UTC(), Status: status, CreatedAt: s.now().UTC()}
	if err := s.store.SaveCalibration(ctx, c); err != nil {
		return model.Calibration{}, err
	}
	return c, nil
}
func (s *Service) CorrectTime(ctx context.Context, stationID string, observed time.Time) (time.Time, model.Calibration, error) {
	c, err := s.store.ActiveCalibration(ctx, stationID)
	if err != nil {
		return time.Time{}, model.Calibration{}, err
	}
	return observed.Add(-time.Duration(c.ClockOffsetMillis) * time.Millisecond).UTC(), c, nil
}
func (s *Service) IsTrusted(c model.Calibration, observed time.Time) bool {
	return !observed.After(c.TrustedUntil)
}
