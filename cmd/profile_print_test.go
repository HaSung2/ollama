package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestProfilePrint(t *testing.T) {
	tests := []struct {
		name string
		res  ProfileResult
		want string
	}{
		{
			name: "normal",
			res: ProfileResult{
				OS:         "darwin",
				Samples:    41,
				CPUAvg:     38,
				CPUPeak:    71,
				RAMPeak:    uint64(112 * 1024 * 1024 * 1024 / 10),
				GPUAvg:     92,
				GPUPeak:    100,
				GPUMemPeak: uint64(26 * 1024 * 1024 * 1024 / 10),
			},
			want: strings.Join([]string{
				"-- profile (darwin, 41 samples) ----------------",
				"CPU      avg 38%    peak 71%",
				"RAM      peak 11.2 GB",
				"GPU      avg 92%    peak 100%",
				"GPU mem  peak 2.6 GB",
			}, "\n") + "\n",
		},
		{
			name: "gpu_na",
			res: ProfileResult{
				OS:      "windows",
				Samples: 12,
				CPUAvg:  25,
				CPUPeak: 40,
				RAMPeak: 2 * 1024 * 1024 * 1024,
				GPUAvg:  -1,
				GPUPeak: -1,
			},
			want: strings.Join([]string{
				"-- profile (windows, 12 samples) ---------------",
				"CPU      avg 25%    peak 40%",
				"RAM      peak 2.0 GB",
				"GPU      N/A",
				"GPU mem  N/A",
			}, "\n") + "\n",
		},
		{
			name: "zero_samples",
			res: ProfileResult{
				OS:      "windows",
				Samples: 0,
				GPUAvg:  -1,
				GPUPeak: -1,
			},
			want: strings.Join([]string{
				"-- profile (windows, 0 samples) ----------------",
				"no samples collected (response finished too quickly)",
			}, "\n") + "\n",
		},
		{
			name: "zero_ram",
			res: ProfileResult{
				OS:      "linux",
				Samples: 5,
				CPUAvg:  10,
				CPUPeak: 15,
				RAMPeak: 0,
				GPUAvg:  -1,
				GPUPeak: -1,
			},
			want: strings.Join([]string{
				"-- profile (linux, 5 samples) ------------------",
				"CPU      avg 10%    peak 15%",
				"RAM      peak 0.0 GB",
				"GPU      N/A",
				"GPU mem  N/A",
			}, "\n") + "\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			tc.res.Print(&buf)
			if got := buf.String(); got != tc.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tc.want)
			}
			firstLine := strings.Split(buf.String(), "\n")[0]
			if len(firstLine) != 48 {
				t.Errorf("first line length = %d, want 48 (%q)", len(firstLine), firstLine)
			}
		})
	}
}
