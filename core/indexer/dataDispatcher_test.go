package indexer

import (
	"context"
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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
)

func TestDataDispatcher(t *testing.T) {
	indexTemplates := make(map[string]io.Reader)
	indexPolicies  := make(map[string]io.Reader)
	transactionsTemplate, _ := core.OpenFile("./testdata/transactions.json")
	indexTemplates["transactions"] = transactionsTemplate
	transactionsPolicy, _ := core.OpenFile("./testdata/transactions_policy.json")
	indexPolicies["transactions_policy"] = transactionsPolicy

	dispatcher, err := NewDataDispatcher(ElasticIndexerArgs{
		Url: "https://search-corcotest-cbtcxoexobkqz4guopvhb64yta.us-east-1.es.amazonaws.com",
		IndexTemplates: indexTemplates,
		IndexPolicies: indexPolicies,
		Options: &Options{
			TxIndexingEnabled: true,
		},
		Marshalizer: &marshal.JsonMarshalizer{},
		Hasher:      &sha256.Sha256{},
		AddressPubkeyConverter: &mock.PubkeyConverterMock{},
	})
	if err != nil {
		fmt.Println(err)
	}

	// Generate transaction and hashes
	numTransactions := 1
	dataSize := 1000
	for i := 0; i < 1000; i++ {
		txs, hashes := generateTransactions(numTransactions, dataSize)

		header := &dataBlock.Header{}
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

		dispatcher.SaveBlock(body, header, txsPool, nil, nil)
	}
}

func TestKinesis(t *testing.T) {
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	}))
	f := firehose.New(sess)
	params := &firehose.ListDeliveryStreamsInput{}
	_, err := f.ListDeliveryStreamsWithContext(ctx, params, func(r *request.Request) {
		r.Handlers.Validate.RemoveByName("core.ValidateParametersHandler")
	})
	if err != nil {
		t.Errorf("expect no error, got %v", err)
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

