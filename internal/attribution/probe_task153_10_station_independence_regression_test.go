package attribution

import (
	"strings"
	"testing"
	"time"

	"task153-rfinterference/internal/model"
)

func TestBug10_SingleStationEvidenceCannotBePresentedAsLocatable(t *testing.T) {
	now := time.Date(2026, 8, 22, 13, 0, 0, 0, time.UTC)
	computer := New()
	computer.SetClock(func() time.Time { return now })
	fragment := model.Fragment{StationID: "site-a", ObservedAt: now, DirectionDeg: 18, StrengthDBm: -35}
	snapshot := computer.Compute(model.Event{ID: "one-site", Revision: 1}, []model.Fragment{fragment}, map[string]model.Calibration{"site-a": {TrustedUntil: now.Add(time.Hour)}})
	if snapshot.Verdict != model.VerdictStationInsufficient {
		t.Fatalf("single-station verdict=%s, want %s", snapshot.Verdict, model.VerdictStationInsufficient)
	}
	if got := BuildConfidence([]model.Fragment{fragment}, 1, 0).StationFactor; got != 1.0/3.0 {
		t.Fatalf("single-station factor=%v, want %v", got, 1.0/3.0)
	}
	if explanation := ExplainEvent(model.Event{ID: "one-site"}, snapshot); !strings.Contains(explanation, "more independent stations") {
		t.Fatalf("single-station explanation=%q", explanation)
	}
}
