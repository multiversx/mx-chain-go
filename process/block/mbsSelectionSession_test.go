package block

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestNewMiniBlocksSelectionSession(t *testing.T) {
	marshaller := &testscommon.MarshallerStub{}
	hasher := &testscommon.HasherStub{}

	t.Run("should return error for nil marshaller", func(t *testing.T) {
		session, err := newMiniBlocksSelectionSession(1, nil, hasher)
		require.Nil(t, session)
		require.Equal(t, process.ErrNilMarshalizer, err)
	})

	t.Run("should return error for nil hasher", func(t *testing.T) {
		session, err := newMiniBlocksSelectionSession(1, marshaller, nil)
		require.Nil(t, session)
		require.Equal(t, process.ErrNilHasher, err)
	})

	t.Run("should create session successfully", func(t *testing.T) {
		session, err := newMiniBlocksSelectionSession(1, marshaller, hasher)
		require.NotNil(t, session)
		require.NoError(t, err)
	})
}

func TestResetSelectionSession(t *testing.T) {
	session := createDummyFilledSession()
	session.ResetSelectionSession()

	require.Empty(t, session.GetMiniBlocks())
	require.Empty(t, session.GetMiniBlockHeaderHandlers())
	require.Empty(t, session.GetMiniBlockHashes())
	require.Empty(t, session.GetReferencedMetaBlockHashes())
	require.Empty(t, session.GetReferencedMetaBlocks())
	require.Nil(t, session.GetLastMetaBlock())
	require.Equal(t, uint64(0), session.GetGasProvided())
	require.Equal(t, uint32(0), session.GetNumTxsAdded())
}

func TestGetters(t *testing.T) {
	session := createDummyFilledSession()

	require.Len(t, session.GetMiniBlockHeaderHandlers(), 1)
	require.Len(t, session.GetMiniBlocks(), 1)
	require.Len(t, session.GetMiniBlockHashes(), 1)
	require.Len(t, session.GetReferencedMetaBlockHashes(), 1)
	require.Len(t, session.GetReferencedMetaBlocks(), 1)
	require.NotNil(t, session.GetLastMetaBlock())
	require.Equal(t, uint64(100), session.GetGasProvided())
	require.Equal(t, uint32(2), session.GetNumTxsAdded())
}

func TestAddMiniBlocksAndHashes(t *testing.T) {
	marshaller := &testscommon.MarshallerStub{}
	hasher := &testscommon.HasherStub{}
	t.Run("should add mini blocks and hashes successfully", func(t *testing.T) {
		session, _ := newMiniBlocksSelectionSession(1, marshaller, hasher)
		miniBlock := &block.MiniBlock{}
		miniBlockHash := []byte("hash")
		err := session.AddMiniBlocksAndHashes([]block.MiniblockAndHash{
			{Miniblock: miniBlock, Hash: miniBlockHash},
		})

		require.NoError(t, err)
		require.Len(t, session.GetMiniBlocks(), 1)
		require.Len(t, session.GetMiniBlockHashes(), 1)
		require.Equal(t, miniBlock, session.GetMiniBlocks()[0])
		require.Equal(t, miniBlockHash, session.GetMiniBlockHashes()[0])
	})
}

func TestCreateAndAddMiniBlockFromTransactions(t *testing.T) {
	tx1Hash := []byte("tx1")
	tx2Hash := []byte("tx2")
	marshaller := &testscommon.MarshallerStub{}
	hasher := &testscommon.HasherStub{}
	t.Run("should create and add mini block from transactions successfully", func(t *testing.T) {
		session, _ := newMiniBlocksSelectionSession(1, marshaller, hasher)
		txHashes := [][]byte{tx1Hash, tx2Hash}
		err := session.CreateAndAddMiniBlockFromTransactions(txHashes)

		require.NoError(t, err)
		require.Len(t, session.GetMiniBlocks(), 1)
		require.Len(t, session.GetMiniBlockHashes(), 1)
	})

	t.Run("should not add mini block for empty transactions", func(t *testing.T) {
		session, _ := newMiniBlocksSelectionSession(1, marshaller, hasher)
		err := session.CreateAndAddMiniBlockFromTransactions(nil)

		require.NoError(t, err)
		require.Empty(t, session.GetMiniBlocks())
		require.Empty(t, session.GetMiniBlockHashes())
	})
}

func createDummyFilledSession() *miniBlocksSelectionSession {
	marshaller := &testscommon.MarshallerStub{}
	hasher := &testscommon.HasherStub{}
	session, _ := newMiniBlocksSelectionSession(1, marshaller, hasher)

	miniBlock := &block.MiniBlock{TxHashes: [][]byte{[]byte("tx1"), []byte("tx2")}}
	miniBlockHash := []byte("dummyHash")
	miniBlockHeader := &block.MiniBlockHeader{
		Hash:            miniBlockHash,
		SenderShardID:   1,
		ReceiverShardID: 1,
		TxCount:         uint32(len(miniBlock.TxHashes)),
		Type:            block.TxBlock,
	}
	metaBlock := &block.MetaBlock{Epoch: 10, Round: 1000}
	metaBlockHash := []byte("metaHash")

	// Add dummy mini blocks and hashes
	session.miniBlocks = append(session.miniBlocks, miniBlock)
	session.miniBlockHashes = append(session.miniBlockHashes, miniBlockHash)
	session.miniBlockHeaderHandlers = append(session.miniBlockHeaderHandlers, miniBlockHeader)
	session.referencedMetaBlockHashes = append(session.referencedMetaBlockHashes, metaBlockHash)
	session.referencedMetaBlocks = append(session.referencedMetaBlocks, metaBlock)
	session.lastMetaBlock = metaBlock
	session.gasProvided = 100
	session.numTxsAdded = 2

	return session
}
