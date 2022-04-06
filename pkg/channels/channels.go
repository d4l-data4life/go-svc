// Package channels contains useful functions for concurrent go
// Most of them originate from the Oreilly's book "Concurrency in Go" - Chapter 4. "Concurrency Patterns in Go"
package channels

import (
	"sync"
	"time"
)

// OrDone forwards data of channel c until c or done close
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

// OrDoneTimeout forwards data of channel c until: (1) c or done close, (2) timeout happens
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

// FanIn merges channels into one output channel
// Data is forwarded and output channel stays open until all channels close
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

// Or returns a channel that closes, when any of the input channels close.
// Data is not forwarded.
//
// Example: Merge multiple done channels into a single one
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

// Barrier returns a channel that closes, when the done channel or any of the input channels close
// Channel emits only after all input channels have emitted data
//
// Example: Merge multiple start-up channels into a single one
func Barrier(done <-chan struct{}, channels ...<-chan struct{}) <-chan struct{} {
	var wg sync.WaitGroup
	var m sync.Mutex
	inProgress := len(channels)
	out := make(chan struct{})
	output := func(c <-chan struct{}) {
		defer wg.Done()
		first := true
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if !ok {
					return
				}
				if first {
					first = false
					m.Lock()
					inProgress -= 1
					// Emit to out channel only after all channels have emitted once
					if inProgress == 0 {
						select {
						case out <- v:
						case <-done:
						}
					}
					m.Unlock()
				}
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
