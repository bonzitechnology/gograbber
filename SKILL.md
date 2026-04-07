---
name: gograbber
description: Guide for executing the gograbber CLI tool for TCP port scanning, directory bruteforcing, and automated screenshotting of web targets. Use this skill when you need to run gograbber or discover web services across IPs, CIDR ranges, and URLs.
---

# Gograbber Execution Skill

This skill provides instructions for AI agents on how to construct commands and execute the `gograbber` CLI tool effectively.

## Core Directives

1. **Understand Input Types (`-i` vs `-U`)**:
   - Use `-i <file>` when you have raw IP addresses, CIDR ranges, or hostnames (e.g., `10.0.0.1`, `10.0.0.0/24`, `example.com`). These targets will be routed through the TCP port scanner first.
   - Use `-U <file>` when you have a list of fully qualified URLs (e.g., `https://example.com:8443/admin`). This completely bypasses the port scanner.
   - For single targets, use `-u <url>` (e.g., `-u http://example.com`).

2. **Always Output Structured Formats**:
   - By default, gograbber outputs Markdown. When acting as an agent, always append `-f json` or `-f csv` to generate machine-readable output that you can easily parse later.
   - Example: `-f md,json,csv`

3. **Concurrency Safety**:
   - `gograbber` is highly concurrent. The default threads are `500` (`-t 500`). For smaller networks or to avoid rate limits, lower this value (`-t 50`).
   - Limit the screenshot workers to avoid CPU/memory spikes. Default is `5` (`-p_procs 5`). Do not increase this excessively unless running on a high-resource server.

## Recommended Workflows

### 1. The "Easy" Discovery Mode
If a user just wants a general scan of a target with standard web ports (80, 443, 8080, 8443) and screenshots, use `-easy`:
```bash
./gograbber -i targets.txt -w wordlist.txt -easy -f md,json
```

### 2. Comprehensive Subnet Scanning
To discover web services across a large CIDR range, specify the ports explicitly and enable all phases:
```bash
./gograbber -i 10.0.0.0/24 -p 80,443,8000,8080,8443 -scan -dirbust -w paths.txt -screenshot -f json,csv
```

### 3. Fast Screenshotting of Existing URLs
If you already possess a list of valid URLs (e.g., output from `httpx`), skip scanning and dirbusting, and go straight to screenshots:
```bash
./gograbber -U discovered_urls.txt -screenshot -f json
```

### 4. Bypassing WAFs / Rate Limits
When running against protected targets, slow down the requests (`-j`), add custom Host headers (`-H`), and supply a session cookie (`-C`):
```bash
./gograbber -i hosts.txt -dirbust -w paths.txt -screenshot -H vhosts.txt -C "sessionid=xyz123" -j 500
```

## Parsing the Output

Once `gograbber` completes, it will write files to `gograbber_output/report/`.
To read the results programmatically, load the JSON report:
```bash
cat gograbber_output/report/hack_*_Report.json | jq .
```
Each JSON object represents a `Host` and contains:
- `protocol`, `host_addr`, `port`, `path`
- `screenshot_filename` (path to the PNG/JPG file)
- `response_body_filename` (path to the raw HTML response)