package factory

import (
	"context"
	"fmt"
	"time"

	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/outport/grpcadapter"
	"github.com/multiversx/mx-chain-core-go/marshal"
	indexerFactory "github.com/multiversx/mx-chain-es-indexer-go/process/factory"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/outport/grpcdriver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OutportFactoryArgs holds the factory arguments of different outport drivers
type OutportFactoryArgs struct {
	IsImportDB                bool
	ShardID                   uint32
	RetrialInterval           time.Duration
	ElasticIndexerFactoryArgs indexerFactory.ArgsIndexerFactory
	EventNotifierFactoryArgs  *EventNotifierFactoryArgs
	HostDriversArgs           []ArgsHostDriverFactory
	GRPCDriversArgs           []ArgsGRPCDriverFactory
	EnableEpochsHandler       common.EnableEpochsHandler
	EnableRoundsHandler       common.EnableRoundsHandler
}

// CreateOutport will create a new instance of OutportHandler
func CreateOutport(args *OutportFactoryArgs) (outport.OutportHandler, error) {
	err := checkArguments(args)
	if err != nil {
		return nil, err
	}

	cfg := outportcore.OutportConfig{
		ShardID:          args.ShardID,
		IsInImportDBMode: args.IsImportDB,
	}

	outportHandler, err := outport.NewOutport(
		args.RetrialInterval,
		cfg,
		args.EnableEpochsHandler,
		args.EnableRoundsHandler,
	)
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

	for idx := 0; idx < len(args.GRPCDriversArgs); idx++ {
		err = createAndSubscribeGRPCDriverIfNeeded(outport, args.GRPCDriversArgs[idx])
		if err != nil {
			return fmt.Errorf("%w when calling createAndSubscribeGRPCDriverIfNeeded, grpc driver index %d", err, idx)
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

type ArgsGRPCDriverFactory struct {
	GRPCClient config.GRPCDriversConfig
	Marshaller marshal.Marshalizer
}

func createAndSubscribeGRPCDriverIfNeeded(
	outport outport.OutportHandler,
	args ArgsGRPCDriverFactory,
) error {
	if !args.GRPCClient.Enabled {
		return nil
	}

	grpcClient, err := grpcadapter.NewOutportGRPCClient(
		args.GRPCClient.URL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(logGRPCOutportCalls),
	)
	if err != nil {
		return err
	}

	grpcDriver, err := grpcdriver.NewGRPCDriver(grpcClient, args.Marshaller)
	if err != nil {
		return err
	}

	return outport.SubscribeDriver(grpcDriver)
}

func logGRPCOutportCalls(
	ctx context.Context,
	method string,
	req any,
	reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Debug("grpc call", "method", method, "duration", time.Since(start))
	return err
}
