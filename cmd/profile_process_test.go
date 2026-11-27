package cmd

import "testing"

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
