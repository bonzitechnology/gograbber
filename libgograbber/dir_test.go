package libgograbber

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
)

func TestHTTPGetter(t *testing.T) {
	InitLogger(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	// Start a local HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/found" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "found it")
		} else if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/found", http.StatusFound)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	hostAddr, portStr, _ := net.SplitHostPort(u.Host)
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	host := Host{
		HostAddr: hostAddr,
		Port:     port,
		Protocol: "http",
	}

	results := make(chan Host, 10)
	threads := make(chan struct{}, 1)
	writeChan := make(chan []byte, 10)
	var wg sync.WaitGroup

	tempDir := t.TempDir()

	// Case 1: Found
	wg.Add(1)
	threads <- struct{}{}
	go HTTPGetter(&wg, host, false, 0, false, IntSet{Set: map[int]bool{}}, 0.95, "/found", results, threads, "testProject", tempDir, writeChan, true)
	wg.Wait()

	gotHost := <-results
	if gotHost.HTTPResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", gotHost.HTTPResp.StatusCode)
	}

	// Case 2: Redirect
	wg.Add(1)
	threads <- struct{}{}
	go HTTPGetter(&wg, host, false, 0, false, IntSet{Set: map[int]bool{}}, 0.95, "/redirect", results, threads, "testProject", tempDir, writeChan, true)
	wg.Wait()

	gotHost = <-results
	if gotHost.HTTPResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK (after redirect), got %d", gotHost.HTTPResp.StatusCode)
	}
	
	// Case 3: Ignored status
	ign := IntSet{Set: map[int]bool{404: true}}
	wg.Add(1)
	threads <- struct{}{}
	go HTTPGetter(&wg, host, false, 0, false, ign, 0.95, "/notfound", results, threads, "testProject", tempDir, writeChan, true)
	wg.Wait()

	select {
	case h := <-results:
		t.Errorf("Expected no result for ignored status, got %d", h.HTTPResp.StatusCode)
	default:
		// success
	}
}

func TestSoft404Detection(t *testing.T) {
	randData := []byte("this is a random page")
	respData := []byte("this is a random page")
	
	ratio := detectSoft404(respData, randData)
	if ratio < 0.99 {
		t.Errorf("Expected ratio close to 1.0, got %f", ratio)
	}

	diffData := []byte("this is something completely different")
	ratio = detectSoft404(diffData, randData)
	if ratio > 0.6 {
		t.Errorf("Expected low ratio, got %f", ratio)
	}
}

func TestPerformSoft404Check(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "random page content")
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	hostAddr, portStr, _ := net.SplitHostPort(u.Host)
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	host := Host{
		HostAddr: hostAddr,
		Port:     port,
		Protocol: "http",
	}

	gotHost := PerformSoft404Check(host, false, "canary123")
	if gotHost.Soft404RandomURL == "" {
		t.Error("Expected Soft404RandomURL to be set")
	}
	if len(gotHost.Soft404RandomPageContents) == 0 {
		t.Error("Expected Soft404RandomPageContents to be set")
	}
}

func TestDirbRunner(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	hostAddr, portStr, _ := net.SplitHostPort(u.Host)
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	s := &State{
		Paths:      StringSet{Set: map[string]bool{"/test": true}},
		Extensions: StringSet{Set: map[string]bool{"": true}},
	}
	h := Host{
		HostAddr: hostAddr,
		Port:     port,
		Protocol: "http",
	}

	results := make(chan Host, 10)
	threads := make(chan struct{}, 10)
	writeChan := make(chan []byte, 10)
	var wg sync.WaitGroup
	wg.Add(1)

	go dirbRunner(s, h, &wg, threads, results, writeChan)
	wg.Wait()
	close(results)

	var count int
	for range results {
		count++
	}
	if count == 0 {
		t.Error("Expected at least one result from dirbRunner")
	}
}
