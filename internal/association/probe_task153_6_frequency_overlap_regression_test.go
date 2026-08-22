package association

import (
	"testing"
	"time"

	"task153-rfinterference/internal/fragment"
	"task153-rfinterference/internal/model"
)

func TestBug06_FrequencyWindowsUseSharedHalfBandwidthRules(t *testing.T) {
	min, max := model.FrequencyRange(1000000, 10000)
	if min != 995000 || max != 1005000 {
		t.Fatalf("frequency range = %d..%d", min, max)
	}
	event := model.Event{StartAt: time.Now(), EndAt: time.Now(), MinHz: min, MaxHz: max}
	f := model.Fragment{CorrectedAt: event.EndAt, CenterHz: 1002000, BandwidthHz: 2000}
	if !FrequencyCompatible(event, f) {
		t.Fatal("overlapping evidence was not associated")
	}
	if !fragment.OverlapsFrequency(model.Fragment{CenterHz: 1000000, BandwidthHz: 10000}, model.Fragment{CenterHz: 1002000, BandwidthHz: 2000}, 0) {
		t.Fatal("fragment overlap rejected compatible spectra")
	}
}
