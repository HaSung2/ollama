package cmd

import "time"

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
