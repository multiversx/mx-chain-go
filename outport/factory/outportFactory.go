package factory

import (
	"os"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	indexerFactory "github.com/multiversx/mx-chain-es-indexer-go/process/factory"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/outport/firehose"
)

// OutportFactoryArgs holds the factory arguments of different outport drivers
type OutportFactoryArgs struct {
	RetrialInterval           time.Duration
	ElasticIndexerFactoryArgs indexerFactory.ArgsIndexerFactory
	EventNotifierFactoryArgs  *EventNotifierFactoryArgs
	HostDriverArgs            ArgsHostDriverFactory
	FireHoseIndexerConfig     config.FireHoseConfig
}

// CreateOutport will create a new instance of OutportHandler
func CreateOutport(args *OutportFactoryArgs) (outport.OutportHandler, error) {
	err := checkArguments(args)
	if err != nil {
		return nil, err
	}

	outportHandler, err := outport.NewOutport(args.RetrialInterval)
	if err != nil {
		return nil, err
	}

	err = createAndSubscribeDrivers(outportHandler, args)
	if err != nil {
		return nil, err
	}

	return outportHandler, nil
}

func createAndSubscribeDrivers(outport outport.OutportHandler, args *OutportFactoryArgs) error {
	err := createAndSubscribeElasticDriverIfNeeded(outport, args.ElasticIndexerFactoryArgs)
	if err != nil {
		return err
	}

	err = createAndSubscribeEventNotifierIfNeeded(outport, args.EventNotifierFactoryArgs)
	if err != nil {
		return err
	}

	err = createAndSubscribeHostDriverIfNeeded(outport, args.HostDriverArgs)
	if err != nil {
		return err
	}

	return createAndSubscribeFirehoseIndexerDriver(outport, args.FireHoseIndexerConfig)
}

func createAndSubscribeElasticDriverIfNeeded(
	outport outport.OutportHandler,
	args indexerFactory.ArgsIndexerFactory,
) error {
	if !args.Enabled {
		return nil
	}

	elasticDriver, err := indexerFactory.NewIndexer(args)
	if err != nil {
		return err
	}

	return outport.SubscribeDriver(elasticDriver)
}

func createAndSubscribeEventNotifierIfNeeded(
	outport outport.OutportHandler,
	args *EventNotifierFactoryArgs,
) error {
	if !args.Enabled {
		return nil
	}

	eventNotifier, err := CreateEventNotifier(args)
	if err != nil {
		return err
	}

	return outport.SubscribeDriver(eventNotifier)
}

func checkArguments(args *OutportFactoryArgs) error {
	if args == nil {
		return outport.ErrNilArgsOutportFactory
	}

	return nil
}

func createAndSubscribeHostDriverIfNeeded(
	outport outport.OutportHandler,
	args ArgsHostDriverFactory,
) error {
	if !args.HostConfig.Enabled {
		return nil
	}

	hostDriver, err := CreateHostDriver(args)
	if err != nil {
		return err
	}

	return outport.SubscribeDriver(hostDriver)
}

func createAndSubscribeFirehoseIndexerDriver(
	outport outport.OutportHandler,
	args config.FireHoseConfig,
) error {
	if !args.Enabled {
		return nil
	}

	container := block.NewEmptyBlockCreatorsContainer()
	err := container.Add(core.ShardHeaderV1, block.NewEmptyHeaderCreator())
	err = container.Add(core.ShardHeaderV2, block.NewEmptyHeaderV2Creator())
	err = container.Add(core.MetaHeader, block.NewEmptyMetaBlockCreator())

	fireHoseIndexer, err := firehose.NewFirehoseIndexer(os.Stdout, container)
	if err != nil {
		return err
	}

	return outport.SubscribeDriver(fireHoseIndexer)
}
