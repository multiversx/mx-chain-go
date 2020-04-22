package resolver

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

const requestEvent = "request"
const resolveEvent = "resolve"
const minIntervalInSeconds = 1
const minThresholdResolve = 1
const minThresholdRequests = 1
const minDebugLineExpiration = 1
const newLineChar = "\n"

var log = logger.GetOrCreate("debug/resolver")

type event struct {
	eventType    string
	hash         []byte
	topic        string
	numReqIntra  int
	numReqCross  int
	numReceived  int
	numProcessed int
	lastErr      error
	numPrints    int
}

func (ev *event) String() string {
	strErr := ""
	if ev.lastErr != nil {
		strErr = ev.lastErr.Error()
	}
	return fmt.Sprintf("type: %s, topic: %s, hash: %s, numReqIntra: %d, numReqCross: %d, "+
		"numReceived: %d, numProcessed: %d, last err: %s",
		ev.eventType,
		ev.topic,
		logger.DisplayByteSlice(ev.hash),
		ev.numReqIntra,
		ev.numReqCross,
		ev.numReceived,
		ev.numProcessed,
		strErr,
	)
}

type interceptorResolver struct {
	mutCriticalArea      sync.RWMutex
	cache                *lrucache.LRUCache
	intervalAutoPrint    time.Duration
	requestsThreshold    int
	resolveFailThreshold int
	maxNumPrints         int
	printEventHandler    func(data string)
}

// NewInterceptorResolver creates a new interceptorResolver able to hold requested-intercepted information
func NewInterceptorResolver(config config.InterceptorResolverDebugConfig) (*interceptorResolver, error) {
	cache, err := lrucache.NewCache(config.CacheSize)
	if err != nil {
		return nil, fmt.Errorf("%w when creating NewInterceptorResolver", err)
	}

	ir := &interceptorResolver{
		cache: cache,
	}

	err = ir.parseConfig(config)
	if err != nil {
		return nil, err
	}

	ir.printEventHandler = ir.printEvent
	if config.EnableAutoPrint {
		go ir.autoPrint()
	}

	return ir, nil
}

func (ir *interceptorResolver) parseConfig(config config.InterceptorResolverDebugConfig) error {
	if !config.EnableAutoPrint {
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

	ir.intervalAutoPrint = time.Second * time.Duration(config.IntervalAutoPrintInSeconds)
	ir.requestsThreshold = config.NumRequestsThreshold
	ir.resolveFailThreshold = config.NumResolveFailureThreshold
	ir.maxNumPrints = config.DebugLineExpiration

	return nil
}

func (ir *interceptorResolver) autoPrint() {
	for {
		time.Sleep(ir.intervalAutoPrint)

		ir.incrementNumOfPrints()

		events := []string{"Requests pending and resolver fails:"}
		events = append(events, ir.getStringEvents(ir.maxNumPrints)...)
		if len(events) == 1 {
			continue
		}

		stringEvent := strings.Join(events, newLineChar)
		ir.printEventHandler(stringEvent)
	}
}

func (ir *interceptorResolver) printEvent(data string) {
	log.Debug(data)
}

func (ir *interceptorResolver) incrementNumOfPrints() {
	ir.mutCriticalArea.Lock()
	defer ir.mutCriticalArea.Unlock()

	keys := ir.cache.Keys()
	for _, key := range keys {
		obj, ok := ir.cache.Get(key)
		if !ok {
			continue
		}

		ev, ok := obj.(*event)
		if !ok {
			continue
		}

		ev.numPrints++
		ir.cache.Put(key, ev)
	}
}

//TODO replace this with a call to Query(search) when a suitable conditional parser will be used. Also replace config parameters
// with a query string so it will be more extensible
func (ir *interceptorResolver) getStringEvents(maxNumPrints int) []string {
	acceptEvent := func(ev *event) bool {
		shouldAcceptRequested := ev.eventType == requestEvent && ev.numReqCross+ev.numReqIntra >= ir.requestsThreshold
		shouldAcceptResolved := ev.eventType == resolveEvent && ev.numReceived >= ir.resolveFailThreshold

		return shouldAcceptRequested || shouldAcceptResolved
	}

	return ir.query(acceptEvent, maxNumPrints)
}

// LogRequestedData is called whenever a hash has been requested
func (ir *interceptorResolver) LogRequestedData(topic string, hash []byte, numReqIntra int, numReqCross int) {
	identifier := ir.computeIdentifier(requestEvent, topic, hash)

	ir.mutCriticalArea.Lock()
	defer ir.mutCriticalArea.Unlock()

	obj, ok := ir.cache.Get(identifier)
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
		}
		ir.cache.Put(identifier, req)

		return
	}

	req, ok := obj.(*event)
	if !ok {
		return
	}

	req.numReqCross += numReqCross
	req.numReqIntra += numReqIntra
	ir.cache.Put(identifier, req)
}

// LogReceivedHash is called whenever a request hash has been received
func (ir *interceptorResolver) LogReceivedHash(topic string, hash []byte) {
	identifier := ir.computeIdentifier(requestEvent, topic, hash)

	ir.mutCriticalArea.Lock()
	defer ir.mutCriticalArea.Unlock()

	obj, ok := ir.cache.Get(identifier)
	if !ok {
		return
	}

	req, ok := obj.(*event)
	if !ok {
		return
	}

	req.numReceived++
	ir.cache.Put(identifier, req)
}

// LogProcessedHash is called whenever a request hash has been processed
func (ir *interceptorResolver) LogProcessedHash(topic string, hash []byte, err error) {
	identifier := ir.computeIdentifier(requestEvent, topic, hash)

	ir.mutCriticalArea.Lock()
	defer ir.mutCriticalArea.Unlock()

	obj, ok := ir.cache.Get(identifier)
	if !ok {
		return
	}

	req, ok := obj.(*event)
	if !ok {
		return
	}

	if err != nil {
		req.numProcessed++
		req.lastErr = err
		ir.cache.Put(identifier, req)

		return
	}

	ir.cache.Remove(identifier)
}

// Enabled returns if this implementation should process data or not. Always true
func (ir *interceptorResolver) Enabled() bool {
	return true
}

func (ir *interceptorResolver) computeIdentifier(eventType string, topic string, hash []byte) []byte {
	return append([]byte(eventType+topic), hash...)
}

// Query returns active requests in a string-ified format having the topic provided
// * will return each and every data
func (ir *interceptorResolver) Query(search string) []string {
	acceptEvent := func(ev *event) bool {
		//TODO replace this rudimentary search pattern with something like
		// github.com/oleksandr/conditions
		return search == "*" || search == ev.topic
	}

	maxNumPrints := math.MaxInt32
	return ir.query(acceptEvent, maxNumPrints)
}

func (ir *interceptorResolver) query(acceptEvent func(ev *event) bool, maxNumPrints int) []string {
	ir.mutCriticalArea.RLock()
	defer ir.mutCriticalArea.RUnlock()

	keys := ir.cache.Keys()
	events := make([]string, 0, len(keys))
	for _, key := range keys {
		obj, ok := ir.cache.Get(key)
		if !ok {
			continue
		}

		ev, ok := obj.(*event)
		if !ok {
			continue
		}

		if ev.numPrints > maxNumPrints {
			continue
		}

		if !acceptEvent(ev) {
			continue
		}

		events = append(events, ev.String())
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i] < events[j]
	})

	return events
}

// LogFailedToResolveData adds a record stating that the resolver was unable to process the data
func (ir *interceptorResolver) LogFailedToResolveData(topic string, hash []byte, err error) {
	identifier := ir.computeIdentifier(resolveEvent, topic, hash)

	ir.mutCriticalArea.Lock()
	defer ir.mutCriticalArea.Unlock()

	obj, ok := ir.cache.Get(identifier)
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
		}
		ir.cache.Put(identifier, req)

		return
	}

	ev, ok := obj.(*event)
	if !ok {
		return
	}

	ev.numReceived++
	ev.lastErr = err
	ir.cache.Put(identifier, ev)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ir *interceptorResolver) IsInterfaceNil() bool {
	return ir == nil
}
