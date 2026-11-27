//go:build !darwin && !windows

package cmd

// readGPU is unavailable on platforms other than macOS and Windows and
// always returns (-1, 0).
func readGPU() (percent float64, memUsed uint64) {
	return -1, 0
}
