# F1-eBPF-Telemetry-Lab

A Go Formula 1 telemetry consumer built with the standard library. It fetches
historical car data from [OpenF1](https://openf1.org/docs/) and prints each sample.

Run from the project root with Go installed and internet access:

```bash
go run ./cmd/pitlane
```

At startup, choose a race weekend, session (practice, qualifying, sprint, or race),
and driver from numbered menus. The app fetches sessions from `/sessions` and
participating drivers from `/drivers`; all names come from the API.
The session menu defaults to the current year. Choose a historical year with:

```bash
go run ./cmd/pitlane -year 2023
```

For scripts, supply IDs directly. Their names are still resolved through the API:

```bash
go run ./cmd/pitlane -session 9159 -driver 55
```

Omitting either ID opens its menu. `-select` opens both menus even when IDs are
provided. Use `-help` for all options. Empty API results and invalid choices produce
an error before telemetry is fetched.

Output looks like:

```text
13:08:19.923 driver=Carlos Sainz speed=315 km/h rpm=11141 gear=8 throttle=99 brake=0
```

The current consumer fetches and prints the historical samples once. The next
stage will turn the historical data into a replayable Go telemetry stream, then
use eBPF hooks to observe the Go process at the Linux kernel level. No eBPF is included
in this initial stage.
