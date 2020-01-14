package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/epochStart/shardchain"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
	containers "github.com/ElrondNetwork/elrond-go/update/container"
	"github.com/ElrondNetwork/elrond-go/update/files"
	"github.com/ElrondNetwork/elrond-go/update/genesis"
)

type ArgsNewExportHandlerFactory struct {
	ShardCoordinator sharding.Coordinator
	Hasher           hashing.Hasher
	Marshalizer      marshal.Marshalizer
	Writer           update.MultiFileWriter
}

type exportHandlerFactory struct {
	shardCoordinator sharding.Coordinator
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer

	writer update.MultiFileWriter
}

func NewExportHandlerFactory(args ArgsNewExportHandlerFactory) (*exportHandlerFactory, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, sharding.ErrNilShardCoordinator
	}
	if check.IfNil(args.Hasher) {
		return nil, sharding.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, data.ErrNilMarshalizer
	}

	argsWriter := files.ArgsNewMultiFileWriter{
		ExportFolder: "",
		ExportStore:  nil,
	}
	writer, err := files.NewMultiFileWriter(argsWriter)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (e *exportHandlerFactory) Create() (update.ExportHandler, error) {
	args := genesis.ArgsNewStateExporter{
		ShardCoordinator: e.shardCoordinator,
		StateSyncer:      nil,
		Marshalizer:      e.marshalizer,
		Writer:           nil,
	}
	exportHandler, err := genesis.NewStateExporter(args)
	if err != nil {
		return nil, err
	}

	return exportHandler, nil
}

func (e *exportHandlerFactory) IsInterfaceNil() bool {
	return e == nil
}

type ArgsExporter struct {
}

func NewExporter(args ArgsExporter) (update.ExportHandler, error) {

	argsEpochTrigger := shardchain.ArgsShardEpochStartTrigger{
		Marshalizer:        args.Marshalizer,
		Hasher:             args.Hasher,
		HeaderValidator:    args.HeaderValidator,
		Uint64Converter:    args.Uint64Converter,
		DataPool:           args.DataPool,
		Storage:            args.StorageService,
		RequestHandler:     args.RequestHandler,
		EpochStartNotifier: notifier.NewEpochStartSubscriptionHandler(),
		Epoch:              0,
		Validity:           process.MetaBlockValidity,
		Finality:           process.MetaBlockFinality,
	}
	epochHandler, err := shardchain.NewEpochStartTrigger(&argsEpochTrigger)
	if err != nil {
		return nil, err
	}

	argsSyncState := genesis.ArgsNewSyncState{
		Hasher:           args.Hasher,
		Marshalizer:      args.Marshalizer,
		ShardCoordinator: args.ShardCoordinator,
		TrieSyncers:      trieSyncersContainer,
		EpochHandler:     epochHandler,
		Storages:         args.StorageService,
		DataPools:        args.DataPool,
		RequestHandler:   args.RequestHandler,
		HeaderValidator:  args.HeaderValidator,
		ActiveDataTries:  activeDataTriesContainer,
	}
	stateSyncer, err := genesis.NewSyncState(argsSyncState)
	if err != nil {
		return nil, err
	}

	argsWriter := files.ArgsNewMultiFileWriter{
		ExportFolder: args.ExportFolder,
		ExportStore:  args.ExportStore,
	}
	writer, err := files.NewMultiFileWriter(argsWriter)
	if err != nil {
		return nil, err
	}

	argsExporter := genesis.ArgsNewStateExporter{
		ShardCoordinator: args.ShardCoordinator,
		StateSyncer:      stateSyncer,
		Marshalizer:      args.Marshalizer,
		Writer:           writer,
	}
	exportHandler, err := genesis.NewStateExporter(argsExporter)
	if err != nil {
		return nil, err
	}

	return exportHandler, nil
}
