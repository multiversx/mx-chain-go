package resolver

import (
	"fmt"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

type request struct {
	hash         []byte
	topic        string
	numReqIntra  int
	numReqCross  int
	numReceived  int
	numProcessed int
	lastErr      error
}

type interceptorResolver struct {
	mutCriticalArea sync.Mutex
	cache           *lrucache.LRUCache
}

// NewInterceptorResolver creates a new interceptorResolver able to hold requested-intercepted information
func NewInterceptorResolver(size int) (*interceptorResolver, error) {
	cache, err := lrucache.NewCache(size)
	if err != nil {
		return nil, fmt.Errorf("%w when creating NewInterceptorResolver", err)
	}

	return &interceptorResolver{
		cache: cache,
	}, nil
}

// RequestedData is called whenever a hash has been requested
func (ir *interceptorResolver) RequestedData(topic string, hash []byte, numReqIntra int, numReqCross int) {
	identifier := ir.computeIdentifier(topic, hash)

	ir.mutCriticalArea.Lock()
	defer ir.mutCriticalArea.Unlock()

	obj, ok := ir.cache.Get(identifier)
	if !ok {
		req := &request{
			hash:         hash,
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

	req, ok := obj.(*request)
	if !ok {
		return
	}

	req.numReqCross += numReqCross
	req.numReqIntra += numReqIntra
	ir.cache.Put(identifier, req)
}

// ReceivedHash is called whenever a request hash has been received
func (ir *interceptorResolver) ReceivedHash(topic string, hash []byte) {
	identifier := ir.computeIdentifier(topic, hash)

	ir.mutCriticalArea.Lock()
	defer ir.mutCriticalArea.Unlock()

	obj, ok := ir.cache.Get(identifier)
	if !ok {
		return
	}

	req, ok := obj.(*request)
	if !ok {
		return
	}

	req.numReceived++
	ir.cache.Put(identifier, req)
}

// ProcessedHash is called whenever a request hash has been processed
func (ir *interceptorResolver) ProcessedHash(topic string, hash []byte, err error) {
	identifier := ir.computeIdentifier(topic, hash)

	ir.mutCriticalArea.Lock()
	defer ir.mutCriticalArea.Unlock()

	obj, ok := ir.cache.Get(identifier)
	if !ok {
		return
	}

	req, ok := obj.(*request)
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

func (ir *interceptorResolver) computeIdentifier(topic string, hash []byte) []byte {
	return append([]byte(topic), hash...)
}

// Query returns active requests in a string-ified format having the topic provided
// * will return each and every data
func (ir *interceptorResolver) Query(topic string) []string {
	ir.mutCriticalArea.Lock()
	defer ir.mutCriticalArea.Unlock()

	keys := ir.cache.Keys()
	stringRequests := make([]string, 0, len(keys))
	for _, key := range keys {
		obj, ok := ir.cache.Get(key)
		if !ok {
			continue
		}

		req, ok := obj.(*request)
		if !ok {
			continue
		}

		topicMatches := topic == "*" || topic == req.topic
		if !topicMatches {
			continue
		}

		strErr := ""
		if req.lastErr != nil {
			strErr = req.lastErr.Error()
		}
		strReq := fmt.Sprintf("hash: %s, topic: %s, numReqIntra: %d, numReqCross: %d, "+
			"numReceived: %d, numProcessed: %d, last err: %s",
			logger.DisplayByteSlice(req.hash),
			req.topic,
			req.numReqIntra,
			req.numReqCross,
			req.numReceived,
			req.numProcessed,
			strErr,
		)

		stringRequests = append(stringRequests, strReq)
	}

	return stringRequests
}

// IsInterfaceNil returns true if there is no value under the interface
func (ir *interceptorResolver) IsInterfaceNil() bool {
	return ir == nil
}
