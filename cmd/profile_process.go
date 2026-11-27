package cmd

import (
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// isOllamaProcessName reports whether a process name belongs to the Ollama
// server or a model runner. Both run the "ollama" binary, so the name is
// exactly "ollama" (macOS/Linux) or "ollama.exe" (Windows). The match is
// case-insensitive and ignores surrounding whitespace.
//
// "ollama app.exe" (the Windows tray app) and "ollama-dev" / "ollama-dev.exe"
// (the CLI we build for testing) are deliberately excluded.
func isOllamaProcessName(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	return n == "ollama" || n == "ollama.exe"
}

// ollamaUsage measures the CPU and RAM of the Ollama process tree. It caches
// process handles so CPU percent is a delta between samples (gopsutil computes
// it against the previous Percent call on the same handle) rather than usage
// since process start.
type ollamaUsage struct {
	mu          sync.Mutex
	self        int32
	seen        map[int32]*process.Process
	slowRefresh bool
	lastRefresh time.Time
}

func newOllamaUsage() *ollamaUsage {
	return &ollamaUsage{self: int32(os.Getpid()), seen: map[int32]*process.Process{}}
}

// refresh rescans the process table for Ollama processes. Newly found
// processes are primed (their first CPU delta is taken on the next sample) and
// processes that have exited are dropped. The caller holds u.mu.
func (u *ollamaUsage) refresh() {
	procs, err := process.Processes()
	if err != nil {
		return
	}
	live := make(map[int32]bool)
	for _, p := range procs {
		if p.Pid == u.self {
			continue
		}
		name, err := p.Name()
		if err != nil || !isOllamaProcessName(name) {
			continue
		}
		live[p.Pid] = true
		if _, ok := u.seen[p.Pid]; !ok {
			_, _ = p.Percent(0) // prime the CPU delta baseline
			u.seen[p.Pid] = p
		}
	}
	for pid := range u.seen {
		if !live[pid] {
			delete(u.seen, pid)
		}
	}
	u.lastRefresh = time.Now()
}

// sample returns the summed CPU percent (normalised to 0-100 across all logical
// cores) and summed RSS of the Ollama processes. The process table is normally
// rescanned every call; if a sample takes longer than 100ms the rescan drops
// to once per second and cached handles are reused in between.
func (u *ollamaUsage) sample() (cpuPercent float64, ramUsed uint64) {
	start := time.Now()
	u.mu.Lock()
	defer u.mu.Unlock()

	if !u.slowRefresh || time.Since(u.lastRefresh) >= time.Second {
		u.refresh()
	}

	var cpuSum float64
	var ramSum uint64
	for pid, p := range u.seen {
		pct, err := p.Percent(0)
		if err != nil {
			delete(u.seen, pid) // gone or inaccessible
			continue
		}
		cpuSum += pct
		if mi, err := p.MemoryInfo(); err == nil {
			ramSum += mi.RSS
		}
	}

	if time.Since(start) > 100*time.Millisecond {
		u.slowRefresh = true
	}

	cores := runtime.NumCPU()
	if cores < 1 {
		cores = 1
	}
	return cpuSum / float64(cores), ramSum
}

// defaultUsage is the process-wide cache used by readSample during real runs.
var defaultUsage = newOllamaUsage()
