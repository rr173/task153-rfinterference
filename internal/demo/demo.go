package demo

import (
	"context"
	"task153-rfinterference/internal/model"
	"task153-rfinterference/internal/service"
	"time"
)

func Import(ctx context.Context, svc *service.Service) ([]model.IngestResult, error) {
	now := svc.Now().Add(-time.Minute)
	for _, station := range []struct {
		id, name string
		offset   int64
		dir      float64
	}{{"north", "North ridge", 120, 22}, {"east", "East hill", 0, 28}, {"west", "West valley", -80, 34}} {
		_, err := svc.RegisterStation(ctx, model.RegisterStationRequest{ID: station.id, Name: station.name, Latitude: 30.1, Longitude: 120.1})
		if err != nil {
			return nil, err
		}
		_, err = svc.CreateCalibration(ctx, station.id, model.CreateCalibrationRequest{ClockOffsetMillis: station.offset, DirectionErrorDeg: 4, TrustedUntil: now.Add(24 * time.Hour), Activate: true})
		if err != nil {
			return nil, err
		}
	}
	return svc.IngestBatch(ctx, model.BatchFragmentsRequest{Fragments: []model.FragmentInput{{StationID: "north", Sequence: "demo-1", ObservedAt: now.Add(120 * time.Millisecond), CenterHz: 433920000, BandwidthHz: 12000, StrengthDBm: -54, DirectionDeg: 22}, {StationID: "east", Sequence: "demo-1", ObservedAt: now, CenterHz: 433922000, BandwidthHz: 10000, StrengthDBm: -57, DirectionDeg: 28}, {StationID: "west", Sequence: "demo-1", ObservedAt: now.Add(-80 * time.Millisecond), CenterHz: 433919000, BandwidthHz: 11000, StrengthDBm: -56, DirectionDeg: 34}}})
}
func SelfCheck(ctx context.Context, svc *service.Service) (model.HealthReport, error) {
	report, err := svc.Health(ctx)
	if err != nil {
		return report, err
	}
	if report.EventCount < 1 || report.FragmentCount < 3 {
		report.OK = false
	}
	return report, nil
}
