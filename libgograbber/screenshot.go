package libgograbber

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func Screenshot(s *State, DirbustChan chan Host, ScreenshotChan chan Host, currTime string, threadChan chan struct{}, wg *sync.WaitGroup) {
	defer func() {
		close(ScreenshotChan)
		wg.Done()
	}()
	var screenshotWg = sync.WaitGroup{}

	if !s.Screenshot {
		// We're not doing Screenshotting here so just pump the values back into the pipeline for the next phase to consume
		for host := range DirbustChan {
			ScreenshotChan <- host
		}
		return
	}
	var cnt int
	screenshotWorkers := make(chan struct{}, s.NumScreenshotWorkers)
	for host := range DirbustChan {
		screenshotWorkers <- struct{}{}
		screenshotWg.Add(1)
		atomic.AddInt64(&s.ScreenshotCounter, 1)
		go ScreenshotAURL(&screenshotWg, s, cnt, host, ScreenshotChan, screenshotWorkers)
		cnt++
	}
	screenshotWg.Wait()
}

// Screenshots a url derived from a Host{} object
func ScreenshotAURL(wg *sync.WaitGroup, s *State, cnt int, host Host, results chan Host, threads chan struct{}) (err error) {
	defer func() {
		<-threads
		wg.Done()
	}()
	
	url := fmt.Sprintf("%v://%v:%v/%v", host.Protocol, host.HostAddr, host.Port, host.Path)

	if strings.HasPrefix(host.Path, "/") && len(host.Path) > 0 {
		host.Path = host.Path[1:] // strip preceding '/' char
	}
	if s.Debug {
		Debug.Printf("Trying to screenshot URL: %v\n", url)
	}
	ApplyJitter(s.Jitter)
	
	timeout := s.Timeout + (time.Second * 5)
	
	page, err := s.Browser.Timeout(timeout).Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		Error.Printf("Unable to create page: %v (%v)\n", url, err)
		return err
	}
	defer page.Close()

	if host.ResponseBodyFilename != "" {
		if body, err := os.ReadFile(host.ResponseBodyFilename); err == nil {
			router := page.HijackRequests()
			defer router.Stop()
			
			router.MustAdd("*", func(ctx *rod.Hijack) {
				reqURL := ctx.Request.URL().String()
				if reqURL == url || reqURL == url+"/" {
					ctx.Response.Payload().ResponseCode = host.HTTPResp.StatusCode
					ctx.Response.SetBody(body)
					if host.HTTPResp != nil && host.HTTPResp.Header != nil {
						for k, v := range host.HTTPResp.Header {
							if len(v) > 0 {
								ctx.Response.SetHeader(k, v[0])
							}
						}
					}
				} else {
					ctx.ContinueRequest(&proto.FetchContinueRequest{})
				}
			})
			go router.Run()
		}
	}

	// Wait for network idle or load
	_ = page.Timeout(timeout).Navigate(url)
	_ = page.Timeout(timeout).WaitLoad()

	// Setup the viewport
	err = page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             s.ImgX,
		Height:            s.ImgY,
		DeviceScaleFactor: 1,
	})
	if err != nil {
		Error.Printf("Unable to set Viewport size: %v (%v)\n", url, err)
		return err
	}
	
	currTime := GetTimeString()
	var screenshotFilename string
	if s.ProjectName != "" {
		screenshotFilename = fmt.Sprintf("%v/%v_%v-%v_%v.%v", s.ScreenshotDirectory, strings.ToLower(SanitiseFilename(s.ProjectName)), SanitiseFilename(url), currTime, rand.Int63(), s.ScreenshotFileType)
	} else {
		screenshotFilename = fmt.Sprintf("%v/%v-%v_%v.%v", s.ScreenshotDirectory, SanitiseFilename(url), currTime, rand.Int63(), s.ScreenshotFileType)
	}
	
	format := proto.PageCaptureScreenshotFormatPng
	if strings.ToLower(s.ScreenshotFileType) == "jpeg" || strings.ToLower(s.ScreenshotFileType) == "jpg" {
		format = proto.PageCaptureScreenshotFormatJpeg
	}
	
	img, err := page.Timeout(timeout).Screenshot(false, &proto.PageCaptureScreenshot{
		Format:  format,
		Quality: &s.ScreenshotQuality,
	})
	if err != nil {
		Error.Printf("Unable to save Screenshot: %v (%v)\n", url, err)
		return err
	}
	
	err = os.WriteFile(screenshotFilename, img, 0644)
	if err != nil {
		Error.Printf("Unable to write Screenshot file: %v (%v)\n", url, err)
		return err
	}

	Good.Printf("Screenshot for [%v] saved to: [%v]\n", g.Sprintf("%s", url), g.Sprintf("%s", screenshotFilename))
	host.ScreenshotFilename = screenshotFilename
	results <- host
	return nil
}
