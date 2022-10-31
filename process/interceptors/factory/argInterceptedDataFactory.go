package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
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
