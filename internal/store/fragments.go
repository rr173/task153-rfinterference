package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rr173/task153-rfinterference/internal/model"
)

func (s *Store) FindFragmentByKey(ctx context.Context, stationID, sequence string) (model.Fragment, error) {
	return scanFragment(s.db.QueryRowContext(ctx, `SELECT id,station_id,sequence,observed_at,corrected_at,center_hz,bandwidth_hz,strength_dbm,direction_deg,status,COALESCE(event_id,''),exclusion_reason,created_at FROM fragments WHERE station_id=? AND sequence=?`, stationID, sequence))
}

func (s *Store) SaveFragment(ctx context.Context, fragment model.Fragment) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO fragments(id,station_id,sequence,observed_at,corrected_at,center_hz,bandwidth_hz,strength_dbm,direction_deg,status,event_id,exclusion_reason,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, fragment.ID, fragment.StationID, fragment.Sequence, dbTime(fragment.ObservedAt), dbTime(fragment.CorrectedAt), fragment.CenterHz, fragment.BandwidthHz, fragment.StrengthDBm, fragment.DirectionDeg, fragment.Status, nullString(fragment.EventID), fragment.ExclusionReason, dbTime(fragment.CreatedAt))
	return err
}

func (s *Store) UpdateFragmentAssociation(ctx context.Context, fragmentID, eventID string, status model.FragmentStatus, reason string) error {
	return s.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `UPDATE fragments SET event_id=?,status=?,exclusion_reason=? WHERE id=?`, nullString(eventID), status, reason, fragmentID); err != nil {
			return err
		}
		if eventID != "" && status == model.FragmentAccepted {
			_, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO event_fragments(event_id,fragment_id) VALUES(?,?)`, eventID, fragmentID)
			return err
		}
		return nil
	})
}

func (s *Store) FragmentsForEvent(ctx context.Context, eventID string) ([]model.Fragment, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,station_id,sequence,observed_at,corrected_at,center_hz,bandwidth_hz,strength_dbm,direction_deg,status,COALESCE(event_id,''),exclusion_reason,created_at FROM fragments WHERE event_id=? AND status=? ORDER BY corrected_at,id`, eventID, model.FragmentAccepted)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Fragment{}
	for rows.Next() {
		v, err := scanFragment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) CountFragments(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM fragments`).Scan(&n)
	return n, err
}

func sameFragment(a, b model.Fragment) bool {
	return a.ObservedAt.Equal(b.ObservedAt) && a.CenterHz == b.CenterHz && a.BandwidthHz == b.BandwidthHz && a.StrengthDBm == b.StrengthDBm && a.DirectionDeg == b.DirectionDeg
}

func isMissing(err error) bool { return errors.Is(err, sql.ErrNoRows) }

type fragmentScanner interface{ Scan(...any) error }

func scanFragment(row fragmentScanner) (model.Fragment, error) {
	var v model.Fragment
	var observed, corrected, created string
	err := row.Scan(&v.ID, &v.StationID, &v.Sequence, &observed, &corrected, &v.CenterHz, &v.BandwidthHz, &v.StrengthDBm, &v.DirectionDeg, &v.Status, &v.EventID, &v.ExclusionReason, &created)
	v.ObservedAt = scanTime(observed)
	v.CorrectedAt = scanTime(corrected)
	v.CreatedAt = scanTime(created)
	return v, err
}
func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
