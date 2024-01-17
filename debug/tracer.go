package debug

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
)

const warningThreshold = time.Second * 2

var log = logger.GetOrCreate("debug")
var PROCESS_TRACER = NewTracer()

type tracer struct {
	mutState         sync.RWMutex
	startTime        time.Time
	cpuFileHandler   *os.File
	mutexFileHandler *os.File
}

func NewTracer() *tracer {
	return &tracer{}
}

func (t *tracer) Start() {
	t.mutState.Lock()
	defer t.mutState.Unlock()

	t.startTime = time.Now()
}

func (t *tracer) Stop() {
	t.mutState.Lock()
	defer t.mutState.Unlock()

	if t.cpuFileHandler == nil {
		return
	}

	pprof.StopCPUProfile()
	_ = t.cpuFileHandler.Close()
	t.cpuFileHandler = nil

	_ = t.mutexFileHandler.Close()
	t.mutexFileHandler = nil

	log.Warn("tracer.stop")
	dump := make([]byte, 1024*1024*100) // 100MB
	num := runtime.Stack(dump, true)

	log.Warn("full stack dump", "dump", string(dump[:num]))
}

func (t *tracer) HaveTimeQueried() {
	t.mutState.Lock()
	defer t.mutState.Unlock()

	if t.cpuFileHandler != nil {
		return
	}
	if time.Since(t.startTime) < warningThreshold {
		return
	}

	log.Warn("tracer.started")
	currentTime := time.Now().Unix()
	t.cpuFileHandler, _ = os.Create(fmt.Sprintf("prof_%d.prof", currentTime))
	t.mutexFileHandler, _ = os.Create(fmt.Sprintf("mutex_%d.prof", currentTime))

	err := pprof.StartCPUProfile(t.cpuFileHandler)
	log.LogIfError(err)

	runtime.SetMutexProfileFraction(5)
	prof := pprof.Lookup("mutex")
	err = prof.WriteTo(t.mutexFileHandler, 1)
	log.LogIfError(err)
}
