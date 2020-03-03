package genesis

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/files"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
)

func createTxsMap() map[string]data.TransactionHandler {
	tx := &transaction.Transaction{Nonce: 1, Value: big.NewInt(100), SndAddr: []byte("snd"), RcvAddr: []byte("rcv")}
	rtx := &rewardTx.RewardTx{Round: 1, Value: big.NewInt(100), RcvAddr: []byte("rcv"), Epoch: 1}
	smrTx := &smartContractResult.SmartContractResult{
		Nonce: 1, SndAddr: []byte("snd"), RcvAddr: []byte("rcv"), Value: big.NewInt(100), TxHash: []byte("hash"),
		GasLimit: 1, GasPrice: 100,
	}

	mt := make(map[string]data.TransactionHandler)
	txBytes, _ := json.Marshal(tx)
	txHash := mock.HasherMock{}.Compute(string(txBytes))

	rtxBytes, _ := json.Marshal(rtx)
	rtxHash := mock.HasherMock{}.Compute(string(rtxBytes))

	smrBytes, _ := json.Marshal(smrTx)
	smrHash := mock.HasherMock{}.Compute(string(smrBytes))

	mt[string(txHash)] = tx
	mt[string(rtxHash)] = rtx
	mt[string(smrHash)] = smrTx
	return mt
}

func createMbMap() map[string]*block.MiniBlock {
	miniBlock1 := &block.MiniBlock{}
	miniBlock2 := &block.MiniBlock{ReceiverShardID: 0, SenderShardID: 1, TxHashes: [][]byte{[]byte("Hash1")}, Type: block.TxBlock}
	miniBlock3 := &block.MiniBlock{ReceiverShardID: 0, SenderShardID: 2, TxHashes: [][]byte{[]byte("Hash2")}, Type: block.RewardsBlock}
	mbs := make(map[string]*block.MiniBlock)

	mbBytes1, _ := json.Marshal(miniBlock1)
	mbHash1 := mock.HasherMock{}.Compute(string(mbBytes1))
	mbBytes2, _ := json.Marshal(miniBlock2)
	mbHash2 := mock.HasherMock{}.Compute(string(mbBytes2))
	mbBytes3, _ := json.Marshal(miniBlock3)
	mbHash3 := mock.HasherMock{}.Compute(string(mbBytes3))

	mbs[string(mbHash1)] = miniBlock1
	mbs[string(mbHash2)] = miniBlock2
	mbs[string(mbHash3)] = miniBlock3
	return mbs
}

func createTestImportFile(t *testing.T, folderName string, storer storage.Storer) {
	argsWriter := files.ArgsNewMultiFileWriter{ExportFolder: folderName, ExportStore: storer}
	writer, _ := files.NewMultiFileWriter(argsWriter)

	metaBlock := &block.MetaBlock{Round: 1, ChainID: []byte("chainId")}

	stateSyncer := &mock.SyncStateStub{
		GetEpochStartMetaBlockCalled: func() (block *block.MetaBlock, err error) {
			return metaBlock, nil
		},
		GetAllMiniBlocksCalled: func() (m map[string]*block.MiniBlock, err error) {

			return createMbMap(), nil
		},
		GetAllTransactionsCalled: func() (m map[string]data.TransactionHandler, err error) {
			return createTxsMap(), nil
		},
	}
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
	if _, err = os.Stat(folderName + "/" + MetaBlockFileName); err != nil {
		require.Fail(t, "file wasn't created"+MetaBlockFileName)
	}
	if _, err = os.Stat(folderName + "/" + MiniBlocksFileName); err != nil {
		require.Fail(t, "file wasn't created"+MiniBlocksFileName)
	}
	if _, err = os.Stat(folderName + "/" + TransactionsFileName); err != nil {
		require.Fail(t, "file wasn't created"+TransactionsFileName)
	}
}

func TestNewStateImport(t *testing.T) {
	tests := []struct {
		name    string
		args    ArgsNewStateImport
		exError error
	}{
		{name: "NilReader", args: ArgsNewStateImport{Reader: nil, Marshalizer: &mock.MarshalizerMock{}, Hasher: &mock.HasherStub{}}, exError: update.ErrNilMultiFileReader},
		{name: "NilMarshalizer", args: ArgsNewStateImport{Reader: &mock.MultiFileReaderStub{}, Marshalizer: nil, Hasher: &mock.HasherStub{}}, exError: update.ErrNilMarshalizer},
		{name: "NilHasher", args: ArgsNewStateImport{Reader: &mock.MultiFileReaderStub{}, Marshalizer: &mock.MarshalizerMock{}, Hasher: nil}, exError: update.ErrNilHasher},
		{name: "Ok", args: ArgsNewStateImport{Reader: &mock.MultiFileReaderStub{}, Marshalizer: &mock.MarshalizerMock{}, Hasher: &mock.HasherStub{}}, exError: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewStateImport(tt.args)
			require.Equal(t, tt.exError, err)
		})
	}
}

func TestImportAll(t *testing.T) {
	t.Parallel()

	folderName := "./files"
	defer func() {
		_ = os.RemoveAll(folderName)
	}()

	storer := mock.NewStorerMock()
	createTestImportFile(t, folderName, storer)

	argsReader := files.ArgsNewMultiFileReader{
		ImportFolder: folderName,
		ImportStore:  storer,
	}
	reader, _ := files.NewMultiFileReader(argsReader)
	args := ArgsNewStateImport{
		Reader:      reader,
		Hasher:      &mock.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
	}

	importState, _ := NewStateImport(args)
	require.False(t, check.IfNil(importState))

	err := importState.ImportAll()
	require.Nil(t, err)
}
