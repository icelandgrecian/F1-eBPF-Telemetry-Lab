package main

import (
	"context"
	"fmt"
	"log"

	"github.com/robertthorne/f1-ebpf-telemetry-lab/internal/openf1"
)

func main() {
	client := openf1.NewClient()
	data, err := client.CarData(context.Background(), 9159, 55)
	if err != nil {
		x
		log.Fatal(err)
	}

	for _, sample := range data {
		fmt.Printf("%s driver=%d speed=%d km/h rpm=%d gear=%d throttle=%d brake=%d\n",
			sample.Date.Format("15:04:05.000"), sample.DriverNumber, sample.Speed,
			sample.RPM, sample.Gear, sample.Throttle, sample.Brake)
	}
}
