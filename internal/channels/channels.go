// Package channels contains unseful functions for concurrent go
// Most of them orgiginates from the Oreilly's book "Concurrency in Go" - Chapter 4. "Concurrency Patterns in Go"
package channels

import (
	"sync"
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

// OrDoneTimeout iterates over channhel c until: (1) c closes, (2) timeout happens, (3) done receives a message
// closing c means that the initialisation procedures has finished
func OrDoneTimeout(done <-chan struct{}, timeout <-chan time.Time, c <-chan struct{}) <-chan struct{} {
	valStream := make(chan struct{})
	go func() {
		defer close(valStream)
		for {
			select {
			case <-timeout:
				return
			case <-done:
				return
			case _, ok := <-c:
				if !ok {
					valStream <- struct{}{}
					return
				}
			}
		}
	}()
	return valStream
}

// FanIn merges channels into one output channel
func FanIn(done <-chan struct{}, channels ...<-chan struct{}) <-chan struct{} {
	var wg sync.WaitGroup
	out := make(chan struct{})
	output := func(c <-chan struct{}) {
		defer wg.Done()
		for i := range c {
			select {
			case <-done:
				return
			case out <- i:
			}
		}
	}
	wg.Add(len(channels))
	for _, c := range channels {
		go output(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
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
