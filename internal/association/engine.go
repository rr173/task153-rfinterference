package association

import (
	"context"
	"github.com/google/uuid"
	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/store"
	"time"
)

type Engine struct {
	store *store.Store
	now   func() time.Time
}

func New(s *store.Store) *Engine                  { return &Engine{store: s, now: time.Now} }
func (e *Engine) SetClock(clock func() time.Time) { e.now = clock }

type Decision struct {
	Event    model.Event
	Accepted bool
	Reason   string
	Created  bool
}

func (e *Engine) Associate(ctx context.Context, f model.Fragment) (Decision, error) {
	events, err := e.store.ActiveEvents(ctx)
	if err != nil {
		return Decision{}, err
	}
	evidence := make(map[string][]model.Fragment, len(events))
	for _, event := range events {
		frags, err := e.store.FragmentsForEvent(ctx, event.ID)
		if err != nil {
			return Decision{}, err
		}
		evidence[event.ID] = frags
	}
	ranked := RankCandidates(events, f, evidence)
	if len(ranked) > 0 {
		event := ranked[0].Event
		Extend(&event, f, e.now())
		return Decision{Event: event, Accepted: true}, nil
	}
	for _, event := range events {
		if FrequencyCompatible(event, f) && TimeCompatible(event, f) {
			return Decision{Event: event, Accepted: false, Reason: "arrival direction conflicts with the active event evidence"}, nil
		}
	}
	min, max := model.FrequencyRange(f.CenterHz, f.BandwidthHz)
	now := e.now().UTC()
	event := model.Event{ID: uuid.NewString(), Status: model.EventObserving, StartAt: f.CorrectedAt, EndAt: f.CorrectedAt, MinHz: min, MaxHz: max, Revision: 1, CreatedAt: now, UpdatedAt: now}
	return Decision{Event: event, Accepted: true, Created: true}, nil
}

func (e *Engine) ArchivedMatch(ctx context.Context, f model.Fragment) (model.Event, bool, error) {
	events, err := e.store.ArchivedEvents(ctx)
	if err != nil {
		return model.Event{}, false, err
	}
	for _, event := range events {
		if FrequencyCompatible(event, f) && TimeCompatible(event, f) {
			return event, true, nil
		}
	}
	return model.Event{}, false, nil
}
func (e *Engine) ApplyLifecycle(event *model.Event, fragments []model.Fragment, now time.Time) {
	// Archived events are immutable: lifecycle processing must never reactivate
	// them, or the frozen historical conclusions downstream rely on would drift.
	if event.Frozen || event.Status == model.EventArchived {
		return
	}
	stations := map[string]struct{}{}
	for _, f := range fragments {
		stations[f.StationID] = struct{}{}
	}
	switch {
	case len(stations) >= 3 && len(fragments) >= 3:
		event.Status = model.EventConfirmed
	case now.Sub(event.EndAt) > 15*time.Minute:
		event.Status = model.EventInsufficientEvidence
	default:
		event.Status = model.EventObserving
	}
	event.UpdatedAt = now.UTC()
}
func (e *Engine) Recover(ctx context.Context) (int, error) {
	events, err := e.store.ActiveEvents(ctx)
	if err != nil {
		return 0, err
	}
	for _, event := range events {
		frags, err := e.store.FragmentsForEvent(ctx, event.ID)
		if err != nil {
			return 0, err
		}
		e.ApplyLifecycle(&event, frags, e.now())
		if err := e.store.UpdateEvent(ctx, event); err != nil {
			return 0, err
		}
	}
	return len(events), nil
}
