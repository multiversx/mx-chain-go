package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/debug/antiflood"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/process"
	antifloodFactory "github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/factory"
)

type networkComponentsFactory struct {
	p2pConfig     config.P2PConfig
	mainConfig    config.Config
	statusHandler core.AppStatusHandler
	listenAddress string
	marshalizer   marshal.Marshalizer
	syncer        p2p.SyncTimer
}

// NewNetworkComponentsFactory returns a new instance of a network components factory
func NewNetworkComponentsFactory(
	p2pConfig config.P2PConfig,
	mainConfig config.Config,
	statusHandler core.AppStatusHandler,
	marshalizer marshal.Marshalizer,
	syncer p2p.SyncTimer,
) (*networkComponentsFactory, error) {
	if check.IfNil(statusHandler) {
		return nil, ErrNilStatusHandler
	}
	if check.IfNil(marshalizer) {
		return nil, fmt.Errorf("%w in NewNetworkComponentsFactory", ErrNilMarshalizer)
	}

	return &networkComponentsFactory{
		p2pConfig:     p2pConfig,
		marshalizer:   marshalizer,
		mainConfig:    mainConfig,
		statusHandler: statusHandler,
		listenAddress: libp2p.ListenAddrWithIp4AndTcp,
		syncer:        syncer,
	}, nil
}

// Create creates and returns the network components
func (ncf *networkComponentsFactory) Create() (*NetworkComponents, error) {
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

	inAntifloodHandler, peerIdBlackList, pkTimeCache, errNewAntiflood := antifloodFactory.NewP2PAntiFloodAndBlackList(
		ncf.mainConfig,
		ncf.statusHandler,
		netMessenger.ID(),
	)
	if errNewAntiflood != nil {
		return nil, errNewAntiflood
	}

	if ncf.mainConfig.Debug.Antiflood.Enabled {
		var debugger process.AntifloodDebugger
		debugger, err = antiflood.NewAntifloodDebugger(ncf.mainConfig.Debug.Antiflood)
		if err != nil {
			return nil, err
		}

		err = inAntifloodHandler.SetDebugger(debugger)
		if err != nil {
			return nil, err
		}
	}

	inputAntifloodHandler, ok := inAntifloodHandler.(P2PAntifloodHandler)
	if !ok {
		return nil, fmt.Errorf("%w when casting input antiflood handler to structs/P2PAntifloodHandler", ErrWrongTypeAssertion)
	}

	outAntifloodHandler, errOutputAntiflood := antifloodFactory.NewP2POutputAntiFlood(ncf.mainConfig)
	if errOutputAntiflood != nil {
		return nil, errOutputAntiflood
	}

	outputAntifloodHandler, ok := outAntifloodHandler.(P2PAntifloodHandler)
	if !ok {
		return nil, fmt.Errorf("%w when casting output antiflood handler to structs/P2PAntifloodHandler", ErrWrongTypeAssertion)
	}

	return &NetworkComponents{
		NetMessenger:           netMessenger,
		InputAntifloodHandler:  inputAntifloodHandler,
		OutputAntifloodHandler: outputAntifloodHandler,
		PeerBlackListHandler:   peerIdBlackList,
		PkTimeCache:            pkTimeCache,
	}, nil
}
