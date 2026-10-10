package openf1

import "time"

// CarData is a single OpenF1 car telemetry sample.
type CarData struct {
	Brake        int       `json:"brake"`
	Date         time.Time `json:"date"`
	DriverNumber int       `json:"driver_number"`
	DRS          int       `json:"drs"`
	Gear         int       `json:"n_gear"`
	RPM          int       `json:"rpm"`
	Speed        int       `json:"speed"`
	Throttle     int       `json:"throttle"`
	SessionKey   int       `json:"session_key"`
	MeetingKey   int       `json:"meeting_key"`
}
