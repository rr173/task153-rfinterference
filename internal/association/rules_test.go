package association

import (
	"task153-rfinterference/internal/model"
	"testing"
	"time"
)

func TestCompatibilityRules(t *testing.T) {
	now := time.Now()
	event := model.Event{StartAt: now.Add(-time.Minute), EndAt: now, MinHz: 990000, MaxHz: 1010000}
	f := model.Fragment{CorrectedAt: now, CenterHz: 1000000, BandwidthHz: 2000, DirectionDeg: 10}
	if !FrequencyCompatible(event, f) || !TimeCompatible(event, f) {
		t.Fatal("expected compatibility")
	}
	if DirectionCompatible([]model.Fragment{{DirectionDeg: 200}}, f) {
		t.Fatal("unexpected direction compatibility")
	}
}

// TestFrequencyWindowSharesHalfBandwidth locks the unification of the frequency
// window and half-bandwidth rules across the association pipeline. Scans with
// different bandwidths whose band edges overlap must be treated as one event,
// not split or merged inconsistently.
func TestFrequencyWindowSharesHalfBandwidth(t *testing.T) {
	narrow := model.Fragment{CenterHz: 1_000_000, BandwidthHz: 2_000}    // [999000,1001000]
	wide := model.Fragment{CenterHz: 1_000_000, BandwidthHz: 6_000}     // [997000,1003000]
	offset := model.Fragment{CenterHz: 1_001_500, BandwidthHz: 4_000}   // [999500,1003500]

	narrowMin, narrowMax := model.FrequencyRange(narrow.CenterHz, narrow.BandwidthHz)
	wideMin, wideMax := model.FrequencyRange(wide.CenterHz, wide.BandwidthHz)
	if narrowMin != 999000 || narrowMax != 1001000 {
		t.Fatalf("narrow half-bandwidth edges = [%d,%d], want [999000,1001000]", narrowMin, narrowMax)
	}
	if wideMin != 997000 || wideMax != 1003000 {
		t.Fatalf("wide half-bandwidth edges = [%d,%d], want [997000,1003000]", wideMin, wideMax)
	}

	// The frequency window rule must not be negated: overlapping bands are compatible.
	if !FrequencyCompatible(model.Event{MinHz: narrowMin, MaxHz: narrowMax}, wide) {
		t.Fatal("overlapping bands were split: FrequencyCompatible returned false")
	}

	// The same primitive must agree between association and fragment paths.
	if !associationOverlapsFragment(narrow, offset) {
		t.Fatal("association and fragment overlap rules disagree on edge-overlapping scans")
	}

	// A scan whose band falls outside the event window must not be compatible.
	far := model.Fragment{CenterHz: 2_000_000, BandwidthHz: 4_000}
	if FrequencyCompatible(model.Event{MinHz: narrowMin, MaxHz: narrowMax}, far) {
		t.Fatal("disjoint band was merged: FrequencyCompatible returned true")
	}
}

func associationOverlapsFragment(a, b model.Fragment) bool {
	// Mirror the fragment.OverlapsFrequency tolerance so both code paths are
	// exercised against the same model.FrequencySpan.Overlaps primitive.
	return model.NewFrequencySpan(a.CenterHz, a.BandwidthHz).
		Overlaps(model.NewFrequencySpan(b.CenterHz, b.BandwidthHz), FrequencyToleranceHz)
}
