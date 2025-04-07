package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/interceptedBlocks"
	"github.com/multiversx/mx-chain-go/sharding"
)

var _ process.InterceptedDataFactory = (*interceptedMetaHeaderDataFactory)(nil)

// ArgInterceptedMetaHeaderFactory is the DTO used to create a new instance of meta header factory
type ArgInterceptedMetaHeaderFactory struct {
	ArgInterceptedDataFactory
	ProofsPool process.ProofsPool
}

type interceptedMetaHeaderDataFactory struct {
	marshalizer             marshal.Marshalizer
	hasher                  hashing.Hasher
	shardCoordinator        sharding.Coordinator
	headerSigVerifier       process.InterceptedHeaderSigVerifier
	headerIntegrityVerifier process.HeaderIntegrityVerifier
	validityAttester        process.ValidityAttester
	epochStartTrigger       process.EpochStartTriggerHandler
	enableEpochsHandler     common.EnableEpochsHandler
	proofsPool              process.ProofsPool
	fieldsSizeChecker       common.FieldsSizeChecker
}

// NewInterceptedMetaHeaderDataFactory creates an instance of interceptedMetaHeaderDataFactory
func NewInterceptedMetaHeaderDataFactory(argument *ArgInterceptedMetaHeaderFactory) (*interceptedMetaHeaderDataFactory, error) {
	if argument == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(argument.CoreComponents) {
		return nil, process.ErrNilCoreComponentsHolder
	}
	if check.IfNil(argument.CoreComponents.InternalMarshalizer()) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(argument.CoreComponents.TxMarshalizer()) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(argument.CoreComponents.Hasher()) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(argument.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(argument.HeaderSigVerifier) {
		return nil, process.ErrNilHeaderSigVerifier
	}
	if check.IfNil(argument.HeaderIntegrityVerifier) {
		return nil, process.ErrNilHeaderIntegrityVerifier
	}
	if check.IfNil(argument.EpochStartTrigger) {
		return nil, process.ErrNilEpochStartTrigger
	}
	if len(argument.CoreComponents.ChainID()) == 0 {
		return nil, process.ErrInvalidChainID
	}
	if check.IfNil(argument.ValidityAttester) {
		return nil, process.ErrNilValidityAttester
	}
	if check.IfNil(argument.ProofsPool) {
		return nil, process.ErrNilProofsPool
	}

	return &interceptedMetaHeaderDataFactory{
		marshalizer:             argument.CoreComponents.InternalMarshalizer(),
		hasher:                  argument.CoreComponents.Hasher(),
		shardCoordinator:        argument.ShardCoordinator,
		headerSigVerifier:       argument.HeaderSigVerifier,
		headerIntegrityVerifier: argument.HeaderIntegrityVerifier,
		validityAttester:        argument.ValidityAttester,
		epochStartTrigger:       argument.EpochStartTrigger,
		enableEpochsHandler:     argument.CoreComponents.EnableEpochsHandler(),
		proofsPool:              argument.ProofsPool,
		fieldsSizeChecker:       argument.CoreComponents.FieldsSizeChecker(),
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (imhdf *interceptedMetaHeaderDataFactory) Create(buff []byte, _ core.PeerID) (process.InterceptedData, error) {
	arg := &interceptedBlocks.ArgInterceptedBlockHeader{
		HdrBuff:                 buff,
		Marshalizer:             imhdf.marshalizer,
		Hasher:                  imhdf.hasher,
		ShardCoordinator:        imhdf.shardCoordinator,
		HeaderSigVerifier:       imhdf.headerSigVerifier,
		HeaderIntegrityVerifier: imhdf.headerIntegrityVerifier,
		ValidityAttester:        imhdf.validityAttester,
		EpochStartTrigger:       imhdf.epochStartTrigger,
		EnableEpochsHandler:     imhdf.enableEpochsHandler,
		ProofsPool:              imhdf.proofsPool,
		FieldsSizeChecker:       imhdf.fieldsSizeChecker,
	}

	return interceptedBlocks.NewInterceptedMetaHeader(arg)
}

// IsInterfaceNil returns true if there is no value under the interface
func (imhdf *interceptedMetaHeaderDataFactory) IsInterfaceNil() bool {
	return imhdf == nil
}
