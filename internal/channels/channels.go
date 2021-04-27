// Package channels contains unseful functions for concurrent go
// Most of them orgiginates from the Oreilly's book "Concurrency in Go" - Chapter 4. "Concurrency Patterns in Go"
package channels

import (
	"time"
)

// OrDone iterates over channhel c until it closes or `done` receives a message
func OrDone(done, c <-chan struct{}) <-chan struct{} {
	valStream := make(chan struct{})
	go func() {
		defer close(valStream)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if !ok {
					return
				}
				select {
				case valStream <- v:
				case <-done:
				}
			}
		}
	}()
	return valStream
}

// OrDoneTimeout iterates over channels c0 and c1 until: (1) c0 AND c1 closes, (2) timeout happens, (3) done receives a message
// values returned from c0 and c1 (while channel is open) mean that the respective DB connection is up
// closing c0 and c1 means that their initialisation procedures have finished
func OrDoneTimeout(done <-chan struct{}, timeout <-chan time.Time, c0 <-chan struct{}, c1 <-chan struct{}) <-chan struct{} {
	valStream := make(chan struct{})
	go func() {
		defer close(valStream)
		channelOpen0 := true
		channelOpen1 := true
		for {
			select {
			case <-timeout:
				return
			case <-done:
				return
			case _, ok := <-c0:
				channelOpen0 = ok
				if !channelOpen0 && !channelOpen1 {
					valStream <- struct{}{}
					return
				}
			case _, ok := <-c1:
				channelOpen1 = ok
				if !channelOpen0 && !channelOpen1 {
					valStream <- struct{}{}
					return
				}
			}
		}
	}()
	return valStream
}

// Or returns when the first of the channels returns
func Or(channels ...<-chan struct{}) <-chan struct{} {
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}
	orDone := make(chan struct{})
	go func() {
		defer close(orDone)

		switch len(channels) {
		case 2:
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default:
			select {
			case <-channels[0]:
			case <-channels[1]:
			case <-channels[2]:
			case <-Or(append(channels[3:], orDone)...):
			}
		}
	}()
	return orDone
}
