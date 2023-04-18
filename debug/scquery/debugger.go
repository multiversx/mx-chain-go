package scquery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const minIntervalAutoPrint = 1
const logStartProcessLoop = "scQueryDebugger debugger processLoop go routine is starting..."
const logClosedProcessLoop = "scQueryDebugger debugger processLoop go routine is stopping..."
const logMetrics = "SC query debugger metrics"
const logMaxConcurrentQueries = "max concurrent queries"
const logDurationOfHeaviestCall = "duration of heaviest call"
const logIndexOfHeaviestCall = "index of heaviest call"
const logFinishedCalls = "finished calls"
const logRunningCalls = "running calls"
const logMetricsInterval = "metrics interval"

// ArgsSCQueryDebugger is the argument DTO for the NewSCQueryDebugger function
type ArgsSCQueryDebugger struct {
	IntervalAutoPrintInSeconds int
	LoggerInstance             logger.Logger
}

type scQueryDebugger struct {
	intervalAutoPrint         time.Duration
	loggerInstance            logger.Logger
	cancelFunc                func()
	mutData                   sync.RWMutex
	runningQueries            map[int]int64
	maxNumOfConcurrentQueries int
	numFinishedCalls          int
	indexOfHeavierSCCall      int
	durationOfHeaviestSCCall  time.Duration
	getCurrentTimeStamp       func() int64
}

// NewSCQueryDebugger creates a new debugger instance used to monitor the SC query service metrics
func NewSCQueryDebugger(args ArgsSCQueryDebugger) (*scQueryDebugger, error) {
	timestampHandler := func() int64 {
		return time.Now().UnixNano()
	}

	return newSCQueryDebugger(args, timestampHandler)
}

func newSCQueryDebugger(args ArgsSCQueryDebugger, timestampHandler func() int64) (*scQueryDebugger, error) {
	err := checkArguments(args)
	if err != nil {
		return nil, fmt.Errorf("%w in NewSCQueryDebugger", err)
	}

	debugger := &scQueryDebugger{
		intervalAutoPrint:   time.Second * time.Duration(args.IntervalAutoPrintInSeconds),
		runningQueries:      make(map[int]int64),
		loggerInstance:      args.LoggerInstance,
		getCurrentTimeStamp: timestampHandler,
	}

	var ctx context.Context
	ctx, debugger.cancelFunc = context.WithCancel(context.Background())
	go debugger.processLoop(ctx)

	return debugger, nil
}

func checkArguments(args ArgsSCQueryDebugger) error {
	if args.IntervalAutoPrintInSeconds < minIntervalAutoPrint {
		return errInvalidIntervalAutoPrint
	}
	if check.IfNil(args.LoggerInstance) {
		return errNilLogger
	}

	return nil
}

func (debugger *scQueryDebugger) processLoop(ctx context.Context) {
	timer := time.NewTimer(debugger.intervalAutoPrint)
	defer timer.Stop()

	debugger.loggerInstance.Trace(logStartProcessLoop)
	for {
		timer.Reset(debugger.intervalAutoPrint)
		select {
		case <-ctx.Done():
			debugger.loggerInstance.Debug(logClosedProcessLoop)
			return
		case <-timer.C:
			debugger.printMetricsThenReset()
		}
	}
}

func (debugger *scQueryDebugger) printMetricsThenReset() {
	debugger.mutData.Lock()
	defer debugger.mutData.Unlock()

	debugger.recomputeMetrics()
	debugger.printMetricsIfRequired()
	debugger.resetMetrics()
}

func (debugger *scQueryDebugger) printMetricsIfRequired() {
	sumRelevantMetrics := debugger.maxNumOfConcurrentQueries + debugger.numFinishedCalls + len(debugger.runningQueries)
	if sumRelevantMetrics == 0 {
		return
	}

	debugger.loggerInstance.Debug(logMetrics,
		logMaxConcurrentQueries, debugger.maxNumOfConcurrentQueries,
		logDurationOfHeaviestCall, debugger.durationOfHeaviestSCCall,
		logIndexOfHeaviestCall, debugger.indexOfHeavierSCCall,
		logFinishedCalls, debugger.numFinishedCalls,
		logRunningCalls, len(debugger.runningQueries),
		logMetricsInterval, debugger.intervalAutoPrint,
	)
}

func (debugger *scQueryDebugger) resetMetrics() {
	debugger.maxNumOfConcurrentQueries = 0
	debugger.indexOfHeavierSCCall = 0
	debugger.durationOfHeaviestSCCall = 0
	debugger.numFinishedCalls = 0
}

// NotifyExecutionStarted will record the fact that the SC query service with the provided index has started the execution
// It also triggers the recomputing of the metrics
// Since multiple SC queries on the same instance are not allowed to be done in parallel, this components won't handle calls like:
//    NotifyExecutionStarted(1)
//    NotifyExecutionStarted(1)
//    NotifyExecutionFinished(1)
//    NotifyExecutionFinished(1)
func (debugger *scQueryDebugger) NotifyExecutionStarted(index int) {
	debugger.mutData.Lock()
	defer debugger.mutData.Unlock()

	debugger.runningQueries[index] = debugger.getCurrentTimeStamp()
	debugger.recomputeMetrics()
}

// NotifyExecutionFinished will record the fact that the SC query service with the provided index has finished the execution
// It also triggers the recomputing of the metrics
func (debugger *scQueryDebugger) NotifyExecutionFinished(index int) {
	debugger.mutData.Lock()
	defer debugger.mutData.Unlock()

	debugger.numFinishedCalls++
	debugger.recomputeMetrics()
	delete(debugger.runningQueries, index)
}

func (debugger *scQueryDebugger) recomputeMetrics() {
	currentLen := len(debugger.runningQueries)
	if currentLen > debugger.maxNumOfConcurrentQueries {
		debugger.maxNumOfConcurrentQueries = currentLen
	}

	currentTimestamp := debugger.getCurrentTimeStamp()
	for index, startTime := range debugger.runningQueries {
		scCallDuration := time.Duration(currentTimestamp - startTime)
		if scCallDuration > debugger.durationOfHeaviestSCCall {
			debugger.durationOfHeaviestSCCall = scCallDuration
			debugger.indexOfHeavierSCCall = index
		}
	}
}

// Close closes the go routine associated with the debugger
func (debugger *scQueryDebugger) Close() error {
	debugger.cancelFunc()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (debugger *scQueryDebugger) IsInterfaceNil() bool {
	return debugger == nil
}
