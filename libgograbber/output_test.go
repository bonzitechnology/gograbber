package libgograbber

import (
	"bufio"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestBuildResponseHeader(t *testing.T) {
	resp := &http.Response{
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Status:     "200 OK",
		StatusCode: 200,
		Header:     make(http.Header),
	}
	resp.Header.Set("X-Test", "Value")
	got := buildResponseHeader(resp)
	if !strings.Contains(got, "HTTP/1.1 200 OK") || !strings.Contains(got, "X-Test: Value") {
		t.Errorf("buildResponseHeader got %q", got)
	}
}

func TestReport(t *testing.T) {
	loggers := InitLogger(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	tempDir := t.TempDir()
	s := &State{
		Log:             loggers,
		ProjectName:     "TestProject",
		ReportDirectory: tempDir,
		OutputFormats:   []string{"md", "json", "csv", "xml"},
	}
	targets := make(chan Host, 1)
	host := Host{
		Protocol: "http",
		HostAddr: "example.com",
		Port:     80,
		Path:     "test",
	}
	targets <- host
	close(targets)

	reportFiles, err := Report(s, targets)
	if err != nil {
		t.Errorf("Report failed: %v", err)
	}
	if len(reportFiles) != 4 {
		t.Errorf("Expected 4 report files, got %d", len(reportFiles))
	}
	for _, f := range reportFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("Report file %s was not created", f)
		}
	}
}

func TestWriterWorker(t *testing.T) {
	loggers := InitLogger(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	tmpfile, err := os.CreateTemp("", "testwriter")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	writeChan := make(chan []byte, 10)
	go writerWorker(loggers, writeChan, tmpfile.Name())

	writeChan <- []byte("test data\n")
	writeChan <- []byte("more data\n")
	close(writeChan)

	// Wait a bit for the worker to finish
	time.Sleep(100 * time.Millisecond)

	file, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) != 2 || lines[0] != "test data" || lines[1] != "more data" {
		t.Errorf("writerWorker output got %v, expected [test data, more data]", lines)
	}
}
