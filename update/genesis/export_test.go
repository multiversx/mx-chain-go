package genesis

import (
	"math/big"
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStateExporter(t *testing.T) {
	tests := []struct {
		name    string
		args    ArgsNewStateExporter
		exError error
	}{
		{
			name: "NilCoordinator",
			args: ArgsNewStateExporter{
				Marshalizer:      &mock.MarshalizerMock{},
				ShardCoordinator: nil,
				Hasher:           &mock.HasherStub{},
				StateSyncer:      &mock.SyncStateStub{},
				HardforkStorer:   &mock.HardforkStorerStub{},
			},
			exError: data.ErrNilShardCoordinator,
		},
		{
			name: "NilStateSyncer",
			args: ArgsNewStateExporter{
				Marshalizer:      &mock.MarshalizerMock{},
				ShardCoordinator: mock.NewOneShardCoordinatorMock(),
				StateSyncer:      nil,
				HardforkStorer:   &mock.HardforkStorerStub{},
				Hasher:           &mock.HasherStub{},
			},
			exError: update.ErrNilStateSyncer,
		},
		{
			name: "NilMarshalizer",
			args: ArgsNewStateExporter{
				Marshalizer:      nil,
				ShardCoordinator: mock.NewOneShardCoordinatorMock(),
				StateSyncer:      &mock.SyncStateStub{},
				HardforkStorer:   &mock.HardforkStorerStub{},
				Hasher:           &mock.HasherStub{},
			},
			exError: data.ErrNilMarshalizer,
		},
		{
			name: "NilHardforkStorer",
			args: ArgsNewStateExporter{
				Marshalizer:      &mock.MarshalizerMock{},
				ShardCoordinator: mock.NewOneShardCoordinatorMock(),
				StateSyncer:      &mock.SyncStateStub{},
				HardforkStorer:   nil,
				Hasher:           &mock.HasherStub{},
			},
			exError: update.ErrNilHardforkStorer,
		},
		{
			name: "NilHasher",
			args: ArgsNewStateExporter{
				Marshalizer:      &mock.MarshalizerMock{},
				ShardCoordinator: mock.NewOneShardCoordinatorMock(),
				StateSyncer:      &mock.SyncStateStub{},
				HardforkStorer:   &mock.HardforkStorerStub{},
				Hasher:           nil,
			},
			exError: update.ErrNilHasher,
		},
		{
			name: "Ok",
			args: ArgsNewStateExporter{
				Marshalizer:              &mock.MarshalizerMock{},
				ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
				StateSyncer:              &mock.SyncStateStub{},
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
			require.Equal(t, tt.exError, err)
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

	metaBlock := &block.MetaBlock{Round: 1, ChainID: []byte("chainId")}
	miniBlock := &block.MiniBlock{}
	tx := &transaction.Transaction{Nonce: 1, Value: big.NewInt(100), SndAddr: []byte("snd"), RcvAddr: []byte("rcv")}
	stateSyncer := &mock.SyncStateStub{
		GetEpochStartMetaBlockCalled: func() (block *block.MetaBlock, err error) {
			return metaBlock, nil
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
	metablockWasWrote := false
	hs := &mock.HardforkStorerStub{
		WriteCalled: func(identifier string, key []byte, value []byte) error {
			switch identifier {
			case TransactionsIdentifier:
				transactionsWereWrote = true
			case MiniBlocksIdentifier:
				miniblocksWereWrote = true
			case MetaBlockIdentifier:
				metablockWasWrote = true
			}

			return nil
		},
	}

	args := ArgsNewStateExporter{
		ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
		Marshalizer:              &mock.MarshalizerMock{},
		StateSyncer:              stateSyncer,
		HardforkStorer:           hs,
		Hasher:                   &mock.HasherMock{},
		AddressPubKeyConverter:   &mock.PubkeyConverterStub{},
		ValidatorPubKeyConverter: &mock.PubkeyConverterStub{},
		ExportFolder:             "test",
		GenesisNodesSetupHandler: &mock.GenesisNodesSetupHandlerStub{},
	}

	stateExporter, _ := NewStateExporter(args)
	require.False(t, check.IfNil(stateExporter))

	err := stateExporter.ExportAll(1)
	require.Nil(t, err)

	assert.True(t, transactionsWereWrote)
	assert.True(t, miniblocksWereWrote)
	assert.True(t, metablockWasWrote)
}

func TestStateExport_ExportTrieShouldExportNodesSetupJson(t *testing.T) {
	t.Parallel()

	testFolderName := "testFilesExportNodes"
	_ = os.Mkdir(testFolderName, 0777)

	defer func() {
		_ = os.RemoveAll(testFolderName)
	}()

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

	args := ArgsNewStateExporter{
		ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
		Marshalizer:              &mock.MarshalizerMock{},
		StateSyncer:              &mock.SyncStateStub{},
		HardforkStorer:           hs,
		Hasher:                   &mock.HasherMock{},
		ExportFolder:             testFolderName,
		AddressPubKeyConverter:   pubKeyConv,
		ValidatorPubKeyConverter: pubKeyConv,
		GenesisNodesSetupHandler: &mock.GenesisNodesSetupHandlerStub{},
	}

	trie := &mock.TrieStub{
		GetAllLeavesCalled: func() (map[string][]byte, error) {
			mm := &mock.MarshalizerMock{}
			valInfo := &state.ValidatorInfo{List: string(core.EligibleList)}
			pacB, _ := mm.Marshal(valInfo)
			return map[string][]byte{"test": pacB}, nil
		},
	}

	stateExporter, err := NewStateExporter(args)
	require.NoError(t, err)

	require.False(t, check.IfNil(stateExporter))

	err = stateExporter.exportTrie("test@1@9", trie)
	require.NoError(t, err)
}
