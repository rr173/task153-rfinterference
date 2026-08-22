package model

import "fmt"

// EventTransitionAllowed protects the immutable archive boundary while keeping
// the transitions explicit for the service layer and API consumers.
func EventTransitionAllowed(from, to EventStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case EventObserving:
		return to == EventConfirmed || to == EventInsufficientEvidence || to == EventArchived
	case EventConfirmed:
		return to == EventInsufficientEvidence || to == EventArchived
	case EventInsufficientEvidence:
		return to == EventObserving || to == EventConfirmed || to == EventArchived
	case EventArchived:
		return false
	default:
		return false
	}
}

func ValidateEventTransition(from, to EventStatus) error {
	if EventTransitionAllowed(from, to) {
		return nil
	}
	return NewError(CodeArchived, "event transition %s -> %s is not allowed", from, to)
}

func IsTerminal(status EventStatus) bool {
	return status == EventArchived
}

func IsStationActive(status StationStatus) bool {
	return status == StationEnabled || status == StationCalibrating
}

func StatusDescription(status EventStatus) string {
	switch status {
	case EventObserving:
		return "evidence is still accumulating"
	case EventConfirmed:
		return "cross-station evidence meets confirmation threshold"
	case EventInsufficientEvidence:
		return "event has aged without enough independent evidence"
	case EventArchived:
		return "report is frozen and later evidence is revision-only"
	default:
		return fmt.Sprintf("unknown event state %q", status)
	}
}
