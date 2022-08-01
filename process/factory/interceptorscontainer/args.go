package interceptorscontainer

import (
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
)

// CommonInterceptorsContainerFactoryArgs holds the arguments needed for the metachain/shard interceptors factories
type CommonInterceptorsContainerFactoryArgs struct {
	CoreComponents               process.CoreComponentsHolder
	CryptoComponents             process.CryptoComponentsHolder
	Accounts                     state.AccountsAdapter
	ShardCoordinator             sharding.Coordinator
	NodesCoordinator             nodesCoordinator.NodesCoordinator
	Messenger                    process.TopicHandler
	Store                        dataRetriever.StorageService
	DataPool                     dataRetriever.PoolsHolder
	MaxTxNonceDeltaAllowed       int
	TxFeeHandler                 process.FeeHandler
	BlockBlackList               process.TimeCacher
	HeaderSigVerifier            process.InterceptedHeaderSigVerifier
	HeaderIntegrityVerifier      process.HeaderIntegrityVerifier
	ValidityAttester             process.ValidityAttester
	EpochStartTrigger            process.EpochStartTriggerHandler
	WhiteListHandler             process.WhiteListHandler
	WhiteListerVerifiedTxs       process.WhiteListHandler
	AntifloodHandler             process.P2PAntifloodHandler
	ArgumentsParser              process.ArgumentsParser
	PreferredPeersHolder         process.PreferredPeersHolderHandler
	SizeCheckDelta               uint32
	EnableSignTxWithHashEpoch    uint32
	RequestHandler               process.RequestHandler
	PeerSignatureHandler         crypto.PeerSignatureHandler
	SignaturesHandler            process.SignaturesHandler
	HeartbeatExpiryTimespanInSec int64
	PeerShardMapper              process.PeerShardMapper
	HardforkTrigger              heartbeat.HardforkTrigger
}
