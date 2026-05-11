package speedtest

import "time"

type TestStep struct {
	Direction string
	Bytes     uint64
	Count     int
}

type TestPlan struct {
	Name            string
	LatencyCount    int
	Steps           []TestStep
	Timeout         time.Duration
	LatencyTimeout  time.Duration
	DownloadTimeout time.Duration
	UploadTimeout   time.Duration
	IncludeRaw      bool
}

var QuickPlan = TestPlan{
	Name:            "quick",
	LatencyCount:    5,
	LatencyTimeout:  5 * time.Second,
	DownloadTimeout: 10 * time.Second,
	Steps: []TestStep{
		{Direction: "download", Bytes: 100_000, Count: 5},
		{Direction: "download", Bytes: 1_000_000, Count: 3},
	},
	Timeout: 15 * time.Second,
}

var DefaultPlan = TestPlan{
	Name:            "default",
	LatencyCount:    20,
	LatencyTimeout:  15 * time.Second,
	DownloadTimeout: 30 * time.Second,
	UploadTimeout:   30 * time.Second,
	Steps: []TestStep{
		{Direction: "download", Bytes: 100_000, Count: 10},
		{Direction: "upload", Bytes: 100_000, Count: 10},
		{Direction: "download", Bytes: 1_000_000, Count: 8},
		{Direction: "upload", Bytes: 1_000_000, Count: 8},
		{Direction: "download", Bytes: 10_000_000, Count: 6},
		{Direction: "upload", Bytes: 10_000_000, Count: 6},
		{Direction: "download", Bytes: 25_000_000, Count: 4},
		{Direction: "upload", Bytes: 25_000_000, Count: 4},
	},
	Timeout: 90 * time.Second,
}

var ThoroughPlan = TestPlan{
	Name:            "thorough",
	LatencyCount:    40,
	LatencyTimeout:  30 * time.Second,
	DownloadTimeout: 90 * time.Second,
	UploadTimeout:   90 * time.Second,
	Steps: []TestStep{
		{Direction: "download", Bytes: 100_000, Count: 10},
		{Direction: "upload", Bytes: 100_000, Count: 10},
		{Direction: "download", Bytes: 1_000_000, Count: 8},
		{Direction: "upload", Bytes: 1_000_000, Count: 8},
		{Direction: "download", Bytes: 10_000_000, Count: 6},
		{Direction: "upload", Bytes: 10_000_000, Count: 6},
		{Direction: "download", Bytes: 25_000_000, Count: 6},
		{Direction: "upload", Bytes: 25_000_000, Count: 6},
		{Direction: "download", Bytes: 100_000_000, Count: 4},
		{Direction: "upload", Bytes: 50_000_000, Count: 4},
	},
	Timeout:    240 * time.Second,
	IncludeRaw: true,
}

func (p TestPlan) DownloadSteps() []TestStep {
	var steps []TestStep
	for _, s := range p.Steps {
		if s.Direction == "download" {
			steps = append(steps, s)
		}
	}
	return steps
}

func (p TestPlan) UploadSteps() []TestStep {
	var steps []TestStep
	for _, s := range p.Steps {
		if s.Direction == "upload" {
			steps = append(steps, s)
		}
	}
	return steps
}
