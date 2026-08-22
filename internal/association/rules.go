package association

import (
	"task153-rfinterference/internal/model"
	"time"
)

const FrequencyToleranceHz int64 = 25000
const EventGap = 3 * time.Minute
const DirectionToleranceDeg = 45.0

func DirectionDistance(a, b float64) float64 {
	return model.DirectionDeviation(a, b)
}
func FrequencyCompatible(event model.Event, f model.Fragment) bool {
	return model.FrequencySpan{MinHz: event.MinHz, MaxHz: event.MaxHz}.Overlaps(model.NewFrequencySpan(f.CenterHz, f.BandwidthHz), FrequencyToleranceHz)
}
func TimeCompatible(event model.Event, f model.Fragment) bool {
	return !f.CorrectedAt.Before(event.StartAt.Add(-EventGap)) && !f.CorrectedAt.After(event.EndAt.Add(EventGap))
}
func DirectionCompatible(existing []model.Fragment, f model.Fragment) bool {
	if len(existing) == 0 {
		return true
	}
	for _, old := range existing {
		if DirectionDistance(old.DirectionDeg, f.DirectionDeg) >= DirectionToleranceDeg {
			return true
		}
	}
	return false
}
func Extend(event *model.Event, f model.Fragment, now time.Time) {
	min, max := model.FrequencyRange(f.CenterHz, f.BandwidthHz)
	if min < event.MinHz {
		event.MinHz = min
	}
	if max > event.MaxHz {
		event.MaxHz = max
	}
	if f.CorrectedAt.Before(event.StartAt) {
		event.StartAt = f.CorrectedAt
	}
	if f.CorrectedAt.After(event.EndAt) {
		event.EndAt = f.CorrectedAt
	}
	event.UpdatedAt = now.UTC()
}
