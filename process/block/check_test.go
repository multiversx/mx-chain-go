package block_test

import (
	"testing"

	bl "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

func createBlockProcessor() process.BlockProcessor {
	txHandler := func(transactionHandler func(txHash []byte)) {}

	tp := &mock.TransactionPoolMock{
		RegisterTransactionHandlerCalled: txHandler,
	}

	bp := block.NewBlockProcessor(
		tp,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		nil,
		2,
	)

	return bp
}

// StateBlockBody
func TestStateBlockBodyWrapper_IntegrityNilProcessorShouldFail(t *testing.T) {
	t.Parallel()

	sbWrapper := block.StateBlockBodyWrapper{
		StateBlockBody: &bl.StateBlockBody{},
	}

	res := sbWrapper.Integrity(nil)
	assert.Equal(t, process.ErrNilProcessor, res)
}

func TestStateBlockBodyWrapper_IntegrityInvalidShardIdShouldFail(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()
	bp.SetNoShards(10)

	sbWrapper := block.StateBlockBodyWrapper{
		StateBlockBody: &bl.StateBlockBody{
			RootHash: []byte("rootHash1"),
			ShardID:  55,
		},
	}

	res := sbWrapper.Integrity(bp)
	assert.Equal(t, process.ErrInvalidShardId, res)
}

func TestStateBlockBodyWrapper_IntegrityNilRootHashShouldFail(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	sbWrapper := block.StateBlockBodyWrapper{
		StateBlockBody: &bl.StateBlockBody{
			RootHash: nil,
			ShardID:  1,
		},
	}

	res := sbWrapper.Integrity(bp)
	assert.Equal(t, process.ErrNilRootHash, res)
}

func TestStateBlockBodyWrapper_IntegrityOK(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	sbWrapper := block.StateBlockBodyWrapper{
		StateBlockBody: &bl.StateBlockBody{
			RootHash: []byte("rootHash1"),
			ShardID:  1,
		},
	}

	res := sbWrapper.Integrity(bp)
	assert.Nil(t, res)
}

func TestStateBlockBodyWrapper_CheckCorruptedBlock(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()
	bp.SetNoShards(2)

	sbWrapper := block.StateBlockBodyWrapper{
		StateBlockBody: &bl.StateBlockBody{
			RootHash: []byte("rootHash1"),
			ShardID:  55,
		},
	}

	res := sbWrapper.Check(bp)
	assert.Equal(t, process.ErrInvalidShardId, res)
}

func TestStateBlockBodyWrapper_CheckWithEmptyAccountsTrieShouldFail(t *testing.T) {
	t.Parallel()

	txHandler := func(transactionHandler func(txHash []byte)) {}

	tp := &mock.TransactionPoolMock{
		RegisterTransactionHandlerCalled: txHandler,
	}

	rootHashStoredTrie := func() []byte {
		return nil
	}

	bp := block.NewBlockProcessor(
		tp,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{RootHashCalled: rootHashStoredTrie},
		2,
	)

	sbWrapper := block.StateBlockBodyWrapper{
		StateBlockBody: &bl.StateBlockBody{
			RootHash: []byte("rootHash1"),
			ShardID:  0,
		},
	}

	res := sbWrapper.Check(bp)
	assert.Equal(t, process.ErrNilRootHash, res)
}

func TestStateBlockBodyWrapper_CheckWithInvalidRootHashShouldFail(t *testing.T) {
	t.Parallel()

	txHandler := func(transactionHandler func(txHash []byte)) {}

	tp := &mock.TransactionPoolMock{
		RegisterTransactionHandlerCalled: txHandler,
	}

	rootHashStoredTrie := func() []byte {
		return []byte("correctRootHash")
	}

	bp := block.NewBlockProcessor(
		tp,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{RootHashCalled: rootHashStoredTrie},
		2,
	)

	sbWrapper := block.StateBlockBodyWrapper{
		StateBlockBody: &bl.StateBlockBody{
			RootHash: []byte("invalidRootHash"),
			ShardID:  0,
		},
	}

	res := sbWrapper.Check(bp)
	assert.Equal(t, process.ErrInvalidRootHash, res)
}

func TestStateBlockBodyWrapper_CheckOK(t *testing.T) {
	t.Parallel()

	txHandler := func(transactionHandler func(txHash []byte)) {}

	tp := &mock.TransactionPoolMock{
		RegisterTransactionHandlerCalled: txHandler,
	}

	rootHashStoredTrie := func() []byte {
		return []byte("correctRootHash")
	}

	bp := block.NewBlockProcessor(
		tp,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{RootHashCalled: rootHashStoredTrie},
		2,
	)

	sbWrapper := block.StateBlockBodyWrapper{
		StateBlockBody: &bl.StateBlockBody{
			RootHash: []byte("correctRootHash"),
			ShardID:  0,
		},
	}

	res := sbWrapper.Check(bp)
	assert.Nil(t, res)
}

// TxBlockBody
func TestTxBlockBodyWrapper_IntegrityInvalidStateBlockShouldFail(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	txbWrapper := block.TxBlockBodyWrapper{
		TxBlockBody: &bl.TxBlockBody{
			StateBlockBody: bl.StateBlockBody{},
		},
	}

	res := txbWrapper.Integrity(bp)
	assert.Equal(t, process.ErrNilRootHash, res)
}

func TestTxBlockBodyWrapper_IntegrityNilMiniblocksShouldFail(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	txbWrapper := block.TxBlockBodyWrapper{
		TxBlockBody: &bl.TxBlockBody{
			StateBlockBody: bl.StateBlockBody{
				RootHash: []byte("rootHash1"),
				ShardID:  1,
			},
			MiniBlocks: nil,
		},
	}

	res := txbWrapper.Integrity(bp)
	assert.Equal(t, process.ErrNilMiniBlocks, res)

}

func TestTxBlockBodyWrapper_IntegrityNilTxHashesShouldFail(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()
	miniBlocks := make([]bl.MiniBlock, 0)
	miniBlock := bl.MiniBlock{
		ShardID:  0,
		TxHashes: nil,
	}

	miniBlocks = append(miniBlocks, miniBlock)

	txbWrapper := block.TxBlockBodyWrapper{
		TxBlockBody: &bl.TxBlockBody{
			StateBlockBody: bl.StateBlockBody{
				RootHash: []byte("rootHash1"),
				ShardID:  1,
			},
			MiniBlocks: miniBlocks,
		},
	}

	res := txbWrapper.Integrity(bp)
	assert.Equal(t, process.ErrNilTxHashes, res)
}

func TestTxBlockBodyWrapper_IntegrityNilTxHashShouldFail(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()
	miniBlocks := make([]bl.MiniBlock, 0)
	txHashes := make([][]byte, 0)
	txHash := []byte(nil)
	txHashes = append(txHashes, txHash)

	miniBlock := bl.MiniBlock{
		ShardID:  0,
		TxHashes: txHashes,
	}

	miniBlocks = append(miniBlocks, miniBlock)

	txbWrapper := block.TxBlockBodyWrapper{
		TxBlockBody: &bl.TxBlockBody{
			StateBlockBody: bl.StateBlockBody{
				RootHash: []byte("rootHash1"),
				ShardID:  1,
			},
			MiniBlocks: miniBlocks,
		},
	}

	res := txbWrapper.Integrity(bp)
	assert.Equal(t, process.ErrNilTxHash, res)
}

func TestTxBlockBodyWrapper_IntegrityOK(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()
	miniBlocks := make([]bl.MiniBlock, 0)
	txHashes := make([][]byte, 0)
	txHash1 := []byte("goodHash1")
	txHash2 := []byte("goodHash2")

	txHashes = append(txHashes, txHash1)
	txHashes = append(txHashes, txHash2)

	miniBlock := bl.MiniBlock{
		ShardID:  0,
		TxHashes: txHashes,
	}

	miniBlocks = append(miniBlocks, miniBlock)

	txbWrapper := block.TxBlockBodyWrapper{
		TxBlockBody: &bl.TxBlockBody{
			StateBlockBody: bl.StateBlockBody{
				RootHash: []byte("rootHash1"),
				ShardID:  1,
			},
			MiniBlocks: miniBlocks,
		},
	}

	res := txbWrapper.Integrity(bp)
	assert.Nil(t, res)
}

func TestTxBlockBodyWrapper_CheckCorruptedBlock(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	txbWrapper := block.TxBlockBodyWrapper{
		TxBlockBody: &bl.TxBlockBody{
			StateBlockBody: bl.StateBlockBody{
				ShardID:  0,
				RootHash: nil,
			},
		},
	}

	res := txbWrapper.Check(bp)
	assert.Equal(t, process.ErrNilRootHash, res)
}

func TestTxBlockBodyWrapper_CheckOK(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()
	miniBlocks := make([]bl.MiniBlock, 0)
	txHashes := make([][]byte, 0)
	txHash1 := []byte("goodHash1")
	txHash2 := []byte("goodHash2")

	txHashes = append(txHashes, txHash1)
	txHashes = append(txHashes, txHash2)

	miniBlock := bl.MiniBlock{
		ShardID:  0,
		TxHashes: txHashes,
	}

	miniBlocks = append(miniBlocks, miniBlock)

	txbWrapper := block.TxBlockBodyWrapper{
		TxBlockBody: &bl.TxBlockBody{
			StateBlockBody: bl.StateBlockBody{
				RootHash: []byte("rootHash1"),
				ShardID:  1,
			},
			MiniBlocks: miniBlocks,
		},
	}

	res := txbWrapper.Check(bp)
	assert.Nil(t, res)
}

// PeerBlockBodyWrapper
func TestPeerBlockBodyWrapper_IntegrityInvalidStateBlockShouldFail(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	peerBlkWrapper := block.PeerBlockBodyWrapper{
		PeerBlockBody: &bl.PeerBlockBody{
			StateBlockBody: bl.StateBlockBody{},
		},
	}

	res := peerBlkWrapper.Integrity(bp)
	assert.Equal(t, process.ErrNilRootHash, res)
}

func TestPeerBlockBodyWrapper_IntegrityNilChangesShouldFail(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()
	peerBlkWrapper := block.PeerBlockBodyWrapper{
		PeerBlockBody: &bl.PeerBlockBody{
			StateBlockBody: bl.StateBlockBody{
				RootHash: []byte("rootHash1"),
				ShardID:  1,
			},
			Changes: nil,
		},
	}

	res := peerBlkWrapper.Integrity(bp)
	assert.Equal(t, process.ErrNilPeerChanges, res)
}

func TestPeerBlockBodyWrapper_IntegrityChangeInvalidShardShouldFail(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	changes := make([]bl.PeerChange, 0)
	change := bl.PeerChange{
		PubKey:      []byte("pubkey1"),
		ShardIdDest: 3,
	}
	changes = append(changes, change)

	peerBlkWrapper := block.PeerBlockBodyWrapper{
		PeerBlockBody: &bl.PeerBlockBody{
			StateBlockBody: bl.StateBlockBody{
				RootHash: []byte("rootHash1"),
				ShardID:  1,
			},
			Changes: changes,
		},
	}

	res := peerBlkWrapper.Integrity(bp)
	assert.Equal(t, process.ErrInvalidShardId, res)
}

func TestPeerBlockBodyWrapper_IntegrityChangeNilPubkeyShouldFail(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	changes := make([]bl.PeerChange, 0)
	change := bl.PeerChange{
		PubKey:      nil,
		ShardIdDest: 0,
	}
	changes = append(changes, change)

	peerBlkWrapper := block.PeerBlockBodyWrapper{
		PeerBlockBody: &bl.PeerBlockBody{
			StateBlockBody: bl.StateBlockBody{
				RootHash: []byte("rootHash1"),
				ShardID:  1,
			},
			Changes: changes,
		},
	}

	res := peerBlkWrapper.Integrity(bp)
	assert.Equal(t, process.ErrNilPublicKey, res)
}

func TestPeerBlockBodyWrapper_IntegrityOK(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	changes := make([]bl.PeerChange, 0)
	change := bl.PeerChange{
		PubKey:      []byte("pubkey1"),
		ShardIdDest: 0,
	}
	changes = append(changes, change)

	peerBlkWrapper := block.PeerBlockBodyWrapper{
		PeerBlockBody: &bl.PeerBlockBody{
			StateBlockBody: bl.StateBlockBody{
				RootHash: []byte("rootHash1"),
				ShardID:  1,
			},
			Changes: changes,
		},
	}

	res := peerBlkWrapper.Integrity(bp)
	assert.Nil(t, res)
}

func TestPeerBlockBodyWrapper_CheckCorruptedBlock(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()
	peerBlkWrapper := block.PeerBlockBodyWrapper{
		PeerBlockBody: &bl.PeerBlockBody{
			StateBlockBody: bl.StateBlockBody{},
		},
	}

	res := peerBlkWrapper.Check(bp)
	assert.Equal(t, process.ErrNilRootHash, res)
}

func TestPeerBlockBodyWrapper_CheckOK(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	changes := make([]bl.PeerChange, 0)
	change := bl.PeerChange{
		PubKey:      []byte("pubkey1"),
		ShardIdDest: 0,
	}
	changes = append(changes, change)

	peerBlkWrapper := block.PeerBlockBodyWrapper{
		PeerBlockBody: &bl.PeerBlockBody{
			StateBlockBody: bl.StateBlockBody{
				RootHash: []byte("rootHash1"),
				ShardID:  1,
			},
			Changes: changes,
		},
	}

	res := peerBlkWrapper.Check(bp)
	assert.Nil(t, res)
}

// Header
func TestHeaderWrapper_IntegrityNilProcessorShouldFail(t *testing.T) {
	t.Parallel()

	headerWrapper := block.HeaderWrapper{
	}

	res := headerWrapper.Integrity(nil)
	assert.Equal(t, process.ErrNilProcessor, res)
}

func TestHeaderWrapper_IntegrityNilBlockBodyHashShouldFail(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	headerWrapper := block.HeaderWrapper{
		Header: &bl.Header{
			BlockBodyHash: nil,
		},
	}

	res := headerWrapper.Integrity(bp)
	assert.Equal(t, process.ErrNilBlockBodyHash, res)
}

func TestHeaderWrapper_IntegrityNilPubKeyBmpShouldFail(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	headerWrapper := block.HeaderWrapper{
		Header: &bl.Header{
			BlockBodyHash: []byte("blockBodyHash"),
			PubKeysBitmap: nil,
		},
	}

	res := headerWrapper.Integrity(bp)
	assert.Equal(t, process.ErrNilPubKeysBitmap, res)
}

func TestHeaderWrapper_IntegrityInvalidShardIdShouldFail(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	headerWrapper := block.HeaderWrapper{
		Header: &bl.Header{
			BlockBodyHash: []byte("blockBodyHash"),
			PubKeysBitmap: []byte("010010"),
			ShardId:       3,
		},
	}

	res := headerWrapper.Integrity(bp)
	assert.Equal(t, process.ErrInvalidShardId, res)
}

func TestHeaderWrapper_IntegrityNilPrevHashShouldFail(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	headerWrapper := block.HeaderWrapper{
		Header: &bl.Header{
			BlockBodyHash: []byte("blockBodyHash"),
			PubKeysBitmap: []byte("010010"),
			ShardId:       0,
			PrevHash:      nil,
		},
	}

	res := headerWrapper.Integrity(bp)
	assert.Equal(t, process.ErrNilPreviousBlockHash, res)
}

func TestHeaderWrapper_IntegrityNilSignatureShouldFail(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	headerWrapper := block.HeaderWrapper{
		Header: &bl.Header{
			BlockBodyHash: []byte("blockBodyHash"),
			PubKeysBitmap: []byte("010010"),
			ShardId:       0,
			PrevHash:      []byte("prevHash"),
			Signature:     nil,
		},
	}

	res := headerWrapper.Integrity(bp)
	assert.Equal(t, process.ErrNilSignature, res)
}

func TestHeaderWrapper_IntegrityOK(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	headerWrapper := block.HeaderWrapper{
		Header: &bl.Header{
			BlockBodyHash: []byte("blockBodyHash"),
			PubKeysBitmap: []byte("010010"),
			ShardId:       0,
			PrevHash:      []byte("prevHash"),
			Signature:     []byte("signature"),
		},
	}

	res := headerWrapper.Integrity(bp)
	assert.Nil(t, res)
}

func TestHeaderWrapper_CheckCorruptedHeader(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	headerWrapper := block.HeaderWrapper{
		Header: &bl.Header{
			BlockBodyHash: nil,
			PubKeysBitmap: []byte("010010"),
			ShardId:       0,
			PrevHash:      []byte("prevHash"),
			Signature:     []byte("signature"),},
	}

	res := headerWrapper.Check(bp)
	assert.Equal(t, process.ErrNilBlockBodyHash, res)
}

func TestHeaderWrapper_CheckOK(t *testing.T) {
	t.Parallel()

	bp := createBlockProcessor()

	headerWrapper := block.HeaderWrapper{
		Header: &bl.Header{
			BlockBodyHash: []byte("blockBodyHash"),
			PubKeysBitmap: []byte("010010"),
			ShardId:       0,
			PrevHash:      []byte("prevHash"),
			Signature:     []byte("signature"),
		},
	}

	res := headerWrapper.Check(bp)
	assert.Nil(t, res)
}

func TestHeaderWrapper_VerifySigNilHeaderShouldFail(t *testing.T) {
	t.Parallel()

	headerWrapper := block.HeaderWrapper{
		Header: nil,
	}

	res := headerWrapper.VerifySig()
	assert.Equal(t, process.ErrNilBlockHeader, res)
}

func TestHeaderWrapper_VerifySigNilSigShouldFail(t *testing.T) {
	t.Parallel()

	headerWrapper := block.HeaderWrapper{
		Header: &bl.Header{
			Signature: nil,
		},
	}

	res := headerWrapper.VerifySig()
	assert.Equal(t, process.ErrNilSignature, res)
}

func TestHeaderWrapper_VerifySigOk(t *testing.T) {
	t.Parallel()

	headerWrapper := block.HeaderWrapper{
		Header: &bl.Header{
			Signature: []byte("signature"),
		},
	}

	res := headerWrapper.VerifySig()
	assert.Nil(t, res)

	// TODO: verify the signature
}
