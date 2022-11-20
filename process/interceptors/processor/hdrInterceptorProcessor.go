package processor

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.InterceptorProcessor = (*HdrInterceptorProcessor)(nil)

// HdrInterceptorProcessor is the processor used when intercepting headers
// (shard headers, meta headers) structs which satisfy HeaderHandler interface.
type HdrInterceptorProcessor struct {
	headers            dataRetriever.HeadersPool
	blackList          process.TimeCacher
	registeredHandlers []func(topic string, hash []byte, data interface{})
	mutHandlers        sync.RWMutex
}

// NewHdrInterceptorProcessor creates a new TxInterceptorProcessor instance
func NewHdrInterceptorProcessor(argument *ArgHdrInterceptorProcessor) (*HdrInterceptorProcessor, error) {
	if argument == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(argument.Headers) {
		return nil, process.ErrNilCacher
	}
	if check.IfNil(argument.BlockBlackList) {
		return nil, process.ErrNilBlackListCacher
	}

	return &HdrInterceptorProcessor{
		headers:            argument.Headers,
		blackList:          argument.BlockBlackList,
		registeredHandlers: make([]func(topic string, hash []byte, data interface{}), 0),
	}, nil
}

// Validate checks if the intercepted data can be processed
func (hip *HdrInterceptorProcessor) Validate(data process.InterceptedData, _ core.PeerID) error {
	interceptedHdr, ok := data.(process.HdrValidatorHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	hip.blackList.Sweep()
	isBlackListed := hip.blackList.Has(string(interceptedHdr.Hash()))
	if isBlackListed {
		return process.ErrHeaderIsBlackListed
	}

	err := hip.checkDevnetHardfork(interceptedHdr.HeaderHandler())
	if err != nil {
		return err
	}

	return nil
}

func (hip *HdrInterceptorProcessor) checkDevnetHardfork(hdr data.HeaderHandler) error {
	if hdr.GetShardID() == 0 {
		return hip.handleShard0Hardfork(hdr)
	}
	if hdr.GetShardID() == 1 {
		return hip.handleShard1Hardfork(hdr)
	}
	if hdr.GetShardID() == 2 {
		return hip.handleShard2Hardfork(hdr)
	}

	return nil
}

func (hip *HdrInterceptorProcessor) handleShard0Hardfork(hdr data.HeaderHandler) error {
	low1 := uint64(3024122)
	high1 := uint64(3028318)
	if hip.isInExcludedRange(hdr, low1, high1) {
		return fmt.Errorf("header is in excluded range, shard %d, round %d, low %d, high %d",
			hdr.GetShardID(), hdr.GetRound(), low1, high1)
	}

	low2 := uint64(3387043)
	high2 := uint64(3404000)
	if hip.isInExcludedRange(hdr, low2, high2) {
		return fmt.Errorf("header is in excluded range, shard %d, round %d, low %d, high %d",
			hdr.GetShardID(), hdr.GetRound(), low2, high2)
	}

	return nil
}

func (hip *HdrInterceptorProcessor) handleShard1Hardfork(hdr data.HeaderHandler) error {
	low1 := uint64(3024122)
	high1 := uint64(3028322)
	if hip.isInExcludedRange(hdr, low1, high1) {
		return fmt.Errorf("header is in excluded range, shard %d, round %d, low %d, high %d",
			hdr.GetShardID(), hdr.GetRound(), low1, high1)
	}

	low2 := uint64(3387043)
	high2 := uint64(3404000)
	if hip.isInExcludedRange(hdr, low2, high2) {
		return fmt.Errorf("header is in excluded range, shard %d, round %d, low %d, high %d",
			hdr.GetShardID(), hdr.GetRound(), low2, high2)
	}

	return nil
}

func (hip *HdrInterceptorProcessor) handleShard2Hardfork(hdr data.HeaderHandler) error {
	low1 := uint64(3024122)
	high1 := uint64(3028326)
	if hip.isInExcludedRange(hdr, low1, high1) {
		return fmt.Errorf("header is in excluded range, shard %d, round %d, low %d, high %d",
			hdr.GetShardID(), hdr.GetRound(), low1, high1)
	}

	low2 := uint64(3387043)
	high2 := uint64(3404000)
	if hip.isInExcludedRange(hdr, low2, high2) {
		return fmt.Errorf("header is in excluded range, shard %d, round %d, low %d, high %d",
			hdr.GetShardID(), hdr.GetRound(), low2, high2)
	}

	return nil
}

func (hip *HdrInterceptorProcessor) isInExcludedRange(hdr data.HeaderHandler, low uint64, high uint64) bool {
	round := hdr.GetRound()
	return round >= low && round <= high && low <= high
}

// Save will save the received data into the headers cacher as hash<->[plain header structure]
// and in headersNonces as nonce<->hash
func (hip *HdrInterceptorProcessor) Save(data process.InterceptedData, _ core.PeerID, topic string) error {
	interceptedHdr, ok := data.(process.HdrValidatorHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	go hip.notify(interceptedHdr.HeaderHandler(), interceptedHdr.Hash(), topic)

	hip.headers.AddHeader(interceptedHdr.Hash(), interceptedHdr.HeaderHandler())

	return nil
}

// RegisterHandler registers a callback function to be notified of incoming headers
func (hip *HdrInterceptorProcessor) RegisterHandler(handler func(topic string, hash []byte, data interface{})) {
	if handler == nil {
		return
	}

	hip.mutHandlers.Lock()
	hip.registeredHandlers = append(hip.registeredHandlers, handler)
	hip.mutHandlers.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (hip *HdrInterceptorProcessor) IsInterfaceNil() bool {
	return hip == nil
}

func (hip *HdrInterceptorProcessor) notify(header data.HeaderHandler, hash []byte, topic string) {
	hip.mutHandlers.RLock()
	for _, handler := range hip.registeredHandlers {
		handler(topic, hash, header)
	}
	hip.mutHandlers.RUnlock()
}
