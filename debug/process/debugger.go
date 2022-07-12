package process

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
)

const minAcceptedValue = 1

const buffSize = 100 * 1024 * 1024 // 100MB
var log = logger.GetOrCreate("debug/process")

type processDebugger struct {
	timer                   *time.Timer
	mut                     sync.RWMutex
	lastCheckedBlockRound   int64
	lastCommittedBlockRound int64
	cancel                  func()
	goRoutinesDumpHandler   func()
	logChangeHandler        func()

	maxRoundsWithoutCommit int64
	pollingTime            time.Duration
	logLevel               string
	dumpGoRoutines         bool
}

// NewProcessDebugger creates a new debugger instance used to monitor the block process flow
func NewProcessDebugger(config config.ProcessDebugConfig) (*processDebugger, error) {
	err := checkConfigs(config)
	if err != nil {
		return nil, err
	}

	pollingTime := time.Duration(config.PollingTimeInSeconds) * time.Second
	d := &processDebugger{
		timer: time.NewTimer(pollingTime),

		pollingTime:    pollingTime,
		logLevel:       config.LogLevelChanger,
		dumpGoRoutines: config.GoRoutinesDump,
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.cancel = cancel
	d.goRoutinesDumpHandler = func() {
		buff := make([]byte, buffSize)
		numBytes := runtime.Stack(buff, true)
		log.Debug(string(buff[:numBytes]))
	}
	d.logChangeHandler = func() {
		errSetLogLevel := logger.SetLogLevel(d.logLevel)
		log.LogIfError(errSetLogLevel)
	}

	go d.processLoop(ctx)

	return d, nil
}

func checkConfigs(config config.ProcessDebugConfig) error {
	if config.PollingTimeInSeconds < minAcceptedValue {
		return fmt.Errorf("%w for PollingTimeInSeconds, minimum %d, got %d",
			errInvalidValue, minAcceptedValue, config.PollingTimeInSeconds)
	}

	return nil
}

func (debugger *processDebugger) processLoop(ctx context.Context) {
	log.Debug("processor debugger processLoop is starting...")

	defer debugger.timer.Stop()

	for {
		debugger.timer.Reset(debugger.pollingTime)

		select {
		case <-ctx.Done():
			log.Debug("processor debugger processLoop is closing...")
			return
		case <-debugger.timer.C:
			debugger.checkRounds()
		}
	}
}

func (debugger *processDebugger) checkRounds() {
	if debugger.shouldTriggerUpdatingLastCheckedRound() {
		debugger.trigger()
	}
}

func (debugger *processDebugger) shouldTriggerUpdatingLastCheckedRound() bool {
	debugger.mut.Lock()
	defer debugger.mut.Unlock()

	isNodeStarting := debugger.lastCheckedBlockRound == 0 && debugger.lastCommittedBlockRound <= 0
	if isNodeStarting {
		log.Debug("processor debugger: node is starting")
		return false
	}

	defer func() {
		// update the last checked round
		debugger.lastCheckedBlockRound = debugger.lastCommittedBlockRound
	}()

	isFirstCommit := debugger.lastCheckedBlockRound == 0 && debugger.lastCommittedBlockRound > 0
	if isFirstCommit {
		log.Debug("processor debugger: first committed block", "round", debugger.lastCommittedBlockRound)
		return false
	}

	isNodeRunning := debugger.lastCheckedBlockRound < debugger.lastCommittedBlockRound
	if isNodeRunning {
		log.Debug("processor debugger: node is running, nothing to do", "round", debugger.lastCommittedBlockRound)
		return false
	}

	return true
}

func (debugger *processDebugger) trigger() {
	debugger.mut.RLock()
	lastCommittedBlockRound := debugger.lastCommittedBlockRound
	debugger.mut.RUnlock()

	log.Warn("processor debugger: node is stuck",
		"last committed round", lastCommittedBlockRound)

	debugger.logChangeHandler()

	if debugger.dumpGoRoutines {
		debugger.goRoutinesDumpHandler()
	}
}

// SetLastCommittedBlockRound sets the last committed block's round
func (debugger *processDebugger) SetLastCommittedBlockRound(round uint64) {
	debugger.mut.Lock()
	defer debugger.mut.Unlock()

	log.Debug("processor debugger: updated last committed block round", "round", round)
	debugger.lastCommittedBlockRound = int64(round)
}

// Close stops any started go routines
func (debugger *processDebugger) Close() error {
	debugger.cancel()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (debugger *processDebugger) IsInterfaceNil() bool {
	return debugger == nil
}
