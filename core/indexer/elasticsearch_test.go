package indexer_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/indexer"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

func buildBlock(transactions []transaction.Transaction) (block.Body, *block.Header) {
	body := block.Body{}
	header := &block.Header{
		TxCount: 4000,
	}

	for i := 0; i < 10; i++ {
		mb := &block.MiniBlock{
			SenderShardID: 0,
			ReceiverShardID: 1,
			TxHashes: make([][]byte, 0),
		}
		for j := 0; j < 400; j++ {
			index := i * 400 + j
			m, _ := marshal.JsonMarshalizer{}.Marshal(transactions[index])
			mb.TxHashes = append(mb.TxHashes, sha256.Sha256{}.Compute(string(m)))
		}
		body = append(body, mb)
	}

	return body, header
}

func TestShit(t *testing.T) {

	log := logger.DefaultLogger()
	es, _ := indexer.NewElasticIndexer("http://localhost:9200", &marshal.JsonMarshalizer{}, &sha256.Sha256{}, log)

	transactions := make([]transaction.Transaction, 4000)
	for i := 0; i < 4000; i++ {
		transactions[i] = transaction.Transaction{
			Nonce: uint64(i),
			Value: big.NewInt(100),
			RcvAddr: []byte("receiver"),
			SndAddr: []byte("sender"),
			Signature: []byte("motherfuckingsignature"),
		}
	}

	getAllCounter := 0
	cs := &mock.ChainStorerMock{
		GetAllCalled: func(unitType dataRetriever.UnitType, keys [][]byte) (b map[string][]byte, e error) {
			res := make(map[string][]byte, 0)
			startingPont := getAllCounter * 400
			keyCounter := 0
			for i := startingPont; i < startingPont + 400; i++ {
				m, _ := marshal.JsonMarshalizer{}.Marshal(transactions[i])
				res[string(keys[keyCounter])] = m
			}

			getAllCounter++
			return res, nil
		},
	}

	body, header := buildBlock(transactions)

	es.SaveBlock(body, header, cs)
}
