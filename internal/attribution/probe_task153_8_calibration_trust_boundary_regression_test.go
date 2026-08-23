package attribution

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"task153-rfinterference/internal/fragment"
	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/station"
	"task153-rfinterference/internal/store"
)

func TestBug08_ExpiredCalibrationIsUntrustedAcrossEvidenceFlow(t *testing.T) {
	now := time.Date(2026, 8, 22, 11, 0, 0, 0, time.UTC)
	cal := model.Calibration{StationID: "site-a", TrustedUntil: now.Add(-time.Minute)}
	if assessment := station.AssessTrust(cal, now, now); assessment.Trusted {
		t.Fatalf("expired calibration unexpectedly trusted: %+v", assessment)
	}

	db, err := store.Open(filepath.Join(t.TempDir(), "trust.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	stations := station.New(db)
	stations.SetClock(func() time.Time { return now })
	ctx := context.Background()
	if _, err := stations.Register(ctx, model.RegisterStationRequest{ID: "site-a", Name: "Site A"}); err != nil {
		t.Fatal(err)
	}
	if _, err := stations.CreateCalibration(ctx, "site-a", model.CreateCalibrationRequest{TrustedUntil: cal.TrustedUntil, Activate: true}); err != nil {
		t.Fatal(err)
	}
	fragments := fragment.New(db, stations)
	fragments.SetClock(func() time.Time { return now })
	f, _, err := fragments.Prepare(ctx, model.FragmentInput{StationID: "site-a", Sequence: "expired-1", ObservedAt: now, CenterHz: 900_000_000, BandwidthHz: 20_000, StrengthDBm: -60, DirectionDeg: 30})
	if err != nil {
		t.Fatal(err)
	}
	if f.ExclusionReason == "" {
		t.Fatal("expired evidence did not retain its calibration-window exclusion")
	}

	c := New()
	c.SetClock(func() time.Time { return now })
	snapshot := c.Compute(model.Event{ID: "event-1", Revision: 1}, []model.Fragment{{StationID: "site-a", ObservedAt: now, CorrectedAt: now, DirectionDeg: 30}}, map[string]model.Calibration{"site-a": cal})
	if snapshot.Verdict != model.VerdictTimeUntrusted {
		t.Fatalf("expired evidence verdict=%s, want %s", snapshot.Verdict, model.VerdictTimeUntrusted)
	}
}
