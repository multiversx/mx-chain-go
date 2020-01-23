package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type interceptedMetaHeaderDataFactory struct {
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	shardCoordinator  sharding.Coordinator
	headerSigVerifier process.InterceptedHeaderSigVerifier
	chainID           []byte
	finalityAttester  process.FinalityAttester
}

// NewInterceptedMetaHeaderDataFactory creates an instance of interceptedMetaHeaderDataFactory
func NewInterceptedMetaHeaderDataFactory(argument *ArgInterceptedDataFactory) (*interceptedMetaHeaderDataFactory, error) {
	if argument == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(argument.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(argument.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(argument.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(argument.HeaderSigVerifier) {
		return nil, process.ErrNilHeaderSigVerifier
	}
	if len(argument.ChainID) == 0 {
		return nil, process.ErrInvalidChainID
	}

	return &interceptedMetaHeaderDataFactory{
		marshalizer:       argument.Marshalizer,
		hasher:            argument.Hasher,
		shardCoordinator:  argument.ShardCoordinator,
		headerSigVerifier: argument.HeaderSigVerifier,
		chainID:           argument.ChainID,
		finalityAttester:  &nilFinalityAttester{},
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (imhdf *interceptedMetaHeaderDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	arg := &interceptedBlocks.ArgInterceptedBlockHeader{
		HdrBuff:           buff,
		Marshalizer:       imhdf.marshalizer,
		Hasher:            imhdf.hasher,
		ShardCoordinator:  imhdf.shardCoordinator,
		HeaderSigVerifier: imhdf.headerSigVerifier,
		ChainID:           imhdf.chainID,
		FinalityAttester:  imhdf.finalityAttester,
	}

	return interceptedBlocks.NewInterceptedMetaHeader(arg)
}

// SetFinalityAttester sets the finality attester
func (imhdf *interceptedMetaHeaderDataFactory) SetFinalityAttester(attester process.FinalityAttester) error {
	if check.IfNil(attester) {
		return process.ErrNilFinalityAttester
	}

	imhdf.finalityAttester = attester
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (imhdf *interceptedMetaHeaderDataFactory) IsInterfaceNil() bool {
	return imhdf == nil
}
