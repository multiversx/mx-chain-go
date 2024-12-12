package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

// interceptedDataCoreComponentsHolder holds the core components required by the intercepted data factory
type interceptedDataCoreComponentsHolder interface {
	InternalMarshalizer() marshal.Marshalizer
	TxMarshalizer() marshal.Marshalizer
	TxVersionChecker() process.TxVersionCheckerHandler
	Hasher() hashing.Hasher
	TxSignHasher() hashing.Hasher
	Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter
	AddressPubKeyConverter() core.PubkeyConverter
	ChainID() string
	MinTransactionVersion() uint32
	IsInterfaceNil() bool
	HardforkTriggerPubKey() []byte
	EnableEpochsHandler() common.EnableEpochsHandler
}

// interceptedDataCryptoComponentsHolder holds the crypto components required by the intercepted data factory
type interceptedDataCryptoComponentsHolder interface {
	TxSignKeyGen() crypto.KeyGenerator
	BlockSignKeyGen() crypto.KeyGenerator
	TxSingleSigner() crypto.SingleSigner
	BlockSigner() crypto.SingleSigner
	GetMultiSigner(epoch uint32) (crypto.MultiSigner, error)
	PublicKey() crypto.PublicKey
	IsInterfaceNil() bool
}

// ArgInterceptedDataFactory holds all dependencies required by the shard and meta intercepted data factory in order to create
// new instances
type ArgInterceptedDataFactory struct {
	CoreComponents               interceptedDataCoreComponentsHolder
	CryptoComponents             interceptedDataCryptoComponentsHolder
	ShardCoordinator             sharding.Coordinator
	NodesCoordinator             nodesCoordinator.NodesCoordinator
	FeeHandler                   process.FeeHandler
	WhiteListerVerifiedTxs       process.WhiteListHandler
	HeaderSigVerifier            process.InterceptedHeaderSigVerifier
	ValidityAttester             process.ValidityAttester
	HeaderIntegrityVerifier      process.HeaderIntegrityVerifier
	EpochStartTrigger            process.EpochStartTriggerHandler
	ArgsParser                   process.ArgumentsParser
	PeerSignatureHandler         crypto.PeerSignatureHandler
	SignaturesHandler            process.SignaturesHandler
	HeartbeatExpiryTimespanInSec int64
	PeerID                       core.PeerID
}
