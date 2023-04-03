package genesis

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStateExporter(t *testing.T) {
	tests := []struct {
		name            string
		args            ArgsNewStateExporter
		requiresErrorIs bool
		exError         error
	}{
		{
			name: "NilCoordinator",
			args: ArgsNewStateExporter{
				Marshalizer:              &mock.MarshalizerMock{},
				ShardCoordinator:         nil,
				Hasher:                   &mock.HasherStub{},
				StateSyncer:              &mock.StateSyncStub{},
				HardforkStorer:           &mock.HardforkStorerStub{},
				AddressPubKeyConverter:   &mock.PubkeyConverterStub{},
				ValidatorPubKeyConverter: &mock.PubkeyConverterStub{},
				GenesisNodesSetupHandler: &mock.GenesisNodesSetupHandlerStub{},
				ExportFolder:             "test",
			},
			exError: data.ErrNilShardCoordinator,
		},
		{
			name: "NilStateSyncer",
			args: ArgsNewStateExporter{
				Marshalizer:              &mock.MarshalizerMock{},
				ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
				StateSyncer:              nil,
				HardforkStorer:           &mock.HardforkStorerStub{},
				Hasher:                   &mock.HasherStub{},
				AddressPubKeyConverter:   &mock.PubkeyConverterStub{},
				ValidatorPubKeyConverter: &mock.PubkeyConverterStub{},
				GenesisNodesSetupHandler: &mock.GenesisNodesSetupHandlerStub{},
				ExportFolder:             "test",
			},
			exError: update.ErrNilStateSyncer,
		},
		{
			name: "NilMarshalizer",
			args: ArgsNewStateExporter{
				Marshalizer:              nil,
				ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
				StateSyncer:              &mock.StateSyncStub{},
				HardforkStorer:           &mock.HardforkStorerStub{},
				Hasher:                   &mock.HasherStub{},
				AddressPubKeyConverter:   &mock.PubkeyConverterStub{},
				ValidatorPubKeyConverter: &mock.PubkeyConverterStub{},
				GenesisNodesSetupHandler: &mock.GenesisNodesSetupHandlerStub{},
				ExportFolder:             "test",
			},
			exError: data.ErrNilMarshalizer,
		},
		{
			name: "NilHardforkStorer",
			args: ArgsNewStateExporter{
				Marshalizer:              &mock.MarshalizerMock{},
				ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
				StateSyncer:              &mock.StateSyncStub{},
				HardforkStorer:           nil,
				Hasher:                   &mock.HasherStub{},
				AddressPubKeyConverter:   &mock.PubkeyConverterStub{},
				ValidatorPubKeyConverter: &mock.PubkeyConverterStub{},
				GenesisNodesSetupHandler: &mock.GenesisNodesSetupHandlerStub{},
				ExportFolder:             "test",
			},
			exError: update.ErrNilHardforkStorer,
		},
		{
			name: "NilHasher",
			args: ArgsNewStateExporter{
				Marshalizer:              &mock.MarshalizerMock{},
				ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
				StateSyncer:              &mock.StateSyncStub{},
				HardforkStorer:           &mock.HardforkStorerStub{},
				Hasher:                   nil,
				AddressPubKeyConverter:   &mock.PubkeyConverterStub{},
				ValidatorPubKeyConverter: &mock.PubkeyConverterStub{},
				GenesisNodesSetupHandler: &mock.GenesisNodesSetupHandlerStub{},
				ExportFolder:             "test",
			},
			exError: update.ErrNilHasher,
		},
		{
			name:            "NilAddressPubKeyConverter",
			requiresErrorIs: true,
			args: ArgsNewStateExporter{
				Marshalizer:              &mock.MarshalizerMock{},
				ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
				StateSyncer:              &mock.StateSyncStub{},
				HardforkStorer:           &mock.HardforkStorerStub{},
				Hasher:                   &mock.HasherStub{},
				AddressPubKeyConverter:   nil,
				ValidatorPubKeyConverter: &mock.PubkeyConverterStub{},
				GenesisNodesSetupHandler: &mock.GenesisNodesSetupHandlerStub{},
				ExportFolder:             "test",
			},
			exError: update.ErrNilPubKeyConverter,
		},
		{
			name:            "NilValidatorPubKeyConverter",
			requiresErrorIs: true,
			args: ArgsNewStateExporter{
				Marshalizer:              &mock.MarshalizerMock{},
				ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
				StateSyncer:              &mock.StateSyncStub{},
				HardforkStorer:           &mock.HardforkStorerStub{},
				Hasher:                   &mock.HasherStub{},
				AddressPubKeyConverter:   &mock.PubkeyConverterStub{},
				ValidatorPubKeyConverter: nil,
				GenesisNodesSetupHandler: &mock.GenesisNodesSetupHandlerStub{},
				ExportFolder:             "test",
			},
			exError: update.ErrNilPubKeyConverter,
		},
		{
			name: "NilGenesisNodesSetupHandler",
			args: ArgsNewStateExporter{
				Marshalizer:              &mock.MarshalizerMock{},
				ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
				StateSyncer:              &mock.StateSyncStub{},
				HardforkStorer:           &mock.HardforkStorerStub{},
				Hasher:                   &mock.HasherStub{},
				AddressPubKeyConverter:   &mock.PubkeyConverterStub{},
				ValidatorPubKeyConverter: &mock.PubkeyConverterStub{},
				GenesisNodesSetupHandler: nil,
				ExportFolder:             "test",
			},
			exError: update.ErrNilGenesisNodesSetupHandler,
		},
		{
			name: "EmptyExportFolder",
			args: ArgsNewStateExporter{
				Marshalizer:              &mock.MarshalizerMock{},
				ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
				StateSyncer:              &mock.StateSyncStub{},
				HardforkStorer:           &mock.HardforkStorerStub{},
				Hasher:                   &mock.HasherStub{},
				AddressPubKeyConverter:   &mock.PubkeyConverterStub{},
				ValidatorPubKeyConverter: &mock.PubkeyConverterStub{},
				GenesisNodesSetupHandler: &mock.GenesisNodesSetupHandlerStub{},
				ExportFolder:             "",
			},
			exError: update.ErrEmptyExportFolderPath,
		},
		{
			name: "Ok",
			args: ArgsNewStateExporter{
				Marshalizer:              &mock.MarshalizerMock{},
				ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
				StateSyncer:              &mock.StateSyncStub{},
				HardforkStorer:           &mock.HardforkStorerStub{},
				Hasher:                   &mock.HasherStub{},
				AddressPubKeyConverter:   &mock.PubkeyConverterStub{},
				ValidatorPubKeyConverter: &mock.PubkeyConverterStub{},
				ExportFolder:             "test",
				GenesisNodesSetupHandler: &mock.GenesisNodesSetupHandlerStub{},
			},
			exError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewStateExporter(tt.args)
			if tt.requiresErrorIs {
				require.True(t, errors.Is(err, tt.exError))
			} else {
				require.Equal(t, err, tt.exError)
			}
		})
	}
}

func TestExportAll(t *testing.T) {
	t.Parallel()

	testFolderName := "testFiles"
	testPath := "./" + testFolderName
	defer func() {
		_ = os.RemoveAll(testPath)
	}()

	metaBlock := &block.MetaBlock{Round: 2, ChainID: []byte("chainId")}
	unFinishedMetaBlocks := map[string]data.MetaHeaderHandler{
		"hash": &block.MetaBlock{Round: 1, ChainID: []byte("chainId")},
	}
	miniBlock := &block.MiniBlock{}
	tx := &transaction.Transaction{Nonce: 1, Value: big.NewInt(100), SndAddr: []byte("snd"), RcvAddr: []byte("rcv")}
	stateSyncer := &mock.StateSyncStub{
		GetEpochStartMetaBlockCalled: func() (block data.MetaHeaderHandler, err error) {
			return metaBlock, nil
		},
		GetUnFinishedMetaBlocksCalled: func() (map[string]data.MetaHeaderHandler, error) {
			return unFinishedMetaBlocks, nil
		},
		GetAllMiniBlocksCalled: func() (m map[string]*block.MiniBlock, err error) {
			mbs := make(map[string]*block.MiniBlock)
			mbs["mb"] = miniBlock
			return mbs, nil
		},
		GetAllTransactionsCalled: func() (m map[string]data.TransactionHandler, err error) {
			mt := make(map[string]data.TransactionHandler)
			mt["tx"] = tx
			return mt, nil
		},
	}

	defer func() {
		_ = os.RemoveAll("./" + testFolderName + "/")
	}()

	transactionsWereWrote := false
	miniblocksWereWrote := false
	epochStartMetablockWasWrote := false
	unFinishedMetablocksWereWrote := false
	hs := &mock.HardforkStorerStub{
		WriteCalled: func(identifier string, key []byte, value []byte) error {
			switch identifier {
			case TransactionsIdentifier:
				transactionsWereWrote = true
			case MiniBlocksIdentifier:
				miniblocksWereWrote = true
			case EpochStartMetaBlockIdentifier:
				epochStartMetablockWasWrote = true
			case UnFinishedMetaBlocksIdentifier:
				unFinishedMetablocksWereWrote = true

			}

			return nil
		},
	}

	args := getDefaultStateExporterArgs()
	args.StateSyncer = stateSyncer
	args.HardforkStorer = hs
	stateExporter, _ := NewStateExporter(args)
	require.False(t, check.IfNil(stateExporter))

	err := stateExporter.ExportAll(1)
	require.Nil(t, err)

	assert.True(t, transactionsWereWrote)
	assert.True(t, miniblocksWereWrote)
	assert.True(t, epochStartMetablockWasWrote)
	assert.True(t, unFinishedMetablocksWereWrote)
}

func TestStateExport_ExportTrieShouldExportNodesSetupJson(t *testing.T) {
	t.Parallel()

	t.Run("should fail on getting trie nodes", func(t *testing.T) {
		t.Parallel()

		args := getDefaultStateExporterArgs()

		expectedErr := errors.New("expected error")
		trie := &trieMock.TrieStub{
			RootCalled: func() ([]byte, error) {
				return []byte{}, nil
			},
			GetAllLeavesOnChannelCalled: func(channels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, keyBuilder common.KeyBuilder) error {
				mm := &mock.MarshalizerMock{}
				valInfo := &state.ValidatorInfo{List: string(common.EligibleList)}
				pacB, _ := mm.Marshal(valInfo)

				go func() {
					channels.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("test"), pacB)
					channels.ErrChan.WriteInChanNonBlocking(expectedErr)
					close(channels.LeavesChan)
				}()

				return nil
			},
		}

		stateExporter, err := NewStateExporter(args)
		require.NoError(t, err)

		err = stateExporter.exportTrie("test@1@9", trie)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work export nodes setup json", func(t *testing.T) {
		t.Parallel()

		testFolderName := t.TempDir()

		hs := &mock.HardforkStorerStub{
			WriteCalled: func(identifier string, key []byte, value []byte) error {
				return nil
			},
		}

		pubKeyConv := &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}

		args := getDefaultStateExporterArgs()
		args.HardforkStorer = hs
		args.ExportFolder = testFolderName
		args.AddressPubKeyConverter = pubKeyConv
		args.ValidatorPubKeyConverter = pubKeyConv

		trie := &trieMock.TrieStub{
			RootCalled: func() ([]byte, error) {
				return []byte{}, nil
			},
			GetAllLeavesOnChannelCalled: func(channels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, keyBuilder common.KeyBuilder) error {
				mm := &mock.MarshalizerMock{}
				valInfo := &state.ValidatorInfo{List: string(common.EligibleList)}
				pacB, _ := mm.Marshal(valInfo)

				go func() {
					channels.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("test"), pacB)
					close(channels.LeavesChan)
					channels.ErrChan.Close()
				}()

				return nil
			},
		}

		stateExporter, err := NewStateExporter(args)
		require.NoError(t, err)

		require.False(t, check.IfNil(stateExporter))

		err = stateExporter.exportTrie("test@1@9", trie)
		require.NoError(t, err)
	})
}

func TestStateExport_ExportNodesSetupJsonShouldExportKeysInAlphabeticalOrder(t *testing.T) {
	t.Parallel()

	testFolderName := t.TempDir()

	hs := &mock.HardforkStorerStub{
		WriteCalled: func(identifier string, key []byte, value []byte) error {
			return nil
		},
	}

	pubKeyConv := &mock.PubkeyConverterStub{
		EncodeCalled: func(pkBytes []byte) string {
			return string(pkBytes)
		},
	}

	args := getDefaultStateExporterArgs()
	args.HardforkStorer = hs
	args.ExportFolder = testFolderName
	args.AddressPubKeyConverter = pubKeyConv
	args.ValidatorPubKeyConverter = pubKeyConv

	stateExporter, err := NewStateExporter(args)
	require.NoError(t, err)

	require.False(t, check.IfNil(stateExporter))

	vals := make(map[uint32][]*state.ValidatorInfo)
	val50 := &state.ValidatorInfo{ShardId: 5, PublicKey: []byte("aaa"), List: string(common.EligibleList)}
	val51 := &state.ValidatorInfo{ShardId: 5, PublicKey: []byte("bbb"), List: string(common.EligibleList)}
	val10 := &state.ValidatorInfo{ShardId: 5, PublicKey: []byte("ccc"), List: string(common.EligibleList)}
	val11 := &state.ValidatorInfo{ShardId: 5, PublicKey: []byte("ddd"), List: string(common.EligibleList)}
	val00 := &state.ValidatorInfo{ShardId: 5, PublicKey: []byte("aaaaaa"), List: string(common.EligibleList)}
	val01 := &state.ValidatorInfo{ShardId: 5, PublicKey: []byte("bbbbbb"), List: string(common.EligibleList)}
	vals[1] = []*state.ValidatorInfo{val50, val51}
	vals[0] = []*state.ValidatorInfo{val00, val01}
	vals[2] = []*state.ValidatorInfo{val10, val11}
	err = stateExporter.exportNodesSetupJson(vals)
	require.Nil(t, err)

	var nodesSetup sharding.NodesSetup

	nsBytes, err := ioutil.ReadFile(filepath.Join(testFolderName, common.NodesSetupJsonFileName))
	require.NoError(t, err)

	err = json.Unmarshal(nsBytes, &nodesSetup)
	require.NoError(t, err)

	initialNodes := nodesSetup.InitialNodes

	// results should be in alphabetical order, sorted by public key
	require.Equal(t, string(val50.PublicKey), initialNodes[0].PubKey) // aaa
	require.Equal(t, string(val00.PublicKey), initialNodes[1].PubKey) // aaaaaa
	require.Equal(t, string(val51.PublicKey), initialNodes[2].PubKey) // bbb
	require.Equal(t, string(val01.PublicKey), initialNodes[3].PubKey) // bbbbbb
	require.Equal(t, string(val10.PublicKey), initialNodes[4].PubKey) // ccc
	require.Equal(t, string(val11.PublicKey), initialNodes[5].PubKey) // ddd
}

func TestStateExport_ExportUnfinishedMetaBlocksShouldWork(t *testing.T) {
	t.Parallel()

	unFinishedMetaBlocks := map[string]data.MetaHeaderHandler{
		"hash": &block.MetaBlock{Round: 1, ChainID: []byte("chainId")},
	}
	stateSyncer := &mock.StateSyncStub{
		GetUnFinishedMetaBlocksCalled: func() (map[string]data.MetaHeaderHandler, error) {
			return unFinishedMetaBlocks, nil
		},
	}

	unFinishedMetablocksWereWrote := false
	hs := &mock.HardforkStorerStub{
		WriteCalled: func(identifier string, key []byte, value []byte) error {
			if strings.Compare(identifier, UnFinishedMetaBlocksIdentifier) == 0 {
				unFinishedMetablocksWereWrote = true
			}
			return nil
		},
	}

	args := getDefaultStateExporterArgs()
	args.StateSyncer = stateSyncer
	args.HardforkStorer = hs

	stateExporter, _ := NewStateExporter(args)
	require.False(t, check.IfNil(stateExporter))

	err := stateExporter.exportUnFinishedMetaBlocks()
	require.Nil(t, err)

	assert.True(t, unFinishedMetablocksWereWrote)
}

func TestStateExport_ExportAllValidatorsInfo(t *testing.T) {
	t.Parallel()

	t.Run("export all validators info with state syncer error", func(t *testing.T) {
		t.Parallel()

		expectedStateSyncerErr := errors.New("state syncer error")
		args := getDefaultStateExporterArgs()
		args.StateSyncer = &mock.StateSyncStub{
			GetAllValidatorsInfoCalled: func() (map[string]*state.ShardValidatorInfo, error) {
				return nil, expectedStateSyncerErr
			},
		}

		stateExporter, _ := NewStateExporter(args)
		err := stateExporter.exportAllValidatorsInfo()
		assert.Equal(t, expectedStateSyncerErr, err)
	})

	t.Run("export all validators info with hardfork storer error", func(t *testing.T) {
		t.Parallel()

		expectedHardforkStorerErr := errors.New("hardfork storer error")
		args := getDefaultStateExporterArgs()
		args.StateSyncer = &mock.StateSyncStub{
			GetAllValidatorsInfoCalled: func() (map[string]*state.ShardValidatorInfo, error) {
				mapShardValidatorInfo := make(map[string]*state.ShardValidatorInfo)
				shardValidatorInfo := &state.ShardValidatorInfo{
					PublicKey: []byte("x"),
				}
				mapShardValidatorInfo["key"] = shardValidatorInfo
				return mapShardValidatorInfo, nil
			},
		}
		args.HardforkStorer = &mock.HardforkStorerStub{
			WriteCalled: func(identifier string, key []byte, value []byte) error {
				return expectedHardforkStorerErr
			},
		}

		stateExporter, _ := NewStateExporter(args)
		err := stateExporter.exportAllValidatorsInfo()
		assert.Equal(t, expectedHardforkStorerErr, err)
	})

	t.Run("export all validators info without error", func(t *testing.T) {
		t.Parallel()

		finishedIdentifierWasCalled := false
		args := getDefaultStateExporterArgs()
		args.HardforkStorer = &mock.HardforkStorerStub{
			FinishedIdentifierCalled: func(identifier string) error {
				finishedIdentifierWasCalled = true
				return nil
			},
		}

		stateExporter, _ := NewStateExporter(args)
		err := stateExporter.exportAllValidatorsInfo()
		assert.Nil(t, err)
		assert.True(t, finishedIdentifierWasCalled)
	})
}

func TestStateExport_ExportValidatorInfo(t *testing.T) {
	t.Parallel()

	t.Run("export validator info with error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("error")
		args := getDefaultStateExporterArgs()
		args.HardforkStorer = &mock.HardforkStorerStub{
			WriteCalled: func(identifier string, key []byte, value []byte) error {
				return expectedErr
			},
		}

		stateExporter, _ := NewStateExporter(args)
		key := "key"
		shardValidatorInfo := &state.ShardValidatorInfo{
			PublicKey: []byte("x"),
		}

		err := stateExporter.exportValidatorInfo(key, shardValidatorInfo)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("export validator info without error", func(t *testing.T) {
		t.Parallel()

		args := getDefaultStateExporterArgs()

		stateExporter, _ := NewStateExporter(args)
		key := "key"
		shardValidatorInfo := &state.ShardValidatorInfo{
			PublicKey: []byte("x"),
		}

		err := stateExporter.exportValidatorInfo(key, shardValidatorInfo)
		assert.Nil(t, err)
	})
}

func getDefaultStateExporterArgs() ArgsNewStateExporter {
	return ArgsNewStateExporter{
		ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
		Marshalizer:              &mock.MarshalizerMock{},
		StateSyncer:              &mock.StateSyncStub{},
		HardforkStorer:           &mock.HardforkStorerStub{},
		Hasher:                   &hashingMocks.HasherMock{},
		AddressPubKeyConverter:   &mock.PubkeyConverterStub{},
		ValidatorPubKeyConverter: &mock.PubkeyConverterStub{},
		ExportFolder:             "test",
		GenesisNodesSetupHandler: &mock.GenesisNodesSetupHandlerStub{},
	}
}
