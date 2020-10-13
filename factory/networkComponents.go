package factory

import (
	"context"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/debug/antiflood"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/rating/peerHonesty"
	antifloodFactory "github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/factory"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
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
	BootstrapWaitSeconds uint32
}

type networkComponentsFactory struct {
	p2pConfig            config.P2PConfig
	mainConfig           config.Config
	ratingsConfig        config.RatingsConfig
	statusHandler        core.AppStatusHandler
	listenAddress        string
	marshalizer          marshal.Marshalizer
	syncer               p2p.SyncTimer
	bootstrapWaitSeconds uint32
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
		bootstrapWaitSeconds: args.BootstrapWaitSeconds,
	}, nil
}

// Create creates and returns the network components
func (ncf *networkComponentsFactory) Create() (*networkComponents, error) {
	arg := libp2p.ArgsNetworkMessenger{
		Marshalizer:   ncf.marshalizer,
		ListenAddress: ncf.listenAddress,
		P2pConfig:     ncf.p2pConfig,
		SyncTimer:     ncf.syncer,
	}

	netMessenger, err := libp2p.NewNetworkMessenger(arg)
	if err != nil {
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	antiFloodComponents, err := antifloodFactory.NewP2PAntiFloodComponents(ctx, ncf.mainConfig, ncf.statusHandler, netMessenger.ID())
	if err != nil {
		return nil, err
	}

	//TODO: move to NewP2PAntiFloodComponents.initP2PAntiFloodComponents
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
		return nil, fmt.Errorf("%w when casting input antiflood handler to structs/P2PAntifloodHandler", errors.ErrWrongTypeAssertion)
	}

	outAntifloodHandler, errOutputAntiflood := antifloodFactory.NewP2POutputAntiFlood(ctx, ncf.mainConfig)
	if errOutputAntiflood != nil {
		return nil, errOutputAntiflood
	}

	outputAntifloodHandler, ok := outAntifloodHandler.(P2PAntifloodHandler)
	if !ok {
		return nil, fmt.Errorf("%w when casting output antiflood handler to structs/P2PAntifloodHandler", errors.ErrWrongTypeAssertion)
	}

	peerHonestyHandler, err := ncf.createPeerHonestyHandler(
		&ncf.mainConfig,
		ncf.ratingsConfig,
		antiFloodComponents.PubKeysCacher,
	)
	if err != nil {
		return nil, err
	}

	err = netMessenger.Bootstrap(ncf.bootstrapWaitSeconds)
	if err != nil {
		return nil, err
	}

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
