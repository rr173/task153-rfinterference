package store

import (
	"context"
	"fmt"
)

func (s *Store) Initialize(ctx context.Context) error {
	statements := []string{
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE IF NOT EXISTS stations (id TEXT PRIMARY KEY, name TEXT NOT NULL, latitude REAL NOT NULL, longitude REAL NOT NULL, status TEXT NOT NULL, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS calibrations (id TEXT PRIMARY KEY, station_id TEXT NOT NULL REFERENCES stations(id), version INTEGER NOT NULL, clock_offset_millis INTEGER NOT NULL, direction_error_deg REAL NOT NULL, trusted_until TEXT NOT NULL, status TEXT NOT NULL, created_at TEXT NOT NULL, UNIQUE(station_id, version))`,
		`CREATE TABLE IF NOT EXISTS fragments (id TEXT PRIMARY KEY, station_id TEXT NOT NULL REFERENCES stations(id), sequence TEXT NOT NULL, observed_at TEXT NOT NULL, corrected_at TEXT NOT NULL, center_hz INTEGER NOT NULL, bandwidth_hz INTEGER NOT NULL, strength_dbm REAL NOT NULL, direction_deg REAL NOT NULL, status TEXT NOT NULL, event_id TEXT, exclusion_reason TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, UNIQUE(station_id, sequence))`,
		`CREATE INDEX IF NOT EXISTS fragments_event_idx ON fragments(event_id, corrected_at)`,
		`CREATE TABLE IF NOT EXISTS events (id TEXT PRIMARY KEY, status TEXT NOT NULL, start_at TEXT NOT NULL, end_at TEXT NOT NULL, min_hz INTEGER NOT NULL, max_hz INTEGER NOT NULL, frozen INTEGER NOT NULL DEFAULT 0, revision INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS events_active_idx ON events(frozen, updated_at)`,
		`CREATE TABLE IF NOT EXISTS event_fragments (event_id TEXT NOT NULL REFERENCES events(id), fragment_id TEXT NOT NULL UNIQUE REFERENCES fragments(id), PRIMARY KEY(event_id, fragment_id))`,
		`CREATE TABLE IF NOT EXISTS attribution_snapshots (id TEXT PRIMARY KEY, event_id TEXT NOT NULL REFERENCES events(id), revision INTEGER NOT NULL, verdict TEXT NOT NULL, confidence REAL NOT NULL, direction_min REAL NOT NULL, direction_max REAL NOT NULL, station_count INTEGER NOT NULL, explanation TEXT NOT NULL, created_at TEXT NOT NULL, UNIQUE(event_id, revision))`,
		`CREATE TABLE IF NOT EXISTS exclusions (id TEXT PRIMARY KEY, event_id TEXT NOT NULL REFERENCES events(id), fragment_id TEXT NOT NULL REFERENCES fragments(id), reason TEXT NOT NULL, created_at TEXT NOT NULL)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("initialize schema: %w", err)
		}
	}
	return nil
}
