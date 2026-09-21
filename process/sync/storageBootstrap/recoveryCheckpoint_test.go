package storageBootstrap

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

func newTestRecoveryCheckpoint(t *testing.T) *common.RecoveryCheckpoint {
	t.Helper()
	hash := make([]byte, 32)
	hash[0] = 1
	cfg := &config.Config{
		HardforkRoundExclusions: []config.HardforkRoundExclusionConfig{{StartRound: 101, EndRound: 199}},
		HardforkRecoveryCheckpoint: config.HardforkRecoveryCheckpointConfig{
			Enabled: true,
			Round:   100,
			Headers: []config.HardforkRecoveryHeaderConfig{
				{ShardID: 0, Hash: hex.EncodeToString(hash)},
				{ShardID: core.MetachainShardId, Hash: hex.EncodeToString(hash)},
			},
		},
	}
	checkpoint, err := common.NewRecoveryCheckpoint(cfg)
	require.NoError(t, err)
	return checkpoint
}

func TestRecoveryCheckpoint_MissingExactRecordDoesNotFallBack(t *testing.T) {
	checkpoint := newTestRecoveryCheckpoint(t)
	savedRound := false
	args := createMockShardStorageBootstrapperArgs()
	args.RecoveryCheckpoint = checkpoint
	args.BootStorer = &mock.BoostrapStorerMock{
		GetHighestRoundCalled: func() int64 { return 101 },
		GetCalled: func(round int64) (bootstrapStorage.BootstrapData, error) {
			require.Equal(t, int64(100), round)
			return bootstrapStorage.BootstrapData{}, storage.ErrKeyNotFound
		},
		SaveLastRoundCalled: func(_ int64) error {
			savedRound = true
			return nil
		},
	}
	bootstrapper, err := NewShardStorageBootstrapper(ArgsShardStorageBootstrapper{ArgsBaseStorageBootstrapper: args})
	require.NoError(t, err)
	err = bootstrapper.LoadFromStorage()
	require.ErrorIs(t, err, ErrRecoveryCheckpointUnavailable)
	require.False(t, savedRound)
}

func TestRecoveryCheckpoint_WrongTargetHashDoesNotChangePointer(t *testing.T) {
	checkpoint := newTestRecoveryCheckpoint(t)
	savedRound := false
	args := createMockShardStorageBootstrapperArgs()
	args.RecoveryCheckpoint = checkpoint
	args.BootStorer = &mock.BoostrapStorerMock{
		GetHighestRoundCalled: func() int64 { return 100 },
		GetCalled: func(_ int64) (bootstrapStorage.BootstrapData, error) {
			return bootstrapStorage.BootstrapData{LastHeader: bootstrapStorage.BootstrapHeaderInfo{ShardId: 0, Hash: []byte("other")}}, nil
		},
		SaveLastRoundCalled: func(_ int64) error {
			savedRound = true
			return nil
		},
	}
	bootstrapper, err := NewShardStorageBootstrapper(ArgsShardStorageBootstrapper{ArgsBaseStorageBootstrapper: args})
	require.NoError(t, err)
	err = bootstrapper.LoadFromStorage()
	require.ErrorIs(t, err, ErrRecoveryCheckpointUnavailable)
	require.False(t, savedRound)
}

func TestRecoveryCheckpoint_SelectorCleanupVerifiesRemoval(t *testing.T) {
	values := make(map[string][]byte)
	converter := uint64ByteSlice.NewBigEndianConverter()
	firstKey := converter.ToByteSlice(11)
	values[string(firstKey)] = []byte("discarded")
	storer := &storageStubs.StorerStub{
		HasCalled: func(key []byte) error {
			if _, ok := values[string(key)]; !ok {
				return storage.ErrKeyNotFound
			}
			return nil
		},
		GetCalled: func(key []byte) ([]byte, error) {
			value, ok := values[string(key)]
			if !ok {
				return nil, storage.ErrKeyNotFound
			}
			return value, nil
		},
		RemoveCalled: func(key []byte) error {
			delete(values, string(key))
			return nil
		},
	}
	bootstrapper := &storageBootstrapper{uint64Converter: converter}
	require.NoError(t, bootstrapper.removeNonceSelectors(storer, 11, 12))
	require.Empty(t, values)

	storer.RemoveCalled = func(_ []byte) error { return nil }
	values[string(firstKey)] = []byte("discarded")
	err := bootstrapper.removeNonceSelectors(storer, 11, 11)
	require.ErrorIs(t, err, ErrRecoveryCheckpointUnavailable)

	storer.HasCalled = func(_ []byte) error { return errors.New("read failed") }
	err = bootstrapper.removeNonceSelectors(storer, 11, 11)
	require.EqualError(t, err, "read failed")
}

func TestRecoveryCheckpoint_SelectorCleanupRemovesAllCopies(t *testing.T) {
	converter := uint64ByteSlice.NewBigEndianConverter()
	copies := 3
	storer := &storageStubs.StorerStub{
		HasCalled: func(_ []byte) error {
			if copies == 0 {
				return storage.ErrKeyNotFound
			}
			return nil
		},
		RemoveCalled: func(_ []byte) error {
			copies--
			return nil
		},
	}
	bootstrapper := &storageBootstrapper{uint64Converter: converter}
	require.NoError(t, bootstrapper.removeNonceSelectors(storer, 11, 11))
	require.Zero(t, copies)
}

func TestRecoveryCheckpoint_RecoverySelectionUsesStoredTipRound(t *testing.T) {
	for _, testCase := range []struct {
		name           string
		highestRound   int64
		expectRecovery bool
		expectError    bool
	}{
		{name: "before target", highestRound: 99, expectError: true},
		{name: "at target", highestRound: 100, expectRecovery: true},
		{name: "inside exclusion", highestRound: 150, expectRecovery: true},
		{name: "after exclusion", highestRound: 200},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			args := createMockShardStorageBootstrapperArgs()
			args.RecoveryCheckpoint = newTestRecoveryCheckpoint(t)
			args.BootStorer = &mock.BoostrapStorerMock{
				GetHighestRoundCalled: func() int64 { return testCase.highestRound },
				GetCalled: func(_ int64) (bootstrapStorage.BootstrapData, error) {
					t.Fatal("recovery selection must not read ancestry")
					return bootstrapStorage.BootstrapData{}, nil
				},
			}
			bootstrapper, err := NewShardStorageBootstrapper(ArgsShardStorageBootstrapper{ArgsBaseStorageBootstrapper: args})
			require.NoError(t, err)
			shouldRecover, checkErr := bootstrapper.shouldRecoverCheckpoint()
			if testCase.expectError {
				require.ErrorIs(t, checkErr, ErrRecoveryCheckpointUnavailable)
				return
			}
			require.NoError(t, checkErr)
			require.Equal(t, testCase.expectRecovery, shouldRecover)
		})
	}
}

func TestRecoveryCheckpoint_PostExclusionFallbackStopsAtTarget(t *testing.T) {
	checkpoint := newTestRecoveryCheckpoint(t)
	for _, testCase := range []struct {
		name      string
		lastRound int64
	}{
		{name: "older fallback", lastRound: 99},
		{name: "missing target", lastRound: 100},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			savedRound := false
			args := createMockShardStorageBootstrapperArgs()
			args.RecoveryCheckpoint = checkpoint
			args.BootStorer = &mock.BoostrapStorerMock{
				GetHighestRoundCalled: func() int64 { return 200 },
				GetCalled: func(round int64) (bootstrapStorage.BootstrapData, error) {
					if round == 200 {
						return bootstrapStorage.BootstrapData{
							LastRound:              testCase.lastRound,
							LastHeader:             bootstrapStorage.BootstrapHeaderInfo{Nonce: 10},
							HighestFinalBlockNonce: 10,
						}, nil
					}
					require.Equal(t, int64(100), round)
					return bootstrapStorage.BootstrapData{}, storage.ErrKeyNotFound
				},
				SaveLastRoundCalled: func(_ int64) error {
					savedRound = true
					return nil
				},
			}
			bootstrapper, err := NewShardStorageBootstrapper(ArgsShardStorageBootstrapper{ArgsBaseStorageBootstrapper: args})
			require.NoError(t, err)
			err = bootstrapper.LoadFromStorage()
			require.ErrorIs(t, err, ErrRecoveryCheckpointUnavailable)
			require.False(t, savedRound)
		})
	}
}

func TestRecoveryCheckpoint_BootstrapHistoryCanReadBeforeTarget(t *testing.T) {
	readRound := int64(0)
	bootstrapper := &storageBootstrapper{
		recoveryCheckpoint: newTestRecoveryCheckpoint(t),
		bootStorer: &mock.BoostrapStorerMock{
			GetCalled: func(round int64) (bootstrapStorage.BootstrapData, error) {
				readRound = round
				return bootstrapStorage.BootstrapData{LastHeader: bootstrapStorage.BootstrapHeaderInfo{Nonce: 1}}, nil
			},
		},
	}
	bootInfos, err := bootstrapper.getBootInfos(bootstrapStorage.BootstrapData{
		LastRound:              99,
		LastHeader:             bootstrapStorage.BootstrapHeaderInfo{Nonce: 2},
		HighestFinalBlockNonce: 1,
	})
	require.NoError(t, err)
	require.Equal(t, int64(99), readRound)
	require.Len(t, bootInfos, 2)
}

func TestRecoveryCheckpoint_RestoresExactTipAndRemovesSuffixSelectors(t *testing.T) {
	marshaller := &marshal.GogoProtoMarshalizer{}
	hasher := sha256.NewSha256()
	converter := uint64ByteSlice.NewBigEndianConverter()
	grandparent := &block.HeaderV3{Round: 98, Nonce: 8, ShardID: 0, ChainID: []byte("1"), LastExecutionResult: &block.ExecutionResultInfo{}}
	grandparentHash, err := core.CalculateHash(marshaller, hasher, grandparent)
	require.NoError(t, err)
	parent := &block.HeaderV3{Round: 99, Nonce: 9, ShardID: 0, ChainID: []byte("1"), PrevHash: grandparentHash, LastExecutionResult: &block.ExecutionResultInfo{}}
	parentHash, err := core.CalculateHash(marshaller, hasher, parent)
	require.NoError(t, err)
	rootHash := make([]byte, 32)
	rootHash[0] = 3
	target := &block.HeaderV3{
		Round: 100, Nonce: 10, ShardID: 0, ChainID: []byte("1"), PrevHash: parentHash,
		LastExecutionResult: &block.ExecutionResultInfo{ExecutionResult: &block.BaseExecutionResult{
			HeaderHash: parentHash, HeaderNonce: 9, RootHash: rootHash,
		}},
		ExecutionResults: []*block.ExecutionResult{{BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash: parentHash, HeaderNonce: 9, RootHash: rootHash,
		}}},
	}
	targetHash, err := core.CalculateHash(marshaller, hasher, target)
	require.NoError(t, err)
	discarded := &block.HeaderV3{Round: 101, Nonce: 11, ShardID: 0, ChainID: []byte("1"), PrevHash: targetHash, LastExecutionResult: &block.ExecutionResultInfo{}}
	discardedHash, err := core.CalculateHash(marshaller, hasher, discarded)
	require.NoError(t, err)
	proofBytes := make(map[string][]byte)
	for _, header := range []*block.HeaderV3{grandparent, parent, target} {
		hash, calculateErr := core.CalculateHash(marshaller, hasher, header)
		require.NoError(t, calculateErr)
		proof := &block.HeaderProof{HeaderHash: hash, HeaderRound: header.GetRound(), HeaderNonce: header.GetNonce(), HeaderShardId: 0}
		proofBytes[string(hash)], calculateErr = marshaller.Marshal(proof)
		require.NoError(t, calculateErr)
	}
	headerBytes := make(map[string][]byte)
	for _, header := range []*block.HeaderV3{grandparent, parent, target, discarded} {
		hash, calculateErr := core.CalculateHash(marshaller, hasher, header)
		require.NoError(t, calculateErr)
		headerBytes[string(hash)], calculateErr = marshaller.Marshal(header)
		require.NoError(t, calculateErr)
	}
	selectors := map[string][]byte{
		string(converter.ToByteSlice(10)): targetHash,
		string(converter.ToByteSlice(11)): discardedHash,
	}
	selectorStorer := &storageStubs.StorerStub{
		HasCalled: func(key []byte) error {
			if _, exists := selectors[string(key)]; !exists {
				return storage.ErrKeyNotFound
			}
			return nil
		},
		GetCalled: func(key []byte) ([]byte, error) {
			value, exists := selectors[string(key)]
			if !exists {
				return nil, storage.ErrKeyNotFound
			}
			return value, nil
		},
		PutCalled: func(key, value []byte) error {
			selectors[string(key)] = value
			return nil
		},
		RemoveCalled: func(key []byte) error {
			delete(selectors, string(key))
			return nil
		},
	}
	checkpointConfig := &config.Config{
		HardforkRoundExclusions: []config.HardforkRoundExclusionConfig{{StartRound: 101, EndRound: 199}},
		HardforkRecoveryCheckpoint: config.HardforkRecoveryCheckpointConfig{
			Enabled: true, Round: 100,
			Headers: []config.HardforkRecoveryHeaderConfig{
				{ShardID: 0, Hash: hex.EncodeToString(targetHash)},
				{ShardID: core.MetachainShardId, Hash: hex.EncodeToString(targetHash)},
			},
		},
	}
	checkpoint, err := common.NewRecoveryCheckpoint(checkpointConfig)
	require.NoError(t, err)
	tip := data.HeaderHandler(nil)
	var tipHash []byte
	savedRound := int64(0)
	var lastProcessedNonce uint64
	var lastProcessedHash []byte
	var finalizedNonce uint64
	var finalizedHash []byte
	finalityChanges := make([]uint64, 0, 2)
	args := createMockShardStorageBootstrapperArgs()
	args.RecoveryCheckpoint = checkpoint
	args.Marshalizer = marshaller
	args.Hasher = hasher
	args.Uint64Converter = converter
	args.ChainID = "1"
	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, _ uint32) bool {
			return flag == common.AndromedaFlag || flag == common.SupernovaFlag
		},
	}
	args.BootStorer = &mock.BoostrapStorerMock{
		GetHighestRoundCalled: func() int64 { return 101 },
		GetCalled: func(round int64) (bootstrapStorage.BootstrapData, error) {
			switch round {
			case 98:
				return bootstrapStorage.BootstrapData{
					LastHeader: bootstrapStorage.BootstrapHeaderInfo{Hash: grandparentHash, ShardId: 0, Nonce: 8},
					LastRound:  97,
				}, nil
			case 99:
				return bootstrapStorage.BootstrapData{
					LastHeader: bootstrapStorage.BootstrapHeaderInfo{Hash: parentHash, ShardId: 0, Nonce: 9},
					LastRound:  98,
				}, nil
			case 100:
				return bootstrapStorage.BootstrapData{
					LastHeader: bootstrapStorage.BootstrapHeaderInfo{Hash: targetHash, ShardId: 0, Nonce: 10},
					LastRound:  99, HighestFinalBlockNonce: 9,
				}, nil
			case 101:
				return bootstrapStorage.BootstrapData{
					LastHeader: bootstrapStorage.BootstrapHeaderInfo{Hash: discardedHash, ShardId: 0, Nonce: 11},
					LastRound:  100,
				}, nil
			default:
				return bootstrapStorage.BootstrapData{}, storage.ErrKeyNotFound
			}
		},
		SaveLastRoundCalled: func(round int64) error {
			savedRound = round
			return nil
		},
	}
	args.Store = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unit dataRetriever.UnitType) (storage.Storer, error) {
			switch unit {
			case dataRetriever.BlockHeaderUnit:
				return &storageStubs.StorerStub{GetCalled: func(key []byte) ([]byte, error) {
					value, exists := headerBytes[string(key)]
					if !exists {
						return nil, storage.ErrKeyNotFound
					}
					return value, nil
				}}, nil
			case dataRetriever.ProofsUnit:
				return &storageStubs.StorerStub{SearchFirstCalled: func(hash []byte) ([]byte, error) {
					value, exists := proofBytes[string(hash)]
					if !exists {
						return nil, storage.ErrKeyNotFound
					}
					return value, nil
				}}, nil
			default:
				return selectorStorer, nil
			}
		},
	}
	args.ChainHandler = &testscommon.ChainHandlerStub{
		SetCurrentBlockHeaderAndHashCalled: func(hash []byte, header data.HeaderHandler) error {
			tip, tipHash = header, hash
			return nil
		},
		GetCurrentBlockHeaderAndHashCalled: func() (data.HeaderHandler, []byte) { return tip, tipHash },
		GetCurrentBlockHeaderCalled:        func() data.HeaderHandler { return tip },
	}
	args.ForkDetector = &mock.ForkDetectorMock{
		RestoreToGenesisCalled: func() {},
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, _ process.BlockHeaderState, _ []data.HeaderHandler, _ [][]byte) error {
			lastProcessedNonce = header.GetNonce()
			lastProcessedHash = hash
			return nil
		},
		SetFinalToLastCheckpointCalled: func() {
			finalizedNonce = lastProcessedNonce
			finalizedHash = lastProcessedHash
			finalityChanges = append(finalityChanges, finalizedNonce)
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return finalizedNonce
		},
		GetHighestFinalBlockHashCalled: func() []byte {
			return finalizedHash
		},
		GetHighestSettledBlockInfoCalled: func() (uint64, []byte) {
			return finalizedNonce, finalizedHash
		},
	}
	args.BlockTracker = &mock.BlockTrackerMock{AddTrackedHeaderCalled: func(_ data.HeaderHandler, _ []byte) {}}
	bootstrapper, err := NewShardStorageBootstrapper(ArgsShardStorageBootstrapper{ArgsBaseStorageBootstrapper: args})
	require.NoError(t, err)
	require.NoError(t, bootstrapper.LoadFromStorage())
	require.Equal(t, int64(100), savedRound)
	require.Equal(t, targetHash, tipHash)
	require.Equal(t, uint64(10), finalizedNonce)
	require.Equal(t, targetHash, finalizedHash)
	require.Equal(t, []uint64{9, 10}, finalityChanges)
	require.Equal(t, targetHash, selectors[string(converter.ToByteSlice(10))])
	require.NotContains(t, selectors, string(converter.ToByteSlice(11)))
}
