package factory

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/sender"
)

// ArgHeartbeatV2ComponentsFactory represents the argument for the heartbeat v2 components factory
type ArgHeartbeatV2ComponentsFactory struct {
	Config            config.Config
	Prefs             config.Preferences
	AppVersion        string
	RedundancyHandler heartbeat.NodeRedundancyHandler
	CoreComponents    CoreComponentsHolder
	DataComponents    DataComponentsHolder
	NetworkComponents NetworkComponentsHolder
	CryptoComponents  CryptoComponentsHolder
}

type heartbeatV2ComponentsFactory struct {
	config            config.Config
	prefs             config.Preferences
	version           string
	redundancyHandler heartbeat.NodeRedundancyHandler
	coreComponents    CoreComponentsHolder
	dataComponents    DataComponentsHolder
	networkComponents NetworkComponentsHolder
	cryptoComponents  CryptoComponentsHolder
}

type heartbeatV2Components struct {
	sender HeartbeatV2Sender
}

// NewHeartbeatV2ComponentsFactory creates a new instance of heartbeatV2ComponentsFactory
func NewHeartbeatV2ComponentsFactory(args ArgHeartbeatV2ComponentsFactory) (*heartbeatV2ComponentsFactory, error) {
	err := checkHeartbeatV2FactoryArgs(args)
	if err != nil {
		return nil, err
	}

	return &heartbeatV2ComponentsFactory{
		config:            args.Config,
		prefs:             args.Prefs,
		version:           args.AppVersion,
		redundancyHandler: args.RedundancyHandler,
		coreComponents:    args.CoreComponents,
		dataComponents:    args.DataComponents,
		networkComponents: args.NetworkComponents,
		cryptoComponents:  args.CryptoComponents,
	}, nil
}

func checkHeartbeatV2FactoryArgs(args ArgHeartbeatV2ComponentsFactory) error {
	if check.IfNil(args.CoreComponents) {
		return errors.ErrNilCoreComponentsHolder
	}
	if check.IfNil(args.DataComponents) {
		return errors.ErrNilDataComponentsHolder
	}
	if check.IfNil(args.NetworkComponents) {
		return errors.ErrNilNetworkComponentsHolder
	}
	if check.IfNil(args.CryptoComponents) {
		return errors.ErrNilCryptoComponentsHolder
	}

	return nil
}

// Create creates the heartbeatV2 components
func (hcf *heartbeatV2ComponentsFactory) Create() (*heartbeatV2Components, error) {
	peerSubType := core.RegularPeer
	if hcf.prefs.Preferences.FullArchive {
		peerSubType = core.FullHistoryObserver
	}

	cfg := hcf.config.HeartbeatV2

	argsSender := sender.ArgSender{
		Messenger:                          hcf.networkComponents.NetworkMessenger(),
		Marshaller:                         hcf.coreComponents.InternalMarshalizer(),
		PeerAuthenticationTopic:            common.PeerAuthenticationTopic,
		HeartbeatTopic:                     common.HeartbeatV2Topic,
		PeerAuthenticationTimeBetweenSends: time.Second * time.Duration(cfg.PeerAuthenticationTimeBetweenSendsInSec),
		PeerAuthenticationTimeBetweenSendsWhenError: time.Second * time.Duration(cfg.PeerAuthenticationTimeBetweenSendsWhenErrorInSec),
		PeerAuthenticationThresholdBetweenSends:     cfg.PeerAuthenticationThresholdBetweenSends,
		HeartbeatTimeBetweenSends:                   time.Second * time.Duration(cfg.HeartbeatTimeBetweenSendsInSec),
		HeartbeatTimeBetweenSendsWhenError:          time.Second * time.Duration(cfg.HeartbeatTimeBetweenSendsWhenErrorInSec),
		HeartbeatThresholdBetweenSends:              cfg.HeartbeatThresholdBetweenSends,
		VersionNumber:                               hcf.version,
		NodeDisplayName:                             hcf.prefs.Preferences.NodeDisplayName,
		Identity:                                    hcf.prefs.Preferences.Identity,
		PeerSubType:                                 peerSubType,
		CurrentBlockProvider:                        hcf.dataComponents.Blockchain(),
		PeerSignatureHandler:                        hcf.cryptoComponents.PeerSignatureHandler(),
		PrivateKey:                                  hcf.cryptoComponents.PrivateKey(),
		RedundancyHandler:                           hcf.redundancyHandler,
	}
	heartbeatV2Sender, err := sender.NewSender(argsSender)
	if err != nil {
		return nil, err
	}

	return &heartbeatV2Components{
		sender: heartbeatV2Sender,
	}, nil
}

// Close closes the heartbeat components
func (hc *heartbeatV2Components) Close() error {
	log.Debug("calling close on heartbeatV2 components")

	if !check.IfNil(hc.sender) {
		log.LogIfError(hc.sender.Close())
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hcf *heartbeatV2ComponentsFactory) IsInterfaceNil() bool {
	return hcf == nil
}
