package cmd

import "strings"

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
