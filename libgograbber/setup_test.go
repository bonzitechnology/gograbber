package libgograbber

import (
	"testing"
	"time"
)

func TestInitialise(t *testing.T) {
	s := &State{Scan: true}
	Initialise(s, "80,443", "", "404", "http", 10, false, false, "", "", "", "md")

	if !s.Protocols.Contains("http") {
		t.Error("Expected protocols to contain http")
	}
	if !s.Ports.Contains(80) || !s.Ports.Contains(443) {
		t.Error("Expected ports to contain 80 and 443")
	}
	if s.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", s.Timeout)
	}
}

func TestInitialiseEasy(t *testing.T) {
	s := &State{}
	Initialise(s, "", "", "404", "", 0, false, true, "", "", "", "md")

	if !s.Scan || !s.Dirbust || !s.Screenshot {
		t.Error("Easy mode should enable Scan, Dirbust, and Screenshot")
	}
	if s.Threads != 1000 {
		t.Errorf("Easy mode threads should be 1000, got %d", s.Threads)
	}
}
