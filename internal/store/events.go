package store

import (
	"context"
	"database/sql"
	"errors"

	"task153-rfinterference/internal/model"
)

func (s *Store) SaveEvent(ctx context.Context, event model.Event) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO events(id,status,start_at,end_at,min_hz,max_hz,frozen,revision,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, event.ID, event.Status, dbTime(event.StartAt), dbTime(event.EndAt), event.MinHz, event.MaxHz, boolInt(event.Frozen), event.Revision, dbTime(event.CreatedAt), dbTime(event.UpdatedAt))
	return err
}
func (s *Store) UpdateEvent(ctx context.Context, event model.Event) error {
	_, err := s.db.ExecContext(ctx, `UPDATE events SET status=?,start_at=?,end_at=?,min_hz=?,max_hz=?,frozen=?,revision=?,updated_at=? WHERE id=?`, event.Status, dbTime(event.StartAt), dbTime(event.EndAt), event.MinHz, event.MaxHz, boolInt(event.Frozen), event.Revision, dbTime(event.UpdatedAt), event.ID)
	return err
}
func (s *Store) GetEvent(ctx context.Context, id string) (model.Event, error) {
	return scanEvent(s.db.QueryRowContext(ctx, `SELECT id,status,start_at,end_at,min_hz,max_hz,frozen,revision,created_at,updated_at FROM events WHERE id=?`, id))
}
func (s *Store) ListEvents(ctx context.Context) ([]model.Event, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,status,start_at,end_at,min_hz,max_hz,frozen,revision,created_at,updated_at FROM events ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Event{}
	for rows.Next() {
		v, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) ActiveEvents(ctx context.Context) ([]model.Event, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,status,start_at,end_at,min_hz,max_hz,frozen,revision,created_at,updated_at FROM events WHERE frozen=0 ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Event{}
	for rows.Next() {
		v, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) ArchivedEvents(ctx context.Context) ([]model.Event, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,status,start_at,end_at,min_hz,max_hz,frozen,revision,created_at,updated_at FROM events WHERE frozen=1 ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Event{}
	for rows.Next() {
		v, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) CountEvents(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM events`).Scan(&n)
	return n, err
}
func (s *Store) SaveExclusion(ctx context.Context, e model.Exclusion) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO exclusions(id,event_id,fragment_id,reason,created_at) VALUES(?,?,?,?,?)`, e.ID, e.EventID, e.FragmentID, e.Reason, dbTime(e.CreatedAt))
	return err
}
func (s *Store) ExclusionsForEvent(ctx context.Context, eventID string) ([]model.Exclusion, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,event_id,fragment_id,reason,created_at FROM exclusions WHERE event_id=? ORDER BY created_at`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Exclusion{}
	for rows.Next() {
		var v model.Exclusion
		var created string
		if err := rows.Scan(&v.ID, &v.EventID, &v.FragmentID, &v.Reason, &created); err != nil {
			return nil, err
		}
		v.CreatedAt = scanTime(created)
		out = append(out, v)
	}
	return out, rows.Err()
}

type eventScanner interface{ Scan(...any) error }

func scanEvent(row eventScanner) (model.Event, error) {
	var v model.Event
	var start, end, created, updated string
	var frozen int
	err := row.Scan(&v.ID, &v.Status, &start, &end, &v.MinHz, &v.MaxHz, &frozen, &v.Revision, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return v, model.NewError(model.CodeNotFound, "event does not exist")
	}
	v.StartAt = scanTime(start)
	v.EndAt = scanTime(end)
	v.CreatedAt = scanTime(created)
	v.UpdatedAt = scanTime(updated)
	v.Frozen = frozen != 0
	return v, err
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
