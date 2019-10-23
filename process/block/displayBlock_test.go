package block

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func createGenesisBlock(shardId uint32) *block.Header {
	rootHash := []byte("roothash")
	return &block.Header{
		Nonce:           0,
		Round:           0,
		Signature:       rootHash,
		RandSeed:        rootHash,
		PrevRandSeed:    rootHash,
		ShardId:         shardId,
		PubKeysBitmap:   rootHash,
		RootHash:        rootHash,
		PrevHash:        rootHash,
		MetaBlockHashes: [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
	}
}

func TestDisplayBlock_GetNumTxFromPool_NilDataPoolReturnZero(t *testing.T) {
	t.Parallel()

	storer := &mock.ChainStorerMock{}
	marshalizer := &mock.MarshalizerMock{}
	transactionCounter := NewTransactionCounter(storer, marshalizer)
	numTxs := transactionCounter.getNumTxsFromPool(0, nil, 1)

	assert.Equal(t, 0, numTxs)
}

func TestDisplayBlock_DisplayMetaHashesIncluded(t *testing.T) {
	t.Parallel()

	shardLines := make([]*display.LineData, 0)
	header := createGenesisBlock(0)
	storer := &mock.ChainStorerMock{}
	marshalizer := &mock.MarshalizerMock{}
	transactionCounter := NewTransactionCounter(storer, marshalizer)
	lines := transactionCounter.displayMetaHashesIncluded(
		shardLines,
		header,
	)

	assert.NotNil(t, lines)
	assert.Equal(t, len(header.MetaBlockHashes), len(lines))
}

func TestDisplayBlock_DisplayTxBlockBody(t *testing.T) {
	t.Parallel()

	shardLines := make([]*display.LineData, 0)
	body := make(block.Body, 0)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
	}
	body = append(body, &miniblock)
	storer := &mock.ChainStorerMock{}
	marshalizer := &mock.MarshalizerMock{}
	transactionCounter := NewTransactionCounter(storer, marshalizer)
	lines := transactionCounter.displayTxBlockBody(
		shardLines,
		body,
	)

	assert.NotNil(t, lines)
	assert.Equal(t, len(miniblock.TxHashes), len(lines))
}
