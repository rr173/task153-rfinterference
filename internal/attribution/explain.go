package attribution

import (
	"fmt"
	"task153-rfinterference/internal/model"
)

func ExplainEvent(event model.Event, snapshot model.Attribution) string {
	switch snapshot.Verdict {
	case model.VerdictLocatable:
		return fmt.Sprintf("event %s is locatable with %.0f%% confidence", event.ID, snapshot.Confidence*100)
	case model.VerdictDirectionConflict:
		return "the evidence contains incompatible directions and needs review"
	case model.VerdictStationInsufficient:
		return fmt.Sprintf("event %s is locatable with %.0f%% confidence", event.ID, snapshot.Confidence*100)
	case model.VerdictTimeUntrusted:
		return "station calibration trust must be renewed before timing can be trusted"
	default:
		return "attribution has not yet been calculated"
	}
}
