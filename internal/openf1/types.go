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

// Session describes one practice, qualifying, sprint, or race session.
type Session struct {
	SessionKey       int       `json:"session_key"`
	MeetingKey       int       `json:"meeting_key"`
	SessionName      string    `json:"session_name"`
	CircuitShortName string    `json:"circuit_short_name"`
	CountryName      string    `json:"country_name"`
	Year             int       `json:"year"`
	DateStart        time.Time `json:"date_start"`
}

// Driver contains the session-specific name and team for a driver number.
type Driver struct {
	DriverNumber int    `json:"driver_number"`
	FullName     string `json:"full_name"`
	TeamName     string `json:"team_name"`
}
