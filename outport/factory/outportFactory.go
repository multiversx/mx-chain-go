package factory

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/outport"
	driversFactory "github.com/ElrondNetwork/elrond-go/outport/drivers/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("outport/factory")

// ArgsOutportFactory is structure used to store all components that are needed to create a new instance of
// OurportHandler
type ArgsOutportFactory struct {
	ArgsElasticDriver  *driversFactory.ArgsElasticDriverFactory
	EpochStartNotifier sharding.EpochStartEventNotifier
	NodesCoordinator   sharding.NodesCoordinator
}

// CreateOutport will create a new instance of OutportHandler
func CreateOutport(args *ArgsOutportFactory) (outport.OutportHandler, error) {
	err := checkArguments(args)
	if err != nil {
		return nil, err
	}

	outportHandler := outport.NewOutport()

	registerEpochHandler(
		outportHandler,
		args.EpochStartNotifier,
		args.ArgsElasticDriver.ShardCoordinator,
		args.NodesCoordinator,
	)

	err = createAndSubscribeDrivers(outportHandler, args)
	if err != nil {
		return nil, err
	}

	return outportHandler, nil
}

func createAndSubscribeDrivers(outport outport.OutportHandler, args *ArgsOutportFactory) error {
	err := createAndSubscribeElasticDriverIfNeeded(outport, args.ArgsElasticDriver)
	if err != nil {
		return err
	}

	return nil
}

func createAndSubscribeElasticDriverIfNeeded(
	outport outport.OutportHandler,
	args *driversFactory.ArgsElasticDriverFactory,
) error {
	if !args.Enabled {
		return nil
	}

	elasticDriver, err := driversFactory.NewElasticClient(args)
	if err != nil {
		return err
	}

	return outport.SubscribeDriver(elasticDriver)
}

func registerEpochHandler(
	outportHandler outport.OutportHandler,
	startNotifier sharding.EpochStartEventNotifier,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,

) {
	shardID := shardCoordinator.SelfId()
	if shardID != core.MetachainShardId {
		return
	}

	handler := epochStartEventHandler(outportHandler.SaveValidatorsPubKeys, nodesCoordinator)
	startNotifier.RegisterHandler(handler)
}

func epochStartEventHandler(
	saveBLSKeys func(validatorsPubKeys map[uint32][][]byte, epoch uint32),
	coordinator sharding.NodesCoordinator,
) epochStart.ActionHandler {
	subscribeHandler := notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {
		currentEpoch := hdr.GetEpoch()
		validatorsPubKeys, err := coordinator.GetAllEligibleValidatorsPublicKeys(currentEpoch)
		if err != nil {
			log.Error("GetAllEligibleValidatorPublicKeys for current epoch failed",
				"epoch", currentEpoch,
				"error", err.Error())
		}

		saveBLSKeys(validatorsPubKeys, currentEpoch)

	}, func(_ data.HeaderHandler) {}, core.OutportOrder)

	return subscribeHandler
}

func checkArguments(args *ArgsOutportFactory) error {
	if args == nil {
		return outport.ErrNilArgsOutportFactory
	}
	if args.ArgsElasticDriver == nil {
		return outport.ErrNilArgsElasticDriverFactory
	}
	if check.IfNil(args.ArgsElasticDriver.ShardCoordinator) {
		return outport.ErrNilShardCoordinator
	}
	if check.IfNil(args.NodesCoordinator) {
		return outport.ErrNilNodesCoordinator
	}
	if check.IfNil(args.EpochStartNotifier) {
		return outport.ErrNilEpochStartNotifier
	}

	return nil
}
