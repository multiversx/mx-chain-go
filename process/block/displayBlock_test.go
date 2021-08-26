package block

import (
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/display"
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
		ShardID:         shardId,
		PubKeysBitmap:   rootHash,
		RootHash:        rootHash,
		PrevHash:        rootHash,
		MetaBlockHashes: [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
	}
}

func TestDisplayBlock_NewTransactionCounterShouldErrWhenHasherIsNil(t *testing.T) {
	t.Parallel()

	txCounter, err := NewTransactionCounter(nil, &mock.MarshalizerMock{})

	assert.Nil(t, txCounter)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestDisplayBlock_NewTransactionCounterShouldErrWhenMarshalizerIsNil(t *testing.T) {
	t.Parallel()

	txCounter, err := NewTransactionCounter(&mock.HasherMock{}, nil)

	assert.Nil(t, txCounter)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestDisplayBlock_NewTransactionCounterShouldWork(t *testing.T) {
	t.Parallel()

	txCounter, err := NewTransactionCounter(&mock.HasherMock{}, &mock.MarshalizerMock{})

	assert.NotNil(t, txCounter)
	assert.Nil(t, err)
}

func TestDisplayBlock_DisplayMetaHashesIncluded(t *testing.T) {
	t.Parallel()

	shardLines := make([]*display.LineData, 0)
	header := createGenesisBlock(0)
	txCounter, _ := NewTransactionCounter(&mock.HasherMock{}, &mock.MarshalizerMock{})
	lines := txCounter.displayMetaHashesIncluded(
		shardLines,
		header,
	)

	assert.NotNil(t, lines)
	assert.Equal(t, len(header.MetaBlockHashes), len(lines))
}

func TestDisplayBlock_DisplayTxBlockBody(t *testing.T) {
	t.Parallel()

	shardLines := make([]*display.LineData, 0)
	body := &block.Body{}
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)
	txCounter, _ := NewTransactionCounter(&mock.HasherMock{}, &mock.MarshalizerMock{})
	lines := txCounter.displayTxBlockBody(
		shardLines,
		&block.Header{},
		body,
	)

	assert.NotNil(t, lines)
	assert.Equal(t, len(miniblock.TxHashes), len(lines))
}
