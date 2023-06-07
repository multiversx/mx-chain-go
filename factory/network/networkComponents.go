package network

import (
	"context"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/debug/antiflood"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/disabled"
	"github.com/multiversx/mx-chain-go/p2p"
	p2pConfig "github.com/multiversx/mx-chain-go/p2p/config"
	p2pFactory "github.com/multiversx/mx-chain-go/p2p/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/rating/peerHonesty"
	antifloodFactory "github.com/multiversx/mx-chain-go/process/throttle/antiflood/factory"
	"github.com/multiversx/mx-chain-go/storage/cache"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// NetworkComponentsFactoryArgs holds the arguments to create a network component handler instance
type NetworkComponentsFactoryArgs struct {
	MainP2pConfig         p2pConfig.P2PConfig
	FullArchiveP2pConfig  p2pConfig.P2PConfig
	MainConfig            config.Config
	RatingsConfig         config.RatingsConfig
	StatusHandler         core.AppStatusHandler
	Marshalizer           marshal.Marshalizer
	Syncer                p2p.SyncTimer
	PreferredPeersSlices  []string
	BootstrapWaitTime     time.Duration
	NodeOperationMode     p2p.NodeOperation
	ConnectionWatcherType string
	CryptoComponents      factory.CryptoComponentsHolder
}

type networkComponentsFactory struct {
	mainP2PConfig         p2pConfig.P2PConfig
	fullArchiveP2PConfig  p2pConfig.P2PConfig
	mainConfig            config.Config
	ratingsConfig         config.RatingsConfig
	statusHandler         core.AppStatusHandler
	listenAddress         string
	marshalizer           marshal.Marshalizer
	syncer                p2p.SyncTimer
	preferredPeersSlices  []string
	bootstrapWaitTime     time.Duration
	nodeOperationMode     p2p.NodeOperation
	connectionWatcherType string
	cryptoComponents      factory.CryptoComponentsHolder
}

type networkComponentsHolder struct {
	netMessenger       p2p.Messenger
	peersRatingHandler p2p.PeersRatingHandler
	peersRatingMonitor p2p.PeersRatingMonitor
}

// networkComponents struct holds the network components
type networkComponents struct {
	mainNetworkHolder        networkComponentsHolder
	fullArchiveNetworkHolder networkComponentsHolder
	inputAntifloodHandler    factory.P2PAntifloodHandler
	outputAntifloodHandler   factory.P2PAntifloodHandler
	pubKeyTimeCacher         process.TimeCacher
	topicFloodPreventer      process.TopicFloodPreventer
	floodPreventers          []process.FloodPreventer
	peerBlackListHandler     process.PeerBlackListCacher
	antifloodConfig          config.AntifloodConfig
	peerHonestyHandler       consensus.PeerHonestyHandler
	peersHolder              factory.PreferredPeersHolderHandler
	closeFunc                context.CancelFunc
}

var log = logger.GetOrCreate("factory")

// NewNetworkComponentsFactory returns a new instance of a network components factory
func NewNetworkComponentsFactory(
	args NetworkComponentsFactoryArgs,
) (*networkComponentsFactory, error) {
	if check.IfNil(args.StatusHandler) {
		return nil, errors.ErrNilStatusHandler
	}
	if check.IfNil(args.Marshalizer) {
		return nil, fmt.Errorf("%w in NewNetworkComponentsFactory", errors.ErrNilMarshalizer)
	}
	if check.IfNil(args.Syncer) {
		return nil, errors.ErrNilSyncTimer
	}
	if check.IfNil(args.CryptoComponents) {
		return nil, errors.ErrNilCryptoComponentsHolder
	}
	if args.NodeOperationMode != p2p.NormalOperation && args.NodeOperationMode != p2p.FullArchiveMode {
		return nil, errors.ErrInvalidNodeOperationMode
	}

	return &networkComponentsFactory{
		mainP2PConfig:         args.MainP2pConfig,
		fullArchiveP2PConfig:  args.FullArchiveP2pConfig,
		ratingsConfig:         args.RatingsConfig,
		marshalizer:           args.Marshalizer,
		mainConfig:            args.MainConfig,
		statusHandler:         args.StatusHandler,
		listenAddress:         p2p.ListenAddrWithIp4AndTcp,
		syncer:                args.Syncer,
		bootstrapWaitTime:     args.BootstrapWaitTime,
		preferredPeersSlices:  args.PreferredPeersSlices,
		nodeOperationMode:     args.NodeOperationMode,
		connectionWatcherType: args.ConnectionWatcherType,
		cryptoComponents:      args.CryptoComponents,
	}, nil
}

// Create creates and returns the network components
func (ncf *networkComponentsFactory) Create() (*networkComponents, error) {
	peersHolder, err := p2pFactory.NewPeersHolder(ncf.preferredPeersSlices)
	if err != nil {
		return nil, err
	}

	mainNetworkComp, err := ncf.createMainNetworkHolder(peersHolder)
	if err != nil {
		return nil, err
	}

	fullArchiveNetworkComp, err := ncf.createFullArchiveNetworkHolder(peersHolder)
	if err != nil {
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancelFunc()
		}
	}()

	var antiFloodComponents *antifloodFactory.AntiFloodComponents
	antiFloodComponents, err = antifloodFactory.NewP2PAntiFloodComponents(ctx, ncf.mainConfig, ncf.statusHandler, mainNetworkComp.netMessenger.ID())
	if err != nil {
		return nil, err
	}

	// TODO: move to NewP2PAntiFloodComponents.initP2PAntiFloodComponents
	if ncf.mainConfig.Debug.Antiflood.Enabled {
		var debugger process.AntifloodDebugger
		debugger, err = antiflood.NewAntifloodDebugger(ncf.mainConfig.Debug.Antiflood)
		if err != nil {
			return nil, err
		}

		err = antiFloodComponents.AntiFloodHandler.SetDebugger(debugger)
		if err != nil {
			return nil, err
		}
	}

	inputAntifloodHandler, ok := antiFloodComponents.AntiFloodHandler.(factory.P2PAntifloodHandler)
	if !ok {
		err = errors.ErrWrongTypeAssertion
		return nil, fmt.Errorf("%w when casting input antiflood handler to P2PAntifloodHandler", err)
	}

	var outAntifloodHandler process.P2PAntifloodHandler
	outAntifloodHandler, err = antifloodFactory.NewP2POutputAntiFlood(ctx, ncf.mainConfig)
	if err != nil {
		return nil, err
	}

	outputAntifloodHandler, ok := outAntifloodHandler.(factory.P2PAntifloodHandler)
	if !ok {
		err = errors.ErrWrongTypeAssertion
		return nil, fmt.Errorf("%w when casting output antiflood handler to P2PAntifloodHandler", err)
	}

	var peerHonestyHandler consensus.PeerHonestyHandler
	peerHonestyHandler, err = ncf.createPeerHonestyHandler(
		&ncf.mainConfig,
		ncf.ratingsConfig,
		antiFloodComponents.PubKeysCacher,
	)
	if err != nil {
		return nil, err
	}

	err = mainNetworkComp.netMessenger.Bootstrap()
	if err != nil {
		return nil, err
	}

	mainNetworkComp.netMessenger.WaitForConnections(ncf.bootstrapWaitTime, ncf.mainP2PConfig.Node.MinNumPeersToWaitForOnBootstrap)

	err = fullArchiveNetworkComp.netMessenger.Bootstrap()
	if err != nil {
		return nil, err
	}

	fullArchiveNetworkComp.netMessenger.WaitForConnections(ncf.bootstrapWaitTime, ncf.fullArchiveP2PConfig.Node.MinNumPeersToWaitForOnBootstrap)

	return &networkComponents{
		mainNetworkHolder:        mainNetworkComp,
		fullArchiveNetworkHolder: fullArchiveNetworkComp,
		inputAntifloodHandler:    inputAntifloodHandler,
		outputAntifloodHandler:   outputAntifloodHandler,
		pubKeyTimeCacher:         antiFloodComponents.PubKeysCacher,
		topicFloodPreventer:      antiFloodComponents.TopicPreventer,
		floodPreventers:          antiFloodComponents.FloodPreventers,
		peerBlackListHandler:     antiFloodComponents.BlacklistHandler,
		antifloodConfig:          ncf.mainConfig.Antiflood,
		peerHonestyHandler:       peerHonestyHandler,
		peersHolder:              peersHolder,
		closeFunc:                cancelFunc,
	}, nil
}

func (ncf *networkComponentsFactory) createPeerHonestyHandler(
	config *config.Config,
	ratingConfig config.RatingsConfig,
	pkTimeCache process.TimeCacher,
) (consensus.PeerHonestyHandler, error) {

	suCache, err := storageunit.NewCache(storageFactory.GetCacherFromConfig(config.PeerHonesty))
	if err != nil {
		return nil, err
	}

	return peerHonesty.NewP2pPeerHonesty(ratingConfig.PeerHonesty, pkTimeCache, suCache)
}

func (ncf *networkComponentsFactory) createNetworkHolder(
	peersHolder p2p.PreferredPeersHolderHandler,
	p2pConfig p2pConfig.P2PConfig,
	logger p2p.Logger,
) (networkComponentsHolder, error) {

	peersRatingCfg := ncf.mainConfig.PeersRatingConfig
	topRatedCache, err := cache.NewLRUCache(peersRatingCfg.TopRatedCacheCapacity)
	if err != nil {
		return networkComponentsHolder{}, err
	}
	badRatedCache, err := cache.NewLRUCache(peersRatingCfg.BadRatedCacheCapacity)
	if err != nil {
		return networkComponentsHolder{}, err
	}

	argsPeersRatingHandler := p2pFactory.ArgPeersRatingHandler{
		TopRatedCache: topRatedCache,
		BadRatedCache: badRatedCache,
		Logger:        logger,
	}
	peersRatingHandler, err := p2pFactory.NewPeersRatingHandler(argsPeersRatingHandler)
	if err != nil {
		return networkComponentsHolder{}, err
	}

	argsMessenger := p2pFactory.ArgsNetworkMessenger{
		ListenAddress:         ncf.listenAddress,
		Marshaller:            ncf.marshalizer,
		P2pConfig:             p2pConfig,
		SyncTimer:             ncf.syncer,
		PreferredPeersHolder:  peersHolder,
		NodeOperationMode:     ncf.nodeOperationMode,
		PeersRatingHandler:    peersRatingHandler,
		ConnectionWatcherType: ncf.connectionWatcherType,
		P2pPrivateKey:         ncf.cryptoComponents.P2pPrivateKey(),
		P2pSingleSigner:       ncf.cryptoComponents.P2pSingleSigner(),
		P2pKeyGenerator:       ncf.cryptoComponents.P2pKeyGen(),
		Logger:                logger,
	}
	networkMessenger, err := p2pFactory.NewNetworkMessenger(argsMessenger)
	if err != nil {
		return networkComponentsHolder{}, err
	}

	argsPeersRatingMonitor := p2pFactory.ArgPeersRatingMonitor{
		TopRatedCache:       topRatedCache,
		BadRatedCache:       badRatedCache,
		ConnectionsProvider: networkMessenger,
	}
	peersRatingMonitor, err := p2pFactory.NewPeersRatingMonitor(argsPeersRatingMonitor)
	if err != nil {
		return networkComponentsHolder{}, err
	}

	return networkComponentsHolder{
		netMessenger:       networkMessenger,
		peersRatingHandler: peersRatingHandler,
		peersRatingMonitor: peersRatingMonitor,
	}, nil
}

func (ncf *networkComponentsFactory) createMainNetworkHolder(peersHolder p2p.PreferredPeersHolderHandler) (networkComponentsHolder, error) {
	loggerInstance := logger.GetOrCreate("main/p2p")
	return ncf.createNetworkHolder(peersHolder, ncf.mainP2PConfig, loggerInstance)
}

func (ncf *networkComponentsFactory) createFullArchiveNetworkHolder(peersHolder p2p.PreferredPeersHolderHandler) (networkComponentsHolder, error) {
	if ncf.nodeOperationMode != p2p.FullArchiveMode {
		return networkComponentsHolder{
			netMessenger:       disabled.NewNetworkMessenger(),
			peersRatingHandler: disabled.NewPeersRatingHandler(),
			peersRatingMonitor: disabled.NewPeersRatingMonitor(),
		}, nil
	}

	loggerInstance := logger.GetOrCreate("full-archive/p2p")

	return ncf.createNetworkHolder(peersHolder, ncf.fullArchiveP2PConfig, loggerInstance)
}

// Close closes all underlying components that need closing
func (nc *networkComponents) Close() error {
	nc.closeFunc()

	if !check.IfNil(nc.inputAntifloodHandler) {
		log.LogIfError(nc.inputAntifloodHandler.Close())
	}
	if !check.IfNil(nc.outputAntifloodHandler) {
		log.LogIfError(nc.outputAntifloodHandler.Close())
	}
	if !check.IfNil(nc.topicFloodPreventer) {
		log.LogIfError(nc.outputAntifloodHandler.Close())
	}
	if !check.IfNil(nc.peerHonestyHandler) {
		log.LogIfError(nc.peerHonestyHandler.Close())
	}

	mainNetMessenger := nc.mainNetworkHolder.netMessenger
	if !check.IfNil(mainNetMessenger) {
		log.Debug("calling close on the main network messenger instance...")
		log.LogIfError(mainNetMessenger.Close())
	}

	fullArchiveNetMessenger := nc.fullArchiveNetworkHolder.netMessenger
	if !check.IfNil(fullArchiveNetMessenger) {
		log.Debug("calling close on the full archive network messenger instance...")
		log.LogIfError(fullArchiveNetMessenger.Close())
	}

	return nil
}
