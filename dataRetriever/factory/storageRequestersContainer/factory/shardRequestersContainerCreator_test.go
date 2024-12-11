package factory

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/common/statistics/disabled"
	storagerequesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/storageRequestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func createFactoryArgs() storagerequesterscontainer.FactoryArgs {
	return storagerequesterscontainer.FactoryArgs{
		ChainID:                  "T",
		WorkingDirectory:         "",
		Hasher:                   &hashingMocks.HasherMock{},
		ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
		Messenger:                &p2pmocks.MessengerStub{},
		Store:                    &storage.ChainStorerStub{},
		Marshalizer:              &mock.MarshalizerMock{},
		Uint64ByteSliceConverter: &mock.Uint64ByteSliceConverterMock{},
		DataPacker:               &mock.DataPackerStub{},
		ManualEpochStartNotifier: &mock.ManualEpochStartNotifierStub{},
		ChanGracefullyClose:      make(chan endProcess.ArgEndProcess),
		EnableEpochsHandler:      &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		StateStatsHandler:        disabled.NewStateStatistics(),
	}
}

func TestShardRequestersContainerCreator_CreateShardRequestersContainerFactory(t *testing.T) {
	t.Parallel()

	creator := NewShardRequestersContainerCreator()
	require.False(t, creator.IsInterfaceNil())

	args := createFactoryArgs()
	container, err := creator.CreateShardRequestersContainerFactory(args)
	require.Nil(t, err)
	require.Equal(t, "*storagerequesterscontainer.shardRequestersContainerFactory", fmt.Sprintf("%T", container))
}
