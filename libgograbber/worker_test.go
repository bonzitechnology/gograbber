package libgograbber

import (
	"sync"
	"testing"
)

func TestRoutineManagerNoOps(t *testing.T) {
	s := &State{
		Threads: 1,
		Targets: make(chan Host),
	}
	close(s.Targets)

	scanChan := make(chan Host)
	dirbChan := make(chan Host)
	screenChan := make(chan Host)
	var wg sync.WaitGroup
	wg.Add(1)

	go RoutineManager(s, scanChan, dirbChan, screenChan, &wg)

	// Consume screenChan until closed
	var count int
	for range screenChan {
		count++
	}
	wg.Wait()

	if count != 0 {
		t.Errorf("Expected 0 results, got %d", count)
	}
}
