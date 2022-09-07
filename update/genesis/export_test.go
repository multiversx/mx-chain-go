package genesis

import (
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
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
