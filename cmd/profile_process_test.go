package cmd

import (
	"os"
	"testing"
	"time"
)

func TestIsOllamaProcessName(t *testing.T) {
	cases := map[string]bool{
		"ollama":         true,
		"Ollama":         true,
		"OLLAMA":         true,
		"ollama.exe":     true,
		"  ollama  ":     true, // trimmed
		"ollama app.exe": false,
		"ollama-dev":     false,
		"ollama-dev.exe": false,
		"ollamacli":      false,
		"chrome":         false,
		"":               false,
	}
	for name, want := range cases {
		if got := isOllamaProcessName(name); got != want {
			t.Errorf("isOllamaProcessName(%q) = %v, want %v", name, got, want)
		}
	}
}

// TestOllamaUsageLive samples the real Ollama process tree. Skipped unless
// OLLAMA_PROFILE_LIVE=1 and a server (`ollama serve`) is running, so CI without
// Ollama is unaffected. CPU percent needs two samples to produce a delta.
func TestOllamaUsageLive(t *testing.T) {
	if os.Getenv("OLLAMA_PROFILE_LIVE") != "1" {
		t.Skip("set OLLAMA_PROFILE_LIVE=1 with ollama serve running to run this")
	}
	u := newOllamaUsage()
	u.sample() // prime CPU deltas
	time.Sleep(300 * time.Millisecond)
	cpu, ram := u.sample()
	if ram == 0 {
		t.Fatal("ram = 0; is `ollama serve` running?")
	}
	if cpu < 0 {
		t.Errorf("cpu = %v, want >= 0", cpu)
	}
	t.Logf("live Ollama usage: cpu=%.1f%% ram=%d bytes", cpu, ram)
}
