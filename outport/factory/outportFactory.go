package factory

import (
	covalentFactory "github.com/ElrondNetwork/covalent-indexer-go/factory"
	indexerFactory "github.com/ElrondNetwork/elastic-indexer-go/factory"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/outport"
	notifierFactory "github.com/ElrondNetwork/notifier-go/factory"
)

// OutportFactoryArgs holds the factory arguments of different outport drivers
type OutportFactoryArgs struct {
	Port                       string
	NodeType                   core.NodeType
	ElasticIndexerFactoryArgs  *indexerFactory.ArgsIndexerFactory
	EventNotifierFactoryArgs   *notifierFactory.EventNotifierFactoryArgs
	CovalentIndexerFactoryArgs *covalentFactory.ArgsCovalentIndexerFactory
}

// CreateOutport will create a new instance of OutportHandler
func CreateOutport(args *OutportFactoryArgs) (outport.OutportHandler, error) {
	err := checkArguments(args)
	if err != nil {
		return nil, err
	}

	outportHandler := outport.NewOutport()
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

	err = createAndSubscribeCovalentDriverIfNeed(outport, args.CovalentIndexerFactoryArgs, args.NodeType)
	if err != nil {
		return err
	}

	return nil
}

func createAndSubscribeCovalentDriverIfNeed(
	outport outport.OutportHandler,
	args *covalentFactory.ArgsCovalentIndexerFactory,
	nodeType core.NodeType,
) error {
	if !args.Enabled || nodeType != core.NodeTypeObserver {
		return nil
	}

	covalentDriver, err := covalentFactory.CreateCovalentIndexer(args)
	if err != nil {
		return err
	}

	return outport.SubscribeDriver(covalentDriver)
}

func createAndSubscribeElasticDriverIfNeeded(
	outport outport.OutportHandler,
	args *indexerFactory.ArgsIndexerFactory,
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
	args *notifierFactory.EventNotifierFactoryArgs,
) error {
	if !args.Enabled {
		return nil
	}

	eventNotifier, err := notifierFactory.CreateEventNotifier(args)
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
