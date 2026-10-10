// Unlike a C# namespace, a Go package is organized by directory. Both this file
// and main.go declare package main, so they share types and functions without
// importing each other. Imports apply only to the file that declares them.
package main

import (
	"context"
	"fmt"

	"github.com/robertthorne/f1-ebpf-telemetry-lab/internal/openf1"
)

// TelemetrySource plays the role of a C# ITelemetrySource interface.
// Go usually omits the I prefix, and does not require one file per type.
//
// In C#, you would declare: class OpenF1Client : ITelemetrySource.
// In Go, *openf1.Client satisfies this interface automatically because it has
// the exact CarData method below. There is no inheritance declaration to add.
// The method uses a pointer receiver, so *openf1.Client satisfies the interface;
// an openf1.Client value does not.
//
// The consumer defines the interface here to describe what it needs. The HTTP
// client package does not need to reference or know about this interface.
// This is a learning example; one concrete source does not require an interface.
type TelemetrySource interface {
	// context.Context carries cancellation/deadlines, similar to a C#
	// CancellationToken, and can also carry request-scoped values.
	// []openf1.CarData is a slice: a sequence of samples, roughly comparable
	// to List<CarData> for this use. The method returns two separate values:
	// the data and an error. This is a synchronous call, not a Task or async method.
	CarData(ctx context.Context, sessionKey int, driverNumber int) ([]openf1.CarData, error)
}

// printTelemetry receives its dependency as a parameter, much like a C# method
// accepting ITelemetrySource. No dependency injection container is needed.
// This is a package-level function; Go does not require a containing class.
// Its lowercase name makes it accessible only within this package. Uppercase
// names, such as TelemetrySource, are exported (subject to internal package rules).
// A future file source with the same method could be passed in unchanged.
func printTelemetry(ctx context.Context, source TelemetrySource, selection telemetrySelection) error {
	// Like calling through a C# interface reference, this invokes the method
	// on the underlying concrete client. := declares data and err with inferred
	// types; two variables receive the method's two return values.
	data, err := source.CarData(ctx, selection.sessionKey, selection.driverNumber)

	// Instead of C# try/catch, ordinary Go failures are returned as error values.
	// nil means no error. %w wraps the original error, loosely like preserving
	// InnerException in C#; errors.Is/As can inspect the wrapped error chain.
	if err != nil {
		return fmt.Errorf("load telemetry: %w", err)
	}

	// This is similar to C# foreach (var sample in data). range produces an
	// index and a value; _ discards the index. sample is a copy of each struct,
	// unlike a C# reference-type object. We only read it here.
	// Samples print immediately; timed replay is a later step.
	for _, sample := range data {
		driverName := selection.driverName
		// Similar to Console.WriteLine formatting: %s inserts a string and %d
		// inserts an integer. Printf needs an explicit \n to end the line.
		// Instead of C# "HH:mm:ss.fff", Go uses "15:04:05.000": the reference
		// time's digits describe hours, minutes, seconds, and milliseconds.
		fmt.Printf("%s driver=%s speed=%d km/h rpm=%d gear=%d throttle=%d brake=%d\n",
			sample.Date.Format("15:04:05.000"), driverName, sample.Speed,
			sample.RPM, sample.Gear, sample.Throttle, sample.Brake)
	}
	// Returning nil reports success to main, which owns the decision to exit.
	return nil
}
