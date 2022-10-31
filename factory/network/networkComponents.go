package network

import (
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/debug/antiflood"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/p2p"
	p2pConfig "github.com/ElrondNetwork/elrond-go/p2p/config"
	p2pFactory "github.com/ElrondNetwork/elrond-go/p2p/factory"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/rating/peerHonesty"
	antifloodFactory "github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/factory"
	"github.com/ElrondNetwork/elrond-go/storage/cache"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageunit"
)

// NetworkComponentsFactoryArgs holds the arguments to create a network component handler instance
type NetworkComponentsFactoryArgs struct {
	P2pConfig             p2pConfig.P2PConfig
	MainConfig            config.Config
	RatingsConfig         config.RatingsConfig
	StatusHandler         core.AppStatusHandler
	Marshalizer           marshal.Marshalizer
	Syncer                p2p.SyncTimer
	PreferredPeersSlices  []string
	BootstrapWaitTime     time.Duration
	NodeOperationMode     p2p.NodeOperation
	ConnectionWatcherType string
	P2pKeyPemFileName     string
}

type networkComponentsFactory struct {
	p2pConfig             p2pConfig.P2PConfig
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
	p2pKeyPemFileName     string
}

// networkComponents struct holds the network components
type networkComponents struct {
	netMessenger           p2p.Messenger
	inputAntifloodHandler  factory.P2PAntifloodHandler
	outputAntifloodHandler factory.P2PAntifloodHandler
	pubKeyTimeCacher       process.TimeCacher
	topicFloodPreventer    process.TopicFloodPreventer
	floodPreventers        []process.FloodPreventer
	peerBlackListHandler   process.PeerBlackListCacher
	antifloodConfig        config.AntifloodConfig
	peerHonestyHandler     consensus.PeerHonestyHandler
	peersHolder            factory.PreferredPeersHolderHandler
	peersRatingHandler     p2p.PeersRatingHandler
	closeFunc              context.CancelFunc
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

	return &networkComponentsFactory{
		p2pConfig:             args.P2pConfig,
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
		p2pKeyPemFileName:     args.P2pKeyPemFileName,
	}, nil
}

// Create creates and returns the network components
func (ncf *networkComponentsFactory) Create() (*networkComponents, error) {
	ph, err := p2pFactory.NewPeersHolder(ncf.preferredPeersSlices)
	if err != nil {
		return nil, err
	}

	topRatedCache, err := cache.NewLRUCache(ncf.mainConfig.PeersRatingConfig.TopRatedCacheCapacity)
	if err != nil {
		return nil, err
	}
	badRatedCache, err := cache.NewLRUCache(ncf.mainConfig.PeersRatingConfig.BadRatedCacheCapacity)
	if err != nil {
		return nil, err
	}
	argsPeersRatingHandler := p2pFactory.ArgPeersRatingHandler{
		TopRatedCache: topRatedCache,
		BadRatedCache: badRatedCache,
	}
	peersRatingHandler, err := p2pFactory.NewPeersRatingHandler(argsPeersRatingHandler)
	if err != nil {
		return nil, err
	}

	p2pPrivateKeyBytes, err := common.GetSkBytesFromP2pKey(ncf.p2pKeyPemFileName)
	if err != nil {
		return nil, err
	}

	arg := p2pFactory.ArgsNetworkMessenger{
		Marshalizer:           ncf.marshalizer,
		ListenAddress:         ncf.listenAddress,
		P2pConfig:             ncf.p2pConfig,
		SyncTimer:             ncf.syncer,
		PreferredPeersHolder:  ph,
		NodeOperationMode:     ncf.nodeOperationMode,
		PeersRatingHandler:    peersRatingHandler,
		ConnectionWatcherType: ncf.connectionWatcherType,
		P2pPrivateKeyBytes:    p2pPrivateKeyBytes,
	}
	netMessenger, err := p2pFactory.NewNetworkMessenger(arg)
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
	antiFloodComponents, err = antifloodFactory.NewP2PAntiFloodComponents(ctx, ncf.mainConfig, ncf.statusHandler, netMessenger.ID())
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

	err = netMessenger.Bootstrap()
	if err != nil {
		return nil, err
	}

	netMessenger.WaitForConnections(ncf.bootstrapWaitTime, ncf.p2pConfig.Node.MinNumPeersToWaitForOnBootstrap)

	return &networkComponents{
		netMessenger:           netMessenger,
		inputAntifloodHandler:  inputAntifloodHandler,
		outputAntifloodHandler: outputAntifloodHandler,
		topicFloodPreventer:    antiFloodComponents.TopicPreventer,
		floodPreventers:        antiFloodComponents.FloodPreventers,
		peerBlackListHandler:   antiFloodComponents.BlacklistHandler,
		pubKeyTimeCacher:       antiFloodComponents.PubKeysCacher,
		antifloodConfig:        ncf.mainConfig.Antiflood,
		peerHonestyHandler:     peerHonestyHandler,
		peersHolder:            ph,
		peersRatingHandler:     peersRatingHandler,
		closeFunc:              cancelFunc,
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

	if nc.netMessenger != nil {
		log.Debug("calling close on the network messenger instance...")
		err := nc.netMessenger.Close()
		log.LogIfError(err)
	}

	return nil
}
