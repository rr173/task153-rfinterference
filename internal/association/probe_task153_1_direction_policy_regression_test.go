package association

import (
	"testing"
	"time"

	"task153-rfinterference/internal/attribution"
	"task153-rfinterference/internal/model"
)

func TestBug01_DirectionPolicyKeepsCompatibleEvidenceTogether(t *testing.T) {
	now := time.Now().UTC()
	fragment := model.Fragment{CorrectedAt: now, CenterHz: 1000000, BandwidthHz: 1000, DirectionDeg: 20}
	if !DirectionCompatible([]model.Fragment{{DirectionDeg: 20}}, fragment) {
		t.Fatal("matching arrival direction was rejected")
	}
	events := []model.Event{
		{ID: "nearest", StartAt: now, EndAt: now, MinHz: 999000, MaxHz: 1001000, CreatedAt: now},
		{ID: "farther", StartAt: now, EndAt: now, MinHz: 999000, MaxHz: 1001000, CreatedAt: now.Add(time.Second)},
	}
	ranked := RankCandidates(events, fragment, map[string][]model.Fragment{
		"nearest": {{DirectionDeg: 20}},
		"farther": {{DirectionDeg: 60}},
	})
	if len(ranked) != 2 || ranked[0].Event.ID != "nearest" {
		t.Fatalf("nearest compatible evidence was not preferred: %+v", ranked)
	}
	calibrations := map[string]model.Calibration{}
	fragments := []model.Fragment{}
	for _, stationID := range []string{"a", "b", "c"} {
		calibrations[stationID] = model.Calibration{StationID: stationID, TrustedUntil: now.Add(time.Hour), CreatedAt: now}
	}
	for index, direction := range []float64{20, 27, 34} {
		fragments = append(fragments, model.Fragment{StationID: []string{"a", "b", "c"}[index], ObservedAt: now, DirectionDeg: direction, StrengthDBm: -50})
	}
	snapshot := attribution.New().Compute(model.Event{ID: "event", Revision: 1}, fragments, calibrations)
	if snapshot.Verdict != model.VerdictLocatable {
		t.Fatalf("compatible directions produced %s", snapshot.Verdict)
	}
}
