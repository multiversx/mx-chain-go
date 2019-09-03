package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// InterceptedDataType represents a type of intercepted data instantiated by the Create func
type InterceptedDataType string

// InterceptedTx is the type for intercepted transaction
const InterceptedTx InterceptedDataType = "intercepted transaction"

type interceptedDataFactory struct {
	marshalizer         marshal.Marshalizer
	hasher              hashing.Hasher
	keyGen              crypto.KeyGenerator
	singleSigner        crypto.SingleSigner
	addrConverter       state.AddressConverter
	shardCoordinator    sharding.Coordinator
	interceptedDataType InterceptedDataType
}

// NewInterceptedDataFactory creates an instance of interceptedDataFactory that can create
// instances of process.InterceptedData
func NewInterceptedDataFactory(
	argument *ArgInterceptedDataFactory,
	dataType InterceptedDataType,
) (*interceptedDataFactory, error) {

	if argument == nil {
		return nil, process.ErrNilArguments
	}
	if check.ForNil(argument.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.ForNil(argument.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.ForNil(argument.KeyGen) {
		return nil, process.ErrNilKeyGen
	}
	if check.ForNil(argument.Signer) {
		return nil, process.ErrNilSingleSigner
	}
	if check.ForNil(argument.AddrConv) {
		return nil, process.ErrNilAddressConverter
	}
	if check.ForNil(argument.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	return &interceptedDataFactory{
		marshalizer:         argument.Marshalizer,
		hasher:              argument.Hasher,
		keyGen:              argument.KeyGen,
		singleSigner:        argument.Signer,
		addrConverter:       argument.AddrConv,
		shardCoordinator:    argument.ShardCoordinator,
		interceptedDataType: dataType,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
// The type of the output instance is provided in the constructor
func (idf *interceptedDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	switch idf.interceptedDataType {
	case InterceptedTx:
		return idf.createInterceptedTx(buff)

	default:
		return nil, process.ErrInterceptedDataTypeNotDefined
	}
}

func (idf *interceptedDataFactory) createInterceptedTx(buff []byte) (process.InterceptedData, error) {
	return transaction.NewInterceptedTransaction(
		buff,
		idf.marshalizer,
		idf.hasher,
		idf.keyGen,
		idf.singleSigner,
		idf.addrConverter,
		idf.shardCoordinator,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (idf *interceptedDataFactory) IsInterfaceNil() bool {
	if idf == nil {
		return true
	}
	return false
}
