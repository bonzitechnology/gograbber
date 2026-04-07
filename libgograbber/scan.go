package libgograbber

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func Scan(s *State, Targets chan Host, ScanChan chan Host, currTime string, threadChan chan struct{}, wg *sync.WaitGroup) {
	defer func() {
		close(ScanChan)
		wg.Done()
	}()
	var scanWg = sync.WaitGroup{}

	if !s.Scan {
		// We're not doing a Scan here so just pump the values back into the pipeline for the next phase to consume
		for host := range Targets {
			ScanChan <- host
		}
		return
	}
	sWriteChan := make(chan []byte)
	var portScanOutFile string

	if s.ProjectName != "" {
		portScanOutFile = fmt.Sprintf("%v/hosts_%v_%v_%v.txt", s.ScanOutputDirectory, strings.ToLower(strings.Replace(s.ProjectName, " ", "_", -1)), currTime, rand.Int63())
	} else {
		portScanOutFile = fmt.Sprintf("%v/hosts_%v_%v.txt", s.ScanOutputDirectory, currTime, rand.Int63())
	}
	go writerWorker(sWriteChan, portScanOutFile)
	for host := range Targets {
		atomic.AddInt64(&s.ScanCounter, 1)
		scanWg.Add(1)
		threadChan <- struct{}{}
		go ConnectHost(&scanWg, s.Timeout, s.Jitter, s.Debug, host, ScanChan, threadChan, sWriteChan)
	}
	scanWg.Wait()
	close(sWriteChan)
}

// connectHost does the actual TCP connection
func ConnectHost(wg *sync.WaitGroup, timeout time.Duration, Jitter int, debug bool, host Host, results chan Host, threads chan struct{}, writeChan chan []byte) {
	defer func() {
		<-threads
		wg.Done()
	}()
	if debug {
		Info.Printf("Port scanning: %v:%v\n", host.HostAddr, host.Port)
	}
	ApplyJitter(Jitter)
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%v", host.HostAddr, host.Port), timeout)
	if err == nil {
		defer conn.Close()
		Good.Printf("%v:%v - %v\n", host.HostAddr, host.Port, g.Sprintf("tcp/%v open", host.Port))
		writeChan <- []byte(fmt.Sprintf("%v,%v\n", host.HostAddr, host.Port))
		results <- host
	} else {
		if debug {
			Debug.Printf("Err connecting [%v:%v]: %v\n", host.HostAddr, host.Port, err)
		}
	}
}
