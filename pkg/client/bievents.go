package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

type BIEventsFilter struct {
	svcAddr   string
	svcSecret string
	caller    *caller
	channel   chan []byte
	wg        *sync.WaitGroup
}

var userAgentBIEventsFilter = "go-svc.client.BIEventsFilter"

func NewBIEventsFilter(runCtx context.Context, svcAddr, svcSecret, caller string) (done chan struct{}, bic BIEventsFilter) {
	c := make(chan []byte)
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
			data := <-c
			err := bic.postEvent(data)
			if err != nil {
				// let worker sleep to reduce cpu load
				time.Sleep(time.Second)
				// requeue data
				go func(data []byte) {
					bic.channel <- data
				}(data)
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
	go func(data []byte) {
		bic.channel <- data
	}(json)
	return nil
}

func (bic BIEventsFilter) postEvent(data []byte) error {
	contentURL := fmt.Sprintf("%s/api/v1/events", bic.svcAddr)
	_, _, err := bic.caller.call(context.Background(), contentURL, "POST", bic.svcSecret, userAgentBIEventsFilter, bytes.NewBuffer(data), http.StatusOK)
	return err
}
