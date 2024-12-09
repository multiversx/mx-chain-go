package epochStartTrigger

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/metachain"
	"github.com/multiversx/mx-chain-go/epochStart/shardchain"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
)

// TODO: MX-15632 Unit tests + fix import cycle

type sovereignEpochStartTriggerFactory struct {
}

// NewSovereignEpochStartTriggerFactory creates a sovereign epoch start trigger. This will be a metachain one, since
// nodes inside sovereign chain will not need to receive meta information, but they will actually execute the meta code.
func NewSovereignEpochStartTriggerFactory() *sovereignEpochStartTriggerFactory {
	return &sovereignEpochStartTriggerFactory{}
}

// CreateEpochStartTrigger creates a meta epoch start trigger for sovereign run type
func (f *sovereignEpochStartTriggerFactory) CreateEpochStartTrigger(args factory.ArgsEpochStartTrigger) (epochStart.TriggerHandler, error) {
	metaTriggerArgs, err := createMetaEpochStartTriggerArgs(args)
	if err != nil {
		return nil, err
	}

	err = checkNilArgs(args)
	if err != nil {
		return nil, err
	}

	argsPeerMiniBlockSyncer := shardchain.ArgPeerMiniBlockSyncer{
		MiniBlocksPool:     args.DataComps.Datapool().MiniBlocks(),
		ValidatorsInfoPool: args.DataComps.Datapool().ValidatorsInfo(),
		RequestHandler:     args.RequestHandler,
	}
	peerMiniBlockSyncer, err := shardchain.NewPeerMiniBlockSyncer(argsPeerMiniBlockSyncer)
	if err != nil {
		return nil, err
	}

	argsSovTrigger := metachain.ArgsSovereignTrigger{
		ArgsNewMetaEpochStartTrigger: metaTriggerArgs,
		ValidatorInfoSyncer:          peerMiniBlockSyncer,
	}

	return metachain.NewSovereignTrigger(argsSovTrigger)
}

func checkNilArgs(args factory.ArgsEpochStartTrigger) error {
	if check.IfNil(args.DataComps) {
		return process.ErrNilDataComponentsHolder
	}
	if check.IfNil(args.DataComps.Datapool()) {
		return process.ErrNilDataPoolHolder
	}
	if check.IfNil(args.DataComps.Datapool().MiniBlocks()) {
		return dataRetriever.ErrNilMiniblocksPool
	}
	if check.IfNil(args.DataComps.Datapool().ValidatorsInfo()) {
		return process.ErrNilValidatorInfoPool
	}
	if check.IfNil(args.RequestHandler) {
		return process.ErrNilRequestHandler
	}

	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignEpochStartTriggerFactory) IsInterfaceNil() bool {
	return f == nil
}
