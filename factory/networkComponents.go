package factory

import (
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/debug/antiflood"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	peersHolder "github.com/ElrondNetwork/elrond-go/p2p/peersHolder"
	"github.com/ElrondNetwork/elrond-go/p2p/rating"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/rating/peerHonesty"
	antifloodFactory "github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/factory"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// NetworkComponentsFactoryArgs holds the arguments to create a network component handler instance
type NetworkComponentsFactoryArgs struct {
	P2pConfig            config.P2PConfig
	MainConfig           config.Config
	RatingsConfig        config.RatingsConfig
	StatusHandler        core.AppStatusHandler
	Marshalizer          marshal.Marshalizer
	Syncer               p2p.SyncTimer
	PreferredPeersSlices []string
	BootstrapWaitTime    time.Duration
	NodeOperationMode    p2p.NodeOperation
}

type networkComponentsFactory struct {
	p2pConfig            config.P2PConfig
	mainConfig           config.Config
	ratingsConfig        config.RatingsConfig
	statusHandler        core.AppStatusHandler
	listenAddress        string
	marshalizer          marshal.Marshalizer
	syncer               p2p.SyncTimer
	preferredPeersSlices []string
	bootstrapWaitTime    time.Duration
	nodeOperationMode    p2p.NodeOperation
}

// networkComponents struct holds the network components
type networkComponents struct {
	netMessenger           p2p.Messenger
	inputAntifloodHandler  P2PAntifloodHandler
	outputAntifloodHandler P2PAntifloodHandler
	pubKeyTimeCacher       process.TimeCacher
	topicFloodPreventer    process.TopicFloodPreventer
	floodPreventers        []process.FloodPreventer
	peerBlackListHandler   process.PeerBlackListCacher
	antifloodConfig        config.AntifloodConfig
	peerHonestyHandler     consensus.PeerHonestyHandler
	peersHolder            PreferredPeersHolderHandler
	peersRatingHandler     p2p.PeersRatingHandler
	closeFunc              context.CancelFunc
}

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
		p2pConfig:            args.P2pConfig,
		ratingsConfig:        args.RatingsConfig,
		marshalizer:          args.Marshalizer,
		mainConfig:           args.MainConfig,
		statusHandler:        args.StatusHandler,
		listenAddress:        libp2p.ListenAddrWithIp4AndTcp,
		syncer:               args.Syncer,
		bootstrapWaitTime:    args.BootstrapWaitTime,
		preferredPeersSlices: args.PreferredPeersSlices,
		nodeOperationMode:    args.NodeOperationMode,
	}, nil
}

// Create creates and returns the network components
func (ncf *networkComponentsFactory) Create() (*networkComponents, error) {
	ph, err := peersHolder.NewPeersHolder(ncf.preferredPeersSlices)
	if err != nil {
		return nil, err
	}

	topRatedCache, err := lrucache.NewCache(ncf.mainConfig.PeersRatingConfig.TopRatedCacheCapacity)
	if err != nil {
		return nil, err
	}
	badRatedCache, err := lrucache.NewCache(ncf.mainConfig.PeersRatingConfig.BadRatedCacheCapacity)
	if err != nil {
		return nil, err
	}
	argsPeersRatingHandler := rating.ArgPeersRatingHandler{
		TopRatedCache: topRatedCache,
		BadRatedCache: badRatedCache,
	}
	peersRatingHandler, err := rating.NewPeersRatingHandler(argsPeersRatingHandler)
	if err != nil {
		return nil, err
	}

	arg := libp2p.ArgsNetworkMessenger{
		Marshalizer:          ncf.marshalizer,
		ListenAddress:        ncf.listenAddress,
		P2pConfig:            ncf.p2pConfig,
		SyncTimer:            ncf.syncer,
		PreferredPeersHolder: ph,
		NodeOperationMode:    ncf.nodeOperationMode,
		PeersRatingHandler:   peersRatingHandler,
	}
	netMessenger, err := libp2p.NewNetworkMessenger(arg)
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

	inputAntifloodHandler, ok := antiFloodComponents.AntiFloodHandler.(P2PAntifloodHandler)
	if !ok {
		err = errors.ErrWrongTypeAssertion
		return nil, fmt.Errorf("%w when casting input antiflood handler to P2PAntifloodHandler", err)
	}

	var outAntifloodHandler process.P2PAntifloodHandler
	outAntifloodHandler, err = antifloodFactory.NewP2POutputAntiFlood(ctx, ncf.mainConfig)
	if err != nil {
		return nil, err
	}

	outputAntifloodHandler, ok := outAntifloodHandler.(P2PAntifloodHandler)
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

	cache, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(config.PeerHonesty))
	if err != nil {
		return nil, err
	}

	return peerHonesty.NewP2pPeerHonesty(ratingConfig.PeerHonesty, pkTimeCache, cache)
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
