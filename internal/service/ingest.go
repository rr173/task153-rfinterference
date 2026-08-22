package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/rr173/task153-rfinterference/internal/model"
)

func (s *Service) IngestBatch(ctx context.Context, req model.BatchFragmentsRequest) ([]model.IngestResult, error) {
	out := make([]model.IngestResult, 0, len(req.Fragments))
	for _, in := range req.Fragments {
		out = append(out, s.ingestOne(ctx, in))
	}
	return out, nil
}
func (s *Service) ingestOne(ctx context.Context, in model.FragmentInput) model.IngestResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, duplicate, err := s.fragments.Prepare(ctx, in)
	if err != nil {
		return model.IngestResult{Sequence: in.Sequence, Status: model.FragmentInvalid, Code: model.CodeOf(err), Reason: err.Error()}
	}
	if duplicate {
		return model.IngestResult{Sequence: in.Sequence, Status: model.FragmentDuplicate, FragmentID: f.ID, EventID: f.EventID}
	}
	if err := s.fragments.Persist(ctx, f); err != nil {
		return model.IngestResult{Sequence: in.Sequence, Status: model.FragmentInvalid, Code: model.CodeInternal, Reason: err.Error()}
	}
	decision, err := s.association.Associate(ctx, f)
	if err != nil {
		return model.IngestResult{Sequence: in.Sequence, Status: model.FragmentInvalid, FragmentID: f.ID, Code: model.CodeInternal, Reason: err.Error()}
	}
	if decision.Created {
		if err := s.store.SaveEvent(ctx, decision.Event); err != nil {
			return model.IngestResult{Sequence: in.Sequence, Status: model.FragmentInvalid, FragmentID: f.ID, Code: model.CodeInternal, Reason: err.Error()}
		}
	}
	if !decision.Accepted {
		_ = s.store.UpdateFragmentAssociation(ctx, f.ID, decision.Event.ID, model.FragmentExcluded, decision.Reason)
		_ = s.store.SaveExclusion(ctx, model.Exclusion{ID: uuid.NewString(), EventID: decision.Event.ID, FragmentID: f.ID, Reason: decision.Reason, CreatedAt: s.now().UTC()})
		_ = s.refresh(ctx, decision.Event.ID)
		return model.IngestResult{Sequence: in.Sequence, Status: model.FragmentExcluded, FragmentID: f.ID, EventID: decision.Event.ID, Reason: decision.Reason}
	}
	if err := s.store.UpdateFragmentAssociation(ctx, f.ID, decision.Event.ID, model.FragmentAccepted, ""); err != nil {
		return model.IngestResult{Sequence: in.Sequence, Status: model.FragmentInvalid, FragmentID: f.ID, Code: model.CodeInternal, Reason: err.Error()}
	}
	if !decision.Created {
		_ = s.store.UpdateEvent(ctx, decision.Event)
	}
	if err := s.refresh(ctx, decision.Event.ID); err != nil {
		return model.IngestResult{Sequence: in.Sequence, Status: model.FragmentInvalid, FragmentID: f.ID, Code: model.CodeInternal, Reason: err.Error()}
	}
	return model.IngestResult{Sequence: in.Sequence, Status: model.FragmentAccepted, FragmentID: f.ID, EventID: decision.Event.ID}
}
