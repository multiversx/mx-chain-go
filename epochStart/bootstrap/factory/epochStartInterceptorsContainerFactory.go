package factory

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	disabledFactory "github.com/multiversx/mx-chain-go/factory/disabled"
	disabledGenesis "github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/update"
)

const timeSpanForBadHeaders = time.Minute

// ArgsEpochStartInterceptorContainer holds the arguments needed for creating a new epoch start interceptors
// container factory
type ArgsEpochStartInterceptorContainer struct {
	CoreComponents          process.CoreComponentsHolder
	CryptoComponents        process.CryptoComponentsHolder
	Config                  config.Config
	ShardCoordinator        sharding.Coordinator
	MainMessenger           process.TopicHandler
	FullArchiveMessenger    process.TopicHandler
	DataPool                dataRetriever.PoolsHolder
	WhiteListHandler        update.WhiteListHandler
	WhiteListerVerifiedTxs  update.WhiteListHandler
	AddressPubkeyConv       core.PubkeyConverter
	NonceConverter          typeConverters.Uint64ByteSliceConverter
	ChainID                 []byte
	ArgumentsParser         process.ArgumentsParser
	HeaderIntegrityVerifier process.HeaderIntegrityVerifier
	RequestHandler          process.RequestHandler
	SignaturesHandler       process.SignaturesHandler
	NodeOperationMode       p2p.NodeOperation
}

// NewEpochStartInterceptorsContainer will return a real interceptors container factory, but with many disabled components
func NewEpochStartInterceptorsContainer(args ArgsEpochStartInterceptorContainer) (process.InterceptorsContainer, error) {
	if check.IfNil(args.CoreComponents) {
		return nil, epochStart.ErrNilCoreComponentsHolder
	}
	if check.IfNil(args.CryptoComponents) {
		return nil, epochStart.ErrNilCryptoComponentsHolder
	}
	if check.IfNil(args.CoreComponents.AddressPubKeyConverter()) {
		return nil, epochStart.ErrNilPubkeyConverter
	}

	cryptoComponents := args.CryptoComponents.Clone().(process.CryptoComponentsHolder)
	err := cryptoComponents.SetMultiSignerContainer(disabled.NewMultiSignerContainer())
	if err != nil {
		return nil, err
	}

	nodesCoordinator := disabled.NewNodesCoordinator()
	storer := disabled.NewChainStorer()
	antiFloodHandler := disabled.NewAntiFloodHandler()
	accountsAdapter := disabled.NewAccountsAdapter()
	blackListHandler := cache.NewTimeCache(timeSpanForBadHeaders)
	feeHandler := &disabledGenesis.FeeHandler{}
	headerSigVerifier := disabled.NewHeaderSigVerifier()
	sizeCheckDelta := 0
	validityAttester := disabled.NewValidityAttester()
	epochStartTrigger := disabled.NewEpochStartTrigger()
	// TODO: move the peerShardMapper creation before boostrapComponents
	peerShardMapper := disabled.NewPeerShardMapper()
	fullArchivePeerShardMapper := disabled.NewPeerShardMapper()
	hardforkTrigger := disabledFactory.HardforkTrigger()

	containerFactoryArgs := interceptorscontainer.CommonInterceptorsContainerFactoryArgs{
		CoreComponents:               args.CoreComponents,
		CryptoComponents:             cryptoComponents,
		Accounts:                     accountsAdapter,
		ShardCoordinator:             args.ShardCoordinator,
		NodesCoordinator:             nodesCoordinator,
		MainMessenger:                args.MainMessenger,
		FullArchiveMessenger:         args.FullArchiveMessenger,
		Store:                        storer,
		DataPool:                     args.DataPool,
		MaxTxNonceDeltaAllowed:       common.MaxTxNonceDeltaAllowed,
		TxFeeHandler:                 feeHandler,
		BlockBlackList:               blackListHandler,
		HeaderSigVerifier:            headerSigVerifier,
		HeaderIntegrityVerifier:      args.HeaderIntegrityVerifier,
		ValidityAttester:             validityAttester,
		EpochStartTrigger:            epochStartTrigger,
		WhiteListHandler:             args.WhiteListHandler,
		WhiteListerVerifiedTxs:       args.WhiteListerVerifiedTxs,
		AntifloodHandler:             antiFloodHandler,
		ArgumentsParser:              args.ArgumentsParser,
		PreferredPeersHolder:         disabled.NewPreferredPeersHolder(),
		SizeCheckDelta:               uint32(sizeCheckDelta),
		RequestHandler:               args.RequestHandler,
		PeerSignatureHandler:         cryptoComponents.PeerSignatureHandler(),
		SignaturesHandler:            args.SignaturesHandler,
		HeartbeatExpiryTimespanInSec: args.Config.HeartbeatV2.HeartbeatExpiryTimespanInSec,
		MainPeerShardMapper:          peerShardMapper,
		FullArchivePeerShardMapper:   fullArchivePeerShardMapper,
		HardforkTrigger:              hardforkTrigger,
	}

	interceptorsContainerFactory, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(containerFactoryArgs)
	if err != nil {
		return nil, err
	}

	container, err := interceptorsContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	err = interceptorsContainerFactory.AddShardTrieNodeInterceptors(container)
	if err != nil {
		return nil, err
	}

	return container, nil
}
