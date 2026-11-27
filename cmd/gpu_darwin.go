//go:build darwin

package cmd

import (
	"context"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

// IOKit exposes a PerformanceStatistics dictionary per accelerator through
// `ioreg`. We read the overall utilization and the in-use system memory.
var (
	reGPUUtil = regexp.MustCompile(`"Device Utilization %"=(\d+)`)
	reGPUMem  = regexp.MustCompile(`"In use system memory"=(\d+)`)
)

// gpuReadTimeout bounds a single ioreg call so a slow read cannot stall the
// 0.2s sampler.
const gpuReadTimeout = 100 * time.Millisecond

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

// readGPU reads GPU utilization (percent, 0-100) and memory in use (bytes)
// from ioreg. It returns (-1, 0) on any failure (command error, timeout, or
// no statistics found). No admin rights are required.
func readGPU() (percent float64, memUsed uint64) {
	ctx, cancel := context.WithTimeout(context.Background(), gpuReadTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, "ioreg", "-r", "-d", "1", "-w", "0", "-c", "IOAccelerator").Output()
	if err != nil {
		return -1, 0
	}
	p, m, ok := parseIOReg(out)
	if !ok {
		return -1, 0
	}
	return p, m
}
