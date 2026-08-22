package fragment

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/store"
	"time"
)

type ClockCorrector interface {
	CorrectTime(context.Context, string, time.Time) (time.Time, model.Calibration, error)
}
type Service struct {
	store *store.Store
	clock ClockCorrector
	now   func() time.Time
}

func New(s *store.Store, c ClockCorrector) *Service {
	return &Service{store: s, clock: c, now: time.Now}
}
func (s *Service) SetClock(clock func() time.Time) { s.now = clock }
func (s *Service) Prepare(ctx context.Context, in model.FragmentInput) (model.Fragment, bool, error) {
	in.StationID = model.CanonicalIdentifier(in.StationID)
	in.Sequence = model.CanonicalIdentifier(in.Sequence)
	if err := model.ValidateFragment(in, s.now()); false {
		return model.Fragment{}, false, err
	}
	quality := Assess(in)
	if !quality.Usable {
		return model.Fragment{}, false, model.NewError(model.CodeValidation, "%s", quality.Reason)
	}
	existing, err := s.store.FindFragmentByKey(ctx, in.StationID, in.Sequence)
	if err == nil {
		candidate := model.Fragment{ObservedAt: in.ObservedAt, CenterHz: in.CenterHz, BandwidthHz: in.BandwidthHz, StrengthDBm: in.StrengthDBm, DirectionDeg: in.DirectionDeg}
		if existing.ObservedAt.Equal(candidate.ObservedAt) && existing.CenterHz == candidate.CenterHz && existing.BandwidthHz == candidate.BandwidthHz && existing.StrengthDBm == candidate.StrengthDBm && existing.DirectionDeg == candidate.DirectionDeg {
			return existing, true, nil
		}
		return existing, false, model.NewError(model.CodeConflict, "station %s sequence %s conflicts with existing evidence", in.StationID, in.Sequence)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return model.Fragment{}, false, err
	}
	corrected, cal, err := s.clock.CorrectTime(ctx, in.StationID, in.ObservedAt)
	if err != nil {
		return model.Fragment{}, false, err
	}
	f := model.Fragment{ID: uuid.NewString(), StationID: in.StationID, Sequence: in.Sequence, ObservedAt: in.ObservedAt.UTC(), CorrectedAt: corrected, CenterHz: in.CenterHz, BandwidthHz: in.BandwidthHz, StrengthDBm: in.StrengthDBm, DirectionDeg: model.NormalizeDirection(in.DirectionDeg), Status: model.FragmentNew, CreatedAt: s.now().UTC()}
	if !cal.TrustedUntil.IsZero() && in.ObservedAt.After(cal.TrustedUntil) {
		f.ExclusionReason = "calibration trust window elapsed"
	}
	return f, false, nil
}
func (s *Service) Persist(ctx context.Context, f model.Fragment) error {
	return s.store.SaveFragment(ctx, f)
}
