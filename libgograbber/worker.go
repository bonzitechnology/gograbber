package libgograbber

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func RoutineManager(ctx context.Context, s *State, ScanChan chan Host, DirbustChan chan Host, ScreenshotChan chan Host, wg *sync.WaitGroup) {
	defer wg.Done()
	threadChan := make(chan struct{}, s.Threads)
	currTime := GetTimeString()

	ticker := time.NewTicker(5 * time.Second)
	startTime := time.Now()
	go func() {
		var currTime time.Duration
		for t := range ticker.C {
			select {
			case <-ctx.Done():
				return
			default:
				currTime = t.Sub(startTime)
				scanCnt := atomic.LoadInt64(&s.ScanCounter)
				dirbCnt := atomic.LoadInt64(&s.DirbustCounter)
				screenCnt := atomic.LoadInt64(&s.ScreenshotCounter)
				
				if s.Debug {
					fmt.Print(LineSep())
					s.Log.Debug.Printf("Elapsed %v | Scanned: %d | Dirbusted: %d | Screenshots: %d\n", currTime, scanCnt, dirbCnt, screenCnt)
					fmt.Print(LineSep())
				} else {
					s.Log.Info.Printf("Progress [%v] - Scanned: %d | Dirbusted: %d | Screenshots: %d\n", currTime.Round(time.Second), scanCnt, dirbCnt, screenCnt)
				}
			}
		}
	}()

	// Start our operations
	wg.Add(1)
	go Scan(ctx, s, s.Targets, ScanChan, currTime, threadChan, wg)
	wg.Add(1)
	go Dirbust(ctx, s, ScanChan, DirbustChan, currTime, threadChan, wg)
	wg.Add(1)
	go Screenshot(ctx, s, DirbustChan, ScreenshotChan, currTime, threadChan, wg)
}
