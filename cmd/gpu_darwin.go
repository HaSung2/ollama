//go:build darwin

package cmd

import (
	"regexp"
	"strconv"
)

// IOKit exposes a PerformanceStatistics dictionary per accelerator through
// `ioreg`. We read the overall utilization and the in-use system memory.
var (
	reGPUUtil = regexp.MustCompile(`"Device Utilization %"=(\d+)`)
	reGPUMem  = regexp.MustCompile(`"In use system memory"=(\d+)`)
)

// parseIOReg extracts GPU utilization and memory from the output of
// `ioreg -c IOAccelerator`. When several accelerators are present, utilization
// is the maximum and memory is the sum. ok is false when no utilization value
// is found.
func parseIOReg(out []byte) (percent float64, mem uint64, ok bool) {
	utils := reGPUUtil.FindAllSubmatch(out, -1)
	if len(utils) == 0 {
		return -1, 0, false
	}

	var maxUtil float64
	for _, m := range utils {
		if v, err := strconv.ParseFloat(string(m[1]), 64); err == nil && v > maxUtil {
			maxUtil = v
		}
	}

	var memSum uint64
	for _, m := range reGPUMem.FindAllSubmatch(out, -1) {
		if v, err := strconv.ParseUint(string(m[1]), 10, 64); err == nil {
			memSum += v
		}
	}
	return maxUtil, memSum, true
}

// readGPU returns GPU utilization (percent, 0-100) and GPU memory in use
// (bytes). On failure it returns (-1, 0).
//
// The real ioreg-backed implementation is added in the next commit.
func readGPU() (percent float64, memUsed uint64) {
	return -1, 0
}
