package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/robertthorne/f1-ebpf-telemetry-lab/internal/openf1"
)

type selectionStub struct {
	sessions                           []openf1.Session
	drivers                            []openf1.Driver
	err                                error
	year, sessionFilter, driverSession int
}

func (s *selectionStub) Sessions(_ context.Context, year, key int) ([]openf1.Session, error) {
	s.year, s.sessionFilter = year, key
	return s.sessions, s.err
}
func (s *selectionStub) Drivers(_ context.Context, key int) ([]openf1.Driver, error) {
	s.driverSession = key
	return s.drivers, nil
}
func fixture() *selectionStub {
	return &selectionStub{
		sessions: []openf1.Session{
			{SessionKey: 101, MeetingKey: 10, Year: 2024, CircuitShortName: "Test Circuit", CountryName: "Test Country", SessionName: "Practice 1", DateStart: time.Date(2024, 5, 1, 10, 0, 0, 0, time.UTC)},
			{SessionKey: 102, MeetingKey: 10, Year: 2024, CircuitShortName: "Test Circuit", CountryName: "Test Country", SessionName: "Race", DateStart: time.Date(2024, 5, 3, 10, 0, 0, 0, time.UTC)},
		},
		drivers: []openf1.Driver{{DriverNumber: 99, FullName: "Zed Driver", TeamName: "Team Z"}, {DriverNumber: 7, FullName: "Alex Driver", TeamName: "Team A"}},
	}
}
func TestSelectionMenusUseAPINamesAndIDs(t *testing.T) {
	source := fixture()
	var output bytes.Buffer
	got, err := parseSelection(context.Background(), source, []string{"-year", "2024"}, strings.NewReader("1\n2\n1\n"), &output)
	if err != nil {
		t.Fatal(err)
	}
	if got.sessionKey != 102 || got.driverNumber != 7 || got.driverName != "Alex Driver" || got.sessionName != "2024 Test Circuit — Race" {
		t.Fatalf("selection: %+v", got)
	}
	if source.year != 2024 || source.sessionFilter != 0 || source.driverSession != 102 {
		t.Fatalf("wrong API filters: %+v", source)
	}
	for _, name := range []string{"Test Country", "Race", "Alex Driver", "Team A"} {
		if !strings.Contains(output.String(), name) {
			t.Fatalf("menu missing %q: %s", name, output.String())
		}
	}
}
func TestExplicitSelectionResolvesNames(t *testing.T) {
	source := fixture()
	got, err := parseSelection(context.Background(), source, []string{"-session", "102", "-driver", "99"}, strings.NewReader(""), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if got.driverName != "Zed Driver" || source.year != 0 || source.sessionFilter != 102 || source.driverSession != 102 {
		t.Fatalf("selection=%+v source=%+v", got, source)
	}
}
func TestSelectionErrors(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		input  string
		source *selectionStub
	}{
		{"bad choice", nil, "3\n", fixture()},
		{"closed input", nil, "", fixture()},
		{"invalid ID", []string{"-driver", "-1"}, "", fixture()},
		{"unknown driver", []string{"-session", "102", "-driver", "8"}, "", fixture()},
		{"unknown session", []string{"-session", "103"}, "", fixture()},
		{"no sessions", nil, "", &selectionStub{}},
		{"API failure", nil, "", &selectionStub{err: errors.New("offline")}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseSelection(context.Background(), tc.source, tc.args, strings.NewReader(tc.input), &bytes.Buffer{})
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
