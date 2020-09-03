package indexer

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

func TestDataIndexer(t *testing.T) {
	//t.Skip("this is not a short test")
	testCreateIndexer(t)
}

func testCreateIndexer(t *testing.T) {
	indexTemplates := make(map[string]io.Reader)
	indexPolicies := make(map[string]io.Reader)
	opendistroTemplate, _ := core.OpenFile("./testdata/opendistro.json")
	indexTemplates["opendistro"] = opendistroTemplate
	transactionsTemplate, _ := core.OpenFile("./testdata/transactions.json")
	indexTemplates["transactions"] = transactionsTemplate
	blocksTemplate, _ := core.OpenFile("./testdata/blocks.json")
	indexTemplates["blocks"] = blocksTemplate
	miniblocksTemplate, _ := core.OpenFile("./testdata/miniblocks.json")
	indexTemplates["miniblocks"] = miniblocksTemplate
	transactionsPolicy, _ := core.OpenFile("./testdata/transactions_policy.json")
	indexPolicies["transactions_policy"] = transactionsPolicy
	blocksPolicy, _ := core.OpenFile("./testdata/blocks_policy.json")
	indexPolicies["blocks_policy"] = blocksPolicy

	tpsTemplate, _ := core.OpenFile("./testdata/tps.json")
	indexTemplates["tps"] = tpsTemplate

	dataIndexer, err := NewDataIndexer(DataIndexerArgs{
		Url:            "https://search-elrond-test-okohrj6g5r575cvmkwfv6jraki.eu-west-1.es.amazonaws.com/",
		IndexTemplates: indexTemplates,
		IndexPolicies:  indexPolicies,
		Options: &Options{
			TxIndexingEnabled: true,
		},
		Marshalizer:              &marshal.JsonMarshalizer{},
		Hasher:                   &sha256.Sha256{},
		AddressPubkeyConverter:   &mock.PubkeyConverterMock{},
		ValidatorPubkeyConverter: &mock.PubkeyConverterMock{},
		NodesCoordinator:         &mock.NodesCoordinatorMock{},
		EpochStartNotifier:       &mock.EpochStartNotifierStub{},
	})
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	// Generate transaction and hashes
	numTransactions := 10
	dataSize := 1000
	signers := []uint64{395, 207, 16, 99, 358, 292, 258, 362, 161, 247, 1, 137, 91, 309, 30, 92, 166, 361, 158, 301, 218, 80, 108, 392, 153, 343, 110, 133, 351, 316, 5, 305, 248, 123, 327, 322, 97, 86, 215, 212, 289, 250, 229, 13, 237, 20, 269, 37, 243, 29, 236, 155, 338, 257, 375, 142, 129, 93, 234, 195, 377, 311, 170}
	for i := 0; i < 100; i++ {
		txs, hashes := generateTransactions(numTransactions, dataSize)

		header := &dataBlock.Header{
			Nonce: uint64(i),
		}
		txsPool := make(map[string]data.TransactionHandler)
		for j := 0; j < numTransactions; j++ {
			txsPool[hashes[j]] = &txs[j]
		}

		miniblock := &dataBlock.MiniBlock{
			TxHashes: make([][]byte, numTransactions),
			Type:     dataBlock.TxBlock,
		}
		for j := 0; j < numTransactions; j++ {
			miniblock.TxHashes[j] = []byte(hashes[j])
		}

		body := &dataBlock.Body{
			MiniBlocks: []*dataBlock.MiniBlock{
				miniblock,
			},
		}
		body.MiniBlocks[0].ReceiverShardID = 2
		body.MiniBlocks[0].SenderShardID = 1

		go dataIndexer.SaveBlock(body, header, txsPool, signers, []string{"aaaaa", "bbbb"})
	}

	for dataIndexer.GetQueueLength() != 0 {
		time.Sleep(time.Second)
	}
}

func generateTransactions(numTxs int, datFieldSize int) ([]transaction.Transaction, []string) {
	txs := make([]transaction.Transaction, numTxs)
	hashes := make([]string, numTxs)

	randomByteArray := make([]byte, datFieldSize)
	_, _ = rand.Read(randomByteArray)

	for i := 0; i < numTxs; i++ {
		txs[i] = transaction.Transaction{
			Nonce:     uint64(i),
			Value:     big.NewInt(int64(i)),
			RcvAddr:   []byte("443e79a8d99ba093262c1db48c58ab3d59bcfeb313ca5cddf2a9d1d06f9894ec"),
			SndAddr:   []byte("443e79a8d99ba093262c1db48c58ab3d59bcfeb313ca5cddf2a9d1d06f9894ec"),
			GasPrice:  200000000000,
			GasLimit:  20000,
			Data:      randomByteArray,
			Signature: []byte("443e79a8d99ba093262c1db48c58ab3d59bcfeb313ca5cddf2a9d1d06f9894ec"),
		}
		hashes[i] = fmt.Sprintf("%v", time.Now())
	}

	return txs, hashes
}
