# gograbber 🕵️‍♂️

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)
![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)

> A high-performance, concurrent horizontal and vertical web content enumerator.

`gograbber` is designed for security professionals, bug bounty hunters, and pentesters who need to quickly discover and document web services across massive network ranges. It unifies **port scanning**, **directory bruteforcing**, and **automated screenshotting** into a single, blazing-fast pipeline.

---

## ✨ Features

- 🎯 **Multi-Target Support:** Seamlessly supply individual hosts, IP addresses, CIDR ranges, or a raw list of URLs.
- ⚡️ **Unified Pipeline:** Perform TCP port scanning, directory discovery, and screenshotting in one continuous execution flow. No need to chain multiple tools!
- 📸 **Modern Screenshot Engine:** Powered by `go-rod` for reliable, headless Chromium-based rendering. Automatically downloads Chromium if it isn't installed.
- 📊 **Flexible Reporting:** Generate beautiful, structured reports in **Markdown, JSON, CSV, and XML** formats simultaneously.
- 🧠 **Smart Soft-404 Detection:** Dynamically detects "wildcard" responses (e.g., sites returning 200 OK for every request) to drastically reduce false positives during directory bruteforcing.
- 🚀 **Highly Concurrent:** Built heavily on Go routines and robust worker pools for maximum speed and stability without deadlocks or race conditions.
- 🛠 **Advanced HTTP Tuning:** Full support for custom HTTP headers, cookies, user-agents, request jitter, and redirect following.

---

## 📦 Installation

`gograbber` requires Go 1.21 or higher.

### Compile from Source

```bash
# Clone the repository
git clone https://github.com/swarley7/gograbber.git

# Navigate into the directory
cd gograbber

# Build the binary
go build gograbber.go

# (Optional) Move to your PATH
mv gograbber /usr/local/bin/
```

*Note: The tool will automatically download a portable headless Chromium instance via `go-rod` upon its first screenshotting execution if no suitable browser is found on your system.*

---

## 🚀 Quick Start Examples

### 1. Easy Mode (The "Just Hack" Approach)
Automatically enables common options: port scan (top ports), directory discovery, and screenshots with optimized threading.
```bash
./gograbber -i targets.txt -w wordlist.txt -easy
```

### 2. Comprehensive Discovery & Documentation
Scan an entire `/24` subnet for common web ports (`80, 443, 8080`), bruteforce directories, and take screenshots of everything discovered. Output the results in Markdown and JSON.
```bash
./gograbber -i 10.0.0.0/24 -p 80,443,8080 -dirbust -w paths.txt -screenshot -f md,json
```

### 3. URL List Mode
If you already have a list of URLs from another tool (e.g., `httpx` or `gau`) and just want screenshots and metadata, bypassing the scanning phase:
```bash
./gograbber -U urls.txt -screenshot -f csv,md
```

### 4. Advanced Bypasses (Headers & Cookies)
Pass a custom Host header file to bypass WAFs/CDNs, supply authenticated session cookies, and add random jitter between requests to avoid rate-limiting.
```bash
./gograbber -i hosts.txt -w paths.txt -dirbust -screenshot -H vhosts.txt -C "sessionid=xyz123" -j 500
```

---

## 🎛 Command Line Arguments

### Core Execution
| Flag | Description | Default |
| :--- | :--- | :--- |
| `-i` | Input filename containing line-separated targets (hosts, IPs, CIDR ranges). | `""` |
| `-U` | Input filename containing line-separated complete URLs (overwrites `-i`, `-p`, `-scan`). | `""` |
| `-u` | Single input URL to test (overwrites `-i`, `-p`, `-scan`). | `""` |
| `-p` | Ports to test (comma separated, ranges, or presets: `top`, `small`, `med`, `large`, `full`). | `80,443` |
| `-P` | Protocols to test each host against. | `http,https` |
| `-w` | Wordlist file for directory bruteforcing. | `""` |

### Operation Modes
| Flag | Description | Default |
| :--- | :--- | :--- |
| `-scan` | Enable the TCP port scanner phase. | `false` |
| `-dirbust` | Enable the directory bruteforcing phase. | `false` |
| `-screenshot`| Enable the automated screenshotting phase. | `false` |
| `-easy` | Enables common options: `-scan -dirbust -screenshot -p top -t 2000 -j 25 -p_procs 7 -T 20`. | `false` |

### HTTP & Network Tuning
| Flag | Description | Default |
| :--- | :--- | :--- |
| `-t` | Number of concurrent threads (goroutines). | `500` |
| `-T` | Timeout in seconds for HTTP/TCP connections. | `20` |
| `-j` | Jitter: Random delay (up to X ms) introduced between requests to avoid rate limits. | `0` |
| `-H` | File containing custom Host headers to issue with each request. | `""` |
| `-headers` | Custom JSON object specifying arbitrary HTTP headers (e.g., `{"X-Forwarded-For":"127.0.0.1"}`). | `""` |
| `-C` | Custom cookies to supply with each request. | `""` |
| `-ua` | Custom User-Agent string. | `gograbber - Beta...` |
| `-e` | Comma-separated list of file extensions to append to directory brute force requests. | `""` |
| `-fr` | Follow HTTP redirects. | `true` |
| `-k` | Ignore SSL/TLS certificate validation errors. | `true` |

### Filtering & Soft-404 Detection
| Flag | Description | Default |
| :--- | :--- | :--- |
| `-s` | HTTP Status codes to ignore (comma separated). | `400,401,403,404...` |
| `-soft404` | Enable Soft-404 detection (compares page content against a random canary URL). | `true` |
| `-r` | Soft-404 comparison ratio (0.0 to 1.0). Pages with similarity above this are ignored. | `0.95` |
| `-canary` | Custom canary string to use for the Soft-404 random endpoint check. | `""` |

### Output & Reporting
| Flag | Description | Default |
| :--- | :--- | :--- |
| `-o` | Base directory to store output files. | `gograbber_output` |
| `-project` | Name of the project (prefixes output files). | `hack` |
| `-f` | Comma-separated list of output formats to generate (`md`, `json`, `csv`, `xml`). | `md` |
| `-p_procs` | Number of concurrent Chromium workers for screenshotting. | `5` |
| `-img_x` | Width of the screenshot image in pixels. | `1024` |
| `-img_y` | Height of the screenshot image in pixels. | `800` |
| `-screenshot_ext`| Image format to save screenshots as (`png`, `jpg`). | `png` |
| `-Q` | Quality percentage for JPEG screenshots. | `50` |
| `-debug` | Enable verbose debug output. | `false` |

---

## 📁 Output Structure

`gograbber` creates a highly organized directory structure inside your specified output folder (default: `gograbber_output/`):

```text
gograbber_output/
├── portscan/
│   └── hosts_hack_20260406_12345.txt        # Raw discovered IP:Port combinations
├── dirbust/
│   └── urls_hack_20260406_12345.txt         # Raw discovered URLs
├── raw_http_response/
│   └── hack_http___example_com.html         # Saved HTML bodies of discovered pages
├── screenshots/
│   └── hack_http___example_com.png          # High-fidelity visual screenshots
└── report/
    ├── hack_20260406_Report.md              # Beautiful Markdown report with inline images
    ├── hack_20260406_Report.json            # Machine-readable JSON metadata
    ├── hack_20260406_Report.csv             # Spreadsheet-ready CSV summary
    └── hack_20260406_Report.xml             # XML formatted output
```

---

## 🧑‍💻 Development & Testing

The codebase relies heavily on standard Go tooling. To run the full suite of unit and integration tests (which includes headless browser automation tests):

```bash
go test -v ./libgograbber/...
```

To run a fast test suite excluding the browser integration tests:

```bash
go test -short ./libgograbber/...
```

---

## 🙏 Acknowledgements

- **OJ Reeves:** This project's architecture borrows heavily from the design of `gobuster`.
- **michenriksen:** Inspired heavily by the screenshotting and reporting workflow of `aquatone`.
- **C_Sto:** Huge thanks for the mentorship, forcing me to learn Golang, and laughing at my extreme incompetence. Check out their awesome tool: [recursebuster](https://github.com/C-Sto/recursebuster).

---

## 📜 License

`gograbber` is licensed under the MIT License. See [LICENSE](LICENSE) for more details.

---

## 🍻 Support & Donate

If you find `gograbber` useful on your pentests or bug bounties, consider shouting a beer!

- **ETH:** `0x486b0faea72a17425ed7241e44dc9ed627f9e492`
- **BTC:** `1Jdz37JDyZYnK7tRDkF9ZW8QJ2bk2DNHzh`