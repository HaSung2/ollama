package cmd

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// ProfileSample is one hardware-usage reading taken every sample interval.
type ProfileSample struct {
	At         time.Time
	CPUPercent float64 // Ollama process CPU sum / logical cores (0-100)
	RAMUsed    uint64  // Ollama process RSS sum (bytes)
	GPUPercent float64 // GPU utilization (%), -1 if unreadable
	GPUMemUsed uint64  // GPU memory (bytes), 0 if unreadable
}

// ProfileResult summarizes every sample collected for one response.
type ProfileResult struct {
	OS         string
	Samples    int
	CPUAvg     float64
	CPUPeak    float64
	RAMPeak    uint64
	GPUAvg     float64 // -1 if no valid GPU sample
	GPUPeak    float64 // -1 if no valid GPU sample
	GPUMemPeak uint64
}

// defaultSampleInterval is how often the sampler reads usage while a response
// is being generated.
const defaultSampleInterval = 200 * time.Millisecond

// summarize reduces raw samples to averages and peaks.
//
// CPU avg/peak use every sample. GPU avg/peak use only samples whose
// GPUPercent is >= 0, and are -1 when no sample has a valid GPU reading.
// With zero samples, Samples is 0, numeric fields are 0, and GPU avg/peak
// are -1.
func summarize(osName string, samples []ProfileSample) ProfileResult {
	r := ProfileResult{OS: osName, Samples: len(samples), GPUAvg: -1, GPUPeak: -1}
	if len(samples) == 0 {
		return r
	}

	var cpuSum, gpuSum float64
	var gpuCount int
	for _, s := range samples {
		cpuSum += s.CPUPercent
		if s.CPUPercent > r.CPUPeak {
			r.CPUPeak = s.CPUPercent
		}
		if s.RAMUsed > r.RAMPeak {
			r.RAMPeak = s.RAMUsed
		}
		if s.GPUPercent >= 0 {
			gpuSum += s.GPUPercent
			gpuCount++
			if s.GPUPercent > r.GPUPeak {
				r.GPUPeak = s.GPUPercent
			}
			if s.GPUMemUsed > r.GPUMemPeak {
				r.GPUMemPeak = s.GPUMemUsed
			}
		}
	}

	r.CPUAvg = cpuSum / float64(len(samples))
	if gpuCount > 0 {
		r.GPUAvg = gpuSum / float64(gpuCount)
	}
	return r
}

// Sampler reads hardware usage on a fixed interval in a background goroutine.
// Create one with newSampler, call Start before the request, and Stop after
// the final response; Stop returns the summarized result.
//
// The read field lets tests inject a fake measurement function. All access to
// samples is guarded by mu so Start and Stop are race-free.
type Sampler struct {
	interval time.Duration
	read     func() ProfileSample

	mu      sync.Mutex
	samples []ProfileSample

	stop     chan struct{}
	stopOnce sync.Once
	done     chan struct{}
}

// newSampler returns a Sampler that reads real usage every interval. An
// interval of 0 or less uses defaultSampleInterval.
func newSampler(interval time.Duration) *Sampler {
	if interval <= 0 {
		interval = defaultSampleInterval
	}
	return &Sampler{
		interval: interval,
		read:     readSample,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// Start begins sampling in a goroutine until Stop is called or ctx is
// cancelled.
func (s *Sampler) Start(ctx context.Context) {
	go func() {
		defer close(s.done)
		t := time.NewTicker(s.interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stop:
				return
			case <-t.C:
				smp := s.read()
				s.mu.Lock()
				s.samples = append(s.samples, smp)
				s.mu.Unlock()
			}
		}
	}()
}

// Stop ends sampling, waits for the goroutine to finish, and returns the
// summary. It is safe to call Stop more than once and safe to call after the
// context passed to Start was cancelled.
func (s *Sampler) Stop() ProfileResult {
	s.stopOnce.Do(func() { close(s.stop) })
	<-s.done
	s.mu.Lock()
	samples := s.samples
	s.mu.Unlock()
	return summarize(runtime.GOOS, samples)
}

// readSample takes one reading across CPU, RAM and GPU.
//
// CPU and RAM come from the Ollama process tree only (server + runner) via
// defaultUsage, so other applications do not affect them. GPU comes from the
// OS-specific readGPU (macOS via ioreg; others return -1, 0).
func readSample() ProfileSample {
	cpuPct, ram := defaultUsage.sample()
	gpuPct, gpuMem := readGPU()
	return ProfileSample{
		At:         time.Now(),
		CPUPercent: cpuPct,
		RAMUsed:    ram,
		GPUPercent: gpuPct,
		GPUMemUsed: gpuMem,
	}
}
