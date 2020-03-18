package metachain

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func createMockEpochEconomicsArguments() ArgsNewEpochEconomics {
	shardCoordinator := mock.NewOneShardCoordinatorMock()

	argsNewEpochEconomics := ArgsNewEpochEconomics{
		Marshalizer:      &mock.MarshalizerMock{},
		Store:            createMetaStore(),
		ShardCoordinator: shardCoordinator,
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		RewardsHandler:   &mock.RewardsHandlerMock{},
		RoundTime:        &mock.RounderMock{},
	}
	return argsNewEpochEconomics
}

//TODO - see why the EpochEconomics returns process.Err
func TestEpochEconomics_NewEndOfEpochEconomicsDataCreatorNilMarshalizer(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochEconomicsArguments()
	arguments.Marshalizer = nil

	esd, err := NewEndOfEpochEconomicsDataCreator(arguments)
	require.Nil(t, esd)
	require.Equal(t, process.ErrNilMarshalizer, err)
}

func TestEpochEconomics_NewEndOfEpochEconomicsDataCreatorNilStore(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochEconomicsArguments()
	arguments.Store = nil

	esd, err := NewEndOfEpochEconomicsDataCreator(arguments)
	require.Nil(t, esd)
	require.Equal(t, process.ErrNilStore, err)
}

func TestEpochEconomics_NewEndOfEpochEconomicsDataCreatorNilShardCoordinator(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochEconomicsArguments()
	arguments.ShardCoordinator = nil

	esd, err := NewEndOfEpochEconomicsDataCreator(arguments)
	require.Nil(t, esd)
	require.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestEpochEconomics_NewEndOfEpochEconomicsDataCreatorNilNodesdCoordinator(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochEconomicsArguments()
	arguments.NodesCoordinator = nil

	esd, err := NewEndOfEpochEconomicsDataCreator(arguments)
	require.Nil(t, esd)
	require.Equal(t, process.ErrNilNodesCoordinator, err)
}

func TestEpochEconomics_NewEndOfEpochEconomicsDataCreatorNilRewardsHandler(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochEconomicsArguments()
	arguments.RewardsHandler = nil

	esd, err := NewEndOfEpochEconomicsDataCreator(arguments)
	require.Nil(t, esd)
	require.Equal(t, process.ErrNilRewardsHandler, err)
}

func TestEpochEconomics_NewEndOfEpochEconomicsDataCreatorNilRounder(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochEconomicsArguments()
	arguments.RoundTime = nil

	esd, err := NewEndOfEpochEconomicsDataCreator(arguments)
	require.Nil(t, esd)
	require.Equal(t, process.ErrNilRounder, err)
}

func TestEpochEconomics_NewEndOfEpochEconomicsDataCreatorShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochEconomicsArguments()

	esd, err := NewEndOfEpochEconomicsDataCreator(arguments)
	require.NotNil(t, esd)
	require.Nil(t, err)
}
