//go:build darwin

package cmd

import (
	"os"
	"testing"
)

// Trimmed real ioreg output: one IOAccelerator with a PerformanceStatistics
// dictionary.
const ioregSingle = `
    +-o IOGPU  <class AGXG15XFamilyUserClient>
      {
        "PerformanceStatistics" = {"In use system memory (driver)"=0,"Tiler Utilization %"=52,"Renderer Utilization %"=52,"Device Utilization %"=54,"Alloc system memory"=5547638784,"In use system memory"=1257127936}
        "IOGPUClassName" = "AGXG15XFamily"
      }
`

func TestParseIORegSingle(t *testing.T) {
	p, m, ok := parseIOReg([]byte(ioregSingle))
	if !ok {
		t.Fatal("ok = false, want true")
	}
	if p != 54 {
		t.Errorf("percent = %v, want 54", p)
	}
	if m != 1257127936 {
		t.Errorf("mem = %d, want 1257127936", m)
	}
}

// Two accelerators: utilization is the max, memory is the sum.
func TestParseIORegMultiple(t *testing.T) {
	in := `"Device Utilization %"=30 ... "In use system memory"=100 ...
	       "Device Utilization %"=70 ... "In use system memory"=200`
	p, m, ok := parseIOReg([]byte(in))
	if !ok {
		t.Fatal("ok = false, want true")
	}
	if p != 70 {
		t.Errorf("percent = %v, want 70 (max)", p)
	}
	if m != 300 {
		t.Errorf("mem = %d, want 300 (sum)", m)
	}
}

func TestParseIORegNone(t *testing.T) {
	_, _, ok := parseIOReg([]byte("no accelerator statistics here"))
	if ok {
		t.Error("ok = true, want false when no utilization present")
	}
}

// Live test against the real machine. Skipped unless OLLAMA_PROFILE_LIVE=1 so
// CI runners without a GPU are not affected.
func TestGPULive(t *testing.T) {
	if os.Getenv("OLLAMA_PROFILE_LIVE") != "1" {
		t.Skip("set OLLAMA_PROFILE_LIVE=1 to run the live GPU test")
	}
	p, m := readGPU()
	if p < 0 {
		t.Errorf("readGPU percent = %v, want >= 0 on a real Mac", p)
	}
	t.Logf("live GPU: util=%.0f%% mem=%d bytes", p, m)
}
