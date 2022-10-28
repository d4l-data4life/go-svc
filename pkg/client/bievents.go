package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

type EventAttempts struct {
	Data     []byte
	Attempts int
}

type BIEventsFilter struct {
	svcAddr   string
	svcSecret string
	caller    *caller
	channel   chan EventAttempts
	wg        *sync.WaitGroup
}

const MaxAttempts = 5

var ErrSendingFailed = errors.New("failed sending BI Event")

var userAgentBIEventsFilter = "go-svc.client.BIEventsFilter"

func NewBIEventsFilter(runCtx context.Context, svcAddr, svcSecret, caller string) (done chan struct{}, bic BIEventsFilter) {
	c := make(chan EventAttempts)
	done = make(chan struct{})

	bic = BIEventsFilter{
		svcAddr:   svcAddr,
		svcSecret: svcSecret,
		channel:   c,
		// prefixing prom metrics
		caller: NewCaller(30*time.Second, "bievents"+caller),
		wg:     &sync.WaitGroup{},
	}

	// worker goroutine loop sending events to the bi service
	go func() {
		for {
			event := <-c
			err := bic.postEvent(event)
			event.Attempts++
			if err != nil {
				// requeue event if attempted less than maximum times
				if event.Attempts < MaxAttempts {
					// let worker sleep with exponential backoff to reduce cpu load
					time.Sleep(time.Second * time.Duration(math.Pow(2.0, float64(event.Attempts+1))))
					go func(event EventAttempts) {
						bic.channel <- event
					}(event)
				} else {
					logging.LogErrorfCtx(context.Background(), ErrSendingFailed, "failed after %d attempts", MaxAttempts)
					bic.wg.Done()
				}
			} else {
				bic.wg.Done()
			}
		}
	}()

	// 1. on context cancel wait for remaining events to be send
	// 2. close done channel to indicate all enqueued events are sent
	// another goroutine can block until the channel is closed
	go func(wg *sync.WaitGroup) {
		<-runCtx.Done()
		logging.LogInfof("Sending remaining bi events")
		wg.Wait()
		logging.LogInfof("Done sending remaining bi events")
		close(done)
	}(bic.wg)

	return done, bic
}

func (bic BIEventsFilter) Send(data interface{}) error {
	bic.wg.Add(1)
	json, err := json.Marshal(data)
	if err != nil {
		return err
	}
	go func(data EventAttempts) {
		bic.channel <- data
	}(EventAttempts{json, 0})
	return nil
}

func (bic BIEventsFilter) postEvent(event EventAttempts) error {
	contentURL := fmt.Sprintf("%s/api/v1/events", bic.svcAddr)
	_, _, err := bic.caller.call(context.Background(), contentURL, "POST", bic.svcSecret, userAgentBIEventsFilter, bytes.NewBuffer(event.Data), http.StatusOK)
	return err
}
