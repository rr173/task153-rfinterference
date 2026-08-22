package store

import (
	"context"
	"github.com/rr173/task153-rfinterference/internal/model"
)

func (s *Store) SaveAttribution(ctx context.Context, a model.Attribution) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO attribution_snapshots(id,event_id,revision,verdict,confidence,direction_min,direction_max,station_count,explanation,created_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, a.ID, a.EventID, a.Revision, a.Verdict, a.Confidence, a.DirectionMin, a.DirectionMax, a.StationCount, a.Explanation, dbTime(a.CreatedAt))
	return err
}
func (s *Store) AttributionForEvent(ctx context.Context, eventID string) ([]model.Attribution, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,event_id,revision,verdict,confidence,direction_min,direction_max,station_count,explanation,created_at FROM attribution_snapshots WHERE event_id=? ORDER BY revision`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Attribution{}
	for rows.Next() {
		var a model.Attribution
		var created string
		if err := rows.Scan(&a.ID, &a.EventID, &a.Revision, &a.Verdict, &a.Confidence, &a.DirectionMin, &a.DirectionMax, &a.StationCount, &a.Explanation, &created); err != nil {
			return nil, err
		}
		a.CreatedAt = scanTime(created)
		out = append(out, a)
	}
	return out, rows.Err()
}
func (s *Store) LatestAttribution(ctx context.Context, eventID string) (model.Attribution, error) {
	var a model.Attribution
	var created string
	err := s.db.QueryRowContext(ctx, `SELECT id,event_id,revision,verdict,confidence,direction_min,direction_max,station_count,explanation,created_at FROM attribution_snapshots WHERE event_id=? ORDER BY revision DESC LIMIT 1`, eventID).Scan(&a.ID, &a.EventID, &a.Revision, &a.Verdict, &a.Confidence, &a.DirectionMin, &a.DirectionMax, &a.StationCount, &a.Explanation, &created)
	a.CreatedAt = scanTime(created)
	return a, err
}
