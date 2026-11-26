package cmd

import (
	"fmt"
	"io"
	"strings"
)

func (r ProfileResult) Print(w io.Writer) {
	header := fmt.Sprintf("-- profile (%s, %d samples) ", r.OS, r.Samples)
	if len(header) < 48 {
		header += strings.Repeat("-", 48-len(header))
	}
	fmt.Fprintln(w, header)

	if r.Samples == 0 {
		fmt.Fprintln(w, "no samples collected (response finished too quickly)")
		return
	}

	const gb = 1024 * 1024 * 1024
	fmt.Fprintf(w, "CPU      avg %.0f%%    peak %.0f%%\n", r.CPUAvg, r.CPUPeak)
	fmt.Fprintf(w, "RAM      peak %.1f GB\n", float64(r.RAMPeak)/gb)

	if r.GPUAvg < 0 {
		fmt.Fprintln(w, "GPU      N/A")
		fmt.Fprintln(w, "GPU mem  N/A")
		return
	}

	fmt.Fprintf(w, "GPU      avg %.0f%%    peak %.0f%%\n", r.GPUAvg, r.GPUPeak)
	fmt.Fprintf(w, "GPU mem  peak %.1f GB\n", float64(r.GPUMemPeak)/gb)
}
