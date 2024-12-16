package epochStartTrigger

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/metachain"
	"github.com/multiversx/mx-chain-go/epochStart/shardchain"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
)

type epochStartTriggerFactory struct {
}

// NewEpochStartTriggerFactory creates a normal run type chain epoch start trigger. It will create epoch start triggers
// (meta/shard) based on shard coordinator. In an ideal case, this should be further broken down into separate factories
// for each shard.
func NewEpochStartTriggerFactory() *epochStartTriggerFactory {
	return &epochStartTriggerFactory{}
}

// CreateEpochStartTrigger creates an epoch start trigger for normal run type
func (f *epochStartTriggerFactory) CreateEpochStartTrigger(args factory.ArgsEpochStartTrigger) (epochStart.TriggerHandler, error) {
	err := checkNilArgs(args)
	if err != nil {
		return nil, err
	}

	shardCoordinator := args.BootstrapComponents.ShardCoordinator()

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return createShardEpochStartTrigger(args)
	}
	if shardCoordinator.SelfId() == core.MetachainShardId {
		return createMetaEpochStartTrigger(args)
	}

	return nil, fmt.Errorf("error creating new start of epoch trigger, errror: %w", process.ErrInvalidShardId)
}

func createShardEpochStartTrigger(args factory.ArgsEpochStartTrigger) (epochStart.TriggerHandler, error) {
	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:      args.CoreData.Hasher(),
		Marshalizer: args.CoreData.InternalMarshalizer(),
	}
	headerValidator, err := args.RunTypeComponentsHolder.HeaderValidatorCreator().CreateHeaderValidator(argsHeaderValidator)
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

	argEpochStart := &shardchain.ArgsShardEpochStartTrigger{
		Marshalizer:                   args.CoreData.InternalMarshalizer(),
		Hasher:                        args.CoreData.Hasher(),
		HeaderValidator:               headerValidator,
		Uint64Converter:               args.CoreData.Uint64ByteSliceConverter(),
		DataPool:                      args.DataComps.Datapool(),
		Storage:                       args.DataComps.StorageService(),
		RequestHandler:                args.RequestHandler,
		Epoch:                         args.BootstrapComponents.EpochBootstrapParams().Epoch(),
		EpochStartNotifier:            args.CoreData.EpochStartNotifierWithConfirm(),
		Validity:                      process.MetaBlockValidity,
		Finality:                      process.BlockFinality,
		PeerMiniBlocksSyncer:          peerMiniBlockSyncer,
		RoundHandler:                  args.CoreData.RoundHandler(),
		AppStatusHandler:              args.StatusCoreComponentsHolder.AppStatusHandler(),
		EnableEpochsHandler:           args.CoreData.EnableEpochsHandler(),
		ExtraDelayForRequestBlockInfo: time.Duration(args.Config.EpochStartConfig.ExtraDelayForRequestBlockInfoInMilliseconds) * time.Millisecond,
	}
	return shardchain.NewEpochStartTrigger(argEpochStart)
}

func createMetaEpochStartTriggerArgs(args factory.ArgsEpochStartTrigger) (*metachain.ArgsNewMetaEpochStartTrigger, error) {
	genesisHeader := args.DataComps.Blockchain().GetGenesisHeader()
	if check.IfNil(genesisHeader) {
		return nil, errorsMx.ErrGenesisBlockNotInitialized
	}

	return &metachain.ArgsNewMetaEpochStartTrigger{
		GenesisTime:        time.Unix(args.CoreData.GenesisNodesSetup().GetStartTime(), 0),
		Settings:           &args.Config.EpochStartConfig,
		Epoch:              args.BootstrapComponents.EpochBootstrapParams().Epoch(),
		EpochStartRound:    genesisHeader.GetRound(),
		EpochStartNotifier: args.CoreData.EpochStartNotifierWithConfirm(),
		Storage:            args.DataComps.StorageService(),
		Marshalizer:        args.CoreData.InternalMarshalizer(),
		Hasher:             args.CoreData.Hasher(),
		AppStatusHandler:   args.StatusCoreComponentsHolder.AppStatusHandler(),
		DataPool:           args.DataComps.Datapool(),
	}, nil
}

func createMetaEpochStartTrigger(args factory.ArgsEpochStartTrigger) (epochStart.TriggerHandler, error) {
	argEpochStart, err := createMetaEpochStartTriggerArgs(args)
	if err != nil {
		return nil, err
	}

	return metachain.NewEpochStartTrigger(argEpochStart)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *epochStartTriggerFactory) IsInterfaceNil() bool {
	return f == nil
}
