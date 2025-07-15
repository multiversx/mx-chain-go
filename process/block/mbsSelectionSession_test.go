package block

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestNewMiniBlocksSelectionSession(t *testing.T) {
	marshaller := &testscommon.MarshallerStub{}
	hasher := &testscommon.HasherStub{}

	t.Run("nil marshaller should return error", func(t *testing.T) {
		session, err := newMiniBlocksSelectionSession(1, nil, hasher)
		require.Nil(t, session)
		require.Equal(t, process.ErrNilMarshalizer, err)
	})

	t.Run("nil hasher should return error", func(t *testing.T) {
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

func TestMiniBlockSelectionSession_ResetSelectionSession(t *testing.T) {
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

func TestMiniBlockSelectionSession_Getters(t *testing.T) {
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

func TestMiniBlockSelectionSession_AddMiniBlocksAndHashes(t *testing.T) {
	marshaller := &testscommon.MarshallerStub{}
	hasher := &testscommon.HasherStub{}
	t.Run("should not add empty mini blocks and hashes", func(t *testing.T) {
		session, _ := newMiniBlocksSelectionSession(1, marshaller, hasher)
		err := session.AddMiniBlocksAndHashes(nil)

		require.NoError(t, err)
		require.Empty(t, session.GetMiniBlocks())
		require.Empty(t, session.GetMiniBlockHashes())
	})
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

func Test_setProcessingTypeAndConstructionStateForProposalMb(t *testing.T) {
	t.Run("error on setProcessingType should error", func(t *testing.T) {
		mbHeaderHandler := &block.MiniBlockHeader{
			Reserved: []byte("invalid should error"),
		}
		err := setProcessingTypeAndConstructionStateForProposalMb(mbHeaderHandler)
		require.Error(t, err)
	})
}

func TestMiniBlockSelectionSession_CreateAndAddMiniBlockFromTransactions(t *testing.T) {
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
	t.Run("should not add mini block for empty transactions slice", func(t *testing.T) {
		session, _ := newMiniBlocksSelectionSession(1, marshaller, hasher)
		err := session.CreateAndAddMiniBlockFromTransactions([][]byte{})

		require.NoError(t, err)
		require.Empty(t, session.GetMiniBlocks())
		require.Empty(t, session.GetMiniBlockHashes())
	})

	t.Run("marshalling error should return error", func(t *testing.T) {
		expectedError := fmt.Errorf("marshalling error")
		marshaller := &testscommon.MarshallerStub{
			MarshalCalled: func(_ interface{}) ([]byte, error) {
				return nil, expectedError
			},
		}
		session, _ := newMiniBlocksSelectionSession(1, marshaller, hasher)
		err := session.CreateAndAddMiniBlockFromTransactions([][]byte{tx1Hash, tx2Hash})

		require.Equal(t, expectedError, err)
		require.Empty(t, session.GetMiniBlocks())
		require.Empty(t, session.GetMiniBlockHashes())
	})
}

func TestMiniBlocksSelectionSession_AddReferencedMetaBlock(t *testing.T) {
	marshaller := &testscommon.MarshallerStub{}
	hasher := &testscommon.HasherStub{}
	t.Run("should add referenced meta block successfully", func(t *testing.T) {
		session, _ := newMiniBlocksSelectionSession(1, marshaller, hasher)
		metaBlock := &block.MetaBlock{Epoch: 1, Round: 1}
		metaBlockHash := []byte("metaHash")
		session.AddReferencedMetaBlock(metaBlock, metaBlockHash)

		require.Len(t, session.GetReferencedMetaBlocks(), 1)
		require.Len(t, session.GetReferencedMetaBlockHashes(), 1)
		require.Equal(t, metaBlock, session.GetReferencedMetaBlocks()[0])
		require.Equal(t, metaBlockHash, session.GetReferencedMetaBlockHashes()[0])
	})
	t.Run("should not add nil meta block", func(t *testing.T) {
		session, _ := newMiniBlocksSelectionSession(1, marshaller, hasher)
		session.AddReferencedMetaBlock(nil, []byte("metaHash"))

		require.Empty(t, session.GetReferencedMetaBlocks())
		require.Empty(t, session.GetReferencedMetaBlockHashes())
	})
	t.Run("should not add empty meta block hash", func(t *testing.T) {
		session, _ := newMiniBlocksSelectionSession(1, marshaller, hasher)
		metaBlock := &block.MetaBlock{Epoch: 1, Round: 1}
		session.AddReferencedMetaBlock(metaBlock, nil)

		require.Empty(t, session.GetReferencedMetaBlocks())
		require.Empty(t, session.GetReferencedMetaBlockHashes())
	})
}

func TestMiniBlocksSelectionSession_IsInterfaceNil(t *testing.T) {
	t.Run("should return true if session is nil", func(t *testing.T) {
		var session *miniBlocksSelectionSession
		require.True(t, session.IsInterfaceNil())
	})

	t.Run("should return false if session is not nil", func(t *testing.T) {
		session := createDummyFilledSession()
		require.False(t, session.IsInterfaceNil())
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
