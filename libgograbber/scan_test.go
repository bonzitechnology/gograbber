package libgograbber

import (
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

func TestConnectHost(t *testing.T) {
	InitLogger(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	// Start a local listener
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}
	defer ln.Close()

	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	host := Host{
		HostAddr: "127.0.0.1",
		Port:     port,
	}

	results := make(chan Host, 1)
	threads := make(chan struct{}, 1)
	threads <- struct{}{}
	writeChan := make(chan []byte, 10)
	var wg sync.WaitGroup
	wg.Add(1)

	go ConnectHost(&wg, time.Second, 0, false, host, results, threads, writeChan)

	wg.Wait()
	close(results)

	gotHost, ok := <-results
	if !ok {
		t.Error("Expected host in results, but channel was closed")
	}
	if gotHost.Port != port {
		t.Errorf("Expected port %d, got %d", port, gotHost.Port)
	}

	// Test a closed port
	host.Port = 1 // Likely closed
	results = make(chan Host, 1)
	threads <- struct{}{}
	wg.Add(1)
	go ConnectHost(&wg, 100*time.Millisecond, 0, false, host, results, threads, writeChan)
	wg.Wait()
	close(results)
	_, ok = <-results
	if ok {
		t.Error("Expected no host in results for closed port")
	}
}
