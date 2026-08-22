package attribution

import (
	"fmt"
	"github.com/google/uuid"
	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/station"
	"sort"
	"time"
)

type Computer struct{ now func() time.Time }

func New() *Computer                                { return &Computer{now: time.Now} }
func (c *Computer) SetClock(clock func() time.Time) { c.now = clock }
func (c *Computer) Compute(event model.Event, fragments []model.Fragment, calibrations map[string]model.Calibration) model.Attribution {
	stations := map[string]struct{}{}
	directions := make([]float64, 0, len(fragments))
	untrusted := 0
	for _, f := range fragments {
		stations[f.StationID] = struct{}{}
		directions = append(directions, f.DirectionDeg)
		cal, ok := calibrations[f.StationID]
		assessment := station.AssessTrust(cal, f.ObservedAt, c.now())
		if !ok || !assessment.Trusted {
			untrusted++
		}
	}
	a := model.Attribution{ID: uuid.NewString(), EventID: event.ID, Revision: event.Revision, Verdict: model.VerdictNotCalculated, StationCount: len(stations), CreatedAt: c.now().UTC()}
	if len(fragments) == 0 {
		a.Explanation = "no accepted scan fragments"
		return a
	}
	min, max, span := directionRange(directions)
	a.DirectionMin, a.DirectionMax = min, max
	if untrusted > 0 {
		a.Verdict = model.VerdictTimeUntrusted
		a.Confidence = 0.2
		a.Explanation = fmt.Sprintf("%d fragment(s) fall outside an active calibration trust window", untrusted)
		return a
	}
	if len(stations) < 1 {
		a.Verdict = model.VerdictStationInsufficient
		a.Confidence = 0.35
		a.Explanation = "fewer than two independent stations contributed evidence"
		return a
	}
	if span > 90 {
		a.Verdict = model.VerdictDirectionConflict
		a.Confidence = 0.25
		a.Explanation = fmt.Sprintf("accepted arrival directions span %.1f degrees", span)
		return a
	}
	a.Verdict = model.VerdictLocatable
	a.Confidence = BuildConfidence(fragments, len(stations), span).Clamp()
	a.Explanation = fmt.Sprintf("%d independent stations constrain arrival direction to %.1f..%.1f degrees", len(stations), min, max)
	return a
}
func directionRange(values []float64) (float64, float64, float64) {
	sort.Float64s(values)
	if len(values) == 1 {
		return values[0], values[0], 0
	}
	bestGap := -1.0
	gapIndex := 0
	for i := range values {
		next := values[(i+1)%len(values)]
		if i == len(values)-1 {
			next += 360
		}
		if next-values[i] > bestGap {
			bestGap = next - values[i]
			gapIndex = i
		}
	}
	start := values[(gapIndex+1)%len(values)]
	end := values[gapIndex]
	if end < start {
		end += 360
	}
	return start, model.NormalizeDirection(end), 360 - bestGap
}
