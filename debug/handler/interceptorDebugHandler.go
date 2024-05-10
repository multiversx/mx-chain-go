package handler

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/debug"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const requestEvent = "request"
const resolveEvent = "resolve"
const minIntervalInSeconds = 1
const minThresholdResolve = 1
const minThresholdRequests = 1
const minDebugLineExpiration = 1
const newLineChar = "\n"
const numIntsInEventStruct = 6
const intSize = 8
const maxKeysToDisplay = 200

var log = logger.GetOrCreate("debug/handler")

type event struct {
	mutEvent     sync.RWMutex
	eventType    string
	hash         []byte
	topic        string
	numReqIntra  int
	numReqCross  int
	numReceived  int
	numProcessed int
	lastErr      error
	numPrints    int
	timestamp    int64
}

// NumPrints returns the current number of prints
func (ev *event) NumPrints() int {
	ev.mutEvent.RLock()
	defer ev.mutEvent.RUnlock()

	return ev.numPrints
}

// Size returns the number of bytes taken by an event line
func (ev *event) Size() int {
	ev.mutEvent.RLock()
	defer ev.mutEvent.RUnlock()

	size := len(ev.eventType) + len(ev.hash) + len(ev.topic) + numIntsInEventStruct*intSize
	if ev.lastErr != nil {
		size += len(ev.lastErr.Error())
	}

	return size
}

func (ev *event) String() string {
	ev.mutEvent.RLock()
	defer ev.mutEvent.RUnlock()

	strErr := ""
	if ev.lastErr != nil {
		strErr = ev.lastErr.Error()
	}

	return fmt.Sprintf("type: %s, topic: %s, hash: %s, numReqIntra: %d, numReqCross: %d, "+
		"numReceived: %d, numProcessed: %d, last err: %s, last query time: %s ",
		ev.eventType,
		ev.topic,
		logger.DisplayByteSlice(ev.hash),
		ev.numReqIntra,
		ev.numReqCross,
		ev.numReceived,
		ev.numProcessed,
		strErr,
		displayTime(ev.timestamp),
	)
}

func displayTime(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04:05.000")
}

type interceptorDebugHandler struct {
	cache                storage.Cacher
	intervalAutoPrint    time.Duration
	requestsThreshold    int
	resolveFailThreshold int
	maxNumPrints         int
	printEventFunc       func(data string)
	timestampHandler     func() int64
	cancelFunc           context.CancelFunc
}

// NewInterceptorDebugHandler creates a new interceptorDebugHandler able to hold requested-intercepted information
func NewInterceptorDebugHandler(config config.InterceptorResolverDebugConfig) (*interceptorDebugHandler, error) {
	lruCache, err := cache.NewLRUCache(config.CacheSize)
	if err != nil {
		return nil, fmt.Errorf("%w when creating NewInterceptorDebugHandler", err)
	}

	idh := &interceptorDebugHandler{
		cache:            lruCache,
		timestampHandler: getCurrentTimeStamp,
	}

	err = idh.parseConfig(config)
	if err != nil {
		return nil, err
	}

	idh.printEventFunc = idh.printEvent
	if config.EnablePrint {
		ctx, cancelFunc := context.WithCancel(context.Background())
		idh.cancelFunc = cancelFunc
		go idh.printContinuously(ctx)
	}

	return idh, nil
}

func getCurrentTimeStamp() int64 {
	return time.Now().Unix()
}

func (idh *interceptorDebugHandler) parseConfig(config config.InterceptorResolverDebugConfig) error {
	if !config.EnablePrint {
		return nil
	}
	if config.IntervalAutoPrintInSeconds < minIntervalInSeconds {
		return fmt.Errorf("%w for IntervalAutoPrintInSeconds, minimum is %d", debug.ErrInvalidValue, minIntervalInSeconds)
	}
	if config.NumRequestsThreshold < minThresholdRequests {
		return fmt.Errorf("%w for NumRequestsThreshold, minimum is %d", debug.ErrInvalidValue, minThresholdRequests)
	}
	if config.NumResolveFailureThreshold < minThresholdResolve {
		return fmt.Errorf("%w for NumResolveFailureThreshold, minimum is %d", debug.ErrInvalidValue, minThresholdResolve)
	}
	if config.DebugLineExpiration < minDebugLineExpiration {
		return fmt.Errorf("%w for DebugLineExpiration, minimum is %d", debug.ErrInvalidValue, minDebugLineExpiration)
	}

	idh.intervalAutoPrint = time.Second * time.Duration(config.IntervalAutoPrintInSeconds)
	idh.requestsThreshold = config.NumRequestsThreshold
	idh.resolveFailThreshold = config.NumResolveFailureThreshold
	idh.maxNumPrints = config.DebugLineExpiration

	return nil
}

func (idh *interceptorDebugHandler) printContinuously(ctx context.Context) {
	for {
		select {
		case <-time.After(idh.intervalAutoPrint):
			idh.incrementNumOfPrints()

			events := []string{"Requests pending and resolver fails:"}
			events = append(events, idh.getStringEvents(idh.maxNumPrints)...)
			if len(events) == 1 {
				continue
			}

			stringEvent := strings.Join(events, newLineChar)
			idh.printEventFunc(stringEvent)
		case <-ctx.Done():
			return
		}
	}
}

func (idh *interceptorDebugHandler) printEvent(data string) {
	log.Debug(data)
}

func (idh *interceptorDebugHandler) incrementNumOfPrints() {
	keys := idh.cache.Keys()
	for _, key := range keys {
		obj, ok := idh.cache.Get(key)
		if !ok {
			continue
		}

		ev, ok := obj.(*event)
		if !ok {
			continue
		}

		ev.mutEvent.Lock()
		ev.numPrints++
		ev.mutEvent.Unlock()
	}
}

// TODO replace this with a call to Query(search) when a suitable conditional parser will be used. Also replace config parameters
// with a query string so it will be more extensible
func (idh *interceptorDebugHandler) getStringEvents(maxNumPrints int) []string {
	acceptEvent := func(ev *event) bool {
		ev.mutEvent.RLock()
		defer ev.mutEvent.RUnlock()

		shouldAcceptRequested := ev.eventType == requestEvent && ev.numReqCross+ev.numReqIntra >= idh.requestsThreshold
		shouldAcceptResolved := ev.eventType == resolveEvent && ev.numReceived >= idh.resolveFailThreshold

		return shouldAcceptRequested || shouldAcceptResolved
	}

	return idh.query(acceptEvent, maxNumPrints)
}

// LogRequestedData is called whenever hashes have been requested
func (idh *interceptorDebugHandler) LogRequestedData(topic string, hashes [][]byte, numReqIntra int, numReqCross int) {
	for _, hash := range hashes {
		idh.logRequestedData(topic, hash, numReqIntra, numReqCross)
	}
}

func (idh *interceptorDebugHandler) logRequestedData(topic string, hash []byte, numReqIntra int, numReqCross int) {
	identifier := idh.computeIdentifier(requestEvent, topic, hash)

	obj, ok := idh.cache.Get(identifier)
	if !ok {
		req := &event{
			hash:         hash,
			eventType:    requestEvent,
			topic:        topic,
			numReqIntra:  numReqIntra,
			numReqCross:  numReqCross,
			numReceived:  0,
			numProcessed: 0,
			lastErr:      nil,
			timestamp:    idh.timestampHandler(),
		}
		idh.cache.Put(identifier, req, req.Size())

		return
	}

	req, ok := obj.(*event)
	if !ok {
		return
	}

	req.mutEvent.Lock()
	req.numReqCross += numReqCross
	req.numReqIntra += numReqIntra
	req.timestamp = idh.timestampHandler()
	req.mutEvent.Unlock()
}

// LogReceivedHashes is called whenever request hashes have been received
func (idh *interceptorDebugHandler) LogReceivedHashes(topic string, hashes [][]byte) {
	for _, hash := range hashes {
		idh.logReceivedHash(topic, hash)
	}
}

func (idh *interceptorDebugHandler) logReceivedHash(topic string, hash []byte) {
	identifier := idh.computeIdentifier(requestEvent, topic, hash)

	obj, ok := idh.cache.Get(identifier)
	if !ok {
		return
	}

	req, ok := obj.(*event)
	if !ok {
		return
	}

	req.mutEvent.Lock()
	req.numReceived++
	req.timestamp = idh.timestampHandler()
	req.mutEvent.Unlock()
}

// LogProcessedHashes is called whenever request hashes have been processed
func (idh *interceptorDebugHandler) LogProcessedHashes(topic string, hashes [][]byte, err error) {
	for _, hash := range hashes {
		idh.logProcessedHash(topic, hash, err)
	}
}

func (idh *interceptorDebugHandler) logProcessedHash(topic string, hash []byte, err error) {
	identifier := idh.computeIdentifier(requestEvent, topic, hash)

	obj, ok := idh.cache.Get(identifier)
	if !ok {
		return
	}

	req, ok := obj.(*event)
	if !ok {
		return
	}

	if err != nil {
		req.mutEvent.Lock()
		req.numProcessed++
		req.timestamp = idh.timestampHandler()
		req.lastErr = err
		req.mutEvent.Unlock()

		return
	}

	idh.cache.Remove(identifier)
}

func (idh *interceptorDebugHandler) computeIdentifier(eventType string, topic string, hash []byte) []byte {
	return append([]byte(eventType+topic), hash...)
}

// Query returns active requests in a string-ified format having the topic provided
// * will return each and every data
func (idh *interceptorDebugHandler) Query(search string) []string {
	acceptEvent := func(ev *event) bool {
		//TODO replace this rudimentary search pattern with something like
		// github.com/oleksandr/conditions
		ev.mutEvent.RLock()
		defer ev.mutEvent.RUnlock()
		return search == "*" || search == ev.topic
	}

	maxNumPrints := math.MaxInt32
	return idh.query(acceptEvent, maxNumPrints)
}

func (idh *interceptorDebugHandler) query(acceptEvent func(ev *event) bool, maxNumPrints int) []string {
	keys := idh.cache.Keys()

	events := make([]string, 0, len(keys))
	trimmed := false
	for _, key := range keys {
		obj, ok := idh.cache.Get(key)
		if !ok {
			continue
		}

		ev, ok := obj.(*event)
		if !ok {
			continue
		}

		if ev.NumPrints() > maxNumPrints {
			continue
		}

		if !acceptEvent(ev) {
			continue
		}

		events = append(events, ev.String())
		if len(events) == maxKeysToDisplay {
			trimmed = true
			break
		}
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i] < events[j]
	})

	if trimmed {
		events = append(events, fmt.Sprintf("[Not all keys are shown here. Showing %d]", len(events)))
	}

	return events
}

// LogFailedToResolveData adds a record stating that the resolver was unable to process the data
func (idh *interceptorDebugHandler) LogFailedToResolveData(topic string, hash []byte, err error) {
	identifier := idh.computeIdentifier(resolveEvent, topic, hash)

	obj, ok := idh.cache.Get(identifier)
	if !ok {
		req := &event{
			hash:         hash,
			eventType:    resolveEvent,
			topic:        topic,
			numReqIntra:  0,
			numReqCross:  0,
			numReceived:  1,
			numProcessed: 0,
			lastErr:      err,
			timestamp:    idh.timestampHandler(),
		}
		idh.cache.Put(identifier, req, req.Size())

		return
	}

	ev, ok := obj.(*event)
	if !ok {
		return
	}

	ev.mutEvent.Lock()
	ev.numReceived++
	ev.timestamp = idh.timestampHandler()
	ev.lastErr = err
	ev.mutEvent.Unlock()
}

// LogSucceededToResolveData removes the recording that the resolver did not resolve a hash in the past
func (idh *interceptorDebugHandler) LogSucceededToResolveData(topic string, hash []byte) {
	identifier := idh.computeIdentifier(resolveEvent, topic, hash)

	idh.cache.Remove(identifier)
}

// Close closes all underlying components
func (idh *interceptorDebugHandler) Close() error {
	idh.cancelFunc()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (idh *interceptorDebugHandler) IsInterfaceNil() bool {
	return idh == nil
}
