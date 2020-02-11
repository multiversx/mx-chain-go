package genesis

import (
	"math/big"
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/files"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
)

func TestNewStateExporter(t *testing.T) {
	tests := []struct {
		name    string
		args    ArgsNewStateExporter
		exError error
	}{
		{name: "NilCoordinator", args: ArgsNewStateExporter{Marshalizer: &mock.MarshalizerMock{}, ShardCoordinator: nil, Hasher: &mock.HasherStub{},
			StateSyncer: &mock.SyncStateStub{}, Writer: &mock.MultiFileWriterStub{}}, exError: data.ErrNilShardCoordinator},
		{name: "NilStateSyncer", args: ArgsNewStateExporter{Marshalizer: &mock.MarshalizerMock{}, ShardCoordinator: mock.NewOneShardCoordinatorMock(),
			StateSyncer: nil, Writer: &mock.MultiFileWriterStub{}, Hasher: &mock.HasherStub{}}, exError: update.ErrNilStateSyncer},
		{name: "NilMarshalizer", args: ArgsNewStateExporter{Marshalizer: nil, ShardCoordinator: mock.NewOneShardCoordinatorMock(),
			StateSyncer: &mock.SyncStateStub{}, Writer: &mock.MultiFileWriterStub{}, Hasher: &mock.HasherStub{}}, exError: data.ErrNilMarshalizer},
		{name: "NilWriter", args: ArgsNewStateExporter{Marshalizer: &mock.MarshalizerMock{}, ShardCoordinator: mock.NewOneShardCoordinatorMock(),
			StateSyncer: &mock.SyncStateStub{}, Writer: nil, Hasher: &mock.HasherStub{}}, exError: epochStart.ErrNilStorage},
		{name: "NilHasher", args: ArgsNewStateExporter{Marshalizer: &mock.MarshalizerMock{}, ShardCoordinator: mock.NewOneShardCoordinatorMock(),
			StateSyncer: &mock.SyncStateStub{}, Writer: &mock.MultiFileWriterStub{}, Hasher: nil}, exError: update.ErrNilHasher},
		{name: "Ok", args: ArgsNewStateExporter{Marshalizer: &mock.MarshalizerMock{}, ShardCoordinator: mock.NewOneShardCoordinatorMock(),
			StateSyncer: &mock.SyncStateStub{}, Writer: &mock.MultiFileWriterStub{}, Hasher: &mock.HasherStub{}}, exError: nil},
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
	_ = os.Mkdir(testFolderName, 0755)
	testPath := "./" + testFolderName
	storer := mock.NewStorerMock()
	argsWriter := files.ArgsNewMultiFileWriter{ExportFolder: testPath, ExportStore: storer}
	writer, _ := files.NewMultiFileWriter(argsWriter)

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

	args := ArgsNewStateExporter{
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		Marshalizer:      &mock.MarshalizerMock{},
		StateSyncer:      stateSyncer,
		Writer:           writer,
		Hasher:           &mock.HasherMock{},
	}

	stateExporter, _ := NewStateExporter(args)
	require.False(t, check.IfNil(stateExporter))

	err := stateExporter.ExportAll(1)
	require.Nil(t, err)

	// check if export files was created
	if _, err := os.Stat(testPath + "/" + MetaBlockFileName); err != nil {
		require.Fail(t, "file wasn't created"+MetaBlockFileName)
	}
	if _, err := os.Stat(testPath + "/" + MiniBlocksFileName); err != nil {
		require.Fail(t, "file wasn't created"+MiniBlocksFileName)
	}
	if _, err := os.Stat(testPath + "/" + TransactionsFileName); err != nil {
		require.Fail(t, "file wasn't created"+TransactionsFileName)
	}
}
