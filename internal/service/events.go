package service

import (
	"context"
	"task153-rfinterference/internal/model"
	"time"
)

func (s *Service) refresh(ctx context.Context, eventID string) error {
	event, err := s.store.GetEvent(ctx, eventID)
	if err != nil {
		return err
	}
	fs, err := s.store.FragmentsForEvent(ctx, eventID)
	if err != nil {
		return err
	}
	if _, err := s.store.LatestAttribution(ctx, eventID); err == nil {
		event.Revision++
	}
	s.association.ApplyLifecycle(&event, fs, s.now())
	if err := s.store.UpdateEvent(ctx, event); err != nil {
		return err
	}
	cal := map[string]model.Calibration{}
	for _, f := range fs {
		if _, ok := cal[f.StationID]; ok {
			continue
		}
		c, err := s.store.ActiveCalibration(ctx, f.StationID)
		if err == nil {
			cal[f.StationID] = c
		}
	}
	snapshot := s.attribution.Compute(event, fs, cal)
	return s.store.SaveAttribution(ctx, snapshot)
}
func (s *Service) Events(ctx context.Context) ([]model.Event, error) { return s.store.ListEvents(ctx) }
func (s *Service) Event(ctx context.Context, id string) (model.EventDetail, error) {
	e, err := s.store.GetEvent(ctx, id)
	if err != nil {
		return model.EventDetail{}, err
	}
	fs, err := s.store.FragmentsForEvent(ctx, id)
	if err != nil {
		return model.EventDetail{}, err
	}
	ex, err := s.store.ExclusionsForEvent(ctx, id)
	if err != nil {
		return model.EventDetail{}, err
	}
	a, err := s.store.AttributionForEvent(ctx, id)
	if err != nil {
		return model.EventDetail{}, err
	}
	return model.EventDetail{Event: e, Fragments: fs, Exclusions: ex, Attribution: a, RevisionNote: BuildRevisionNote(e, fs, ex, a)}, nil
}
func (s *Service) Archive(ctx context.Context, id string) (model.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	event, err := s.store.GetEvent(ctx, id)
	if err != nil {
		return model.Event{}, err
	}
	event.Frozen = true
	event.Status = model.EventArchived
	event.Revision++
	event.UpdatedAt = s.now().UTC()
	if err := s.store.UpdateEvent(ctx, event); err != nil {
		return model.Event{}, err
	}
	if err := s.refresh(ctx, id); err != nil {
		return model.Event{}, err
	}
	return event, nil
}
func (s *Service) Health(ctx context.Context) (model.HealthReport, error) {
	events, err := s.store.CountEvents(ctx)
	if err != nil {
		return model.HealthReport{}, err
	}
	fragments, err := s.store.CountFragments(ctx)
	if err != nil {
		return model.HealthReport{}, err
	}
	return model.HealthReport{OK: true, EventCount: events, FragmentCount: fragments, RecoveredActive: s.recovered, CheckedAt: s.now().UTC()}, nil
}
func (s *Service) Now() time.Time { return s.now().UTC() }
