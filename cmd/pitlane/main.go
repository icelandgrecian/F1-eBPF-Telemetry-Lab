// Package main builds an executable. Go starts it by calling main below.
package main

import (
	"context"
	"fmt"
	"log"

	// This import combines the module path from go.mod with the package directory.
	"github.com/robertthorne/f1-ebpf-telemetry-lab/internal/openf1"
)

func main() {
	// Construct one API client so its underlying HTTP client can be reused.
	// := declares a variable and lets Go infer its type from the returned value.
	client := openf1.NewClient()

	// Background returns a context with no cancellation or deadline. The HTTP
	// client's timeout still limits the request. A caller could instead supply
	// a context with a deadline or cancellation when using this package elsewhere.
	// CarData fetches historical samples for session 9159 and driver 55,
	// returning two values: a slice of samples and an error.
	data, err := client.CarData(context.Background(), 9159, 55)

	// Go handles errors explicitly: nil means the operation succeeded.
	// Fatal prints the error and exits with a nonzero status. Exiting is a
	// decision made here in main; the reusable API package only returns errors.
	if err != nil {
		log.Fatal(err)
	}

	// range visits each sample in the slice. The blank identifier (_) discards
	// the index because we only need the sample value. This prints all samples
	// immediately; replaying them at their original timing is a later step.
	for _, sample := range data {
		// Printf substitutes arguments in order: %s is a string, %d is an integer,
		// and \n ends the line. Speed is measured in kilometres per hour.
		// Go formats time using a reference-time layout: 15:04:05.000 means
		// hours:minutes:seconds with three digits of fractional seconds.
		fmt.Printf("%s driver=%d speed=%d km/h rpm=%d gear=%d throttle=%d brake=%d\n",
			sample.Date.Format("15:04:05.000"), sample.DriverNumber, sample.Speed,
			sample.RPM, sample.Gear, sample.Throttle, sample.Brake)
	}
}
