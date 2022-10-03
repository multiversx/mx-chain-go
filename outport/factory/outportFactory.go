package factory

import (
	"time"

	covalentFactory "github.com/ElrondNetwork/covalent-indexer-go/factory"
	indexerFactory "github.com/ElrondNetwork/elastic-indexer-go/process/factory"
	wsDriverFactory "github.com/ElrondNetwork/elrond-go-core/websocketOutportDriver/factory"
	"github.com/ElrondNetwork/elrond-go/outport"
)

// WrappedOutportDriverWebSocketSenderFactoryArgs extends the wsDriverFactory.OutportDriverWebSocketSenderFactoryArgs structure with the Enabled field
type WrappedOutportDriverWebSocketSenderFactoryArgs struct {
	Enabled bool
	wsDriverFactory.OutportDriverWebSocketSenderFactoryArgs
}

// OutportFactoryArgs holds the factory arguments of different outport drivers
type OutportFactoryArgs struct {
	RetrialInterval                  time.Duration
	ElasticIndexerFactoryArgs        *indexerFactory.ArgsIndexerFactory
	EventNotifierFactoryArgs         *EventNotifierFactoryArgs
	CovalentIndexerFactoryArgs       *covalentFactory.ArgsCovalentIndexerFactory
	WebSocketSenderDriverFactoryArgs WrappedOutportDriverWebSocketSenderFactoryArgs
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

	err = createAndSubscribeCovalentDriverIfNeeded(outport, args.CovalentIndexerFactoryArgs)
	if err != nil {
		return err
	}

	return createAndSubscribeWebSocketDriver(outport, args.WebSocketSenderDriverFactoryArgs)
}

func createAndSubscribeCovalentDriverIfNeeded(
	outport outport.OutportHandler,
	args *covalentFactory.ArgsCovalentIndexerFactory,
) error {
	if !args.Enabled {
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

func createAndSubscribeWebSocketDriver(
	outport outport.OutportHandler,
	args WrappedOutportDriverWebSocketSenderFactoryArgs,
) error {
	if !args.Enabled {
		return nil
	}

	wsFactory, err := wsDriverFactory.NewOutportDriverWebSocketSenderFactory(args.OutportDriverWebSocketSenderFactoryArgs)
	if err != nil {
		return err
	}

	wsDriver, err := wsFactory.Create()
	if err != nil {
		return err
	}

	return outport.SubscribeDriver(wsDriver)
}
