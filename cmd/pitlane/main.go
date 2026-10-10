// Package main builds an executable. Go starts it by calling main below.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	// This import combines the module path from go.mod with the package directory.
	"github.com/robertthorne/f1-ebpf-telemetry-lab/internal/openf1"
)

func main() {
	// Construct one API client so its underlying HTTP client can be reused.
	// := declares a variable and lets Go infer its type from the returned value.
	client := openf1.NewClient()
	ctx := context.Background()
	selection, err := parseSelection(ctx, client, os.Args[1:], os.Stdin, os.Stderr)
	if err != nil {
		if err == flag.ErrHelp {
			return
		}
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "Loading %s\n", selection.sessionName+" — "+selection.driverName)

	// client has type *openf1.Client. Its CarData method has a pointer receiver
	// and matches the interface, so this assignment compiles automatically.
	// source holds the concrete client, but exposes only the interface's methods.
	var source TelemetrySource = client

	// Background returns a context with no cancellation or deadline. The HTTP
	// client's timeout still limits the request. A caller could instead supply
	// a context with a deadline or cancellation when using this package elsewhere.
	// Pass the interface to a function that only needs TelemetrySource behavior.
	if err := printTelemetry(ctx, source, selection); err != nil {
		// main decides to print the returned error and exit with a nonzero status.
		log.Fatal(err)
	}
}
