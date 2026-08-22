package service

import (
	"fmt"
	"sort"

	"github.com/rr173/task153-rfinterference/internal/model"
)

// BuildRevisionNote makes the event's immutable/reportable history explicit.
// It intentionally derives counts from persisted evidence rather than mutable
// in-memory counters so a post-restart detail query reads the same way.
func BuildRevisionNote(event model.Event, fragments []model.Fragment, exclusions []model.Exclusion, snapshots []model.Attribution) model.RevisionNote {
	note := model.RevisionNote{
		Revision:         event.Revision,
		Frozen:           event.Frozen,
		AcceptedEvidence: len(fragments),
		ExcludedEvidence: len(exclusions),
	}
	if event.Frozen {
		note.Summary = fmt.Sprintf("archived report revision %d is frozen; later historical evidence is supplementary", event.Revision)
		return note
	}
	if len(snapshots) == 0 {
		note.Summary = "evidence is persisted but no attribution snapshot has been calculated"
		return note
	}
	latest := snapshots[len(snapshots)-1]
	note.Summary = fmt.Sprintf("working revision %d currently reports %s", event.Revision, latest.Verdict)
	if len(exclusions) > 0 {
		note.Summary += fmt.Sprintf(" with %d retained exclusion(s)", len(exclusions))
	}
	return note
}

func LatestVerdict(snapshots []model.Attribution) (model.Attribution, bool) {
	if len(snapshots) == 0 {
		return model.Attribution{}, false
	}
	ordered := append([]model.Attribution(nil), snapshots...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Revision == ordered[j].Revision {
			return ordered[i].CreatedAt.After(ordered[j].CreatedAt)
		}
		return ordered[i].Revision > ordered[j].Revision
	})
	return ordered[0], true
}

func NeedsHumanReview(event model.Event, snapshots []model.Attribution, exclusions []model.Exclusion) bool {
	if event.Frozen {
		return false
	}
	latest, ok := LatestVerdict(snapshots)
	if !ok {
		return true
	}
	if latest.Verdict == model.VerdictDirectionConflict || latest.Verdict == model.VerdictTimeUntrusted {
		return true
	}
	return len(exclusions) > 0 && latest.Confidence < 0.7
}
