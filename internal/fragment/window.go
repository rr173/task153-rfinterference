package fragment

import (
	"github.com/rr173/task153-rfinterference/internal/model"
	"time"
)

const LateWindow = 20 * time.Minute

func IsLate(now time.Time, fragment model.Fragment) bool {
	return now.Sub(fragment.CorrectedAt) > LateWindow
}
func OverlapsFrequency(a, b model.Fragment, toleranceHz int64) bool {
	amin, amax := model.FrequencyRange(a.CenterHz, a.BandwidthHz)
	bmin, bmax := model.FrequencyRange(b.CenterHz, b.BandwidthHz)
	return amin <= bmax+toleranceHz && bmin <= amax+toleranceHz
}
