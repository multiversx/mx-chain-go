package consensus

import (
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/sync"
)

type consensusComponentsFactoryV2 struct {
	*consensusComponentsFactory
}

// NewConsensusComponentsFactoryV2 creates an instance of consensusComponentsFactoryV2
func NewConsensusComponentsFactoryV2(consensusComponentsFactory *consensusComponentsFactory) (*consensusComponentsFactoryV2, error) {
	if consensusComponentsFactory == nil {
		return nil, sposFactory.ErrNilConsensusComponentFactory
	}

	ccf := &consensusComponentsFactoryV2{
		consensusComponentsFactory,
	}

	ccf.getSubroundsFactoryMethod = ccf.getSubroundsFactory
	ccf.createShardSyncBootstrapperMethod = ccf.createShardSyncBootstrapper

	return ccf, nil
}

func (ccf *consensusComponentsFactoryV2) getSubroundsFactory(
	consensusDataContainer *spos.ConsensusCore,
	consensusState *spos.ConsensusState,
	cc *consensusComponents,
) (spos.SubroundsFactory, error) {
	return sposFactory.GetSubroundsFactoryV2(
		consensusDataContainer,
		consensusState,
		cc.worker,
		ccf.config.Consensus.Type,
		ccf.coreComponents.StatusHandler(),
		ccf.statusComponents.OutportHandler(),
		[]byte(ccf.coreComponents.ChainID()),
		ccf.networkComponents.NetworkMessenger().ID(),
	)
}

func (ccf *consensusComponentsFactoryV2) createShardSyncBootstrapper(argsBaseBootstrapper sync.ArgBaseBootstrapper) (process.Bootstrapper, error) {
	argsShardBootstrapper := sync.ArgShardBootstrapper{
		ArgBaseBootstrapper: argsBaseBootstrapper,
	}

	bootstrap, err := sync.NewShardBootstrap(argsShardBootstrapper)
	if err != nil {
		return nil, err
	}

	return sync.NewSovereignChainShardBootstrap(bootstrap)
}
