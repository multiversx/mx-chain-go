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

type interceptedTxDataFactory struct {
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	keyGen           crypto.KeyGenerator
	singleSigner     crypto.SingleSigner
	addrConverter    state.AddressConverter
	shardCoordinator sharding.Coordinator
	feeHandler       process.FeeHandler
}

// NewInterceptedTxDataFactory creates an instance of interceptedTxDataFactory
func NewInterceptedTxDataFactory(argument *ArgInterceptedDataFactory) (*interceptedTxDataFactory, error) {
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
	if check.IfNil(argument.FeeHandler) {
		return nil, process.ErrNilEconomicsFeeHandler
	}

	return &interceptedTxDataFactory{
		marshalizer:      argument.Marshalizer,
		hasher:           argument.Hasher,
		keyGen:           argument.KeyGen,
		singleSigner:     argument.Signer,
		addrConverter:    argument.AddrConv,
		shardCoordinator: argument.ShardCoordinator,
		feeHandler:       argument.FeeHandler,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (itdf *interceptedTxDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	return transaction.NewInterceptedTransaction(
		buff,
		itdf.marshalizer,
		itdf.hasher,
		itdf.keyGen,
		itdf.singleSigner,
		itdf.addrConverter,
		itdf.shardCoordinator,
		itdf.feeHandler,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (itdf *interceptedTxDataFactory) IsInterfaceNil() bool {
	if itdf == nil {
		return true
	}
	return false
}
