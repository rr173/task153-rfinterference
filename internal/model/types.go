package model

import "time"

type StationStatus string

const (
	StationEnabled     StationStatus = "enabled"
	StationCalibrating StationStatus = "calibrating"
	StationDisabled    StationStatus = "disabled"
)

type CalibrationStatus string

const (
	CalibrationDraft   CalibrationStatus = "draft"
	CalibrationActive  CalibrationStatus = "active"
	CalibrationRetired CalibrationStatus = "retired"
)

type FragmentStatus string

const (
	FragmentNew       FragmentStatus = "new"
	FragmentAccepted  FragmentStatus = "accepted"
	FragmentDuplicate FragmentStatus = "duplicate"
	FragmentInvalid   FragmentStatus = "invalid"
	FragmentExcluded  FragmentStatus = "excluded"
)

type EventStatus string

const (
	EventObserving            EventStatus = "observing"
	EventConfirmed            EventStatus = "confirmed"
	EventInsufficientEvidence EventStatus = "insufficient_evidence"
	EventArchived             EventStatus = "archived"
)

type AttributionVerdict string

const (
	VerdictNotCalculated       AttributionVerdict = "not_calculated"
	VerdictLocatable           AttributionVerdict = "locatable"
	VerdictDirectionConflict   AttributionVerdict = "direction_conflict"
	VerdictStationInsufficient AttributionVerdict = "station_insufficient"
	VerdictTimeUntrusted       AttributionVerdict = "time_untrusted"
)

type Station struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Latitude  float64       `json:"latitude"`
	Longitude float64       `json:"longitude"`
	Status    StationStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
}

type Calibration struct {
	ID                string            `json:"id"`
	StationID         string            `json:"station_id"`
	Version           int               `json:"version"`
	ClockOffsetMillis int64             `json:"clock_offset_millis"`
	DirectionErrorDeg float64           `json:"direction_error_deg"`
	TrustedUntil      time.Time         `json:"trusted_until"`
	Status            CalibrationStatus `json:"status"`
	CreatedAt         time.Time         `json:"created_at"`
}

type Fragment struct {
	ID              string         `json:"id"`
	StationID       string         `json:"station_id"`
	Sequence        string         `json:"sequence"`
	ObservedAt      time.Time      `json:"observed_at"`
	CorrectedAt     time.Time      `json:"corrected_at"`
	CenterHz        int64          `json:"center_hz"`
	BandwidthHz     int64          `json:"bandwidth_hz"`
	StrengthDBm     float64        `json:"strength_dbm"`
	DirectionDeg    float64        `json:"direction_deg"`
	Status          FragmentStatus `json:"status"`
	EventID         string         `json:"event_id,omitempty"`
	ExclusionReason string         `json:"exclusion_reason,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
}

type Event struct {
	ID        string      `json:"id"`
	Status    EventStatus `json:"status"`
	StartAt   time.Time   `json:"start_at"`
	EndAt     time.Time   `json:"end_at"`
	MinHz     int64       `json:"min_hz"`
	MaxHz     int64       `json:"max_hz"`
	Frozen    bool        `json:"frozen"`
	Revision  int         `json:"revision"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type Attribution struct {
	ID           string             `json:"id"`
	EventID      string             `json:"event_id"`
	Revision     int                `json:"revision"`
	Verdict      AttributionVerdict `json:"verdict"`
	Confidence   float64            `json:"confidence"`
	DirectionMin float64            `json:"direction_min"`
	DirectionMax float64            `json:"direction_max"`
	StationCount int                `json:"station_count"`
	Explanation  string             `json:"explanation"`
	CreatedAt    time.Time          `json:"created_at"`
}

type Exclusion struct {
	ID         string    `json:"id"`
	EventID    string    `json:"event_id"`
	FragmentID string    `json:"fragment_id"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}
