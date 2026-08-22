package service

import (
	"context"
	"github.com/rr173/task153-rfinterference/internal/association"
	"github.com/rr173/task153-rfinterference/internal/attribution"
	"github.com/rr173/task153-rfinterference/internal/fragment"
	"github.com/rr173/task153-rfinterference/internal/model"
	"github.com/rr173/task153-rfinterference/internal/station"
	"github.com/rr173/task153-rfinterference/internal/store"
	"sync"
	"time"
)

type Service struct {
	store       *store.Store
	stations    *station.Service
	fragments   *fragment.Service
	association *association.Engine
	attribution *attribution.Computer
	now         func() time.Time
	mu          sync.Mutex
	recovered   int
}

func New(s *store.Store) *Service {
	st := station.New(s)
	return &Service{store: s, stations: st, fragments: fragment.New(s, st), association: association.New(s), attribution: attribution.New(), now: time.Now}
}
func (s *Service) SetClock(clock func() time.Time) {
	s.now = clock
	s.stations.SetClock(clock)
	s.fragments.SetClock(clock)
	s.association.SetClock(clock)
	s.attribution.SetClock(clock)
}
func (s *Service) Recover(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, err := s.association.Recover(ctx)
	s.recovered = n
	return err
}
func (s *Service) RegisterStation(ctx context.Context, req model.RegisterStationRequest) (model.Station, error) {
	return s.stations.Register(ctx, req)
}
func (s *Service) CreateCalibration(ctx context.Context, id string, req model.CreateCalibrationRequest) (model.Calibration, error) {
	return s.stations.CreateCalibration(ctx, id, req)
}
