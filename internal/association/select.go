package association

import (
	"sort"
	"time"

	"task153-rfinterference/internal/model"
)

type Candidate struct {
	Event          model.Event
	FrequencyScore float64
	TimeScore      float64
	DirectionScore float64
	Score          float64
}

func RankCandidates(events []model.Event, fragment model.Fragment, evidence map[string][]model.Fragment) []Candidate {
	candidates := make([]Candidate, 0, len(events))
	for _, event := range events {
		if !FrequencyCompatible(event, fragment) || !TimeCompatible(event, fragment) {
			continue
		}
		existing := evidence[event.ID]
		if !DirectionCompatible(existing, fragment) {
			continue
		}
		candidate := Candidate{Event: event}
		candidate.FrequencyScore = frequencyScore(event, fragment)
		candidate.TimeScore = timeScore(event, fragment)
		candidate.DirectionScore = directionScore(existing, fragment)
		candidate.Score = candidate.FrequencyScore*0.45 + candidate.TimeScore*0.35 + candidate.DirectionScore*0.20
		candidates = append(candidates, candidate)
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].Event.CreatedAt.Before(candidates[j].Event.CreatedAt)
		}
		return candidates[i].Score > candidates[j].Score
	})
	return candidates
}

func frequencyScore(event model.Event, fragment model.Fragment) float64 {
	span := model.FrequencySpan{MinHz: event.MinHz, MaxHz: event.MaxHz}
	candidate := model.NewFrequencySpan(fragment.CenterHz, fragment.BandwidthHz)
	if !span.Overlaps(candidate, FrequencyToleranceHz) {
		return 0
	}
	distance := abs64(span.CenterHz() - candidate.CenterHz())
	width := span.WidthHz() + candidate.WidthHz() + FrequencyToleranceHz
	return 1 - minFloat(1, float64(distance)/float64(width))
}

func timeScore(event model.Event, fragment model.Fragment) float64 {
	distance := durationToWindow(fragment.CorrectedAt, event.StartAt, event.EndAt)
	return 1 - minFloat(1, float64(distance)/float64(EventGap))
}

func directionScore(existing []model.Fragment, fragment model.Fragment) float64 {
	if len(existing) == 0 {
		return 1
	}
	best := 0.0
	for _, old := range existing {
		deviation := DirectionDistance(old.DirectionDeg, fragment.DirectionDeg) / DirectionToleranceDeg
		score := 1 - minFloat(1, deviation)
		if score > best {
			best = score
		}
	}
	return best
}

func durationToWindow(value, start, end time.Time) time.Duration {
	if value.Before(start) {
		return start.Sub(value)
	}
	if value.After(end) {
		return value.Sub(end)
	}
	return 0
}

func abs64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
