package model

import "time"

type RegisterStationRequest struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type CreateCalibrationRequest struct {
	ClockOffsetMillis int64     `json:"clock_offset_millis"`
	DirectionErrorDeg float64   `json:"direction_error_deg"`
	TrustedUntil      time.Time `json:"trusted_until"`
	Activate          bool      `json:"activate"`
}

type FragmentInput struct {
	StationID    string    `json:"station_id"`
	Sequence     string    `json:"sequence"`
	ObservedAt   time.Time `json:"observed_at"`
	CenterHz     int64     `json:"center_hz"`
	BandwidthHz  int64     `json:"bandwidth_hz"`
	StrengthDBm  float64   `json:"strength_dbm"`
	DirectionDeg float64   `json:"direction_deg"`
}

type BatchFragmentsRequest struct {
	Fragments []FragmentInput `json:"fragments"`
}

type IngestResult struct {
	Sequence   string         `json:"sequence"`
	Status     FragmentStatus `json:"status"`
	FragmentID string         `json:"fragment_id,omitempty"`
	EventID    string         `json:"event_id,omitempty"`
	Code       string         `json:"code,omitempty"`
	Reason     string         `json:"reason,omitempty"`
}

type EventDetail struct {
	Event        Event         `json:"event"`
	Fragments    []Fragment    `json:"fragments"`
	Exclusions   []Exclusion   `json:"exclusions"`
	Attribution  []Attribution `json:"attribution"`
	RevisionNote RevisionNote  `json:"revision_note"`
}

type RevisionNote struct {
	Revision         int    `json:"revision"`
	Frozen           bool   `json:"frozen"`
	AcceptedEvidence int    `json:"accepted_evidence"`
	ExcludedEvidence int    `json:"excluded_evidence"`
	Summary          string `json:"summary"`
}

type HealthReport struct {
	OK              bool      `json:"ok"`
	EventCount      int       `json:"event_count"`
	FragmentCount   int       `json:"fragment_count"`
	RecoveredActive int       `json:"recovered_active"`
	CheckedAt       time.Time `json:"checked_at"`
}
