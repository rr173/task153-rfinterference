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
	if len(stations) < 2 {
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
	if len(values) == 0 {
		return 0, 0, 0
	}
	norm := make([]float64, len(values))
	for i, v := range values {
		norm[i] = model.NormalizeDirection(v)
	}
	sort.Float64s(norm)
	if len(norm) == 1 {
		return norm[0], norm[0], 0
	}
	// Find the largest gap between consecutive directions (including the
	// wrap-around across north). The covered arc is the complement of that
	// gap, and its endpoints sit on either side of the largest gap so the
	// reported min/max stay continuous when evidence straddles north.
	maxGap := 0.0
	maxIdx := 0
	for i := 0; i < len(norm); i++ {
		next := norm[(i+1)%len(norm)]
		gap := next - norm[i]
		if gap < 0 {
			gap += 360
		}
		if gap > maxGap {
			maxGap = gap
			maxIdx = i
		}
	}
	span := 360 - maxGap
	min := norm[(maxIdx+1)%len(norm)]
	max := norm[maxIdx]
	return min, max, span
}
