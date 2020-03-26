package block

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/display"
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

func TestDisplayBlock_DisplayMetaHashesIncluded(t *testing.T) {
	t.Parallel()

	shardLines := make([]*display.LineData, 0)
	header := createGenesisBlock(0)
	txCounter := NewTransactionCounter()
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
	txCounter := NewTransactionCounter()
	lines := txCounter.displayTxBlockBody(
		shardLines,
		body,
	)

	assert.NotNil(t, lines)
	assert.Equal(t, len(miniblock.TxHashes), len(lines))
}
