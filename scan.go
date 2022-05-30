package main

import (
	"context"
	"log"
	"time"

	"github.com/borud/hdlc"
	"go.bug.st/serial"
)

// scanner that looks for HDLC framed serial devices.
type scanner struct {
	timeout  time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	framedCh chan string
}

const (
	// the miniumum number of frames we need to see in order to be convinced.
	minFramesSeen = 1
)

// Scan for serial ports that seem to be emitting HDLC frames. Will try to scan until we
// time out all the ports we found.  Since we open them all in parallell, the timeout
// should be the approximate running time of this call.
//
// If abortAfterFirst is true we abort as soon as we have found one device.
func Scan(timeout time.Duration, abortAfterFirst bool) []string {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(timeout))

	s := &scanner{
		timeout:  timeout,
		framedCh: make(chan string),
		ctx:      ctx,
		cancel:   cancel,
	}

	candidates, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}

	var framed []string

	go func() {
		for path := range s.framedCh {
			framed = append(framed, path)
			if abortAfterFirst {
				s.cancel()
				return
			}
		}
	}()

	log.Printf("scanning %d ports", len(candidates))
	for _, path := range candidates {
		go s.detectFramed(path)
	}

	<-s.ctx.Done()
	return framed
}

// detectFramed opens serial device at `path` and tries to detect whether it
// emits something that looks like an HDLC framed protocol.
func (s *scanner) detectFramed(path string) {
	port, err := serial.Open(path, &serial.Mode{})
	if err != nil {
		log.Printf("error opening %s: %s", path, err)
		return
	}
	defer port.Close()

	unframer := hdlc.NewUnframer(port)
	countDown := minFramesSeen
	for {
		select {
		case frame := <-unframer.Frames():
			log.Printf("got a frame %s: [%x]", path, frame)
			countDown--
			if countDown == 0 {
				s.framedCh <- path
				return
			}

		case <-s.ctx.Done():
			return
		}
	}
}
