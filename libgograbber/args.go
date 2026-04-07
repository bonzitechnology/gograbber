package libgograbber

import (
	"log"
	"net/http"
	"time"

	"github.com/go-rod/rod"
)

type Loggers struct {
	Good    *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Debug   *log.Logger
	Error   *log.Logger
}

type State struct {
	Canary                 string
	Cookies                string
	Expanded               bool
	Extensions             StringSet
	FollowRedirect         bool
	ScreenshotQuality      int
	ScreenshotDirectory    string
	ScreenshotFileType     string
	FollowRedirects        bool
	MaxRedirects           int
	ReportDirectory        string
	HostHeaders            StringSet
	HttpHeaders            map[string]string
	ScanOutputDirectory    string
	ProjectName            string
	DirbustOutputDirectory string
	IncludeLength          bool
	NumScreenshotWorkers   int
	HttpClient             *http.Client
	HTTPResponseDirectory  string
	Ratio                  float64
	Soft404Detection       bool
	Soft404Method          int
	PrefetchedHosts        map[string]bool
	Soft404edHosts         map[string]bool
	NoStatus               bool
	Hosts                  StringSet
	InputFile              string
	Debug                  bool
	ExcludeList            []string
	Password               string
	Ports                  IntSet
	Jitter                 int
	Sleep                  float64
	Timeout                time.Duration
	VerbosityLevel         int
	ShowIPs                bool
	Protocols              StringSet
	StatusCodes            IntSet
	IgnoreSSLErrors        bool
	StatusCodesIgn         IntSet
	ImgX                   int
	ImgY                   int
	Screenshot             bool
	Threads                int
	URLFile                string
	URLProvided            bool
	Dirbust                bool
	SingleURL              string
	Browser                *rod.Browser
	Targets                chan Host
	Paths                  StringSet
	UseSlash               bool
	Scan                   bool
	UserAgent              string
	Username               string
	OutputChan             chan string
	Verbose                bool
	Wordlist               string
	Version                string
	OutputDirectory        string
	OutputFormats          []string
	StartTime              time.Time
	ScanCounter            int64
	DirbustCounter         int64
	ScreenshotCounter      int64
	Log                    Loggers
}