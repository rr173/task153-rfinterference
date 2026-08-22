package model

import "sort"

type EventSummary struct {
	ID            string             `json:"id"`
	Status        EventStatus        `json:"status"`
	FrequencyHz   int64              `json:"frequency_hz"`
	FragmentCount int                `json:"fragment_count"`
	Verdict       AttributionVerdict `json:"verdict"`
	Confidence    float64            `json:"confidence"`
}

func Summarize(event Event, fragments []Fragment, snapshot *Attribution) EventSummary {
	s := EventSummary{ID: event.ID, Status: event.Status, FrequencyHz: (event.MinHz + event.MaxHz) / 2, FragmentCount: len(fragments), Verdict: VerdictNotCalculated}
	if snapshot != nil {
		s.Verdict, s.Confidence = snapshot.Verdict, snapshot.Confidence
	}
	return s
}

func SortEvents(events []Event) {
	sort.Slice(events, func(i, j int) bool { return events[i].UpdatedAt.After(events[j].UpdatedAt) })
}
