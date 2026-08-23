package station

import (
	"testing"
	"time"

	"task153-rfinterference/internal/model"
)

func TestAssessTrust(t *testing.T) {
	now := time.Now()
	for _, tc := range []struct {
		name   string
		cal    model.Calibration
		when   time.Time
		trust  bool
		reason bool
	}{
		{
			name:   "within trust window is trusted",
			cal:    model.Calibration{TrustedUntil: now.Add(time.Hour), CreatedAt: now.Add(-time.Hour)},
			when:   now,
			trust:  true,
			reason: false,
		},
		{
			name:   "expired calibration is not trusted",
			cal:    model.Calibration{TrustedUntil: now.Add(-time.Hour), CreatedAt: now.Add(-2 * time.Hour)},
			when:   now,
			trust:  false,
			reason: true,
		},
		{
			name:   "missing trust window is not trusted",
			cal:    model.Calibration{CreatedAt: now.Add(-time.Hour)},
			when:   now,
			trust:  false,
			reason: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := AssessTrust(tc.cal, tc.when, now)
			if a.Trusted != tc.trust {
				t.Fatalf("trusted=%v want %v", a.Trusted, tc.trust)
			}
			if a.Trusted == false && !tc.reason {
				t.Fatalf("expected no reason, got %q", a.Reason)
			}
			if a.Trusted == false && a.Reason == "" {
				t.Fatalf("expected a reason for untrusted assessment")
			}
		})
	}
}
