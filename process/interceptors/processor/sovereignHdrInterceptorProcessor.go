package processor

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

// ArgsSovereignHeaderInterceptorProcessor is a struct placeholder used to create a new sovereign extended header interceptor processor
type ArgsSovereignHeaderInterceptorProcessor struct {
	BlockBlackList           process.TimeCacher
	Hasher                   hashing.Hasher
	Marshaller               marshal.Marshalizer
	IncomingHeaderSubscriber process.IncomingHeaderSubscriber
	HeadersPool              dataRetriever.HeadersPool
}

type sovereignHeaderInterceptorProcessor struct {
	blackList                process.TimeCacher
	hasher                   hashing.Hasher
	marshaller               marshal.Marshalizer
	incomingHeaderSubscriber process.IncomingHeaderSubscriber
	headersPool              dataRetriever.HeadersPool
}

// NewSovereignHdrInterceptorProcessor creates a new sovereign extended header interceptor processor
func NewSovereignHdrInterceptorProcessor(args *ArgsSovereignHeaderInterceptorProcessor) (*sovereignHeaderInterceptorProcessor, error) {
	err := checkSovHdrProcArgs(args)
	if err != nil {
		return nil, err
	}

	return &sovereignHeaderInterceptorProcessor{
		blackList:                args.BlockBlackList,
		hasher:                   args.Hasher,
		marshaller:               args.Marshaller,
		incomingHeaderSubscriber: args.IncomingHeaderSubscriber,
		headersPool:              args.HeadersPool,
	}, nil
}

func checkSovHdrProcArgs(args *ArgsSovereignHeaderInterceptorProcessor) error {
	if args == nil {
		return process.ErrNilArgumentStruct
	}
	if check.IfNil(args.BlockBlackList) {
		return process.ErrNilBlackListCacher
	}
	if check.IfNil(args.Hasher) {
		return errors.ErrNilHasher
	}
	if check.IfNil(args.Marshaller) {
		return errors.ErrNilMarshalizer
	}
	if check.IfNil(args.IncomingHeaderSubscriber) {
		return errors.ErrNilIncomingHeaderSubscriber
	}
	if check.IfNil(args.HeadersPool) {
		return process.ErrNilHeadersDataPool
	}

	return nil
}

// Validate checks if the intercepted data can be processed
func (hip *sovereignHeaderInterceptorProcessor) Validate(data process.InterceptedData, _ core.PeerID) error {
	interceptedHdr, ok := data.(process.ExtendedHeaderValidatorHandler)
	if !ok {
		return fmt.Errorf("sovereignHeaderInterceptorProcessor.Validate error: %w", process.ErrWrongTypeAssertion)
	}

	hip.blackList.Sweep()
	isBlackListed := hip.blackList.Has(string(interceptedHdr.Hash()))
	if isBlackListed {
		return process.ErrHeaderIsBlackListed
	}

	return hip.validateReceivedHeader(interceptedHdr.GetExtendedHeader(), interceptedHdr.Hash())
}

func (hip *sovereignHeaderInterceptorProcessor) validateReceivedHeader(
	extendedHdr data.ShardHeaderExtendedHandler,
	hash []byte,
) error {
	computedExtendedHeader, err := hip.incomingHeaderSubscriber.CreateExtendedHeader(extendedHdr)
	if err != nil {
		return err
	}

	computedHeaderHash, err := core.CalculateHash(hip.marshaller, hip.hasher, computedExtendedHeader)
	if err != nil {
		return err
	}

	if !bytes.Equal(computedHeaderHash, hash) {
		return fmt.Errorf("%w, computed hash: %s, received hash: %s",
			errors.ErrInvalidReceivedExtendedShardHeader,
			hex.EncodeToString(computedHeaderHash),
			hex.EncodeToString(hash),
		)
	}

	return nil
}

// Save will save the received data into headers pool, if it doesn't already exist
func (hip *sovereignHeaderInterceptorProcessor) Save(data process.InterceptedData, _ core.PeerID, _ string) error {
	interceptedHdr, ok := data.(process.ExtendedHeaderValidatorHandler)
	if !ok {
		return fmt.Errorf("sovereignHeaderInterceptorProcessor.Save error: %w", process.ErrWrongTypeAssertion)
	}

	extendedHeaderNonce := interceptedHdr.GetExtendedHeader().GetNonce()
	extendedHeaderHash := interceptedHdr.Hash()
	_, hashes, err := hip.headersPool.GetHeadersByNonceAndShardId(extendedHeaderNonce, core.MainChainShardId)

	// do not add header again + create scrs and mbs if already received
	if err == nil && contains(hashes, extendedHeaderHash) {
		log.Debug("sovereignHeaderInterceptorProcessor.Save skipping already received extended header",
			"nonce", extendedHeaderNonce, "hash", hex.EncodeToString(extendedHeaderHash))
		return nil
	}

	return hip.incomingHeaderSubscriber.AddHeader(extendedHeaderHash, interceptedHdr.GetExtendedHeader())
}

func contains(hashes [][]byte, hash []byte) bool {
	for _, h := range hashes {
		if bytes.Equal(h, hash) {
			return true
		}
	}

	return false
}

// RegisterHandler does nothing
func (hip *sovereignHeaderInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (hip *sovereignHeaderInterceptorProcessor) IsInterfaceNil() bool {
	return hip == nil
}
