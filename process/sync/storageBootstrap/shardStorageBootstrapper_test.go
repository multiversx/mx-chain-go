package storageBootstrap

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dataRetrieverMocks "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	epochNotifierMock "github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageMock "github.com/multiversx/mx-chain-go/testscommon/storage"
)

func TestShardStorageBootstrapper_LoadFromStorageShouldWork(t *testing.T) {
	t.Parallel()

	wasCalledBlockchainSetHash := false
	wasCalledBlockchainSetHeader := false
	wasCalledForkDetectorAddHeader := false
	wasCalledBlockTrackerAddTrackedHeader := false
	wasCalledEpochNotifier := false
	savedLastRound := int64(0)

	marshaller := &marshallerMock.MarshalizerMock{}
	startRound := 4000
	hdr := &block.HeaderV2{
		Header: &block.Header{
			Nonce:    3999,
			Round:    3999,
			RootHash: []byte("roothash"),
			ShardID:  0,
			ChainID:  []byte("1"),
		},
	}
	hdrHash := []byte("header hash")
	hdrBytes, _ := marshaller.Marshal(hdr)
	blockStorerMock := genericMocks.NewStorerMock()
	_ = blockStorerMock.Put(hdrHash, hdrBytes)

	args := ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper{
			BootStorer: &mock.BoostrapStorerMock{
				GetHighestRoundCalled: func() int64 {
					return int64(startRound)
				},
				GetCalled: func(round int64) (bootstrapStorage.BootstrapData, error) {
					return bootstrapStorage.BootstrapData{
						LastHeader: bootstrapStorage.BootstrapHeaderInfo{
							ShardId: hdr.GetShardID(),
							Epoch:   hdr.GetEpoch(),
							Nonce:   hdr.GetNonce(),
							Hash:    hdrHash,
						},
						HighestFinalBlockNonce: 3999,
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
					assert.Equal(t, hdr, header)
					assert.Equal(t, hdrHash, hash)
					assert.Equal(t, process.BHProcessed, state)

					wasCalledForkDetectorAddHeader = true
					return nil
				},
			},
			BlockProcessor: &testscommon.BlockProcessorStub{},
			ChainHandler: &testscommon.ChainHandlerStub{
				GetGenesisHeaderCalled: func() data.HeaderHandler {
					return nil
				},
				SetCurrentBlockHeaderHashCalled: func(bytes []byte) {
					assert.Equal(t, hdrHash, bytes)
					wasCalledBlockchainSetHash = true
				},
				SetCurrentBlockHeaderAndRootHashCalled: func(header data.HeaderHandler, rootHash []byte) error {
					assert.Equal(t, hdr, header)
					assert.Equal(t, hdr.GetRootHash(), rootHash)
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
			Marshalizer: &marshallerMock.MarshalizerMock{},
			Store: &storageMock.ChainStorerStub{
				GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
					return blockStorerMock, nil
				},
			},
			Uint64Converter:     testscommon.NewNonceHashConverterMock(),
			BootstrapRoundIndex: uint64(startRound - 1),
			ShardCoordinator:    testscommon.NewMultiShardsCoordinatorMock(1),
			NodesCoordinator:    &shardingMocks.NodesCoordinatorMock{},
			EpochStartTrigger:   &mock.EpochStartTriggerStub{},
			BlockTracker: &mock.BlockTrackerMock{
				AddTrackedHeaderCalled: func(header data.HeaderHandler, hash []byte) {
					assert.Equal(t, hdr, header)
					assert.Equal(t, hdrHash, hash)

					wasCalledBlockTrackerAddTrackedHeader = true
				},
			},
			ChainID:                      string(hdr.GetChainID()),
			ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
			MiniblocksProvider:           &mock.MiniBlocksProviderStub{},
			EpochNotifier: &epochNotifierMock.EpochNotifierStub{
				CheckEpochCalled: func(header data.HeaderHandler) {
					assert.Equal(t, hdr, header)
					wasCalledEpochNotifier = true
				},
			},
			ProcessedMiniBlocksTracker: &testscommon.ProcessedMiniBlocksTrackerStub{},
			AppStatusHandler:           &statusHandler.AppStatusHandlerMock{},
			EnableEpochsHandler:        &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			ProofsPool:                 &dataRetrieverMocks.ProofsPoolMock{},
			ExecutionManager:           &processMocks.ExecutionManagerMock{},
		},
	}

	ssb, err := NewShardStorageBootstrapper(args)
	require.Nil(t, err)

	err = ssb.LoadFromStorage()
	assert.Nil(t, err)
	assert.True(t, wasCalledBlockchainSetHash)
	assert.True(t, wasCalledBlockchainSetHeader)
	assert.True(t, wasCalledForkDetectorAddHeader)
	assert.True(t, wasCalledBlockTrackerAddTrackedHeader)
	assert.Equal(t, int64(3999), savedLastRound)
	assert.True(t, wasCalledEpochNotifier)
}

func TestShardStorageBootstrapper_CleanupNotarizedStorageForHigherNoncesIfExist(t *testing.T) {
	baseArgs := createMockShardStorageBootstrapperArgs()

	bForceError := true
	numCalled := 0
	numKeysNotFound := 0
	metaNonce := uint64(2)
	nonceToByteSlice := []byte("nonceToByteSlice")
	metaHash := []byte("meta_hash")

	metaNonceToDelete := metaNonce + maxNumOfConsecutiveNoncesNotFoundAccepted + 2
	metaBlock := &block.MetaBlock{Nonce: metaNonceToDelete}
	marshalledMetaBlock, _ := baseArgs.Marshalizer.Marshal(metaBlock)

	baseArgs.Uint64Converter = &mock.Uint64ByteSliceConverterMock{
		ToByteSliceCalled: func(u uint64) []byte {
			if u == metaNonceToDelete {
				return nonceToByteSlice
			}
			return []byte("")
		},
	}
	baseArgs.Store = &storageMock.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageMock.StorerStub{
				RemoveCalled: func(key []byte) error {
					if bForceError {
						return errors.New("forced error")
					}

					if bytes.Equal(key, nonceToByteSlice) {
						numCalled++
						return nil
					}
					if bytes.Equal(key, metaHash) {
						numCalled++
						return nil
					}

					return errors.New("error")
				},
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByteSlice) {
						return metaHash, nil
					}
					if bytes.Equal(key, metaHash) {
						return marshalledMetaBlock, nil
					}
					numKeysNotFound++
					return nil, errors.New("error")
				},
			}, nil
		},
	}

	args := ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper: baseArgs,
	}
	ssb, _ := NewShardStorageBootstrapper(args)

	crossNotarizedHeaders := make([]bootstrapStorage.BootstrapHeaderInfo, 0)

	crossNotarizedHeaders = append(crossNotarizedHeaders, bootstrapStorage.BootstrapHeaderInfo{ShardId: 0, Nonce: 1})
	ssb.cleanupNotarizedStorageForHigherNoncesIfExist(crossNotarizedHeaders)
	assert.Equal(t, 0, numCalled)

	crossNotarizedHeaders = append(crossNotarizedHeaders, bootstrapStorage.BootstrapHeaderInfo{ShardId: core.MetachainShardId, Nonce: metaNonce})
	ssb.cleanupNotarizedStorageForHigherNoncesIfExist(crossNotarizedHeaders)
	assert.Equal(t, 0, numCalled)
	assert.Equal(t, maxNumOfConsecutiveNoncesNotFoundAccepted, numKeysNotFound-1)

	numKeysNotFound = 0
	metaNonceToDelete = metaNonce + maxNumOfConsecutiveNoncesNotFoundAccepted + 1
	metaBlock = &block.MetaBlock{Nonce: metaNonceToDelete}
	marshalledMetaBlock, _ = baseArgs.Marshalizer.Marshal(metaBlock)

	ssb.cleanupNotarizedStorageForHigherNoncesIfExist(crossNotarizedHeaders)
	assert.Equal(t, 0, numCalled)
	assert.Equal(t, maxNumOfConsecutiveNoncesNotFoundAccepted*2, numKeysNotFound-1)

	numKeysNotFound = 0
	bForceError = false

	ssb.cleanupNotarizedStorageForHigherNoncesIfExist(crossNotarizedHeaders)
	assert.Equal(t, 2, numCalled)
	assert.Equal(t, maxNumOfConsecutiveNoncesNotFoundAccepted*2, numKeysNotFound-1)
}

func TestShardStorageBootstrapper_GetCrossNotarizedHeaderNonceShouldWork(t *testing.T) {
	crossNotarizedHeaders := make([]bootstrapStorage.BootstrapHeaderInfo, 0)

	crossNotarizedHeaders = append(crossNotarizedHeaders, bootstrapStorage.BootstrapHeaderInfo{ShardId: 0, Nonce: 1})
	nonce, err := getLastCrossNotarizedHeaderNonce(crossNotarizedHeaders)
	assert.Equal(t, sync.ErrHeaderNotFound, err)
	assert.Equal(t, uint64(0), nonce)

	crossNotarizedHeaders = append(crossNotarizedHeaders, bootstrapStorage.BootstrapHeaderInfo{ShardId: core.MetachainShardId, Nonce: 2})
	nonce, err = getLastCrossNotarizedHeaderNonce(crossNotarizedHeaders)
	assert.Nil(t, err)
	assert.Equal(t, uint64(2), nonce)
}

func TestShardStorageBootstrapper_LoadFromStorageShouldCleanupRoundsAboveBootstrapRoundIndex(t *testing.T) {
	t.Parallel()

	t.Run("multiple rounds above bootstrapRoundIndex are cleaned before valid one", func(t *testing.T) {
		t.Parallel()

		marshaller := &marshallerMock.MarshalizerMock{}

		// bootstrapRoundIndex is 97, highest round is 100
		// rounds 100, 99, 98 are above the index and should be cleaned
		// round 97 should be processed
		bootstrapRoundIdx := uint64(97)

		hdr := &block.HeaderV3{
			Nonce:   96,
			Round:   96,
			ShardID: 0,
			ChainID: []byte("1"),
		}
		hdrHash := []byte("header hash 96")
		hdrBytes, _ := marshaller.Marshal(hdr)

		cleanupCount := 0
		savedLastRound := int64(0)
		wasCalledBlockchainSetHeader := false

		args := ArgsShardStorageBootstrapper{
			ArgsBaseStorageBootstrapper{
				BootStorer: &mock.BoostrapStorerMock{
					GetHighestRoundCalled: func() int64 {
						return 100
					},
					GetCalled: func(round int64) (bootstrapStorage.BootstrapData, error) {
						return bootstrapStorage.BootstrapData{
							LastHeader: bootstrapStorage.BootstrapHeaderInfo{
								ShardId: hdr.GetShardID(),
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
		}

		ssb, err := NewShardStorageBootstrapper(args)
		require.Nil(t, err)

		err = ssb.LoadFromStorage()
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

		hdr := &block.HeaderV3{
			Nonce:   96,
			Round:   96,
			ShardID: 0,
			ChainID: []byte("1"),
		}
		hdrHash := []byte("header hash 96")
		hdrBytes, _ := marshaller.Marshal(hdr)

		cleanupCount := 0
		savedLastRound := int64(0)
		wasCalledBlockchainSetHeader := false

		args := ArgsShardStorageBootstrapper{
			ArgsBaseStorageBootstrapper{
				BootStorer: &mock.BoostrapStorerMock{
					GetHighestRoundCalled: func() int64 {
						return 100
					},
					GetCalled: func(round int64) (bootstrapStorage.BootstrapData, error) {
						return bootstrapStorage.BootstrapData{
							LastHeader: bootstrapStorage.BootstrapHeaderInfo{
								ShardId: hdr.GetShardID(),
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
		}

		ssb, err := NewShardStorageBootstrapper(args)
		require.Nil(t, err)

		err = ssb.LoadFromStorage()
		assert.Nil(t, err)
		assert.Equal(t, int64(0), savedLastRound)

		// all rounds info should be cleaned
		assert.Equal(t, 100, cleanupCount)
		assert.True(t, wasCalledBlockchainSetHeader)
	})
}

func TestShardStorageBootstrapper_LoadFromStorageRestoresPersistedFinalUnderSupernova(t *testing.T) {
	t.Parallel()

	marshaller := &mock.MarshalizerMock{}
	chainID := []byte("1")

	newHeader := func(nonce uint64, round uint64, prevHash []byte) *block.HeaderV2 {
		return &block.HeaderV2{Header: &block.Header{Nonce: nonce, Round: round, PrevHash: prevHash, ChainID: chainID}}
	}

	hash3997, hash3998, hash3999 := []byte("hash3997"), []byte("hash3998"), []byte("hash3999")
	headers := map[string]*block.HeaderV2{
		string(hash3997): newHeader(3997, 3997, []byte("hash3996")),
		string(hash3998): newHeader(3998, 3998, hash3997),
		// contended head: skipped rounds before it, persisted final stayed behind
		string(hash3999): newHeader(3999, 4005, hash3998),
	}

	blockStorer := genericMocks.NewStorerMock()
	for hash, header := range headers {
		headerBytes, _ := marshaller.Marshal(header)
		_ = blockStorer.Put([]byte(hash), headerBytes)
	}

	proofBytes, _ := marshaller.Marshal(&block.HeaderProof{})
	proofsStorer := &storageMock.StorerStub{
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			return proofBytes, nil
		},
	}

	bootstrapDataByRound := map[int64]bootstrapStorage.BootstrapData{
		4000: {
			LastHeader:             bootstrapStorage.BootstrapHeaderInfo{ShardId: 0, Nonce: 3999, Hash: hash3999},
			HighestFinalBlockNonce: 3998,
			LastRound:              3999,
		},
		3999: {
			LastHeader:             bootstrapStorage.BootstrapHeaderInfo{ShardId: 0, Nonce: 3998, Hash: hash3998},
			HighestFinalBlockNonce: 3998,
			LastRound:              3998,
		},
		3998: {
			LastHeader:             bootstrapStorage.BootstrapHeaderInfo{ShardId: 0, Nonce: 3997, Hash: hash3997},
			HighestFinalBlockNonce: 3997,
			LastRound:              3997,
		},
	}

	createArgs := func(events *[]string, restoredFinalNonce uint64, addHeaderErrNonce uint64) ArgsShardStorageBootstrapper {
		args := ArgsShardStorageBootstrapper{createMockShardStorageBootstrapperArgs()}
		args.Marshalizer = marshaller
		args.ChainID = string(chainID)
		args.BootstrapRoundIndex = 4000
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag || flag == common.SupernovaFlag
			},
		}
		args.BlockTracker = &mock.BlockTrackerMock{
			AddTrackedHeaderCalled:       func(header data.HeaderHandler, hash []byte) {},
			AddSelfNotarizedHeaderCalled: func(shardID uint32, selfNotarizedHeader data.HeaderHandler, selfNotarizedHeaderHash []byte) {},
		}
		args.BootStorer = &mock.BoostrapStorerMock{
			GetHighestRoundCalled: func() int64 {
				return 4000
			},
			GetCalled: func(round int64) (bootstrapStorage.BootstrapData, error) {
				bootstrapData, ok := bootstrapDataByRound[round]
				if !ok {
					return bootstrapStorage.BootstrapData{}, errors.New("not found")
				}
				return bootstrapData, nil
			},
			SaveLastRoundCalled: func(round int64) error {
				return nil
			},
		}
		args.Store = &storageMock.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				if unitType == dataRetriever.ProofsUnit {
					return proofsStorer, nil
				}

				return blockStorer, nil
			},
		}
		args.ForkDetector = &mock.ForkDetectorMock{
			AddCheckpointCalled: func(nonce uint64, round uint64, hash []byte) {
				*events = append(*events, fmt.Sprintf("addCheckpoint-%d", nonce))
			},
			SetFinalToLastCheckpointCalled: func() {
				*events = append(*events, "setFinal")
			},
			AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, _ []data.HeaderHandler, _ [][]byte) error {
				*events = append(*events, fmt.Sprintf("addHeader-%d", header.GetNonce()))
				if header.GetNonce() == addHeaderErrNonce {
					return errors.New("add header failed")
				}
				return nil
			},
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return restoredFinalNonce
			},
		}

		return args
	}

	t.Run("contended head restores as non-final, final checkpoint set at the persisted nonce", func(t *testing.T) {
		t.Parallel()

		events := make([]string, 0)
		ssb, err := NewShardStorageBootstrapper(createArgs(&events, 3998, 0))
		require.Nil(t, err)

		err = ssb.LoadFromStorage()
		require.Nil(t, err)

		expectedEvents := []string{
			"addCheckpoint-3999",
			"addHeader-3997",
			"addHeader-3998",
			"setFinal",
			"addHeader-3999",
		}
		require.Equal(t, expectedEvents, events)
	})

	t.Run("final checkpoint falls back to head when the persisted nonce cannot be restored", func(t *testing.T) {
		t.Parallel()

		events := make([]string, 0)
		ssb, err := NewShardStorageBootstrapper(createArgs(&events, 0, 3998))
		require.Nil(t, err)

		err = ssb.LoadFromStorage()
		require.Nil(t, err)

		expectedEvents := []string{
			"addCheckpoint-3999",
			"addHeader-3997",
			"addHeader-3998",
			"addHeader-3999",
			"setFinal",
		}
		require.Equal(t, expectedEvents, events)
	})
}
