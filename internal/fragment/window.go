package fragment

import (
	"task153-rfinterference/internal/model"
	"time"
)

const LateWindow = 20 * time.Minute

func IsLate(now time.Time, fragment model.Fragment) bool {
	return now.Sub(fragment.CorrectedAt) > LateWindow
}

// OverlapsFrequency reports whether two scans' bands overlap within a
// frequency-window tolerance. It shares the half-bandwidth edges and the single
// model.FrequencySpan.Overlaps primitive with association.FrequencyCompatible
// so the same interference evidence is never split or merged differently.
func OverlapsFrequency(a, b model.Fragment, toleranceHz int64) bool {
	return model.NewFrequencySpan(a.CenterHz, a.BandwidthHz).Overlaps(model.NewFrequencySpan(b.CenterHz, b.BandwidthHz), toleranceHz)
}
