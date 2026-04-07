package libgograbber

import (
	"bufio"
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
)

// Get yo colours sorted
var g = color.New(color.FgGreen, color.Bold)
var y = color.New(color.FgYellow, color.Bold)
var r = color.New(color.FgRed, color.Bold)
var m = color.New(color.FgMagenta, color.Bold)
var b = color.New(color.FgBlue, color.Bold)

var full = "0-65535"
var large = "1,3-4,6-7,9,13,17,19-26,30,32-33,37,42-43,49,53,70,79-85,88-90,99-100,106,109-111,113,119,125,135,139,143-144,146,161,163,179,199,211-212,222,254-256,259,264,280,301,306,311,340,366,389,406-407,416-417,425,427,443-445,458,464-465,481,497,500,512-515,524,541,543-545,548,554-555,563,587,593,616-617,625,631,636,646,648,666-668,683,687,691,700,705,711,714,720,722,726,749,765,777,783,787,800-801,808,843,873,880,888,898,900-903,911-912,981,987,990,992-993,995,999-1002,1007,1009-1011,1021-1100,1102,1104-1108,1110-1114,1117,1119,1121-1124,1126,1130-1132,1137-1138,1141,1145,1147-1149,1151-1152,1154,1163-1166,1169,1174-1175,1183,1185-1187,1192,1198-1199,1201,1213,1216-1218,1233-1234,1236,1244,1247-1248,1259,1271-1272,1277,1287,1296,1300-1301,1309-1311,1322,1328,1334,1352,1417,1433-1434,1443,1455,1461,1494,1500-1501,1503,1521,1524,1533,1556,1580,1583,1594,1600,1641,1658,1666,1687-1688,1700,1717-1721,1723,1755,1761,1782-1783,1801,1805,1812,1839-1840,1862-1864,1875,1900,1914,1935,1947,1971-1972,1974,1984,1998-2010,2013,2020-2022,2030,2033-2035,2038,2040-2043,2045-2049,2065,2068,2099-2100,2103,2105-2107,2111,2119,2121,2126,2135,2144,2160-2161,2170,2179,2190-2191,2196,2200,2222,2251,2260,2288,2301,2323,2366,2381-2383,2393-2394,2399,2401,2492,2500,2522,2525,2557,2601-2602,2604-2605,2607-2608,2638,2701-2702,2710,2717-2718,2725,2800,2809,2811,2869,2875,2909-2910,2920,2967-2968,2998,3000-3001,3003,3005-3007,3011,3013,3017,3030-3031,3052,3071,3077,3128,3168,3211,3221,3260-3261,3268-3269,3283,3300-3301,3306,3322-3325,3333,3351,3367,3369-3372,3389-3390,3404,3476,3493,3517,3527,3546,3551,3580,3659,3689-3690,3703,3737,3766,3784,3800-3801,3809,3814,3826-3828,3851,3869,3871,3878,3880,3889,3905,3914,3918,3920,3945,3971,3986,3995,3998,4000-4006,4045,4111,4125-4126,4129,4224,4242,4279,4321,4343,4443-4446,4449,4550,4567,4662,4848,4899-4900,4998,5000-5004,5009,5030,5033,5050-5051,5054,5060-5061,5080,5087,5100-5102,5120,5190,5200,5214,5221-5222,5225-5226,5269,5280,5298,5357,5405,5414,5431-5432,5440,5500,5510,5544,5550,5555,5560,5566,5631,5633,5666,5678-5679,5718,5730,5800-5802,5810-5811,5815,5822,5825,5850,5859,5862,5877,5900-5904,5906-5907,5910-5911,5915,5922,5925,5950,5952,5959-5963,5987-5989,5998-6007,6009,6025,6059,6100-6101,6106,6112,6123,6129,6156,6346,6389,6502,6510,6543,6547,6565-6567,6580,6646,6666-6669,6689,6692,6699,6779,6788-6789,6792,6839,6881,6901,6969,7000-7002,7004,7007,7019,7025,7070,7100,7103,7106,7200-7201,7402,7435,7443,7496,7512,7625,7627,7676,7741,7777-7778,7800,7911,7920-7921,7937-7938,7999-8002,8007-8011,8021-8022,8031,8042,8045,8080-8090,8093,8099-8100,8180-8181,8192-8194,8200,8222,8254,8290-8292,8300,8333,8383,8400,8402,8443,8500,8600,8649,8651-8652,8654,8701,8800,8873,8888,8899,8994,9000-9003,9009-9011,9040,9050,9071,9080-9081,9090-9091,9099-9103,9110-9111,9200,9207,9220,9290,9415,9418,9485,9500,9502-9503,9535,9575,9593-9595,9618,9666,9876-9878,9898,9900,9917,9929,9943-9944,9968,9998-10004,10009-10010,10012,10024-10025,10082,10180,10215,10243,10566,10616-10617,10621,10626,10628-10629,10778,11110-11111,11967,12000,12174,12265,12345,13456,13722,13782-13783,14000,14238,14441-14442,15000,15002-15004,15660,15742,16000-16001,16012,16016,16018,16080,16113,16992-16993,17877,17988,18040,18101,18988,19101,19283,19315,19350,19780,19801,19842,20000,20005,20031,20221-20222,20828,21571,22939,23502,24444,24800,25734-25735,26214,27000,27352-27353,27355-27356,27715,28201,30000,30718,30951,31038,31337,32768-32785,33354,33899,34571-34573,35500,38292,40193,40911,41511,42510,44176,44442-44443,44501,45100,48080,49152-49161,49163,49165,49167,49175-49176,49400,49999-50003,50006,50300,50389,50500,50636,50800,51103,51493,52673,52822,52848,52869,54045,54328,55055-55056,55555,55600,56737-56738,57294,57797,58080,60020,60443,61532,61900,62078,63331,64623,64680,65000,65129,65389"
var medium = "7,9,13,21-23,25-26,37,53,79-81,88,106,110-111,113,119,135,139,143-144,179,199,389,427,443-445,465,513-515,543-544,548,554,587,631,646,873,990,993,995,1025-1029,1110,1433,1720,1723,1755,1900,2000-2001,2049,2121,2717,3000,3128,3306,3389,3986,4899,5000,5009,5051,5060,5101,5190,5357,5432,5631,5666,5800,5900,6000-6001,6646,7070,8000,8008-8009,8080-8081,8443,8888,9100,9999-10000,32768,49152-49157"
var small = "21-23,25,53,80,110-111,135,139,143,199,443,445,587,993,995,1025,1720,1723,3306,3389,5900,8080,8888"
var top = "80-81,443-444,591-593,832,981,1582-1583,2087-2095,2480,4444-4445,4567,4711,5000,5104,5433,5555,5800,7000-7002,8008,8042,8080-8090,8222,8243,8280-8281,8530-8531,8443,8843,8887-8888,9080-9095,9443,9981,11371,12043,12443,16080,18091-18092,20000,24465"

type TargetHost chan struct{}

// Shim type for "set" containing ints
type IntSet struct {
	Set map[int]bool
}

// Shim type for "set" containing strings
type StringSet struct {
	Set map[string]bool
}

type Host struct {
	Path                      string            `json:"path" xml:"path"`
	HostAddr                  string            `json:"host_addr" xml:"host_addr"`
	Port                      int               `json:"port" xml:"port"`
	Protocol                  string            `json:"protocol" xml:"protocol"`
	ScreenshotFilename        string            `json:"screenshot_filename,omitempty" xml:"screenshot_filename,omitempty"`
	ResponseBodyFilename      string            `json:"response_body_filename,omitempty" xml:"response_body_filename,omitempty"`
	Soft404RandomURL          string            `json:"soft404_random_url,omitempty" xml:"soft404_random_url,omitempty"`
	Soft404RandomPageContents []byte            `json:"-" xml:"-"`
	PrefetchDone              bool              `json:"-" xml:"-"`
	Soft404Done               bool              `json:"-" xml:"-"`
	HTTPResp                  *http.Response    `json:"-" xml:"-"`
	HTTPReq                   *http.Request     `json:"-" xml:"-"`
	HostHeader                string            `json:"host_header,omitempty" xml:"host_header,omitempty"`
	UserAgent                 string            `json:"user_agent,omitempty" xml:"user_agent,omitempty"`
	Cookies                   string            `json:"cookies,omitempty" xml:"cookies,omitempty"`
	RequestHeaders            map[string]string `json:"request_headers,omitempty" xml:"request_headers,omitempty"`
}

func (host *Host) PrefetchHash() (h string) {
	hash := sha1.New()
	io.WriteString(hash, host.HostAddr)
	io.WriteString(hash, fmt.Sprintf("%d", host.Port))
	return fmt.Sprintf("%x", hash.Sum(nil))
}
func (host *Host) PrefetchDoneCheck(hashes map[string]bool) bool {
	if _, ok := hashes[host.PrefetchHash()]; ok {
		return true
	}
	return false
}

func (host *Host) Soft404Hash() (h string) {
	hash := sha1.New()
	io.WriteString(hash, host.HostAddr)
	io.WriteString(hash, fmt.Sprintf("%d", host.Port))
	io.WriteString(hash, host.Protocol)
	return fmt.Sprintf("%x", hash.Sum(nil))
}
func (host *Host) Soft404DoneCheck(hashes map[string]bool) bool {
	if _, ok := hashes[host.Soft404Hash()]; ok {
		return true
	}
	return false
}

func InitColours() {

}

func InitLogger(
	goodHandle io.Writer,
	infoHandle io.Writer,
	debugHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) Loggers {

	return Loggers{
		Good: log.New(goodHandle,
			g.Sprintf("GOOD: "),
			log.Ldate|log.Ltime),

		Info: log.New(infoHandle,
			b.Sprintf("INFO: "),
			log.Ldate|log.Ltime),

		Debug: log.New(debugHandle,
			y.Sprintf("DEBUG: "),
			log.Ldate|log.Ltime),

		Warning: log.New(warningHandle,
			m.Sprintf("WARNING: "),
			log.Ldate|log.Ltime),

		Error: log.New(errorHandle,
			r.Sprintf("ERROR: "),
			log.Ldate|log.Ltime),
	}
}

var d = net.Dialer{}
var tx = &http.Transport{}

var cl = http.Client{}

func ApplyJitter(Jitter int) {
	if Jitter > 0 {
		jitter := time.Duration(rand.Intn(Jitter)) * time.Millisecond
		time.Sleep(jitter)
	}
}

func GetTimeString() (currTime string) {
	t := time.Now()
	currTime = fmt.Sprintf("%d%d%d%d%d%d", t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	return
}

func Hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	return ips, nil
}

//  http://play.golang.org/p/m8TNTtygK0
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// ExpandHosts takes a string array of IP addresses/CIDR masks and converts into a string array of pure IP addresses
func ExpandHosts(targets []string) (allHosts StringSet) {
	allHosts = StringSet{Set: map[string]bool{}} // Initialise the hosts list... nfi why this is a thing?
	for _, target := range targets {
		ips, err := Hosts(target)
		if err != nil { // Not a CIDR... Might be a straight IP or hostname
			ip := net.ParseIP(target)
			if ip != nil {
				allHosts.Add(ip.String())
			} else {
				// could be hostname, i'll add it anyway... fuckit. DNS will solv this problem later
				allHosts.Add(target)
			}

		}
		allHosts.AddRange(ips)
	}
	return allHosts
}

// LeftPad2Len https://github.com/DaddyOh/golang-samples/blob/master/pad.go
func LeftPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr) - overallLen):]
}

// RightPad2Len https://github.com/DaddyOh/golang-samples/blob/master/pad.go
func RightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
func GetDataFromFile(l Loggers, fileName string) (data []string, err error) {
	if fileName != "" {
		data, err = readLines(fileName)
		if err != nil {
			l.Error.Printf("File: %v does not exist, or you do not have permz (%v)\n", fileName, err)
			return nil, err
		}
	}
	return data, err
}

// Taken from gobuster - THANKS! /**/
// StrArrToInt takes an array of strings and *hopefully* returns an array of ints?
func StrArrToInt(t []string) (t2 []int) {
	for _, i := range t {
		if i == "" {
			continue
		}
		j, err := strconv.Atoi(i)
		if err != nil {
			return t2
		}
		t2 = append(t2, j)
	}
	return t2
}

// Add an element to a set
func (set *StringSet) Add(s string) bool {
	_, found := set.Set[s]
	set.Set[s] = true
	return !found
}

// Add a list of elements to a set
func (set *StringSet) AddRange(ss []string) {
	for _, s := range ss {
		set.Set[s] = true
	}
}

// Test if an element is in a set
func (set *StringSet) Contains(s string) bool {
	_, found := set.Set[s]
	return found
}

// Check if any of the elements exist
func (set *StringSet) ContainsAny(ss []string) bool {
	for _, s := range ss {
		if set.Set[s] {
			return true
		}
	}
	return false
}

// Stringify the set
func (set *StringSet) Stringify() string {
	values := []string{}
	for s, _ := range set.Set {
		values = append(values, s)
	}
	return strings.Join(values, ",")
}

// Add an element to a set
func (set *IntSet) Add(i int) bool {
	_, found := set.Set[i]
	set.Set[i] = true
	return !found
}

// Test if an element is in a set
func (set *IntSet) Contains(i int) bool {
	_, found := set.Set[i]
	return found
}

// Stringify the set
func (set *IntSet) Stringify() string {
	values := []string{}
	for s, _ := range set.Set {
		values = append(values, strconv.Itoa(s))
	}
	return strings.Join(values, ",")
}

/**/
func ChunkString(s string, chunkSize int) []string {
	var chunks []string
	runes := []rune(s)

	if len(runes) == 0 {
		return []string{s}
	}

	for i := 0; i < len(runes); i += chunkSize {
		nn := i + chunkSize
		if nn > len(runes) {
			nn = len(runes)
		}
		chunks = append(chunks, string(runes[i:nn]))
	}
	return chunks
}

func GenerateURLs(targetList StringSet, Ports IntSet, Paths *StringSet, targets chan Host) {
	defer close(targets)
	for target, _ := range targetList.Set {
		for port, _ := range Ports.Set {
			targets <- Host{Port: port, HostAddr: target}
		}
	}
}

func ParseURLToHost(l Loggers, URL string, targets chan Host) {
	URL = strings.TrimSpace(URL)
	if URL == "" {
		return
	}

	URLObj, err := url.ParseRequestURI(URL)
	if err != nil || URLObj.Scheme == "" {
		// Attempt to add http:// if no scheme was provided
		URLObj, err = url.ParseRequestURI("http://" + URL)
		if err != nil {
			l.Error.Printf("Invalid URL provided and could not be parsed: %s\n", URL)
			return
		}
	}

	port := URLObj.Port()
	var Port int
	if port != "" {
		Port, err = strconv.Atoi(port)
	} else {
		if strings.ToLower(URLObj.Scheme) == "http" {
			Port = 80
		} else if strings.ToLower(URLObj.Scheme) == "https" {
			Port = 443
		}
	}
	path := URLObj.EscapedPath()
	targets <- Host{HostAddr: URLObj.Hostname(), Path: path, Protocol: URLObj.Scheme, Port: Port}
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandString() string {
	return fmt.Sprintf("%v", uuid.New())
}

func UnpackPortString(ports string) (ProcessedPorts IntSet) {
	ProcessedPorts = IntSet{Set: map[int]bool{}}
	for _, portRange := range strings.Split(ports, ",") {
		if x := StrArrToInt(strings.Split(portRange, "-")); len(x) > 1 {
			if x[0] > x[1] { // some dumbass will do this, surely
				tmp := x[0]
				x[0] = x[1]
				x[1] = tmp
			} else if x[0] == x[1] {
				ProcessedPorts.Add(x[0])
				continue
			}
			if x[0] < 0 {
				x[0] = 0
			}
			if x[1] > 65535 {
				x[1] = 65535 // some other dumbass will do this
			}
			for _, i := range makeRange(x[0], x[1]) {
				ProcessedPorts.Add(i)
			}
		} else {

			for _, i := range x {
				if i > 0 || i < 65536 {
					ProcessedPorts.Add(i)
				}
			}
		}
	}
	return
}

func (host *Host) makeHTTPRequest(ctx context.Context, url string) (req *http.Request, resp *http.Response, err error) {
	req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}
	if host.Cookies != "" {
		req.Header.Set("Cookie", host.Cookies)
	}
	if host.HostHeader != "" {
		req.Host = host.HostHeader
	}
	// TODO: support additional custom headers
	//      - Will probably require a cmdline arg for JSON header object or something?
	for header, value := range host.RequestHeaders {
		req.Header.Set(header, value)
	}
	req.Header.Set("User-Agent", host.UserAgent)
	resp, err = cl.Do(req)
	if err != nil {
		return
	}
	return
}
