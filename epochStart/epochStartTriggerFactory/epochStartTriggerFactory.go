package epochStartTriggerFactory

import (
	"errors"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
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

// ArgsEpochStartTrigger is a struct placeholder for arguments needed to create an epoch start trigger
type ArgsEpochStartTrigger struct {
	RequestHandler             epochStart.RequestHandler
	CoreData                   factory.CoreComponentsHolder
	BootstrapComponents        factory.BootstrapComponentsHandler
	DataComps                  factory.DataComponentsHolder
	StatusCoreComponentsHolder process.StatusCoreComponentsHolder
	RunTypeComponentsHolder    factory.RunTypeComponentsHolder
	Config                     config.Config
}

// NewEpochStartTriggerFactory creates a normal run type chain epoch start trigger. It will create epoch start triggers
// (meta/shard) based on shard coordinator. In an ideal case, this should be further broken down into separate factories
// for each shard.
func NewEpochStartTriggerFactory() *epochStartTriggerFactory {
	return &epochStartTriggerFactory{}
}

// CreateEpochStartTrigger creates an epoch start trigger for normal run type
func (f *epochStartTriggerFactory) CreateEpochStartTrigger(args ArgsEpochStartTrigger) (process.EpochStartTriggerHandler, error) {
	shardCoordinator := args.BootstrapComponents.ShardCoordinator()

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return createShardEpochStartTrigger(args)
	}
	if shardCoordinator.SelfId() == core.MetachainShardId {
		return createMetaEpochStartTrigger(args)
	}

	return nil, errors.New("error creating new start of epoch trigger because of invalid shard id")
}

func createShardEpochStartTrigger(args ArgsEpochStartTrigger) (process.EpochStartTriggerHandler, error) {
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

func createMetaEpochStartTrigger(args ArgsEpochStartTrigger) (process.EpochStartTriggerHandler, error) {
	genesisHeader := args.DataComps.Blockchain().GetGenesisHeader()
	if check.IfNil(genesisHeader) {
		return nil, errorsMx.ErrGenesisBlockNotInitialized
	}

	argEpochStart := &metachain.ArgsNewMetaEpochStartTrigger{
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
	}

	return metachain.NewEpochStartTrigger(argEpochStart)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *epochStartTriggerFactory) IsInterfaceNil() bool {
	return f == nil
}
