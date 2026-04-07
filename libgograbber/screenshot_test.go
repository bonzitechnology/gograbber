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
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func TestScreenshotAURL(t *testing.T) {
	InitLogger(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	if testing.Short() {
		t.Skip("skipping screenshot test in short mode")
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "<html><body><h1>Hello Test</h1></body></html>")
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
		Path:     "",
	}

	l := launcher.New().Headless(true)
	browserURL, err := l.Launch()
	if err != nil {
		t.Skipf("Failed to launch browser: %v", err)
	}
	browser := rod.New().ControlURL(browserURL).MustConnect()
	defer browser.Close()

	s := &State{
		Browser:              browser,
		ScreenshotDirectory:  t.TempDir(),
		ImgX:                 1024,
		ImgY:                 800,
		ScreenshotFileType:   "png",
		ScreenshotQuality:    50,
		Timeout:              time.Second * 10,
	}

	results := make(chan Host, 1)
	threads := make(chan struct{}, 1)
	threads <- struct{}{}
	var wg sync.WaitGroup
	wg.Add(1)

	err = ScreenshotAURL(&wg, s, 0, host, results, threads)
	if err != nil {
		t.Errorf("ScreenshotAURL failed: %v", err)
	}

	gotHost, ok := <-results
	if !ok {
		t.Error("Expected host in results")
	}
	if gotHost.ScreenshotFilename == "" {
		t.Error("Expected screenshot filename to be set")
	}
	if _, err := os.Stat(gotHost.ScreenshotFilename); os.IsNotExist(err) {
		t.Errorf("Screenshot file does not exist: %s", gotHost.ScreenshotFilename)
	}
}
