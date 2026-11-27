//go:build darwin

package cmd

// readGPU returns GPU utilization (percent, 0-100) and GPU memory in use
// (bytes). On failure it returns (-1, 0) and must return within 100ms.
//
// H-1 placeholder: always unavailable. H-2 implements this for macOS by
// parsing `ioreg -r -d 1 -w 0 -c IOAccelerator`.
func readGPU() (percent float64, memUsed uint64) {
	return -1, 0
}
