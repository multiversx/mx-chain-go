package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/process/unsigned"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type shardInterceptedDataFactory struct {
	marshalizer         marshal.Marshalizer
	hasher              hashing.Hasher
	keyGen              crypto.KeyGenerator
	singleSigner        crypto.SingleSigner
	addrConverter       state.AddressConverter
	shardCoordinator    sharding.Coordinator
	interceptedDataType InterceptedDataType
	multiSigVerifier    crypto.MultiSigVerifier
	chronologyValidator process.ChronologyValidator
}

// NewShardInterceptedDataFactory creates an instance of interceptedDataFactory that can create
// instances of process.InterceptedData and is used on shard nodes
func NewShardInterceptedDataFactory(
	argument *ArgShardInterceptedDataFactory,
	dataType InterceptedDataType,
) (*shardInterceptedDataFactory, error) {

	if argument == nil {
		return nil, process.ErrNilArguments
	}
	if check.IfNil(argument.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(argument.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(argument.KeyGen) {
		return nil, process.ErrNilKeyGen
	}
	if check.IfNil(argument.Signer) {
		return nil, process.ErrNilSingleSigner
	}
	if check.IfNil(argument.AddrConv) {
		return nil, process.ErrNilAddressConverter
	}
	if check.IfNil(argument.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(argument.MultiSigVerifier) {
		return nil, process.ErrNilMultiSigVerifier
	}
	if check.IfNil(argument.ChronologyValidator) {
		return nil, process.ErrNilChronologyValidator
	}

	return &shardInterceptedDataFactory{
		marshalizer:         argument.Marshalizer,
		hasher:              argument.Hasher,
		keyGen:              argument.KeyGen,
		singleSigner:        argument.Signer,
		addrConverter:       argument.AddrConv,
		shardCoordinator:    argument.ShardCoordinator,
		interceptedDataType: dataType,
		multiSigVerifier:    argument.MultiSigVerifier,
		chronologyValidator: argument.ChronologyValidator,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
// The type of the output instance is provided in the constructor
func (sidf *shardInterceptedDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	switch sidf.interceptedDataType {
	case InterceptedTx:
		return sidf.createInterceptedTx(buff)
	case InterceptedShardHeader:
		return sidf.createInterceptedShardHeader(buff)
	case InterceptedUnsignedTx:
		return sidf.createInterceptedUnsignedTx(buff)
	case InterceptedMetaHeader:
		return sidf.createInterceptedMetaHeader(buff)
	default:
		return nil, process.ErrInterceptedDataTypeNotDefined
	}
}

func (sidf *shardInterceptedDataFactory) createInterceptedTx(buff []byte) (process.InterceptedData, error) {
	return transaction.NewInterceptedTransaction(
		buff,
		sidf.marshalizer,
		sidf.hasher,
		sidf.keyGen,
		sidf.singleSigner,
		sidf.addrConverter,
		sidf.shardCoordinator,
	)
}

func (sidf *shardInterceptedDataFactory) createInterceptedUnsignedTx(buff []byte) (process.InterceptedData, error) {
	return unsigned.NewInterceptedUnsignedTransaction(
		buff,
		sidf.marshalizer,
		sidf.hasher,
		sidf.addrConverter,
		sidf.shardCoordinator,
	)
}

func (sidf *shardInterceptedDataFactory) createInterceptedShardHeader(buff []byte) (process.InterceptedData, error) {
	arg := &interceptedBlocks.ArgInterceptedBlockHeader{
		HdrBuff:             buff,
		Marshalizer:         sidf.marshalizer,
		Hasher:              sidf.hasher,
		MultiSigVerifier:    sidf.multiSigVerifier,
		ChronologyValidator: sidf.chronologyValidator,
		ShardCoordinator:    sidf.shardCoordinator,
	}

	return interceptedBlocks.NewInterceptedHeader(arg)
}

func (sidf *shardInterceptedDataFactory) createInterceptedMetaHeader(buff []byte) (process.InterceptedData, error) {
	arg := &interceptedBlocks.ArgInterceptedBlockHeader{
		HdrBuff:             buff,
		Marshalizer:         sidf.marshalizer,
		Hasher:              sidf.hasher,
		MultiSigVerifier:    sidf.multiSigVerifier,
		ChronologyValidator: sidf.chronologyValidator,
		ShardCoordinator:    sidf.shardCoordinator,
	}

	return interceptedBlocks.NewInterceptedMetaHeader(arg)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sidf *shardInterceptedDataFactory) IsInterfaceNil() bool {
	if sidf == nil {
		return true
	}
	return false
}
