# Gograbber Agent Skills & Architecture Guide

This file provides technical context, architectural patterns, and guidelines for AI agents, LLMs, or new developers interacting with the `gograbber` codebase.

## 1. Core Pipeline & Concurrency Model
`gograbber` operates on a strict, unidirectional channel-based pipeline:
`Targets -> ScanChan -> DirbustChan -> ScreenshotChan -> Report`

- **Bounded Worker Pools:** We use buffered channels (e.g., `screenshotWorkers := make(chan struct{}, s.NumScreenshotWorkers)`) and `sync.WaitGroup` to throttle concurrency. *Never spawn unbounded goroutines inside the pipeline loops,* as this will cause OOM crashes (especially with Chromium).
- **Pass-By-Value:** The `Host` struct is passed **by value** through the channels. This is a critical security feature to prevent race conditions. When a worker mutates a `Host` (e.g., updating the path or response body), it only mutates its local copy.
- **Read-Only State:** The global `State` struct (defined in `libgograbber/args.go`) holds all configuration. Slices (like Wordlists) and Maps (like Headers) must be treated as **100% read-only** after `Initialise()` completes.

## 2. Headless Browser & Screenshotting (`go-rod`)
- **Network Hijacking (Crucial):** `gograbber` does NOT make secondary network requests for the primary HTML document during the screenshot phase. In `libgograbber/screenshot.go`, we use `go-rod`'s `HijackRequests()` router to intercept Chromium's navigation request and instantly serve it the HTML body/headers that were already saved to disk during the `Dirbust` phase. 
- *Agent Rule:* If modifying the screenshot engine, ensure the hijack router remains intact to preserve speed and avoid breaking authenticated session states.

## 3. Soft-404 Heuristics
- **Byte-Length Comparison:** Soft-404 detection (`libgograbber/dir.go`) compares the raw byte length of the target response against a randomized canary endpoint. It does *not* use `O(N^2)` sequence matching (like `difflib`). 
- *Agent Rule:* Do not revert back to string/word sequence matching, as large wildcard error pages will bottleneck the dirbust workers.

## 4. Building and Testing
- **Build:** `go build gograbber.go`
- **Unit Tests (Fast):** `go test ./libgograbber/... -short` (Skips `go-rod` browser integration tests).
- **Integration Tests (Slow):** `go test -v ./libgograbber/...` (Requires network connectivity and will automatically download a Chromium binary if missing).

## 5. Input Parsing (-i vs -U)
- **`-i` (Input Hosts):** Expects raw IP addresses, CIDR ranges, or hostnames. These targets are passed through the TCP port scanner (`ScanChan`) and multiplied against the provided `-p` (ports) and `-P` (protocols) to generate thousands of combinations.
- **`-U` (Input URLs):** Expects fully qualified URLs (e.g., `https://example.com:8443/admin`). Targets passed here **bypass the TCP scanner entirely** and are injected directly into the `DirbustChan`. 

## 6. Output & Reporting
- All output is flushed to a single directory (default: `gograbber_output/`).
- The `Report()` function in `libgograbber/output.go` aggregates the final results from the `ScreenshotChan` in a single thread to avoid race conditions, then generates Markdown, JSON, CSV, and XML files simultaneously.