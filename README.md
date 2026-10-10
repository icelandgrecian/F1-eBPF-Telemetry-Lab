# F1-eBPF-Telemetry-Lab

A Go Formula 1 telemetry consumer built with the standard library. It fetches
historical car data from [OpenF1](https://openf1.org/docs/) and prints each sample
for session `9159`, driver `55`.

Run from the project root with Go installed and internet access:

```bash
go run ./cmd/pitlane
```

Output looks like:

```text
13:08:19.923 driver=55 speed=315 km/h rpm=11141 gear=8 throttle=99 brake=0
```

The current consumer fetches and prints the historical samples once. The next
stage will turn the historical data into a replayable Go telemetry stream, then
use eBPF hooks to observe the Go process at the Linux kernel level. No eBPF is included
in this initial stage.
