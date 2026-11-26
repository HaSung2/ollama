package cmd

import "testing"

func TestProfileFlagRegistered(t *testing.T) {
	root := NewCLI()
	runCmd, _, err := root.Find([]string{"run"})
	if err != nil {
		t.Fatalf("failed to find run command: %v", err)
	}

	flag := runCmd.Flags().Lookup("profile")
	if flag == nil {
		t.Fatal("expected --profile flag to be registered on run command")
	}
	if flag.DefValue != "false" {
		t.Errorf("expected --profile default value false, got %q", flag.DefValue)
	}
}
