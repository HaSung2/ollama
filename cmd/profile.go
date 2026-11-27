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
