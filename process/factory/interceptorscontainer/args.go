package interceptorscontainer

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type coreComponentsHolder interface {
	InternalMarshalizer() marshal.Marshalizer
	SetInternalMarshalizer(marshalizer marshal.Marshalizer) error
	TxMarshalizer() marshal.Marshalizer
	Hasher() hashing.Hasher
	Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter
	AddressPubKeyConverter() state.PubkeyConverter
	ChainID() string
	IsInterfaceNil() bool
}

type cryptoComponentsHolder interface {
	TxSignKeyGen() crypto.KeyGenerator
	BlockSignKeyGen() crypto.KeyGenerator
	TxSingleSigner() crypto.SingleSigner
	BlockSigner() crypto.SingleSigner
	MultiSigner() crypto.MultiSigner
	IsInterfaceNil() bool
}

// ShardInterceptorsContainerFactoryArgs holds the arguments needed for ShardInterceptorsContainerFactory
type ShardInterceptorsContainerFactoryArgs struct {
	Accounts               state.AccountsAdapter
	ShardCoordinator       sharding.Coordinator
	NodesCoordinator       sharding.NodesCoordinator
	Messenger              process.TopicHandler
	Store                  dataRetriever.StorageService
	CoreComponents         coreComponentsHolder
	CryptoComponents       cryptoComponentsHolder
	DataPool               dataRetriever.PoolsHolder
	MaxTxNonceDeltaAllowed int
	TxFeeHandler           process.FeeHandler
	BlackList              process.BlackListHandler
	HeaderSigVerifier      process.InterceptedHeaderSigVerifier
	SizeCheckDelta         uint32
	ValidityAttester       process.ValidityAttester
	EpochStartTrigger      process.EpochStartTriggerHandler
	WhiteListHandler       process.WhiteListHandler
	WhiteListerVerifiedTxs process.WhiteListHandler
	AntifloodHandler       process.P2PAntifloodHandler
}

// MetaInterceptorsContainerFactoryArgs holds the arguments needed for MetaInterceptorsContainerFactory
type MetaInterceptorsContainerFactoryArgs struct {
	CoreComponents         coreComponentsHolder
	CryptoComponents       cryptoComponentsHolder
	ShardCoordinator       sharding.Coordinator
	NodesCoordinator       sharding.NodesCoordinator
	Messenger              process.TopicHandler
	Store                  dataRetriever.StorageService
	DataPool               dataRetriever.PoolsHolder
	Accounts               state.AccountsAdapter
	MaxTxNonceDeltaAllowed int
	TxFeeHandler           process.FeeHandler
	BlackList              process.BlackListHandler
	HeaderSigVerifier      process.InterceptedHeaderSigVerifier
	SizeCheckDelta         uint32
	ValidityAttester       process.ValidityAttester
	EpochStartTrigger      process.EpochStartTriggerHandler
	WhiteListHandler       process.WhiteListHandler
	WhiteListerVerifiedTxs process.WhiteListHandler
	AntifloodHandler       process.P2PAntifloodHandler
}
