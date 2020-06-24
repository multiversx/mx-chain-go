package factory

import (
	"context"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/debug/antiflood"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/process"
	antifloodFactory "github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/factory"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

type networkComponentsFactory struct {
	p2pConfig     config.P2PConfig
	mainConfig    config.Config
	statusHandler core.AppStatusHandler
	listenAddress string
}

// networkComponents struct holds the network components
type networkComponents struct {
	netMessenger           p2p.Messenger
	inputAntifloodHandler  P2PAntifloodHandler
	outputAntifloodHandler P2PAntifloodHandler
	topicFloodPreventer    process.TopicFloodPreventer
	floodPreventers        []process.FloodPreventer
	peerBlackListHandler   process.PeerBlackListHandler
	antifloodConfig        config.AntifloodConfig
	closeFunc              context.CancelFunc
}

// newNetworkComponentsFactory returns a new instance of a network components factory
func newNetworkComponentsFactory(
	p2pConfig config.P2PConfig,
	mainConfig config.Config,
	statusHandler core.AppStatusHandler,
) (*networkComponentsFactory, error) {
	if check.IfNil(statusHandler) {
		return nil, ErrNilStatusHandler
	}

	return &networkComponentsFactory{
		p2pConfig:     p2pConfig,
		mainConfig:    mainConfig,
		statusHandler: statusHandler,
		listenAddress: libp2p.ListenAddrWithIp4AndTcp,
	}, nil
}

// Create creates and returns the network components
func (ncf *networkComponentsFactory) Create() (*networkComponents, error) {
	arg := libp2p.ArgsNetworkMessenger{
		ListenAddress: ncf.listenAddress,
		P2pConfig:     ncf.p2pConfig,
	}

	netMessenger, err := libp2p.NewNetworkMessenger(arg)
	if err != nil {
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	antiFloodComponents, err := antifloodFactory.NewP2PAntiFloodComponents(
		ncf.mainConfig,
		ncf.statusHandler,
		netMessenger.ID(),
		ctx,
	)
	if err != nil {
		return nil, err
	}

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
		return nil, fmt.Errorf("%w when casting input antiflood handler to structs/P2PAntifloodHandler", ErrWrongTypeAssertion)
	}

	outAntifloodHandler, errOutputAntiflood := antifloodFactory.NewP2POutputAntiFlood(ncf.mainConfig, ctx)
	if errOutputAntiflood != nil {
		return nil, errOutputAntiflood
	}

	outputAntifloodHandler, ok := outAntifloodHandler.(P2PAntifloodHandler)
	if !ok {
		return nil, fmt.Errorf("%w when casting output antiflood handler to structs/P2PAntifloodHandler", ErrWrongTypeAssertion)
	}

	err = netMessenger.SetPeerBlackListHandler(antiFloodComponents.BlacklistHandler)
	if err != nil {
		return nil, err
	}

	cache, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(ncf.mainConfig.P2PMessageIDAdditionalCache))
	if err != nil {
		return nil, fmt.Errorf("%w while creating p2p cacher", err)
	}

	err = netMessenger.SetMessageIdsCacher(cache)
	if err != nil {
		return nil, fmt.Errorf("%w while setting p2p cacher", err)
	}

	return &networkComponents{
		netMessenger:           netMessenger,
		inputAntifloodHandler:  inputAntifloodHandler,
		outputAntifloodHandler: outputAntifloodHandler,
		topicFloodPreventer:    antiFloodComponents.TopicPreventer,
		floodPreventers:        antiFloodComponents.FloodPreventers,
		peerBlackListHandler:   antiFloodComponents.BlacklistHandler,
		antifloodConfig:        ncf.mainConfig.Antiflood,
		closeFunc:              cancelFunc,
	}, nil
}
