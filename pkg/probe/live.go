package probe

import (
	"sync"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

var onceLive sync.Once
var instaceLive *Live

type Live struct {
	mutex *sync.Mutex
	value bool
}

func (l *Live) GetValue() bool {
	return l.value
}

func (l *Live) SetLive() {
	logging.LogDebugf("liveness set to true")
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.value = true
}

func (l *Live) SetDead() {
	logging.LogDebugf("liveness set to false")
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.value = false
}

// Liveness is a singleton to mark the status of the K8s liveness probe
func Liveness() *Live {
	onceLive.Do(func() {
		instaceLive = &Live{value: false, mutex: &sync.Mutex{}}
	})
	return instaceLive
}
