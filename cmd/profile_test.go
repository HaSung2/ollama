package cmd

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestSummarizeBasic(t *testing.T) {
	samples := []ProfileSample{
		{CPUPercent: 10, RAMUsed: 100, GPUPercent: 20, GPUMemUsed: 1000},
		{CPUPercent: 30, RAMUsed: 300, GPUPercent: 60, GPUMemUsed: 3000},
		{CPUPercent: 20, RAMUsed: 200, GPUPercent: 40, GPUMemUsed: 2000},
	}
	r := summarize("darwin", samples)

	if r.OS != "darwin" {
		t.Errorf("OS = %q, want darwin", r.OS)
	}
	if r.Samples != 3 {
		t.Errorf("Samples = %d, want 3", r.Samples)
	}
	if r.CPUAvg != 20 {
		t.Errorf("CPUAvg = %v, want 20", r.CPUAvg)
	}
	if r.CPUPeak != 30 {
		t.Errorf("CPUPeak = %v, want 30", r.CPUPeak)
	}
	if r.RAMPeak != 300 {
		t.Errorf("RAMPeak = %d, want 300", r.RAMPeak)
	}
	if r.GPUAvg != 40 {
		t.Errorf("GPUAvg = %v, want 40", r.GPUAvg)
	}
	if r.GPUPeak != 60 {
		t.Errorf("GPUPeak = %v, want 60", r.GPUPeak)
	}
	if r.GPUMemPeak != 3000 {
		t.Errorf("GPUMemPeak = %d, want 3000", r.GPUMemPeak)
	}
}

// Samples where only some have a valid GPU reading: CPU uses all samples,
// GPU uses only the valid ones.
func TestSummarizeGPUPartial(t *testing.T) {
	samples := []ProfileSample{
		{CPUPercent: 10, RAMUsed: 100, GPUPercent: -1},
		{CPUPercent: 20, RAMUsed: 200, GPUPercent: 50, GPUMemUsed: 500},
		{CPUPercent: 30, RAMUsed: 300, GPUPercent: -1},
	}
	r := summarize("linux", samples)

	if r.CPUAvg != 20 {
		t.Errorf("CPUAvg = %v, want 20", r.CPUAvg)
	}
	if r.GPUAvg != 50 {
		t.Errorf("GPUAvg = %v, want 50 (only valid sample)", r.GPUAvg)
	}
	if r.GPUPeak != 50 {
		t.Errorf("GPUPeak = %v, want 50", r.GPUPeak)
	}
	if r.GPUMemPeak != 500 {
		t.Errorf("GPUMemPeak = %d, want 500", r.GPUMemPeak)
	}
}

func TestSummarizeNoValidGPU(t *testing.T) {
	samples := []ProfileSample{
		{CPUPercent: 10, GPUPercent: -1},
		{CPUPercent: 20, GPUPercent: -1},
	}
	r := summarize("linux", samples)

	if r.GPUAvg != -1 {
		t.Errorf("GPUAvg = %v, want -1", r.GPUAvg)
	}
	if r.GPUPeak != -1 {
		t.Errorf("GPUPeak = %v, want -1", r.GPUPeak)
	}
	if r.GPUMemPeak != 0 {
		t.Errorf("GPUMemPeak = %d, want 0", r.GPUMemPeak)
	}
}

func TestSummarizeEmpty(t *testing.T) {
	r := summarize("windows", nil)

	if r.Samples != 0 {
		t.Errorf("Samples = %d, want 0", r.Samples)
	}
	if r.CPUAvg != 0 || r.CPUPeak != 0 || r.RAMPeak != 0 {
		t.Errorf("empty result should have zero CPU/RAM, got %+v", r)
	}
	if r.GPUAvg != -1 || r.GPUPeak != -1 {
		t.Errorf("empty result GPU should be -1, got avg=%v peak=%v", r.GPUAvg, r.GPUPeak)
	}
}

// Sampler collects whatever the injected read function returns and summarizes
// it on Stop.
func TestSamplerCollectsInjected(t *testing.T) {
	s := newSampler(5 * time.Millisecond)
	var n int64
	s.read = func() ProfileSample {
		i := atomic.AddInt64(&n, 1)
		return ProfileSample{CPUPercent: float64(i * 10), RAMUsed: uint64(i * 100), GPUPercent: -1}
	}

	s.Start(context.Background())
	time.Sleep(60 * time.Millisecond) // expect several ticks
	r := s.Stop()

	if r.Samples == 0 {
		t.Fatal("no samples collected")
	}
	if r.CPUPeak <= 0 || r.RAMPeak == 0 {
		t.Errorf("expected positive CPU/RAM peaks, got %+v", r)
	}
	if r.GPUAvg != -1 {
		t.Errorf("GPUAvg = %v, want -1 (injected GPU is -1)", r.GPUAvg)
	}
}

// Stop can be called more than once without panicking or blocking.
func TestSamplerStopTwice(t *testing.T) {
	s := newSampler(5 * time.Millisecond)
	s.read = func() ProfileSample { return ProfileSample{GPUPercent: -1} }

	s.Start(context.Background())
	time.Sleep(20 * time.Millisecond)
	first := s.Stop()
	second := s.Stop()

	if first.Samples != second.Samples {
		t.Errorf("Stop twice gave different sample counts: %d vs %d", first.Samples, second.Samples)
	}
}

// After the context is cancelled the goroutine stops on its own, and Stop
// still returns without hanging.
func TestSamplerContextCancel(t *testing.T) {
	s := newSampler(5 * time.Millisecond)
	s.read = func() ProfileSample { return ProfileSample{GPUPercent: -1} }

	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)
	time.Sleep(20 * time.Millisecond)
	cancel()

	done := make(chan struct{})
	go func() { s.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop hung after context cancel")
	}
}
