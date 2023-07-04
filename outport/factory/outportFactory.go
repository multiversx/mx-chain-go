package factory

import (
	"fmt"
	"time"

	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	indexerFactory "github.com/multiversx/mx-chain-es-indexer-go/process/factory"
	"github.com/multiversx/mx-chain-go/outport"
)

// OutportFactoryArgs holds the factory arguments of different outport drivers
type OutportFactoryArgs struct {
	IsImportDB                bool
	RetrialInterval           time.Duration
	ElasticIndexerFactoryArgs indexerFactory.ArgsIndexerFactory
	EventNotifierFactoryArgs  *EventNotifierFactoryArgs
	HostDriversArgs           []ArgsHostDriverFactory
}

// CreateOutport will create a new instance of OutportHandler
func CreateOutport(args *OutportFactoryArgs) (outport.OutportHandler, error) {
	err := checkArguments(args)
	if err != nil {
		return nil, err
	}

	cfg := outportcore.OutportConfig{
		IsInImportDBMode: args.IsImportDB,
	}

	outportHandler, err := outport.NewOutport(args.RetrialInterval, cfg)
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

	for idx := 0; idx < len(args.HostDriversArgs); idx++ {
		err = createAndSubscribeHostDriverIfNeeded(outport, args.HostDriversArgs[idx])
		if err != nil {
			return fmt.Errorf("%w when calling createAndSubscribeHostDriverIfNeeded, host driver index %d", err, idx)
		}
	}

	return nil
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
