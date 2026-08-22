package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/rr173/task153-rfinterference/internal/model"
)

func (s *Store) SaveStation(ctx context.Context, station model.Station) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO stations(id,name,latitude,longitude,status,created_at) VALUES(?,?,?,?,?,?)`, station.ID, station.Name, station.Latitude, station.Longitude, station.Status, dbTime(station.CreatedAt))
	return err
}

func (s *Store) GetStation(ctx context.Context, id string) (model.Station, error) {
	var station model.Station
	var created string
	err := s.db.QueryRowContext(ctx, `SELECT id,name,latitude,longitude,status,created_at FROM stations WHERE id=?`, id).Scan(&station.ID, &station.Name, &station.Latitude, &station.Longitude, &station.Status, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return station, model.NewError(model.CodeUnknownStation, "station %s does not exist", id)
	}
	station.CreatedAt = scanTime(created)
	return station, err
}

func (s *Store) ListStations(ctx context.Context) ([]model.Station, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,name,latitude,longitude,status,created_at FROM stations ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.Station{}
	for rows.Next() {
		var v model.Station
		var created string
		if err := rows.Scan(&v.ID, &v.Name, &v.Latitude, &v.Longitude, &v.Status, &created); err != nil {
			return nil, err
		}
		v.CreatedAt = scanTime(created)
		result = append(result, v)
	}
	return result, rows.Err()
}

func (s *Store) SaveCalibration(ctx context.Context, calibration model.Calibration) error {
	return s.WithTx(ctx, func(tx *sql.Tx) error {
		if calibration.Status == model.CalibrationActive {
			if _, err := tx.ExecContext(ctx, `UPDATE calibrations SET status=? WHERE station_id=? AND status=?`, model.CalibrationRetired, calibration.StationID, model.CalibrationActive); err != nil {
				return err
			}
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO calibrations(id,station_id,version,clock_offset_millis,direction_error_deg,trusted_until,status,created_at) VALUES(?,?,?,?,?,?,?,?)`, calibration.ID, calibration.StationID, calibration.Version, calibration.ClockOffsetMillis, calibration.DirectionErrorDeg, dbTime(calibration.TrustedUntil), calibration.Status, dbTime(calibration.CreatedAt))
		return err
	})
}

func (s *Store) NextCalibrationVersion(ctx context.Context, stationID string) (int, error) {
	var v int
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version),0)+1 FROM calibrations WHERE station_id=?`, stationID).Scan(&v)
	return v, err
}

func (s *Store) ActiveCalibration(ctx context.Context, stationID string) (model.Calibration, error) {
	var c model.Calibration
	var trusted, created string
	err := s.db.QueryRowContext(ctx, `SELECT id,station_id,version,clock_offset_millis,direction_error_deg,trusted_until,status,created_at FROM calibrations WHERE station_id=? AND status=? ORDER BY version DESC LIMIT 1`, stationID, model.CalibrationActive).Scan(&c.ID, &c.StationID, &c.Version, &c.ClockOffsetMillis, &c.DirectionErrorDeg, &trusted, &c.Status, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return c, model.NewError(model.CodeNotFound, "station %s has no active calibration", stationID)
	}
	c.TrustedUntil = scanTime(trusted)
	c.CreatedAt = scanTime(created)
	return c, err
}

func (s *Store) CalibrationForStation(ctx context.Context, stationID string) ([]model.Calibration, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,station_id,version,clock_offset_millis,direction_error_deg,trusted_until,status,created_at FROM calibrations WHERE station_id=? ORDER BY version`, stationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Calibration{}
	for rows.Next() {
		var c model.Calibration
		var trusted, created string
		if err := rows.Scan(&c.ID, &c.StationID, &c.Version, &c.ClockOffsetMillis, &c.DirectionErrorDeg, &trusted, &c.Status, &created); err != nil {
			return nil, fmt.Errorf("scan calibration: %w", err)
		}
		c.TrustedUntil = scanTime(trusted)
		c.CreatedAt = scanTime(created)
		out = append(out, c)
	}
	return out, rows.Err()
}
