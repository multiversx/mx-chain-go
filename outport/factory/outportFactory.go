package factory

import (
	indexerFactory "github.com/ElrondNetwork/elastic-indexer-go/factory"
	"github.com/ElrondNetwork/elrond-go/outport"
)

// ArgsOutportFactory is structure used to store all components that are needed to create a new instance of
// OurportHandler
type ArgsOutportFactory struct {
	ArgsElasticDriver *indexerFactory.ArgsIndexerFactory
}

// CreateOutport will create a new instance of OutportHandler
func CreateOutport(args *ArgsOutportFactory) (outport.OutportHandler, error) {
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

func createAndSubscribeDrivers(outport outport.OutportHandler, args *ArgsOutportFactory) error {
	err := createAndSubscribeElasticDriverIfNeeded(outport, args.ArgsElasticDriver)
	if err != nil {
		return err
	}

	return nil
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

func checkArguments(args *ArgsOutportFactory) error {
	if args == nil {
		return outport.ErrNilArgsOutportFactory
	}
	if args.ArgsElasticDriver == nil {
		return outport.ErrNilArgsElasticDriverFactory
	}

	return nil
}
