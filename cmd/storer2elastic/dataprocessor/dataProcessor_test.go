package dataprocessor_test

import (
	"testing"

	storer2ElasticData "github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/data"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/dataprocessor"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/mock"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewDataProcessor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argsFunc func() dataprocessor.ArgsDataProcessor
		exError  error
	}{
		{
			name: "NilElasticIndexer",
			argsFunc: func() dataprocessor.ArgsDataProcessor {
				args := getDataProcessorArgs()
				args.OutportHandler = nil
				return args
			},
			exError: dataprocessor.ErrNilOutportHandler,
		},
		{
			name: "NilDataReplayer",
			argsFunc: func() dataprocessor.ArgsDataProcessor {
				args := getDataProcessorArgs()
				args.DataReplayer = nil
				return args
			},
			exError: dataprocessor.ErrNilDataReplayer,
		},
		{
			name: "NilMarshalizer",
			argsFunc: func() dataprocessor.ArgsDataProcessor {
				args := getDataProcessorArgs()
				args.Marshalizer = nil
				return args
			},
			exError: dataprocessor.ErrNilMarshalizer,
		},
		{
			name: "NilHasher",
			argsFunc: func() dataprocessor.ArgsDataProcessor {
				args := getDataProcessorArgs()
				args.Hasher = nil
				return args
			},
			exError: dataprocessor.ErrNilHasher,
		},
		{
			name: "NilShardCoordinator",
			argsFunc: func() dataprocessor.ArgsDataProcessor {
				args := getDataProcessorArgs()
				args.ShardCoordinator = nil
				return args
			},
			exError: dataprocessor.ErrNilShardCoordinator,
		},
		{
			name: "NilGenesisNodesSetupHandler",
			argsFunc: func() dataprocessor.ArgsDataProcessor {
				args := getDataProcessorArgs()
				args.GenesisNodesSetup = nil
				return args
			},
			exError: dataprocessor.ErrNilGenesisNodesSetup,
		},
		{
			name: "NilTPSBenchmarkUpdater",
			argsFunc: func() dataprocessor.ArgsDataProcessor {
				args := getDataProcessorArgs()
				args.TPSBenchmarkUpdater = nil
				return args
			},
			exError: dataprocessor.ErrNilTPSBenchmarkUpdater,
		},
		{
			name: "NilTPSBenchmarkUpdater",
			argsFunc: func() dataprocessor.ArgsDataProcessor {
				args := getDataProcessorArgs()
				args.TPSBenchmarkUpdater = nil
				return args
			},
			exError: dataprocessor.ErrNilTPSBenchmarkUpdater,
		},
		{
			name: "NilRatingsProcessor",
			argsFunc: func() dataprocessor.ArgsDataProcessor {
				args := getDataProcessorArgs()
				args.RatingsProcessor = nil
				return args
			},
			exError: dataprocessor.ErrNilRatingProcessor,
		},
		{
			name: "All arguments ok",
			argsFunc: func() dataprocessor.ArgsDataProcessor {
				return getDataProcessorArgs()
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dataprocessor.NewDataProcessor(tt.argsFunc())
			require.Equal(t, err, tt.exError)
		})
	}
}

func TestDataProcessor_Index(t *testing.T) {
	t.Parallel()

	args := getDataProcessorArgs()
	args.DataReplayer = &mock.DataReplayerStub{
		RangeCalled: func(handler func(persistedData storer2ElasticData.RoundPersistedData) bool) error {
			handler(storer2ElasticData.RoundPersistedData{
				MetaBlockData: &storer2ElasticData.HeaderData{
					Header:           &block.MetaBlock{},
					Body:             &block.Body{},
					BodyTransactions: map[string]data.TransactionHandler{},
				},
				ShardHeaders: map[uint32][]*storer2ElasticData.HeaderData{
					0: {
						{Header: &block.Header{ShardID: 0}},
					},
				},
			})
			return nil
		},
	}
	dp, _ := dataprocessor.NewDataProcessor(args)
	require.NotNil(t, dp)

	err := dp.Index()
	require.NoError(t, err)
}

func getDataProcessorArgs() dataprocessor.ArgsDataProcessor {
	return dataprocessor.ArgsDataProcessor{
		OutportHandler: &testscommon.OutportStub{},
		DataReplayer:   &mock.DataReplayerStub{},
		GenesisNodesSetup: &mock.GenesisNodesSetupHandlerStub{
			InitialNodesInfoCalled: func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
				nodeEligible := mock.NewNodeInfo([]byte("addr1"), []byte("pubKey1"), 0, 10)
				return map[uint32][]sharding.GenesisNodeInfoHandler{
					core.MetachainShardId: {nodeEligible},
				}, nil
			},
		},
		ShardCoordinator:    &mock.ShardCoordinatorMock{},
		Marshalizer:         &mock.MarshalizerMock{},
		Hasher:              &mock.HasherMock{},
		TPSBenchmarkUpdater: &mock.TPSBenchmarkUpdaterStub{},
		RatingsProcessor:    &mock.RatingsProcessorStub{},
		RatingConfig: config.RatingsConfig{
			ShardChain: config.ShardChain{
				RatingSteps: config.RatingSteps{
					HoursToMaxRatingFromStartRating: 50,
					ProposerValidatorImportance:     1,
					ProposerDecreaseFactor:          -4,
					ValidatorDecreaseFactor:         -4,
					ConsecutiveMissedBlocksPenalty:  1.1,
				},
			},
			MetaChain: config.MetaChain{
				RatingSteps: config.RatingSteps{
					HoursToMaxRatingFromStartRating: 50,
					ProposerValidatorImportance:     1,
					ProposerDecreaseFactor:          -4,
					ValidatorDecreaseFactor:         -4,
					ConsecutiveMissedBlocksPenalty:  1.1,
				},
			},
			General: config.General{
				StartRating:           500000,
				MaxRating:             1000000,
				MinRating:             1,
				SignedBlocksThreshold: 0.025,
				SelectionChances: []*config.SelectionChance{
					{
						MaxThreshold:  0,
						ChancePercent: 5,
					},
					{
						MaxThreshold:  100000,
						ChancePercent: 0,
					},
					{
						MaxThreshold:  200000,
						ChancePercent: 16,
					},
					{
						MaxThreshold:  300000,
						ChancePercent: 17,
					},
					{
						MaxThreshold:  400000,
						ChancePercent: 18,
					},
					{
						MaxThreshold:  500000,
						ChancePercent: 19,
					},
					{
						MaxThreshold:  600000,
						ChancePercent: 20,
					},
					{
						MaxThreshold:  700000,
						ChancePercent: 21,
					},
					{
						MaxThreshold:  800000,
						ChancePercent: 22,
					},
					{
						MaxThreshold:  900000,
						ChancePercent: 23,
					},
					{
						MaxThreshold:  1000000,
						ChancePercent: 24,
					},
				},
			},
		},
		StartingEpoch: 0,
	}
}
