package cmd

import (
	"math"
	"testing"
)

const edgeGB = uint64(1024 * 1024 * 1024)

// summarizeCase is one table-driven case for summarize. want holds the full
// expected ProfileResult so every field is checked, not only the one the case
// is about.
type summarizeCase struct {
	name    string
	samples []ProfileSample
	want    ProfileResult
}

func checkSummarize(t *testing.T, tc summarizeCase) {
	t.Helper()
	got := summarize("test", tc.samples)
	tc.want.OS = "test"
	if got.Samples != tc.want.Samples {
		t.Errorf("Samples = %d, want %d", got.Samples, tc.want.Samples)
	}
	floats := []struct {
		field     string
		got, want float64
	}{
		{"CPUAvg", got.CPUAvg, tc.want.CPUAvg},
		{"CPUPeak", got.CPUPeak, tc.want.CPUPeak},
		{"GPUAvg", got.GPUAvg, tc.want.GPUAvg},
		{"GPUPeak", got.GPUPeak, tc.want.GPUPeak},
	}
	for _, f := range floats {
		if math.Abs(f.got-f.want) > 1e-9 {
			t.Errorf("%s = %v, want %v", f.field, f.got, f.want)
		}
	}
	if got.RAMPeak != tc.want.RAMPeak {
		t.Errorf("RAMPeak = %d, want %d", got.RAMPeak, tc.want.RAMPeak)
	}
	if got.GPUMemPeak != tc.want.GPUMemPeak {
		t.Errorf("GPUMemPeak = %d, want %d", got.GPUMemPeak, tc.want.GPUMemPeak)
	}
	if got.OS != tc.want.OS {
		t.Errorf("OS = %q, want %q", got.OS, tc.want.OS)
	}
}

// TestSummarizeEdgeCases covers sample counts and CPU/RAM values at the edges
// of the 7.2 contract.
func TestSummarizeEdgeCases(t *testing.T) {
	cases := []summarizeCase{
		{
			name:    "nil samples",
			samples: nil,
			want:    ProfileResult{Samples: 0, GPUAvg: -1, GPUPeak: -1},
		},
		{
			name:    "one sample",
			samples: []ProfileSample{{CPUPercent: 37.5, RAMUsed: 3 * edgeGB, GPUPercent: 81, GPUMemUsed: 2 * edgeGB}},
			want: ProfileResult{Samples: 1, CPUAvg: 37.5, CPUPeak: 37.5, RAMPeak: 3 * edgeGB,
				GPUAvg: 81, GPUPeak: 81, GPUMemPeak: 2 * edgeGB},
		},
		{
			name: "CPU fixed at 0 percent",
			samples: []ProfileSample{
				{CPUPercent: 0, RAMUsed: edgeGB, GPUPercent: -1},
				{CPUPercent: 0, RAMUsed: edgeGB, GPUPercent: -1},
				{CPUPercent: 0, RAMUsed: edgeGB, GPUPercent: -1},
			},
			want: ProfileResult{Samples: 3, CPUAvg: 0, CPUPeak: 0, RAMPeak: edgeGB, GPUAvg: -1, GPUPeak: -1},
		},
		{
			name: "CPU at 100 percent",
			samples: []ProfileSample{
				{CPUPercent: 100, GPUPercent: -1},
				{CPUPercent: 100, GPUPercent: -1},
			},
			want: ProfileResult{Samples: 2, CPUAvg: 100, CPUPeak: 100, GPUAvg: -1, GPUPeak: -1},
		},
		{
			name: "RAM 64 GB",
			samples: []ProfileSample{
				{CPUPercent: 10, RAMUsed: 63 * edgeGB, GPUPercent: -1},
				{CPUPercent: 20, RAMUsed: 64 * edgeGB, GPUPercent: -1},
				{CPUPercent: 30, RAMUsed: 60 * edgeGB, GPUPercent: -1},
			},
			want: ProfileResult{Samples: 3, CPUAvg: 20, CPUPeak: 30, RAMPeak: 64 * edgeGB, GPUAvg: -1, GPUPeak: -1},
		},
		{
			name: "RAM 0 bytes",
			samples: []ProfileSample{
				{CPUPercent: 5, RAMUsed: 0, GPUPercent: -1},
			},
			want: ProfileResult{Samples: 1, CPUAvg: 5, CPUPeak: 5, RAMPeak: 0, GPUAvg: -1, GPUPeak: -1},
		},
		{
			name: "peak in the first sample",
			samples: []ProfileSample{
				{CPUPercent: 90, RAMUsed: 8 * edgeGB, GPUPercent: -1},
				{CPUPercent: 10, RAMUsed: 2 * edgeGB, GPUPercent: -1},
				{CPUPercent: 20, RAMUsed: 4 * edgeGB, GPUPercent: -1},
			},
			want: ProfileResult{Samples: 3, CPUAvg: 40, CPUPeak: 90, RAMPeak: 8 * edgeGB, GPUAvg: -1, GPUPeak: -1},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { checkSummarize(t, tc) })
	}
}

// TestSummarizeManySamples checks a long response: 5 minutes at 0.2 s is 1500
// samples, and the average must not drift.
func TestSummarizeManySamples(t *testing.T) {
	samples := make([]ProfileSample, 1500)
	for i := range samples {
		cpu := 20.0
		if i%2 == 1 {
			cpu = 40.0
		}
		samples[i] = ProfileSample{CPUPercent: cpu, RAMUsed: uint64(i), GPUPercent: -1}
	}
	checkSummarize(t, summarizeCase{
		samples: samples,
		want:    ProfileResult{Samples: 1500, CPUAvg: 30, CPUPeak: 40, RAMPeak: 1499, GPUAvg: -1, GPUPeak: -1},
	})
}

// TestSummarizeGPUMixed checks that samples with GPUPercent -1 are left out of
// the GPU average and peak, while CPU and RAM still use every sample.
func TestSummarizeGPUMixed(t *testing.T) {
	cases := []summarizeCase{
		{
			name: "all GPU -1",
			samples: []ProfileSample{
				{CPUPercent: 10, RAMUsed: edgeGB, GPUPercent: -1},
				{CPUPercent: 30, RAMUsed: 2 * edgeGB, GPUPercent: -1},
			},
			want: ProfileResult{Samples: 2, CPUAvg: 20, CPUPeak: 30, RAMPeak: 2 * edgeGB,
				GPUAvg: -1, GPUPeak: -1, GPUMemPeak: 0},
		},
		{
			name: "one GPU -1 among valid samples",
			samples: []ProfileSample{
				{CPUPercent: 10, GPUPercent: 60, GPUMemUsed: edgeGB},
				{CPUPercent: 20, GPUPercent: -1},
				{CPUPercent: 30, GPUPercent: 90, GPUMemUsed: 2 * edgeGB},
			},
			// GPU average is (60+90)/2, not (60-1+90)/3 or (60+0+90)/3.
			want: ProfileResult{Samples: 3, CPUAvg: 20, CPUPeak: 30,
				GPUAvg: 75, GPUPeak: 90, GPUMemPeak: 2 * edgeGB},
		},
		{
			name: "GPU -1 first, then valid",
			samples: []ProfileSample{
				{CPUPercent: 10, GPUPercent: -1},
				{CPUPercent: 10, GPUPercent: -1},
				{CPUPercent: 10, GPUPercent: 40, GPUMemUsed: edgeGB},
			},
			want: ProfileResult{Samples: 3, CPUAvg: 10, CPUPeak: 10,
				GPUAvg: 40, GPUPeak: 40, GPUMemPeak: edgeGB},
		},
		{
			name: "valid first, then GPU -1",
			samples: []ProfileSample{
				{CPUPercent: 10, GPUPercent: 40, GPUMemUsed: edgeGB},
				{CPUPercent: 10, GPUPercent: -1},
			},
			want: ProfileResult{Samples: 2, CPUAvg: 10, CPUPeak: 10,
				GPUAvg: 40, GPUPeak: 40, GPUMemPeak: edgeGB},
		},
		{
			// 0% is a valid reading (GPUPercent >= 0), so the result is 0, not N/A.
			name: "GPU idle at 0 percent is valid",
			samples: []ProfileSample{
				{CPUPercent: 5, GPUPercent: 0},
				{CPUPercent: 5, GPUPercent: 0},
			},
			want: ProfileResult{Samples: 2, CPUAvg: 5, CPUPeak: 5, GPUAvg: 0, GPUPeak: 0},
		},
		{
			name: "single sample with GPU -1",
			samples: []ProfileSample{
				{CPUPercent: 50, RAMUsed: edgeGB, GPUPercent: -1},
			},
			want: ProfileResult{Samples: 1, CPUAvg: 50, CPUPeak: 50, RAMPeak: edgeGB, GPUAvg: -1, GPUPeak: -1},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { checkSummarize(t, tc) })
	}
}
