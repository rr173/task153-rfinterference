package metrics

import (
	"context"
	"task153-rfinterference/internal/model"
	"strings"
	"testing"
	"time"
)

type fake struct{}

func (fake) Events(context.Context) ([]model.Event, error) {
	return []model.Event{{Status: model.EventConfirmed}}, nil
}
func (fake) Health(context.Context) (model.HealthReport, error) {
	return model.HealthReport{FragmentCount: 2, CheckedAt: time.Now()}, nil
}
func TestRender(t *testing.T) {
	out, err := Render(context.Background(), fake{})
	if err != nil || !strings.Contains(out, "rf_events_total") {
		t.Fatal(out, err)
	}
}
