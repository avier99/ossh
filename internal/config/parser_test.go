package config_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/kevinburke/ssh_config"
)

func TestRoundTrip(t *testing.T) {
	original, err := os.ReadFile("testdata/sample.ssh_config")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := ssh_config.Decode(bytes.NewReader(original))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	got := cfg.String()

	if got != string(original) {
		t.Errorf("round-trip failed — library does not preserve the file faithfully\n\n--- original (%d bytes) ---\n%s\n\n--- got (%d bytes) ---\n%s",
			len(original), original, len(got), got)
	}
}
