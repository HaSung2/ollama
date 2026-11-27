//go:build windows

package cmd

// readGPU returns GPU utilization (percent, 0-100) and GPU memory in use
// (bytes). On failure it returns (-1, 0) and must return within 100ms.
//
// H-1 placeholder: always unavailable. H-4 (optional) implements this for
// Windows via nvidia-smi or performance counters.
func readGPU() (percent float64, memUsed uint64) {
	return -1, 0
}
