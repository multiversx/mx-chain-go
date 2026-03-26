package storageBootstrap

import (
	"bytes"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	epochNotifierMock "github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageMock "github.com/multiversx/mx-chain-go/testscommon/storage"
)

func TestMetaStorageBootstrapper_LoadFromStorageShouldCleanupRoundsAboveBootstrapRoundIndex(t *testing.T) {
	t.Parallel()

	t.Run("multiple rounds above bootstrapRoundIndex are cleaned before valid one", func(t *testing.T) {
		t.Parallel()

		marshaller := &marshallerMock.MarshalizerMock{}

		// bootstrapRoundIndex is 97, highest round is 100
		// rounds 100, 99, 98 are above the index and should be cleaned
		// round 97 should be processed
		bootstrapRoundIdx := uint64(97)

		hdr := &block.MetaBlockV3{
			Nonce:   96,
			Round:   96,
			ChainID: []byte("1"),
		}
		hdrHash := []byte("header hash 96")
		hdrBytes, _ := marshaller.Marshal(hdr)

		cleanupCount := 0
		savedLastRound := int64(0)
		wasCalledBlockchainSetHeader := false

		args := ArgsMetaStorageBootstrapper{
			ArgsBaseStorageBootstrapper: ArgsBaseStorageBootstrapper{
				BootStorer: &mock.BoostrapStorerMock{
					GetHighestRoundCalled: func() int64 {
						return 100
					},
					GetCalled: func(round int64) (bootstrapStorage.BootstrapData, error) {
						return bootstrapStorage.BootstrapData{
							LastHeader: bootstrapStorage.BootstrapHeaderInfo{
								ShardId: core.MetachainShardId,
								Epoch:   hdr.GetEpoch(),
								Nonce:   hdr.GetNonce(),
								Hash:    hdrHash,
							},
							HighestFinalBlockNonce: hdr.GetNonce(),
							LastRound:              round - 1,
						}, nil
					},
					SaveLastRoundCalled: func(round int64) error {
						savedLastRound = round
						return nil
					},
				},
				ForkDetector: &mock.ForkDetectorMock{
					AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error {
						return nil
					},
				},
				BlockProcessor: &testscommon.BlockProcessorStub{},
				ChainHandler: &testscommon.ChainHandlerStub{
					GetGenesisHeaderCalled: func() data.HeaderHandler {
						return nil
					},
					SetCurrentBlockHeaderAndRootHashCalled: func(header data.HeaderHandler, rootHash []byte) error {
						wasCalledBlockchainSetHeader = true
						return nil
					},
					GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
						if wasCalledBlockchainSetHeader {
							return hdr
						}
						return nil
					},
				},
				Marshalizer: marshaller,
				Store: &storageMock.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return &storageMock.StorerStub{
							GetCalled: func(key []byte) ([]byte, error) {
								if bytes.Equal(key, hdrHash) {
									return hdrBytes, nil
								}
								return nil, errors.New("key not found")
							},
							RemoveCalled: func(key []byte) error {
								cleanupCount++
								return nil
							},
							SearchFirstCalled: func(key []byte) ([]byte, error) {
								return nil, errors.New("not found")
							},
						}, nil
					},
				},
				Uint64Converter:     testscommon.NewNonceHashConverterMock(),
				BootstrapRoundIndex: bootstrapRoundIdx,
				ShardCoordinator:    testscommon.NewMultiShardsCoordinatorMock(1),
				NodesCoordinator:    &shardingMocks.NodesCoordinatorMock{},
				EpochStartTrigger:   &mock.EpochStartTriggerStub{},
				BlockTracker: &mock.BlockTrackerMock{
					AddTrackedHeaderCalled: func(header data.HeaderHandler, hash []byte) {},
				},
				ChainID:                      "1",
				ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
				MiniblocksProvider:           &mock.MiniBlocksProviderStub{},
				EpochNotifier:                &epochNotifierMock.EpochNotifierStub{},
				ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
				AppStatusHandler:             &statusHandler.AppStatusHandlerMock{},
				EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
				ProofsPool:                   &dataRetrieverMocks.ProofsPoolMock{},
				ExecutionManager:             &processMocks.ExecutionManagerMock{},
			},
			PendingMiniBlocksHandler: &mock.PendingMiniBlocksHandlerStub{},
		}

		msb, err := NewMetaStorageBootstrapper(args)
		require.Nil(t, err)

		err = msb.LoadFromStorage()
		assert.Nil(t, err)
		assert.Equal(t, int64(97), savedLastRound)
		// rounds 100, 99, 98 were above bootstrapRoundIndex=97, each triggers a Remove in cleanupStorage
		assert.True(t, cleanupCount >= 3)
		assert.True(t, wasCalledBlockchainSetHeader)
	})

	t.Run("bootstrapRoundIndex zero should cleaned all", func(t *testing.T) {
		t.Parallel()

		marshaller := &marshallerMock.MarshalizerMock{}

		bootstrapRoundIdx := uint64(0)

		hdr := &block.MetaBlockV3{
			Nonce:   96,
			Round:   96,
			ChainID: []byte("1"),
		}
		hdrHash := []byte("header hash 96")
		hdrBytes, _ := marshaller.Marshal(hdr)

		cleanupCount := 0
		savedLastRound := int64(0)
		wasCalledBlockchainSetHeader := false

		args := ArgsMetaStorageBootstrapper{
			ArgsBaseStorageBootstrapper: ArgsBaseStorageBootstrapper{
				BootStorer: &mock.BoostrapStorerMock{
					GetHighestRoundCalled: func() int64 {
						return 100
					},
					GetCalled: func(round int64) (bootstrapStorage.BootstrapData, error) {
						return bootstrapStorage.BootstrapData{
							LastHeader: bootstrapStorage.BootstrapHeaderInfo{
								ShardId: core.MetachainShardId,
								Epoch:   hdr.GetEpoch(),
								Nonce:   hdr.GetNonce(),
								Hash:    hdrHash,
							},
							HighestFinalBlockNonce: hdr.GetNonce(),
							LastRound:              round - 1,
						}, nil
					},
					SaveLastRoundCalled: func(round int64) error {
						savedLastRound = round
						return nil
					},
				},
				ForkDetector: &mock.ForkDetectorMock{
					AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error {
						return nil
					},
				},
				BlockProcessor: &testscommon.BlockProcessorStub{},
				ChainHandler: &testscommon.ChainHandlerStub{
					GetGenesisHeaderCalled: func() data.HeaderHandler {
						return nil
					},
					SetCurrentBlockHeaderAndRootHashCalled: func(header data.HeaderHandler, rootHash []byte) error {
						wasCalledBlockchainSetHeader = true
						return nil
					},
					GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
						if wasCalledBlockchainSetHeader {
							return hdr
						}
						return nil
					},
				},
				Marshalizer: marshaller,
				Store: &storageMock.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return &storageMock.StorerStub{
							GetCalled: func(key []byte) ([]byte, error) {
								if bytes.Equal(key, hdrHash) {
									return hdrBytes, nil
								}
								return nil, errors.New("key not found")
							},
							RemoveCalled: func(key []byte) error {
								cleanupCount++
								return nil
							},
							SearchFirstCalled: func(key []byte) ([]byte, error) {
								return nil, errors.New("not found")
							},
						}, nil
					},
				},
				Uint64Converter:     testscommon.NewNonceHashConverterMock(),
				BootstrapRoundIndex: bootstrapRoundIdx,
				ShardCoordinator:    testscommon.NewMultiShardsCoordinatorMock(1),
				NodesCoordinator:    &shardingMocks.NodesCoordinatorMock{},
				EpochStartTrigger:   &mock.EpochStartTriggerStub{},
				BlockTracker: &mock.BlockTrackerMock{
					AddTrackedHeaderCalled: func(header data.HeaderHandler, hash []byte) {},
				},
				ChainID:                      "1",
				ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
				MiniblocksProvider:           &mock.MiniBlocksProviderStub{},
				EpochNotifier:                &epochNotifierMock.EpochNotifierStub{},
				ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
				AppStatusHandler:             &statusHandler.AppStatusHandlerMock{},
				EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
				ProofsPool:                   &dataRetrieverMocks.ProofsPoolMock{},
				ExecutionManager:             &processMocks.ExecutionManagerMock{},
			},
			PendingMiniBlocksHandler: &mock.PendingMiniBlocksHandlerStub{},
		}

		msb, err := NewMetaStorageBootstrapper(args)
		require.Nil(t, err)

		err = msb.LoadFromStorage()
		assert.Nil(t, err)
		assert.Equal(t, int64(0), savedLastRound)

		// all rounds info should be cleaned
		assert.Equal(t, 100, cleanupCount)
		assert.True(t, wasCalledBlockchainSetHeader)
	})
}
