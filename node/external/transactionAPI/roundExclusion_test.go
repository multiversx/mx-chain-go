package transactionAPI

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/storage"
	dbmock "github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	storagemock "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestTransactionExclusionsWithoutHistoryPreserveStorageLookup(t *testing.T) {
	args := createMockArgAPITransactionProcessor()
	args.HistoryRepository = &dbmock.HistoryRepositoryStub{
		IsEnabledCalled: func() bool { return false },
		GetMiniblockMetadataByTxHashCalled: func([]byte) (*dblookupext.MiniblockMetadata, error) {
			t.Fatal("must not query disabled history")
			return nil, nil
		},
	}
	var err error
	args.RoundExclusionHandler, err = common.NewRoundExclusionHandler([]config.HardforkRoundExclusionConfig{{StartRound: 100, EndRound: 199}})
	require.NoError(t, err)
	txBytes, err := args.Marshalizer.Marshal(&transaction.Transaction{Value: big.NewInt(1), SndAddr: []byte("alice"), RcvAddr: []byte("bob")})
	require.NoError(t, err)
	scrBytes, err := args.Marshalizer.Marshal(&smartContractResult.SmartContractResult{Value: big.NewInt(1), SndAddr: []byte("alice"), RcvAddr: []byte("bob")})
	require.NoError(t, err)
	args.StorageService = &storagemock.ChainStorerStub{GetStorerCalled: func(dataRetriever.UnitType) (storage.Storer, error) {
		return &storagemock.StorerStub{
			SearchFirstCalled:  func([]byte) ([]byte, error) { return txBytes, nil },
			GetFromEpochCalled: func([]byte, uint32) ([]byte, error) { return scrBytes, nil },
		}, nil
	}}
	proc, err := NewAPITransactionProcessor(args)
	require.NoError(t, err)
	tx, err := proc.GetTransaction(hex.EncodeToString([]byte("tx")), false)
	require.NoError(t, err)
	require.Zero(t, tx.Round)
	results, err := proc.transactionResultsProcessor.getSmartContractResultsInTransactionByHashesAndEpoch([][]byte{[]byte("scr")}, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
}
