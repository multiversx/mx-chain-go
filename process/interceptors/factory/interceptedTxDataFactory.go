package factory

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ process.InterceptedDataFactory = (*interceptedTxDataFactory)(nil)

type interceptedTxDataFactory struct {
	protoMarshalizer       marshal.Marshalizer
	signMarshalizer        marshal.Marshalizer
	hasher                 hashing.Hasher
	keyGen                 crypto.KeyGenerator
	singleSigner           crypto.SingleSigner
	pubkeyConverter        core.PubkeyConverter
	shardCoordinator       sharding.Coordinator
	feeHandler             process.FeeHandler
	whiteListerVerifiedTxs process.WhiteListHandler
	argsParser             process.ArgumentsParser
	chainID                []byte
	minTransactionVersion  uint32
}

// NewInterceptedTxDataFactory creates an instance of interceptedTxDataFactory
func NewInterceptedTxDataFactory(argument *ArgInterceptedDataFactory) (*interceptedTxDataFactory, error) {
	if argument == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(argument.CoreComponents) {
		return nil, process.ErrNilCoreComponentsHolder
	}
	if check.IfNil(argument.CryptoComponents) {
		return nil, process.ErrNilCryptoComponentsHolder
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
	if check.IfNil(argument.CryptoComponents.TxSignKeyGen()) {
		return nil, process.ErrNilKeyGen
	}
	if check.IfNil(argument.CryptoComponents.TxSingleSigner()) {
		return nil, process.ErrNilSingleSigner
	}
	if check.IfNil(argument.CoreComponents.AddressPubKeyConverter()) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(argument.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(argument.FeeHandler) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(argument.WhiteListerVerifiedTxs) {
		return nil, process.ErrNilWhiteListHandler
	}
	if check.IfNil(argument.ArgsParser) {
		return nil, process.ErrNilArgumentParser
	}
	if len(argument.CoreComponents.ChainID()) == 0 {
		return nil, process.ErrInvalidChainID
	}
	if argument.CoreComponents.MinTransactionVersion() == 0 {
		return nil, process.ErrInvalidTransactionVersion
	}

	return &interceptedTxDataFactory{
		protoMarshalizer:       argument.CoreComponents.InternalMarshalizer(),
		signMarshalizer:        argument.CoreComponents.TxMarshalizer(),
		hasher:                 argument.CoreComponents.Hasher(),
		keyGen:                 argument.CryptoComponents.TxSignKeyGen(),
		singleSigner:           argument.CryptoComponents.TxSingleSigner(),
		pubkeyConverter:        argument.CoreComponents.AddressPubKeyConverter(),
		shardCoordinator:       argument.ShardCoordinator,
		feeHandler:             argument.FeeHandler,
		whiteListerVerifiedTxs: argument.WhiteListerVerifiedTxs,
		argsParser:             argument.ArgsParser,
		chainID:                []byte(argument.CoreComponents.ChainID()),
		minTransactionVersion:  argument.CoreComponents.MinTransactionVersion(),
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (itdf *interceptedTxDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	return transaction.NewInterceptedTransaction(
		buff,
		itdf.protoMarshalizer,
		itdf.signMarshalizer,
		itdf.hasher,
		itdf.keyGen,
		itdf.singleSigner,
		itdf.pubkeyConverter,
		itdf.shardCoordinator,
		itdf.feeHandler,
		itdf.whiteListerVerifiedTxs,
		itdf.argsParser,
		itdf.chainID,
		itdf.minTransactionVersion,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (itdf *interceptedTxDataFactory) IsInterfaceNil() bool {
	return itdf == nil
}
