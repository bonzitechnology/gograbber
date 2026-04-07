package libgograbber

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

func Dirbust(ctx context.Context, s *State, ScanChan chan Host, DirbustChan chan Host, currTime string, threadChan chan struct{}, wg *sync.WaitGroup) {
	defer func() {
		close(DirbustChan)
		wg.Done()
	}()
	var dirbWg = sync.WaitGroup{}

	if !s.Dirbust {
		// We're not doing a dirbust here so just pump the values back into the pipeline for the next phase to consume
		for host := range ScanChan {
			if !s.URLProvided {
				for scheme := range s.Protocols.Set {
					host.Protocol = scheme
					DirbustChan <- host
				}
			} else {
				DirbustChan <- host
			}
		}
		return
	}
	// Do dirbusting
	var dirbustOutFile string

	dWriteChan := make(chan []byte)

	if s.ProjectName != "" {
		dirbustOutFile = fmt.Sprintf("%v/urls_%v_%v_%v.txt", s.DirbustOutputDirectory, strings.ToLower(SanitiseFilename(s.ProjectName)), currTime, rand.Int63())
	} else {
		dirbustOutFile = fmt.Sprintf("%v/urls_%v_%v.txt", s.DirbustOutputDirectory, currTime, rand.Int63())
	}
	go writerWorker(s.Log, dWriteChan, dirbustOutFile)
	
	for host := range ScanChan {
		select {
		case <-ctx.Done():
			return
		default:
			dirbWg.Add(1)
			host.RequestHeaders = s.HttpHeaders
			host.UserAgent = s.UserAgent
			host.Cookies = s.Cookies
			for hostHeader, _ := range s.HostHeaders.Set {
				host.HostHeader = hostHeader
				if s.URLProvided {
					var h Host
					h = host
					dirbWg.Add(1)
					go dirbRunner(ctx, s, h, &dirbWg, threadChan, DirbustChan, dWriteChan)

				} else {
					for scheme := range s.Protocols.Set {
						var h Host
						h = host
						h.Protocol = scheme // Weird hack to fix a random race condition...
						dirbWg.Add(1)
						go dirbRunner(ctx, s, h, &dirbWg, threadChan, DirbustChan, dWriteChan)
					}
				}
			}
			dirbWg.Done()
		}
	}
	dirbWg.Wait()
	close(dWriteChan)
}

func dirbRunner(ctx context.Context, s *State, h Host, dirbWg *sync.WaitGroup, threadChan chan struct{}, DirbustChan chan Host, dWriteChan chan []byte) {
	defer dirbWg.Done()

	if s.Soft404Detection {
		h = PerformSoft404Check(s, h, s.Debug, s.Canary)
	}
	for path, _ := range s.Paths.Set {
		var p string
		p = fmt.Sprintf("%v/%v", strings.TrimSuffix(h.Path, "/"), strings.TrimPrefix(path, "/"))
		// Add custom file extension to each request specified using -e flag
		for ext, _ := range s.Extensions.Set {
			var extPath string
			ext = strings.TrimPrefix(ext, ".")
			if len(ext) == 0 {
				extPath = p
			} else {
				extPath = fmt.Sprintf("%s.%s", p, ext)
			}
			select {
			case <-ctx.Done():
				return
			default:
				dirbWg.Add(1)
				threadChan <- struct{}{}
				atomic.AddInt64(&s.DirbustCounter, 1)
				go HTTPGetter(ctx, dirbWg, s, h, extPath, DirbustChan, threadChan, dWriteChan)
			}
		}
	}
}

func HTTPGetter(ctx context.Context, wg *sync.WaitGroup, s *State, host Host, path string, results chan Host, threads chan struct{}, writeChan chan []byte) {
	defer func() {
		<-threads
		wg.Done()
	}()

	if strings.HasPrefix(path, "/") && len(path) > 0 {
		path = path[1:] // strip preceding '/' char
	}
	Url := fmt.Sprintf("%v://%v:%v/%v", host.Protocol, host.HostAddr, host.Port, path)
	if s.Debug {
		s.Log.Debug.Printf("Trying URL: %v\n", Url)
	}
	ApplyJitter(s.Jitter)

	var err error
	nextUrl := Url
	var i int
	var redirs []string
	numRedirects := s.MaxRedirects
	for i < numRedirects { // number of times to follow redirect
		i++

		host.HTTPReq, host.HTTPResp, err = host.makeHTTPRequest(ctx, nextUrl)
		if err != nil {
			return
		}
		if s.StatusCodesIgn.Contains(host.HTTPResp.StatusCode) {
			host.HTTPResp.Body.Close()
			return
		}
		
		if host.HTTPResp.StatusCode >= 300 && host.HTTPResp.StatusCode < 400 && s.FollowRedirects {
			host.HTTPResp.Body.Close()
			x, err := host.HTTPResp.Location()
			if err == nil {
				redirs = append(redirs, fmt.Sprintf("[%v - %s]", y.Sprintf("%d", host.HTTPResp.StatusCode), g.Sprintf("%s", nextUrl)))
				writeChan <- []byte(fmt.Sprintf("%v\n", nextUrl))
				nextUrl = x.String()
			} else {
				break
			}
		} else {
			if s.FollowRedirects && len(redirs) > 0 {
				s.Log.Good.Printf("Redirect %v->[%v - %v]", strings.Join(redirs, "->"), y.Sprintf("%d", host.HTTPResp.StatusCode), g.Sprintf("%s", nextUrl))
			}
			Url = nextUrl
			break
		}
	}
	
	if host.HTTPResp == nil {
		return
	}
	defer host.HTTPResp.Body.Close()
	
	buf, err := ioutil.ReadAll(host.HTTPResp.Body)
	if err != nil {
		return
	}

	if s.Soft404Detection && path != "" && host.Soft404RandomPageContents != nil {
		soft404Ratio := detectSoft404(buf, host.Soft404RandomPageContents)
		if soft404Ratio > s.Ratio {
			if s.Debug {
				s.Log.Debug.Printf("[%v] is very similar to [%v] (%v match)\n", y.Sprintf("%s", Url), y.Sprintf("%s", host.Soft404RandomURL), y.Sprintf("%.4f%%", (soft404Ratio*100)))
			}
			return
		}
	}

	if host.HostHeader != "" {
		s.Log.Good.Printf("%v - %v [%v bytes] (HostHeader: %v)\n", Url, g.Sprintf("%d", host.HTTPResp.StatusCode), len(buf), host.HostHeader)
	} else {
		s.Log.Good.Printf("%v - %v [%v bytes]\n", Url, g.Sprintf("%d", host.HTTPResp.StatusCode), len(buf))
	}
	currTime := GetTimeString()

	var responseFilename string
	if s.ProjectName != "" {
		responseFilename = fmt.Sprintf("%v/%v_%v-%v_%v.html", s.HTTPResponseDirectory, strings.ToLower(SanitiseFilename(s.ProjectName)), SanitiseFilename(Url), currTime, rand.Int63())
	} else {
		responseFilename = fmt.Sprintf("%v/%v-%v_%v.html", s.HTTPResponseDirectory, SanitiseFilename(Url), currTime, rand.Int63())
	}
	file, err := os.Create(responseFilename)
	if err != nil {
		s.Log.Error.Printf("%v\n", err)
	} else {
		if len(buf) > 0 {
			file.Write(buf)
			host.ResponseBodyFilename = responseFilename
		}
		file.Close()
		if len(buf) == 0 {
			_ = os.Remove(responseFilename)
		}
	}
	host.Path = path
	writeChan <- []byte(fmt.Sprintf("%v\n", Url))
	results <- host
}

func PerformSoft404Check(s *State, h Host, debug bool, canary string) Host {
	var knary string
	if canary != "" {
		knary = canary
	} else {
		knary = RandString()
	}
	randURL := fmt.Sprintf("%v://%v:%v/%v", h.Protocol, h.HostAddr, h.Port, knary)
	if debug {
		s.Log.Debug.Printf("Soft404 checking [%v]\n", randURL)
	}
	_, randResp, err := h.makeHTTPRequest(context.Background(), randURL)
	if err != nil {
		if debug {
			s.Log.Error.Printf("Soft404 check failed... [%v] Err:[%v] \n", randURL, err)
		}
	} else {
		defer randResp.Body.Close()
		data, err := ioutil.ReadAll(randResp.Body)
		if err != nil {
			s.Log.Error.Printf("uhhh... [%v]\n", err)
			return h
		}
		h.Soft404RandomURL = randURL
		h.Soft404RandomPageContents = data
	}
	return h
}

func detectSoft404(responseData []byte, randRespData []byte) (ratio float64) {
	lenResp := len(responseData)
	lenRand := len(randRespData)
	
	if lenResp == lenRand {
		return 1.0
	}
	
	var diff int
	if lenResp > lenRand {
		diff = lenResp - lenRand
	} else {
		diff = lenRand - lenResp
	}
	
	maxLen := lenResp
	if lenRand > maxLen {
		maxLen = lenRand
	}
	
	if maxLen == 0 {
		return 1.0
	}
	
	return 1.0 - (float64(diff) / float64(maxLen))
}
