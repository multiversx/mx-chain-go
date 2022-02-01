package blockAPI

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process/txstatus"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

func createMockArgumentsWithTx(
	srcShardID uint32,
	destShardID uint32,
	recvAddress string,
	miniblockBytes []byte,
	miniblockType block.Type,
	txHash string,
	txBytes []byte,
	marshalizer *mock.MarshalizerFake,
) baseAPIBlockProcessor {
	storerMock := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &storageStubs.StorerStub{
				GetBulkFromEpochCalled: func(keys [][]byte, epoch uint32) (map[string][]byte, error) {
					return map[string][]byte{txHash: txBytes}, nil
				},
				GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
					return miniblockBytes, nil
				},
			}
		},
	}
	statusComputer, _ := txstatus.NewStatusComputer(srcShardID, mock.NewNonceHashConverterMock(), storerMock)
	return baseAPIBlockProcessor{
		selfShardID: srcShardID,
		marshalizer: marshalizer,
		store:       storerMock,
		historyRepo: &dblookupext.HistoryRepositoryStub{
			IsEnabledCalled: func() bool {
				return false
			},
		},
		unmarshalTx: func(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
			var unmarshalledTx transaction.Transaction
			_ = marshalizer.Unmarshal(&unmarshalledTx, txBytes)
			return &transaction.ApiTransactionResult{
				Tx:               &unmarshalledTx,
				Type:             string(txType),
				Nonce:            unmarshalledTx.Nonce,
				Hash:             txHash,
				MiniBlockType:    string(miniblockType),
				SourceShard:      srcShardID,
				DestinationShard: destShardID,
				Receiver:         recvAddress,
				Data:             []byte{},
			}, nil
		},
		txStatusComputer: statusComputer,
	}
}

func TestBaseApiBlockProcessor_GetNormalTxFromMiniBlock(t *testing.T) {
	t.Parallel()

	epoch := uint32(1)
	nonce := uint64(0)
	sourceShardID := uint32(1)
	destShardID := uint32(1)
	marshalizer := &mock.MarshalizerFake{}

	txHash := "d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00"
	mbHash := "f08089d2ab739520598ff7aeed08c427460fe94f286383047f3f61951afc4e04"
	recvAddress := "a08089d2ab739520598ff7aeed08c427460fe94f286383047f3f61951afc4e04"

	mbType := block.TxBlock
	txType := transaction.TxTypeNormal

	tx := transaction.Transaction{
		Nonce: nonce,
	}

	mb := block.MiniBlock{
		TxHashes: [][]byte{[]byte(txHash)},
		Type:     mbType,
	}

	mbHeader := block.MiniBlockHeader{
		Hash: []byte(mbHash),
		Type: mbType,
	}

	txBytes, _ := marshalizer.Marshal(&tx)
	mbBytes, _ := marshalizer.Marshal(&mb)

	baseAPIBlock := createMockArgumentsWithTx(
		sourceShardID,
		destShardID,
		recvAddress,
		mbBytes,
		mbType,
		txHash,
		txBytes,
		marshalizer,
	)

	mbTxs := baseAPIBlock.getTxsByMb(&mbHeader, epoch)

	assert.Equal(t, mbTxs[0].Nonce, tx.Nonce)
	assert.EqualValues(t, mbTxs[0].Type, txType)
	assert.EqualValues(t, mbTxs[0].Receiver, recvAddress)
}

func TestBaseApiBlockProcessor_GetRewardsTxFromMiniBlock(t *testing.T) {
	t.Parallel()

	epoch := uint32(1)
	nonce := uint64(0)
	sourceShardID := uint32(1)
	destShardID := uint32(1)
	marshalizer := &mock.MarshalizerFake{}

	txHash := "d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00"
	mbHash := "f08089d2ab739520598ff7aeed08c427460fe94f286383047f3f61951afc4e04"
	recvAddress := "a08089d2ab739520598ff7aeed08c427460fe94f286383047f3f61951afc4e04"

	mbType := block.RewardsBlock
	txType := transaction.TxTypeReward

	tx := transaction.Transaction{
		Nonce: nonce,
	}

	mb := block.MiniBlock{
		TxHashes: [][]byte{[]byte(txHash)},
		Type:     mbType,
	}

	mbHeader := block.MiniBlockHeader{
		Hash: []byte(mbHash),
		Type: mbType,
	}

	txBytes, _ := marshalizer.Marshal(&tx)
	mbBytes, _ := marshalizer.Marshal(&mb)

	baseAPIBlock := createMockArgumentsWithTx(
		sourceShardID,
		destShardID,
		recvAddress,
		mbBytes,
		mbType,
		txHash,
		txBytes,
		marshalizer,
	)

	mbTxs := baseAPIBlock.getTxsByMb(&mbHeader, epoch)

	assert.Equal(t, mbTxs[0].Nonce, tx.Nonce)
	assert.EqualValues(t, mbTxs[0].Type, txType)
	assert.EqualValues(t, mbTxs[0].Receiver, recvAddress)
}

func TestBaseApiBlockProcessor_GetUnsignedTxFromMiniBlock(t *testing.T) {
	t.Parallel()

	epoch := uint32(1)
	nonce := uint64(0)
	sourceShardID := uint32(1)
	destShardID := uint32(1)
	marshalizer := &mock.MarshalizerFake{}

	txHash := "d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00"
	mbHash := "f08089d2ab739520598ff7aeed08c427460fe94f286383047f3f61951afc4e04"
	recvAddress := "a08089d2ab739520598ff7aeed08c427460fe94f286383047f3f61951afc4e04"

	mbType := block.SmartContractResultBlock
	txType := transaction.TxTypeUnsigned

	tx := transaction.Transaction{
		Nonce: nonce,
	}

	mb := block.MiniBlock{
		TxHashes: [][]byte{[]byte(txHash)},
		Type:     mbType,
	}

	mbHeader := block.MiniBlockHeader{
		Hash: []byte(mbHash),
		Type: mbType,
	}

	txBytes, _ := marshalizer.Marshal(&tx)
	mbBytes, _ := marshalizer.Marshal(&mb)

	baseAPIBlock := createMockArgumentsWithTx(
		sourceShardID,
		destShardID,
		recvAddress,
		mbBytes,
		mbType,
		txHash,
		txBytes,
		marshalizer,
	)

	mbTxs := baseAPIBlock.getTxsByMb(&mbHeader, epoch)

	assert.Equal(t, mbTxs[0].Nonce, tx.Nonce)
	assert.EqualValues(t, mbTxs[0].Type, txType)
	assert.EqualValues(t, mbTxs[0].Receiver, recvAddress)
}

func TestBaseApiBlockProcessor_GetInvalidTxFromMiniBlock(t *testing.T) {
	t.Parallel()

	epoch := uint32(1)
	nonce := uint64(0)
	sourceShardID := uint32(1)
	destShardID := uint32(1)
	marshalizer := &mock.MarshalizerFake{}

	txHash := "d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00"
	mbHash := "f08089d2ab739520598ff7aeed08c427460fe94f286383047f3f61951afc4e04"
	recvAddress := "a08089d2ab73952059  8ff7aeed08c427460fe94f286383047f3f61951afc4e04"

	mbType := block.InvalidBlock
	txType := transaction.TxTypeInvalid

	tx := transaction.Transaction{
		Nonce: nonce,
	}

	mb := block.MiniBlock{
		TxHashes: [][]byte{[]byte(txHash)},
		Type:     mbType,
	}

	mbHeader := block.MiniBlockHeader{
		Hash: []byte(mbHash),
		Type: mbType,
	}

	txBytes, _ := marshalizer.Marshal(&tx)
	mbBytes, _ := marshalizer.Marshal(&mb)

	baseAPIBlock := createMockArgumentsWithTx(
		sourceShardID,
		destShardID,
		recvAddress,
		mbBytes,
		mbType,
		txHash,
		txBytes,
		marshalizer,
	)

	mbTxs := baseAPIBlock.getTxsByMb(&mbHeader, epoch)

	assert.Equal(t, mbTxs[0].Nonce, tx.Nonce)
	assert.EqualValues(t, mbTxs[0].Type, txType)
	assert.EqualValues(t, mbTxs[0].Receiver, recvAddress)
}
