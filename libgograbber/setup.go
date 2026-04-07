package libgograbber

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// Initialise sets up the program's state
func Initialise(s *State, ports string, wordlist string, statusCodesIgn string, protocols string, timeout int, AdvancedUsage bool, easy bool, HostHeaderFile string, httpHeaders string, extensions string, outputFormats string) {
	if AdvancedUsage {

		var Usage = func() {
			fmt.Print(LineSep())
			fmt.Fprintf(os.Stderr, "Advanced usage of %s:\n", os.Args[0])
			flag.PrintDefaults()
			fmt.Print(LineSep())
			fmt.Printf("Examples for %s:\n", os.Args[0])
			fmt.Printf(">> Scan and dirbust the hosts from hosts.txt.\n")
			fmt.Printf("%v -i hosts.txt -w wordlist.txt -t 2000 -scan -dirbust\n", os.Args[0])
			fmt.Printf(">> Scan and dirbust the hosts from hosts.txt, and screenshot discovered web resources.\n")
			fmt.Printf("%v -i hosts.txt -w wordlist.txt -t 2000 -scan  -dirbust -screenshot\n", os.Args[0])
			fmt.Printf(">> Scan, dirbust, and screenshot the hosts from hosts.txt on common web application ports. Additionally, set the number of screenshot workers to 3.\n")
			fmt.Printf("%v -i hosts.txt -w wordlist.txt -t 2000 -p_procs=3 -p top -scan -dirbust -screenshot\n", os.Args[0])
			fmt.Printf(">> Screenshot the URLs from urls.txt.\n")
			fmt.Printf("%v -U urls.txt -t 200 -j 400 -screenshot\n", os.Args[0])
			fmt.Printf(">> Screenshot the supplied URL.\n")
			fmt.Printf("%v -u http://example.com/test -t 200 -j 400 -screenshot\n", os.Args[0])
			fmt.Printf(">> EASY MODE/I DON'T WANT TO READ STUFF LEMME HACK OK?.\n")
			fmt.Printf("%v -i hosts.txt -w wordlist.txt -easy\n", os.Args[0])

			fmt.Print(LineSep())
		}
		Usage()
		os.Exit(0)
	}

	if easy { // user wants to use easymode... lol?
		s.Timeout = 20
		s.Jitter = 25
		s.Scan = true
		s.Dirbust = true
		s.Screenshot = true
		s.Threads = 1000
		s.NumScreenshotWorkers = 7
		ports = "top"
	}
	s.OutputFormats = strings.Split(outputFormats, ",")
	s.Extensions = StringSet{map[string]bool{}}
	for _, p := range strings.Split(extensions, ",") {
		s.Extensions.Add(p)
	}
	s.Extensions.Add("")
	s.HostHeaders = StringSet{map[string]bool{}}
	if HostHeaderFile != "" {
		hostHeaders, err := GetDataFromFile(HostHeaderFile)
		if err != nil {
			Error.Println(err)
			panic(err)
		}
		hostHeadersExpanded := ExpandHosts(hostHeaders)
		for hostHeader, _ := range hostHeadersExpanded.Set {
			s.HostHeaders.Add(hostHeader)
		}
	} else {
		s.HostHeaders.Add("")
	}
	s.HttpHeaders = map[string]string{}
	if httpHeaders != "" {
		err := json.Unmarshal([]byte(httpHeaders), &s.HttpHeaders)
		if err != nil {
			Error.Printf("Your JSON looks pretty bad eh. You should do something about that: [%v]", httpHeaders)
		}
	}

	s.Timeout = time.Duration(timeout) * time.Second

	d.Timeout = s.Timeout
	tx = &http.Transport{
		DialContext:        (d).DialContext,
		DisableCompression: true,
		MaxIdleConns:       100,

		TLSClientConfig: &tls.Config{InsecureSkipVerify: s.IgnoreSSLErrors}}
	cl = http.Client{
		Transport: tx,
		Timeout:   s.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	s.Targets = make(chan Host)
	s.ScreenshotFileType = strings.ToLower(s.ScreenshotFileType)

	s.URLProvided = false
	if s.URLFile != "" || s.SingleURL != "" {
		s.URLProvided = true
		s.Scan = false
	}
	s.Paths = StringSet{Set: map[string]bool{}}

	if wordlist != "" {
		pathData, err := GetDataFromFile(wordlist)
		if err != nil {
			Error.Println(err)
			panic(err)
		}
		for _, path := range pathData {
			s.Paths.Add(path)
		}
	} else {
		s.Paths.Add("")
	}
	s.StatusCodesIgn = IntSet{map[int]bool{}}
	for _, code := range StrArrToInt(strings.Split(statusCodesIgn, ",")) {
		s.StatusCodesIgn.Add(code)
	}

	if s.URLProvided { // A url and/or file full of urls was supplied - treat them as gospel
		go func() {
			defer func() {
				close(s.Targets)
			}()
			if s.URLFile != "" {
				inputData, err := GetDataFromFile(s.URLFile)
				if err != nil {
					Error.Println(err)
					panic(err)
				}
				for _, item := range inputData {
					ParseURLToHost(item, s.Targets)
				}
			}
			if s.SingleURL != "" {
				s.URLProvided = true
				Info.Println(s.SingleURL)
				ParseURLToHost(s.SingleURL, s.Targets)
			}
		}()
		return
	}

	if ports != "" {
		if strings.ToLower(ports) == "full" {
			ports = full
		} else if strings.ToLower(ports) == "med" {
			ports = medium
		} else if strings.ToLower(ports) == "small" {
			ports = small
		} else if strings.ToLower(ports) == "large" {
			ports = large
		} else if strings.ToLower(ports) == "top" {
			ports = top
		}
		s.Ports = UnpackPortString(ports)

	}
	if s.InputFile != "" {
		inputData, err := GetDataFromFile(s.InputFile)
		if err != nil {
			panic(err)
		}
		targetList := ExpandHosts(inputData)
		if s.Debug {
			for target := range targetList.Set {
				Debug.Printf("Target: %v\n", target)
			}
		}
		s.Hosts = targetList
	}
	s.Protocols = StringSet{map[string]bool{}}
	for _, p := range strings.Split(protocols, ",") {
		s.Protocols.Add(p)
	}

	go GenerateURLs(s.Hosts, s.Ports, &s.Paths, s.Targets)
	if !s.Dirbust && !s.Scan && !s.Screenshot && !s.URLProvided {
		flag.Usage()
		os.Exit(1)
	}
	return
}

// Start does the thing
func Start(s *State) {
	fmt.Print(LineSep())

	os.Mkdir(path.Join(s.OutputDirectory), 0755) // drwxr-xr-x
	ScanChan := make(chan Host)
	DirbChan := make(chan Host)
	ScreenshotChan := make(chan Host)
	if s.Scan {
		s.ScanOutputDirectory = path.Join(s.OutputDirectory, "portscan")
		os.Mkdir(s.ScanOutputDirectory, 0755) // drwxr-xr-x
	}
	if s.Dirbust {
		s.HTTPResponseDirectory = path.Join(s.OutputDirectory, "raw_http_response")
		os.Mkdir(s.HTTPResponseDirectory, 0755) // drwxr-xr-x
		s.DirbustOutputDirectory = path.Join(s.OutputDirectory, "dirbust")
		os.Mkdir(s.DirbustOutputDirectory, 0755) // drwxr-xr-x
	}
	if s.Screenshot {
		if s.Debug {
			Debug.Printf("Creating Chromium browser instance... This could take a second\n")
		}
		l := launcher.New().Headless(true)
		if s.IgnoreSSLErrors {
			l = l.Set("ignore-certificate-errors", "true")
		}
		u := l.MustLaunch()
		s.Browser = rod.New().ControlURL(u).MustConnect()

		s.ScreenshotDirectory = path.Join(s.OutputDirectory, "screenshots")
		os.Mkdir(s.ScreenshotDirectory, 0755) // drwxr-xr-x
		if s.Debug {
			Debug.Printf("Screenshot engine initialized.\n")
		}
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go RoutineManager(s, ScanChan, DirbChan, ScreenshotChan, &wg)

	s.ReportDirectory = path.Join(s.OutputDirectory, "report")
	os.Mkdir(s.ReportDirectory, 0755) // drwxr-xr-x
	reportFiles := Report(s, ScreenshotChan)
	wg.Wait()
	
	if s.Browser != nil {
		s.Browser.MustClose()
	}
	
	currentTime := time.Now()
	fmt.Print(LineSep())
	Info.Printf("Gograbber completed in [%v] seconds\n", g.Sprintf("%v", currentTime.Sub(s.StartTime)))
	for _, reportFile := range reportFiles {
		Info.Printf("Report written to: [%v]\n", g.Sprintf("%s", reportFile))
	}
	fmt.Print(LineSep())
}
