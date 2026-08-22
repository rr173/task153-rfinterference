package metrics

import (
	"context"
	"fmt"
	"github.com/rr173/task153-rfinterference/internal/model"
	"strings"
)

type EventSource interface {
	Events(context.Context) ([]model.Event, error)
	Health(context.Context) (model.HealthReport, error)
}

func Render(ctx context.Context, source EventSource) (string, error) {
	events, err := source.Events(ctx)
	if err != nil {
		return "", err
	}
	health, err := source.Health(ctx)
	if err != nil {
		return "", err
	}
	counts := map[model.EventStatus]int{}
	for _, event := range events {
		counts[event.Status]++
	}
	lines := []string{"# HELP rf_events_total Number of interference events by status", "# TYPE rf_events_total gauge"}
	for _, status := range []model.EventStatus{model.EventObserving, model.EventConfirmed, model.EventInsufficientEvidence, model.EventArchived} {
		lines = append(lines, fmt.Sprintf("rf_events_total{status=%q} %d", status, counts[status]))
	}
	lines = append(lines, "# HELP rf_fragments_total Number of persisted scan fragments", "# TYPE rf_fragments_total gauge", fmt.Sprintf("rf_fragments_total %d", health.FragmentCount))
	return strings.Join(lines, "\n") + "\n", nil
}
