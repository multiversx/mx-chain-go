package processor

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
)

// IncomingHeaderSubscriber defines a subscriber to incoming headers
type IncomingHeaderSubscriber interface {
	AddHeader(headerHash []byte, header sovereign.IncomingHeaderHandler) error
	IsInterfaceNil() bool
}

type ArgsSovereignHeaderInterceptorProcessor struct {
	Headers                  dataRetriever.HeadersPool
	BlockBlackList           process.TimeCacher
	Hasher                   hashing.Hasher
	Marshaller               marshal.Marshalizer
	IncomingHeaderSubscriber IncomingHeaderSubscriber
}

type sovereignHeaderInterceptorProcessor struct {
	headers                  dataRetriever.HeadersPool
	blackList                process.TimeCacher
	registeredHandlers       []func(topic string, hash []byte, data interface{})
	mutHandlers              sync.RWMutex
	Hasher                   hashing.Hasher
	Marshaller               marshal.Marshalizer
	IncomingHeaderSubscriber IncomingHeaderSubscriber
}

func NewSovereignHdrInterceptorProcessor(argument *ArgsSovereignHeaderInterceptorProcessor) (*sovereignHeaderInterceptorProcessor, error) {
	if argument == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(argument.Headers) {
		return nil, process.ErrNilCacher
	}
	if check.IfNil(argument.BlockBlackList) {
		return nil, process.ErrNilBlackListCacher
	}

	return &sovereignHeaderInterceptorProcessor{
		headers:                  argument.Headers,
		blackList:                argument.BlockBlackList,
		registeredHandlers:       make([]func(topic string, hash []byte, data interface{}), 0),
		Hasher:                   argument.Hasher,
		Marshaller:               argument.Marshaller,
		IncomingHeaderSubscriber: argument.IncomingHeaderSubscriber,
	}, nil
}

// Validate checks if the intercepted data can be processed
func (hip *sovereignHeaderInterceptorProcessor) Validate(data process.InterceptedData, _ core.PeerID) error {
	interceptedHdr, ok := data.(sovereign.Proof)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	//hip.blackList.Sweep()
	//isBlackListed := hip.blackList.Has(string(interceptedHdr.Hash()))
	//if isBlackListed {
	//	return process.ErrHeaderIsBlackListed
	//}

	_ = interceptedHdr
	return nil
}

// Save will save the received data into the headers cacher as hash<->[plain header structure]
// and in headersNonces as nonce<->hash
func (hip *sovereignHeaderInterceptorProcessor) Save(data process.InterceptedData, _ core.PeerID, topic string) error {
	interceptedHdr, ok := data.(sovereign.Proof)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	hdr, ok := interceptedHdr.(*block.ShardHeaderExtended)
	if !ok {
		return fmt.Errorf("%w for ShardHeaderExtended", process.ErrWrongTypeAssertion)
	}

	hash, _ := core.CalculateHash(hip.Marshaller, hip.Hasher, hdr)

	return hip.IncomingHeaderSubscriber.AddHeader(hash, hdr)
}

// RegisterHandler registers a callback function to be notified of incoming headers
func (hip *sovereignHeaderInterceptorProcessor) RegisterHandler(handler func(topic string, hash []byte, data interface{})) {

}

// IsInterfaceNil returns true if there is no value under the interface
func (hip *sovereignHeaderInterceptorProcessor) IsInterfaceNil() bool {
	return hip == nil
}
