package factory

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/outport"
	driversFactory "github.com/ElrondNetwork/elrond-go/outport/drivers/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("outport/factory")

// ArgsOutportFactory is structure used to store all components that are needed to create a new instance of
// DriverHandler
type ArgsOutportFactory struct {
	ArgsElasticDriver  *driversFactory.ArgsElasticDriverFactory
	EpochStartNotifier sharding.EpochStartEventNotifier
	NodesCoordinator   sharding.NodesCoordinator
}

// CreateOutport will create a new instance of OutportHandler
func CreateOutport(args *ArgsOutportFactory) (outport.OutportHandler, error) {
	outportHandler := outport.NewOutport()

	shardID := args.ArgsElasticDriver.ShardCoordinator.SelfId()

	if shardID == core.MetachainShardId {
		handler := epochStartEventHandler(outportHandler.SaveValidatorsPubKeys, args.NodesCoordinator)
		args.EpochStartNotifier.RegisterHandler(handler)
	}

	err := subscribeDrivers(outportHandler, args)
	if err != nil {
		return nil, err
	}

	return outportHandler, nil
}

func subscribeDrivers(outport outport.OutportHandler, args *ArgsOutportFactory) error {
	err := createElasticDriverAndRegisterIfNeeded(outport, args.ArgsElasticDriver)
	if err != nil {
		return err
	}

	return nil
}

func createElasticDriverAndRegisterIfNeeded(
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

func epochStartEventHandler(
	saveBLSKeys func(validatorsPubKeys map[uint32][][]byte, epoch uint32),
	coordinator sharding.NodesCoordinator,
) epochStart.ActionHandler {
	subscribeHandler := notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {
		currentEpoch := hdr.GetEpoch()
		validatorsPubKeys, err := coordinator.GetAllEligibleValidatorsPublicKeys(currentEpoch)
		if err != nil {
			log.Warn("GetAllEligibleValidatorPublicKeys for current epoch failed",
				"epoch", currentEpoch,
				"error", err.Error())
		}

		saveBLSKeys(validatorsPubKeys, currentEpoch)

	}, func(_ data.HeaderHandler) {}, core.IndexerOrder)

	return subscribeHandler
}
