// Files in the same directory share a Go package, roughly like types sharing a
// C# namespace. Imports below are file-specific, similar to ordinary using directives.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/robertthorne/f1-ebpf-telemetry-lab/internal/openf1"
)

// A struct groups the selected IDs and display names, like a small C# DTO.
// Go structs have value semantics: assignment copies these fields, as with a C#
// struct. Lowercase type and field names are accessible only inside this package.
type telemetrySelection struct {
	sessionKey   int
	driverNumber int
	sessionName  string
	driverName   string
}

// Similar to a C# ISelectionSource, this describes the API methods the menu needs.
// *openf1.Client implements it implicitly by having matching methods; Go needs
// no ": ISelectionSource" declaration. Tests can pass another implementation.
// Context is similar to CancellationToken; each method synchronously returns
// a slice (roughly List<T> here) and an error, rather than Task<T>.
type SelectionSource interface {
	Sessions(ctx context.Context, year, sessionKey int) ([]openf1.Session, error)
	Drivers(ctx context.Context, sessionKey int) ([]openf1.Driver, error)
}

// A package-level function is like a C# static helper without a containing class.
// Sprintf returns formatted text, similar to string.Format or interpolation.
func sessionDisplayName(session openf1.Session) string {
	return fmt.Sprintf("%d %s — %s", session.Year, session.CircuitShortName, session.SessionName)
}

// Missing flags trigger API-backed menus. Explicit IDs are resolved to names too.
// Dependencies arrive as parameters, similar to constructor/method injection in C#.
// io.Reader and io.Writer let callers supply stdin/stderr or in-memory test buffers,
// much like using TextReader/TextWriter instead of calling Console everywhere.
// The two return values are the result and an error; the caller checks the error
// instead of catching an exception. A nil error reports success.
func parseSelection(ctx context.Context, source SelectionSource, args []string, input io.Reader, output io.Writer) (telemetrySelection, error) {
	// Like default(T) for a C# struct: ints start at 0 and strings at "".
	// This app uses 0 to mean "no ID supplied; show the selection menu."
	var selection telemetrySelection
	// := declares a local variable with an inferred type, similar to C# var.
	// FlagSet is the standard-library command-line parser. ContinueOnError makes
	// parsing return an error so main owns the decision to exit.
	flags := flag.NewFlagSet("pitlane", flag.ContinueOnError)
	flags.SetOutput(output)
	// & takes a field's address. IntVar stores parsed input through that pointer,
	// somewhat like passing a field by ref in C#.
	flags.IntVar(&selection.sessionKey, "session", 0, "OpenF1 session key (omit to choose by name)")
	flags.IntVar(&selection.driverNumber, "driver", 0, "driver number (omit to choose by name)")
	// Int and Bool allocate flag values and return *int/*bool pointers instead.
	// *year and *menu dereference them to read the values after Parse runs.
	year := flags.Int("year", time.Now().Year(), "year for the session menu")
	menu := flags.Bool("select", false, "choose session and driver by name (overrides IDs)")
	// The initializer before the semicolon scopes err to this if and its branches.
	// err != nil is Go's ordinary failure check; errors are values, not exceptions.
	if err := flags.Parse(args); err != nil {
		return selection, err
	}
	// NArg counts positional arguments left over after parsing named flags.
	// This command accepts flags only, so leftover arguments indicate a mistake.
	if flags.NArg() != 0 {
		return selection, fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	if selection.sessionKey < 0 || selection.driverNumber < 0 {
		return selection, fmt.Errorf("session key and driver number must be positive integers")
	}
	if *year < 2023 {
		return selection, fmt.Errorf("OpenF1 session data starts in 2023")
	}
	// Multiple assignment resets both IDs in one statement to force both menus.
	if *menu {
		selection.sessionKey, selection.driverNumber = 0, 0
	}
	// Scanner reads lines, similar to repeated TextReader.ReadLine calls. Reuse
	// one scanner across menus so buffered input is preserved between prompts.
	scanner := bufio.NewScanner(input)
	// With an explicit session ID, omit the year filter (the client treats 0 as
	// omitted), allowing a historical session regardless of the menu year.
	queryYear := 0
	if selection.sessionKey == 0 {
		queryYear = *year
	}
	// Like calling an injected C# interface, this dispatches to the actual source.
	// %w below wraps the error, similar in purpose to preserving InnerException.
	sessions, err := source.Sessions(ctx, queryYear, selection.sessionKey)
	if err != nil {
		return selection, fmt.Errorf("load sessions: %w", err)
	}
	if len(sessions) == 0 {
		return selection, fmt.Errorf("no sessions found for the requested year or session")
	}
	// Sort in place using a comparison function, similar to List<T>.Sort with a
	// lambda. LINQ OrderBy would instead produce a separately ordered sequence.
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].DateStart.Before(sessions[j].DateStart) })
	var session openf1.Session
	if selection.sessionKey == 0 {
		// Group by meeting so the user first chooses a weekend, then its session.
		// This is an explicit version of LINQ GroupBy(s => s.MeetingKey).
		// []int is a slice, used here like List<int>. Its zero value is nil, and
		// append works on it. Keep this list because map iteration is unordered.
		var meetings []int
		// make initializes a writable map, like new Dictionary<int, List<Session>>().
		grouped := make(map[int][]openf1.Session)
		// range is similar to foreach. _ discards the index; item is a struct copy.
		for _, item := range sessions {
			// Map lookup can return both value and presence, like TryGetValue.
			// A missing key yields a nil slice, so the append below creates its group.
			if _, exists := grouped[item.MeetingKey]; !exists {
				meetings = append(meetings, item.MeetingKey)
			}
			// append returns the updated slice; assign it back because its backing
			// array may grow. Unlike List<T>.Add, append does not update the variable.
			grouped[item.MeetingKey] = append(grouped[item.MeetingKey], item)
		}
		fmt.Fprintln(output, "Choose a race weekend:")
		// range can give both index and value. Display i+1 for human-friendly
		// choices while keeping the slices indexed from 0, like C# arrays.
		// Go dates use the reference layout "2006-01-02", rather than "yyyy-MM-dd".
		for i, key := range meetings {
			first := grouped[key][0]
			fmt.Fprintf(output, "%d) %s — %s (%s)\n", i+1, first.CountryName, first.CircuitShortName, first.DateStart.Format("2006-01-02"))
		}
		index, err := readChoice(scanner, output, "Race weekend: ", len(meetings))
		if err != nil {
			return selection, err
		}
		// Dictionary lookup retrieves only the sessions from the chosen weekend.
		options := grouped[meetings[index]]
		fmt.Fprintln(output, "Choose a session:")
		for i, item := range options {
			fmt.Fprintf(output, "%d) %s\n", i+1, item.SessionName)
		}
		// = reuses existing variables; := above declared index and err locally.
		index, err = readChoice(scanner, output, "Session: ", len(options))
		if err != nil {
			return selection, err
		}
		session = options[index]
	} else {
		// This search resembles FirstOrDefault plus a presence check. The explicit
		// bool distinguishes "not found" from a zero-valued Session struct.
		found := false
		// range is similar to foreach. _ discards the index; item is a struct copy.
		for _, item := range sessions {
			if item.SessionKey == selection.sessionKey {
				session, found = item, true
				break
			}
		}
		if !found {
			return selection, fmt.Errorf("session %d not found", selection.sessionKey)
		}
	}
	selection.sessionKey = session.SessionKey
	selection.sessionName = sessionDisplayName(session)
	// Fetch participants for this session, rather than assuming a fixed driver
	// roster. IDs remain for API calls; names are used for the displayed choices.
	drivers, err := source.Drivers(ctx, selection.sessionKey)
	if err != nil {
		return selection, fmt.Errorf("load drivers: %w", err)
	}
	if len(drivers) == 0 {
		return selection, fmt.Errorf("no drivers found for %s", selection.sessionName)
	}
	// Go compares strings lexically here; this is not a culture-aware comparison
	// like some C# string sorting overloads.
	sort.Slice(drivers, func(i, j int) bool { return drivers[i].FullName < drivers[j].FullName })
	if selection.driverNumber == 0 {
		fmt.Fprintln(output, "Choose a driver:")
		for i, driver := range drivers {
			fmt.Fprintf(output, "%d) %s — %s\n", i+1, driver.FullName, driver.TeamName)
		}
		index, err := readChoice(scanner, output, "Driver: ", len(drivers))
		if err != nil {
			return selection, err
		}
		selection.driverNumber = drivers[index].DriverNumber
		selection.driverName = drivers[index].FullName
	} else {
		for _, driver := range drivers {
			if driver.DriverNumber == selection.driverNumber {
				selection.driverName = driver.FullName
				break
			}
		}
		if selection.driverName == "" {
			return selection, fmt.Errorf("driver %d did not participate in %s", selection.driverNumber, selection.sessionName)
		}
	}
	return selection, nil
}

// readChoice is a shared input helper, similar to a C# TryParse-based prompt.
// *bufio.Scanner is a pointer so each call advances the same scanner state.
func readChoice(scanner *bufio.Scanner, output io.Writer, prompt string, count int) (int, error) {
	fmt.Fprint(output, prompt)
	// Scan returns false for both end-of-input and a read failure. Err separates
	// those cases; io.EOF is returned here when there is no next line.
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return 0, fmt.Errorf("read selection: %w", err)
		}
		return 0, fmt.Errorf("read selection: %w", io.EOF)
	}
	// TrimSpace is like string.Trim(). Atoi parses an int but returns an error
	// instead of throwing FormatException, closer to int.TryParse in its usage.
	choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || choice < 1 || choice > count {
		return 0, fmt.Errorf("choice must be between 1 and %d", count)
	}
	// Convert the displayed 1-based choice into a valid 0-based slice index.
	return choice - 1, nil
}
