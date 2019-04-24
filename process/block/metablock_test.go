package block_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	blproc "github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewMetaProcessor

func TestNewMetaProcessor_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, err := blproc.NewMetaProcessor(
		nil,
		tdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()
	be, err := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		nil,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilForkDetectorShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		tdp,
		nil,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	assert.Equal(t, process.ErrNilForkDetector, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		tdp,
		&mock.ForkDetectorMock{},
		nil,
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilMarshalizerShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		tdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		nil,
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilChainStorerShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		tdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		nil,
		func(shardID uint32, hdrHash []byte) {},
	)
	assert.Equal(t, process.ErrNilStorage, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilrequestHeaderHandlerShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		tdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		nil,
	)
	assert.Equal(t, process.ErrNilRequestHeaderHandler, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	mp, err := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		tdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	assert.Nil(t, err)
	assert.NotNil(t, mp)
}

////------- ProcessBlock
//
//func TestMetaProcessor_ProcessBlockWithNilBlockchainShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	blk := make(block.Body, 0)
//	err := mp.ProcessBlock(nil, &block.Header{}, blk, haveTime)
//	assert.Equal(t, process.ErrNilBlockChain, err)
//}
//
//func TestMetaProcessor_ProcessBlockWithNilHeaderShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	body := make(block.Body, 0)
//	err := mp.ProcessBlock(&blockchain.BlockChain{}, nil, body, haveTime)
//	assert.Equal(t, process.ErrNilBlockHeader, err)
//}
//
//func TestMetaProcessor_ProcessBlockWithNilBlockBodyShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	err := mp.ProcessBlock(&blockchain.BlockChain{}, &block.Header{}, nil, haveTime)
//	assert.Equal(t, process.ErrNilMiniBlocks, err)
//}
//
//func TestMetaProcessor_ProcessBlockWithNilHaveTimeFuncShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	blk := make(block.Body, 0)
//	err := mp.ProcessBlock(&blockchain.BlockChain{}, &block.Header{}, blk, nil)
//	assert.Equal(t, process.ErrNilHaveTimeHandler, err)
//}
//
//func TestMetaProcessor_ProcessWithDirtyAccountShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	// set accounts dirty
//	journalLen := func() int { return 3 }
//	revToSnapshot := func(snapshot int) error { return nil }
//	blkc := &blockchain.BlockChain{}
//	hdr := block.Header{
//		Nonce:         1,
//		PubKeysBitmap: []byte("0100101"),
//		PrevHash:      []byte(""),
//		Signature:     []byte("signature"),
//		RootHash:      []byte("roothash"),
//	}
//	body := make(block.Body, 0)
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{
//			JournalLenCalled:       journalLen,
//			RevertToSnapshotCalled: revToSnapshot,
//		},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	// should return err
//	err := mp.ProcessBlock(blkc, &hdr, body, haveTime)
//	assert.NotNil(t, err)
//	assert.Equal(t, err, process.ErrAccountStateDirty)
//}
//
//func TestMetaProcessor_ProcessBlockWithInvalidTransactionShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	txHash := []byte("tx_hash1")
//	blkc := &blockchain.BlockChain{}
//	hdr := block.Header{
//		Nonce:         1,
//		PrevHash:      []byte(""),
//		Signature:     []byte("signature"),
//		PubKeysBitmap: []byte("00110"),
//		ShardId:       0,
//		RootHash:      []byte("rootHash"),
//	}
//	body := make(block.Body, 0)
//	txHashes := make([][]byte, 0)
//	txHashes = append(txHashes, txHash)
//	miniblock := block.MiniBlock{
//		ReceiverShardID: 0,
//		SenderShardID:   0,
//		TxHashes:        txHashes,
//	}
//	body = append(body, &miniblock)
//	// set accounts not dirty
//	journalLen := func() int { return 0 }
//	revertToSnapshot := func(snapshot int) error { return nil }
//	rootHashCalled := func() []byte {
//		return []byte("rootHash")
//	}
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{
//			JournalLenCalled:       journalLen,
//			RevertToSnapshotCalled: revertToSnapshot,
//			RootHashCalled:         rootHashCalled,
//		},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	go func() {
//		mp.ChRcvAllHdrs <- true
//	}()
//	// should return err
//	err := mp.ProcessBlock(blkc, &hdr, body, haveTime)
//	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
//}
//
//func TestMetaProcessor_ProcessWithHeaderNotFirstShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	hdr := &block.Header{
//		Nonce:         0,
//		Round:         1,
//		PubKeysBitmap: []byte("0100101"),
//		PrevHash:      []byte(""),
//		Signature:     []byte("signature"),
//		RootHash:      []byte("root hash"),
//	}
//	body := make(block.Body, 0)
//	blkc := &blockchain.BlockChain{}
//	err := mp.ProcessBlock(blkc, hdr, body, haveTime)
//	assert.Equal(t, process.ErrWrongNonceInBlock, err)
//}
//
//func TestMetaProcessor_ProcessWithHeaderNotCorrectNonceShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	hdr := &block.Header{
//		Nonce:         0,
//		Round:         1,
//		PubKeysBitmap: []byte("0100101"),
//		PrevHash:      []byte(""),
//		Signature:     []byte("signature"),
//		RootHash:      []byte("root hash"),
//	}
//	body := make(block.Body, 0)
//	blkc := &blockchain.BlockChain{}
//	err := mp.ProcessBlock(blkc, hdr, body, haveTime)
//	assert.Equal(t, process.ErrWrongNonceInBlock, err)
//}
//
//func TestMetaProcessor_ProcessWithHeaderNotCorrectPrevHashShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	hdr := &block.Header{
//		Nonce:         1,
//		Round:         1,
//		PubKeysBitmap: []byte("0100101"),
//		PrevHash:      []byte("zzz"),
//		Signature:     []byte("signature"),
//		RootHash:      []byte("root hash"),
//	}
//	body := make(block.Body, 0)
//	blkc := &blockchain.BlockChain{
//		CurrentBlockHeader: &block.Header{
//			Nonce: 0,
//		},
//	}
//	err := mp.ProcessBlock(blkc, hdr, body, haveTime)
//	assert.Equal(t, process.ErrInvalidBlockHash, err)
//}
//
//func TestMetaProcessor_ProcessBlockWithErrOnProcessBlockTransactionsCallShouldRevertState(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	txHash := []byte("tx_hash1")
//	err := errors.New("process block transaction error")
//	blkc := &blockchain.BlockChain{
//		CurrentBlockHeader: &block.Header{
//			Nonce: 0,
//		},
//	}
//	hdr := block.Header{
//		Nonce:         1,
//		PrevHash:      []byte(""),
//		Signature:     []byte("signature"),
//		PubKeysBitmap: []byte("00110"),
//		ShardId:       0,
//		RootHash:      []byte("rootHash"),
//	}
//	body := make(block.Body, 0)
//	txHashes := make([][]byte, 0)
//	txHashes = append(txHashes, txHash)
//	miniblock := block.MiniBlock{
//		ReceiverShardID: 0,
//		SenderShardID:   0,
//		TxHashes:        txHashes,
//	}
//	body = append(body, &miniblock)
//	// set accounts not dirty
//	journalLen := func() int { return 0 }
//	wasCalled := false
//	revertToSnapshot := func(snapshot int) error {
//		wasCalled = true
//		return nil
//	}
//	rootHashCalled := func() []byte {
//		return []byte("rootHash")
//	}
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{
//			JournalLenCalled:       journalLen,
//			RevertToSnapshotCalled: revertToSnapshot,
//			RootHashCalled:         rootHashCalled,
//		},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	go func() {
//		mp.ChRcvAllHdrs <- true
//	}()
//	// should return err
//	err2 := mp.ProcessBlock(blkc, &hdr, body, haveTime)
//	assert.Equal(t, err, err2)
//	assert.True(t, wasCalled)
//}
//
//func TestMetaProcessor_ProcessBlockWithErrOnVerifyStateRootCallShouldRevertState(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	txHash := []byte("tx_hash1")
//	blkc := &blockchain.BlockChain{
//		CurrentBlockHeader: &block.Header{
//			Nonce: 0,
//		},
//	}
//	hdr := block.Header{
//		Nonce:         1,
//		PrevHash:      []byte(""),
//		Signature:     []byte("signature"),
//		PubKeysBitmap: []byte("00110"),
//		ShardId:       0,
//		RootHash:      []byte("rootHash"),
//	}
//	body := make(block.Body, 0)
//	txHashes := make([][]byte, 0)
//	txHashes = append(txHashes, txHash)
//	miniblock := block.MiniBlock{
//		ReceiverShardID: 0,
//		SenderShardID:   0,
//		TxHashes:        txHashes,
//	}
//	body = append(body, &miniblock)
//	// set accounts not dirty
//	journalLen := func() int { return 0 }
//	wasCalled := false
//	revertToSnapshot := func(snapshot int) error {
//		wasCalled = true
//		return nil
//	}
//	rootHashCalled := func() []byte {
//		return []byte("rootHashX")
//	}
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{
//			JournalLenCalled:       journalLen,
//			RevertToSnapshotCalled: revertToSnapshot,
//			RootHashCalled:         rootHashCalled,
//		},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	go func() {
//		mp.ChRcvAllHdrs <- true
//	}()
//	// should return err
//	err := mp.ProcessBlock(blkc, &hdr, body, haveTime)
//	assert.Equal(t, process.ErrRootStateMissmatch, err)
//	assert.True(t, wasCalled)
//}
//
////------- CommitBlock
//
//func TestMetaProcessor_CommitBlockNilBlockchainShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	accounts := &mock.AccountsStub{}
//	accounts.RevertToSnapshotCalled = func(snapshot int) error {
//		return nil
//	}
//	mp, _ := blproc.NewMetaProcessor(
//		accounts,
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	blk := make(block.Body, 0)
//	err := mp.CommitBlock(nil, &block.Header{}, blk)
//	assert.Equal(t, process.ErrNilBlockChain, err)
//}
//
//func TestMetaProcessor_CommitBlockMarshalizerFailForHeaderShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	rootHash := []byte("root hash to be tested")
//	accounts := &mock.AccountsStub{
//		RootHashCalled: func() []byte {
//			return rootHash
//		},
//		RevertToSnapshotCalled: func(snapshot int) error {
//			return nil
//		},
//	}
//	errMarshalizer := errors.New("failure")
//	hdr := &block.Header{
//		Nonce:         1,
//		Round:         1,
//		PubKeysBitmap: []byte("0100101"),
//		PrevHash:      []byte("zzz"),
//		Signature:     []byte("signature"),
//		RootHash:      rootHash,
//	}
//	body := make(block.Body, 0)
//	marshalizer := &mock.MarshalizerStub{
//		MarshalCalled: func(obj interface{}) (i []byte, e error) {
//			if reflect.DeepEqual(obj, hdr) {
//				return nil, errMarshalizer
//			}
//
//			return []byte("obj"), nil
//		},
//	}
//	mp, _ := blproc.NewMetaProcessor(
//		accounts,
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		marshalizer,
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	blkc := createTestBlockchain()
//	err := mp.CommitBlock(blkc, hdr, body)
//	assert.Equal(t, errMarshalizer, err)
//}
//
//func TestMetaProcessor_CommitBlockStorageFailsForHeaderShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	errPersister := errors.New("failure")
//	rootHash := []byte("root hash to be tested")
//	accounts := &mock.AccountsStub{
//		RootHashCalled: func() []byte {
//			return rootHash
//		},
//		RevertToSnapshotCalled: func(snapshot int) error {
//			return nil
//		},
//	}
//	hdr := &block.Header{
//		Nonce:         1,
//		Round:         1,
//		PubKeysBitmap: []byte("0100101"),
//		PrevHash:      []byte("zzz"),
//		Signature:     []byte("signature"),
//		RootHash:      rootHash,
//	}
//	body := make(block.Body, 0)
//	hdrUnit := &mock.StorerStub{
//		PutCalled: func(key, data []byte) error {
//			return errPersister
//		},
//	}
//	store := initStore()
//	store.AddStorer(dataRetriever.BlockHeaderUnit, hdrUnit)
//
//	mp, _ := blproc.NewMetaProcessor(
//		accounts,
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		store,
//		func(shardID uint32, hdrHash []byte) {},
//	)
//
//	blkc, _ := blockchain.NewBlockChain(
//		generateTestCache(),
//	)
//	err := mp.CommitBlock(blkc, hdr, body)
//	assert.Equal(t, errPersister, err)
//}
//
//func TestMetaProcessor_CommitBlockStorageFailsForBodyShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	errPersister := errors.New("failure")
//	rootHash := []byte("root hash to be tested")
//	accounts := &mock.AccountsStub{
//		RootHashCalled: func() []byte {
//			return rootHash
//		},
//		CommitCalled: func() (i []byte, e error) {
//			return nil, nil
//		},
//		RevertToSnapshotCalled: func(snapshot int) error {
//			return nil
//		},
//	}
//	hdr := &block.Header{
//		Nonce:         1,
//		Round:         1,
//		PubKeysBitmap: []byte("0100101"),
//		PrevHash:      []byte("zzz"),
//		Signature:     []byte("signature"),
//		RootHash:      rootHash,
//	}
//	mb := block.MiniBlock{}
//	body := make(block.Body, 0)
//	body = append(body, &mb)
//
//	miniBlockUnit := &mock.StorerStub{
//		PutCalled: func(key, data []byte) error {
//			return errPersister
//		},
//	}
//	store := initStore()
//	store.AddStorer(dataRetriever.MiniBlockUnit, miniBlockUnit)
//
//	mp, _ := blproc.NewMetaProcessor(
//		accounts,
//		tdp,
//		&mock.ForkDetectorMock{
//			AddHeaderCalled: func(header data.HeaderHandler, hash []byte, isProcessed bool) error {
//				return nil
//			},
//		},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		store,
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	blkc, _ := blockchain.NewBlockChain(
//		generateTestCache(),
//	)
//	err := mp.CommitBlock(blkc, hdr, body)
//
//	assert.Equal(t, errPersister, err)
//}
//
//func TestMetaProcessor_CommitBlockNilNoncesDataPoolShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	rootHash := []byte("root hash to be tested")
//	accounts := &mock.AccountsStub{
//		RootHashCalled: func() []byte {
//			return rootHash
//		},
//		RevertToSnapshotCalled: func(snapshot int) error {
//			return nil
//		},
//	}
//	hdr := &block.Header{
//		Nonce:         1,
//		Round:         1,
//		PubKeysBitmap: []byte("0100101"),
//		PrevHash:      []byte("zzz"),
//		Signature:     []byte("signature"),
//		RootHash:      rootHash,
//	}
//	body := make(block.Body, 0)
//	store := initStore()
//
//	mp, _ := blproc.NewMetaProcessor(
//		accounts,
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		store,
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	tdp.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
//		return nil
//	}
//	blkc := createTestBlockchain()
//	err := mp.CommitBlock(blkc, hdr, body)
//
//	assert.Equal(t, process.ErrNilDataPoolHolder, err)
//}
//
//func TestMetaProcessor_CommitBlockNoTxInPoolShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	txHash := []byte("txHash")
//	rootHash := []byte("root hash")
//	hdrHash := []byte("header hash")
//	hdr := &block.Header{
//		Nonce:         1,
//		Round:         1,
//		PubKeysBitmap: []byte("0100101"),
//		PrevHash:      []byte("zzz"),
//		Signature:     []byte("signature"),
//		RootHash:      rootHash,
//	}
//	mb := block.MiniBlock{
//		TxHashes: [][]byte{[]byte(txHash)},
//	}
//	body := block.Body{&mb}
//	accounts := &mock.AccountsStub{
//		CommitCalled: func() (i []byte, e error) {
//			return rootHash, nil
//		},
//		RootHashCalled: func() []byte {
//			return rootHash
//		},
//		RevertToSnapshotCalled: func(snapshot int) error {
//			return nil
//		},
//	}
//	fd := &mock.ForkDetectorMock{
//		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, isProcessed bool) error {
//			return nil
//		},
//	}
//	hasher := &mock.HasherStub{}
//	hasher.ComputeCalled = func(s string) []byte {
//		return hdrHash
//	}
//	store := initStore()
//
//	mp, _ := blproc.NewMetaProcessor(
//		accounts,
//		tdp,
//		fd,
//		hasher,
//		&mock.MarshalizerMock{},
//		store,
//		func(shardID uint32, hdrHash []byte) {},
//	)
//
//	txCache := &mock.CacherStub{
//		PeekCalled: func(key []byte) (value interface{}, ok bool) {
//			return nil, false
//		},
//		LenCalled: func() int {
//			return 0
//		},
//	}
//	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
//		return &mock.ShardedDataStub{
//			ShardDataStoreCalled: func(id string) (c storage.Cacher) {
//				return txCache
//			},
//
//			RemoveSetOfDataFromPoolCalled: func(keys [][]byte, id string) {
//			},
//
//			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
//				if reflect.DeepEqual(key, []byte("tx1_hash")) {
//					return &transaction.Transaction{Nonce: 10}, true
//				}
//				return nil, false
//			},
//		}
//	}
//	blkc := createTestBlockchain()
//	err := mp.CommitBlock(blkc, hdr, body)
//	assert.Equal(t, process.ErrMissingTransaction, err)
//}
//
//func TestMetaProcessor_CommitBlockOkValsShouldWork(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	txHash := []byte("txHash")
//	tx := &transaction.Transaction{}
//	rootHash := []byte("root hash")
//	hdrHash := []byte("header hash")
//	hdr := &block.Header{
//		Nonce:         1,
//		Round:         1,
//		PubKeysBitmap: []byte("0100101"),
//		PrevHash:      []byte("zzz"),
//		Signature:     []byte("signature"),
//		RootHash:      rootHash,
//	}
//	mb := block.MiniBlock{
//		TxHashes: [][]byte{[]byte(txHash)},
//	}
//	body := block.Body{&mb}
//	accounts := &mock.AccountsStub{
//		CommitCalled: func() (i []byte, e error) {
//			return rootHash, nil
//		},
//		RootHashCalled: func() []byte {
//			return rootHash
//		},
//	}
//	forkDetectorAddCalled := false
//	fd := &mock.ForkDetectorMock{
//		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, isProcessed bool) error {
//			if header == hdr {
//				forkDetectorAddCalled = true
//				return nil
//			}
//
//			return errors.New("should have not got here")
//		},
//	}
//	hasher := &mock.HasherStub{}
//	hasher.ComputeCalled = func(s string) []byte {
//		return hdrHash
//	}
//	store := initStore()
//
//	mp, _ := blproc.NewMetaProcessor(
//		accounts,
//		tdp,
//		fd,
//		hasher,
//		&mock.MarshalizerMock{},
//		store,
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	txCache := &mock.CacherStub{
//		PeekCalled: func(key []byte) (value interface{}, ok bool) {
//			if bytes.Equal(txHash, key) {
//				return tx, true
//			}
//			return nil, false
//		},
//		LenCalled: func() int {
//			return 0
//		},
//	}
//	removeTxWasCalled := false
//	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
//		return &mock.ShardedDataStub{
//			ShardDataStoreCalled: func(id string) (c storage.Cacher) {
//				return txCache
//			},
//
//			RemoveSetOfDataFromPoolCalled: func(keys [][]byte, id string) {
//				if bytes.Equal(keys[0], []byte(txHash)) && len(keys) == 1 {
//					removeTxWasCalled = true
//				}
//			},
//			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
//				if reflect.DeepEqual(key, []byte(txHash)) {
//					return &transaction.Transaction{Nonce: 10}, true
//				}
//				return nil, false
//			},
//		}
//
//	}
//	blkc := createTestBlockchain()
//	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
//		return hdr
//	}
//	blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
//		return hdrHash
//	}
//	err := mp.CommitBlock(blkc, hdr, body)
//	assert.Nil(t, err)
//	assert.True(t, removeTxWasCalled)
//	assert.True(t, forkDetectorAddCalled)
//	assert.True(t, blkc.GetCurrentBlockHeader() == hdr)
//	assert.Equal(t, hdrHash, blkc.GetCurrentBlockHeaderHash())
//	//this should sleep as there is an async call to display current header and block in CommitBlock
//	time.Sleep(time.Second)
//}
//
//func TestMetaProcessor_GetTransactionFromPool(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	txHash := []byte("tx1_hash")
//	hdr := mp.GetHeaderFromPool(1, txHash)
//	assert.NotNil(t, hdr)
//	assert.Equal(t, uint64(10), hdr.Nonce)
//}
//
//func TestBlockProc_RequestTransactionFromNetwork(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	//shardId := uint32(1)
//	//txHash1 := []byte("tx1_hash1")
//	header := &block.MetaBlock{}
//	//	body := make(block.Body, 0)
//	//	txHashes := make([][]byte, 0)
//	//	txHashes = append(txHashes, txHash1)
//	//	mBlk := block.MiniBlock{ReceiverShardID: shardId, TxHashes: txHashes}
//	//	body = append(body, &mBlk)
//	//TODO refactor the test
//	//if mp.RequestBlockHeaders(body) > 0 {
//	if mp.RequestBlockHeaders(header) > 0 {
//		mp.WaitForHdrHashes(haveTime())
//	}
//}
//
//func TestBlockProc_CreateTxBlockBodyWithDirtyAccStateShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	journalLen := func() int { return 3 }
//	revToSnapshot := func(snapshot int) error { return nil }
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{
//			JournalLenCalled:       journalLen,
//			RevertToSnapshotCalled: revToSnapshot,
//		},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//
//	bl, err := mp.CreateBlockBody(0, func() bool { return true })
//	// nil block
//	assert.Nil(t, bl)
//	// error
//	assert.Equal(t, process.ErrAccountStateDirty, err)
//}
//
//func TestMetaProcessor_CreateTxBlockBodyWithNoTimeShouldEmptyBlock(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	journalLen := func() int { return 0 }
//	rootHashfunc := func() []byte { return []byte("roothash") }
//	revToSnapshot := func(snapshot int) error { return nil }
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{
//			JournalLenCalled:       journalLen,
//			RootHashCalled:         rootHashfunc,
//			RevertToSnapshotCalled: revToSnapshot,
//		},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//
//	haveTime := func() bool {
//		return false
//	}
//	bl, err := mp.CreateBlockBody(0, haveTime)
//	// no error
//	assert.Nil(t, err)
//	// no miniblocks
//	assert.Equal(t, len(bl.(block.Body)), 0)
//}
//
//func TestMetaProcessor_CreateTxBlockBodyOK(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	journalLen := func() int { return 0 }
//	rootHashfunc := func() []byte { return []byte("roothash") }
//	haveTime := func() bool {
//		return true
//	}
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{
//			JournalLenCalled: journalLen,
//			RootHashCalled:   rootHashfunc,
//		},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		&mock.ChainStorerMock{},
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	blk, err := mp.CreateBlockBody(0, haveTime)
//	assert.NotNil(t, blk)
//	assert.Nil(t, err)
//}
//
//func TestMetaProcessor_CreateGenesisBlockBodyWithFailSetBalanceShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	_, err := mp.CreateGenesisBlock(nil)
//	assert.Equal(t, process.ErrAccountStateDirty, err)
//}
//
//func TestMetaProcessor_CreateGenesisBlockBodyOK(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	rootHash, err := mp.CreateGenesisBlock(nil)
//	assert.Nil(t, err)
//	assert.NotNil(t, rootHash)
//	assert.Equal(t, rootHash, []byte("rootHash"))
//}
//
//func TestMetaProcessor_RemoveBlockTxsFromPoolNilBlockShouldErr(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	err := mp.RemoveBlockInfoFromPool(nil)
//	assert.NotNil(t, err)
//	assert.Equal(t, err, process.ErrNilTxBlockBody)
//}
//
//func TestMetaProcessor_RemoveBlockTxsFromPoolOK(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	header := &block.MetaBlock{}
//	//body := make(block.Body, 0)
//	//txHash := []byte("txHash")
//	//txHashes := make([][]byte, 0)
//	//txHashes = append(txHashes, txHash)
//	//miniblock := block.MiniBlock{
//	//	ReceiverShardID: 0,
//	//	SenderShardID:   0,
//	//	TxHashes:        txHashes,
//	//}
//	//body = append(body, &miniblock)
//	err := mp.RemoveBlockInfoFromPool(header)
//	assert.Nil(t, err)
//}
//
//func createTestMetaHdr() *block.MetaBlock {
//	hasher := mock.HasherMock{}
//	hdr := &block.MetaBlock{
//		Nonce:         1,
//		Epoch:         3,
//		Round:         4,
//		TimeStamp:     uint64(11223344),
//		PrevHash:      hasher.Compute("prev hash"),
//		PubKeysBitmap: []byte{255, 0, 128},
//		Signature:     hasher.Compute("signature"),
//		RootHash:      hasher.Compute("root hash"),
//	}
//
//	//txBlock := block.Body{
//	//	{
//	//		ReceiverShardID: 0,
//	//		SenderShardID:   0,
//	//		TxHashes: [][]byte{
//	//			hasher.Compute("txHash_0_1"),
//	//			hasher.Compute("txHash_0_2"),
//	//		},
//	//	},
//	//	{
//	//		ReceiverShardID: 1,
//	//		SenderShardID:   0,
//	//		TxHashes: [][]byte{
//	//			hasher.Compute("txHash_1_1"),
//	//			hasher.Compute("txHash_1_2"),
//	//		},
//	//	},
//	//	{
//	//		ReceiverShardID: 2,
//	//		SenderShardID:   0,
//	//		TxHashes: [][]byte{
//	//			hasher.Compute("txHash_2_1"),
//	//		},
//	//	},
//	//	{
//	//		ReceiverShardID: 3,
//	//		SenderShardID:   0,
//	//		TxHashes:        make([][]byte, 0),
//	//	},
//	//}
//	return hdr
//}
//
////------- ComputeNewNoncePrevHash
//
//func TestMetaProcessor_DisplayLogInfo(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	hasher := mock.HasherMock{}
//	hdr := createTestMetaHdr()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	hdr.PrevHash = hasher.Compute("prev hash")
//	mp.DisplayMetaBlock(hdr)
//}
//
//func TestMetaProcessor_CreateBlockHeaderShouldNotReturnNil(t *testing.T) {
//	t.Parallel()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		initDataPool(),
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	mbHeaders, err := mp.CreateBlockHeader(nil, 0, func() bool {
//		return true
//	})
//	assert.Nil(t, err)
//	assert.NotNil(t, mbHeaders)
//	assert.Equal(t, 0, len(mbHeaders.(*block.Header).MiniBlockHeaders))
//}
//
//func TestMetaProcessor_CreateBlockHeaderShouldErrWhenMarshalizerErrors(t *testing.T) {
//	t.Parallel()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		initDataPool(),
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	mbHeaders, err := mp.CreateBlockHeader(nil, 0, func() bool {
//		return true
//	})
//	assert.NotNil(t, err)
//	assert.Nil(t, mbHeaders)
//}
//
//func TestMetaProcessor_CreateBlockHeaderReturnsOK(t *testing.T) {
//	t.Parallel()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		initDataPool(),
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	header, err := mp.CreateBlockHeader(nil, 0, func() bool {
//		return true
//	})
//	assert.Nil(t, err)
//	assert.NotNil(t, header)
//
//	//mbHeaders, err := mp.CreateBlockHeader(nil, 0, func() bool {
//	//	return true
//	//})
//	//assert.Nil(t, err)
//	//assert.Equal(t, len(body), len(mbHeaders.(*block.Header).MiniBlockHeaders))
//}
//
//func TestMetaProcessor_CommitBlockShouldRevertAccountStateWhenErr(t *testing.T) {
//	t.Parallel()
//	// set accounts dirty
//	journalEntries := 3
//	revToSnapshot := func(snapshot int) error {
//		journalEntries = 0
//		return nil
//	}
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{
//			RevertToSnapshotCalled: revToSnapshot,
//		},
//		initDataPool(),
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	err := mp.CommitBlock(nil, nil, nil)
//	assert.NotNil(t, err)
//	assert.Equal(t, 0, journalEntries)
//}
//
//func TestMetaProcessor_MarshalizedDataForCrossShardShouldWork(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	txHash0 := []byte("txHash0")
//	mb0 := block.MiniBlock{
//		ReceiverShardID: 0,
//		SenderShardID:   0,
//		TxHashes:        [][]byte{[]byte(txHash0)},
//	}
//	txHash1 := []byte("txHash1")
//	mb1 := block.MiniBlock{
//		ReceiverShardID: 1,
//		SenderShardID:   0,
//		TxHashes:        [][]byte{[]byte(txHash1)},
//	}
//	body := make(block.Body, 0)
//	body = append(body, &mb0)
//	body = append(body, &mb1)
//	body = append(body, &mb0)
//	body = append(body, &mb1)
//	marshalizer := &mock.MarshalizerMock{
//		Fail: false,
//	}
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		marshalizer,
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//
//	msh, mstx, err := mp.MarshalizedDataForCrossShard(body)
//	assert.Nil(t, err)
//	assert.NotNil(t, msh)
//	assert.NotNil(t, mstx)
//	_, found := msh[0]
//	assert.False(t, found)
//
//	expectedBody := make(block.Body, 0)
//	err = marshalizer.Unmarshal(&expectedBody, msh[1])
//	assert.Nil(t, err)
//	assert.Equal(t, 2, len(expectedBody))
//	assert.Equal(t, &mb1, expectedBody[0])
//	assert.Equal(t, &mb1, expectedBody[1])
//}
//
//func TestMetaProcessor_MarshalizedDataWrongType(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	marshalizer := &mock.MarshalizerMock{
//		Fail: false,
//	}
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		marshalizer,
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//
//	wr := wrongBody{}
//	msh, mstx, err := mp.MarshalizedDataForCrossShard(wr)
//	assert.Equal(t, process.ErrWrongTypeAssertion, err)
//	assert.Nil(t, msh)
//	assert.Nil(t, mstx)
//}
//
//func TestMetaProcessor_MarshalizedDataNilInput(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	marshalizer := &mock.MarshalizerMock{
//		Fail: false,
//	}
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		marshalizer,
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	msh, mstx, err := mp.MarshalizedDataForCrossShard(nil)
//	assert.Equal(t, process.ErrNilMiniBlocks, err)
//	assert.Nil(t, msh)
//	assert.Nil(t, mstx)
//}
//
//func TestMetaProcessor_MarshalizedDataMarshalWithoutSuccess(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	txHash0 := []byte("txHash0")
//	mb0 := block.MiniBlock{
//		ReceiverShardID: 1,
//		SenderShardID:   0,
//		TxHashes:        [][]byte{[]byte(txHash0)},
//	}
//	body := make(block.Body, 0)
//	body = append(body, &mb0)
//	marshalizer := &mock.MarshalizerStub{
//		MarshalCalled: func(obj interface{}) ([]byte, error) {
//			return nil, process.ErrMarshalWithoutSuccess
//		},
//	}
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		marshalizer,
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	msh, mstx, err := mp.MarshalizedDataForCrossShard(body)
//	assert.Equal(t, process.ErrMarshalWithoutSuccess, err)
//	assert.Nil(t, msh)
//	assert.Nil(t, mstx)
//}
//
////------- receivedTransaction
//
//func TestMetaProcessor_ReceivedTransactionShouldEraseRequested(t *testing.T) {
//	t.Parallel()
//
//	hasher := mock.HasherMock{}
//	marshalizer := &mock.MarshalizerMock{}
//	dataPool := mock.NewPoolsHolderFake()
//
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		dataPool,
//		&mock.ForkDetectorMock{},
//		hasher,
//		marshalizer,
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//
//	//add 3 tx hashes on requested list
//	hdrHash1 := []byte("hdr hash 1")
//	hdrHash2 := []byte("hdr hash 2")
//	hdrHash3 := []byte("hdr hash 3")
//
//	mp.AddHdrHashToRequestedList(hdrHash1)
//	mp.AddHdrHashToRequestedList(hdrHash2)
//	mp.AddHdrHashToRequestedList(hdrHash3)
//
//	//received txHash2
//	mp.ReceivedHeader(hdrHash2)
//
//	assert.True(t, mp.IsHdrHashRequested(hdrHash1))
//	assert.False(t, mp.IsHdrHashRequested(hdrHash2))
//	assert.True(t, mp.IsHdrHashRequested(hdrHash3))
//}
//
////------- createMiniBlocks
//
//func TestMetaProcessor_CreateMiniBlocksShouldWorkWithIntraShardTxs(t *testing.T) {
//	t.Parallel()
//
//	hasher := mock.HasherMock{}
//	marshalizer := &mock.MarshalizerMock{}
//	dataPool := mock.NewPoolsHolderFake()
//
//	//we will have a 3 txs in pool
//
//	txHash1 := []byte("tx hash 1")
//	txHash2 := []byte("tx hash 2")
//	txHash3 := []byte("tx hash 3")
//
//	senderShardId := uint32(0)
//	receiverShardId := uint32(0)
//
//	tx1Nonce := uint64(45)
//	tx2Nonce := uint64(46)
//	tx3Nonce := uint64(47)
//
//	//put the existing tx inside datapool
//	cacheId := process.ShardCacherIdentifier(senderShardId, receiverShardId)
//	dataPool.Transactions().AddData(txHash1, &transaction.Transaction{
//		Nonce: tx1Nonce,
//		Data:  txHash1,
//	}, cacheId)
//	dataPool.Transactions().AddData(txHash2, &transaction.Transaction{
//		Nonce: tx2Nonce,
//		Data:  txHash2,
//	}, cacheId)
//	dataPool.Transactions().AddData(txHash3, &transaction.Transaction{
//		Nonce: tx3Nonce,
//		Data:  txHash3,
//	}, cacheId)
//
//	tx1ExecutionResult := uint64(0)
//	tx2ExecutionResult := uint64(0)
//	tx3ExecutionResult := uint64(0)
//
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{
//			RevertToSnapshotCalled: func(snapshot int) error {
//				assert.Fail(t, "revert should have not been called")
//				return nil
//			},
//			JournalLenCalled: func() int {
//				return 0
//			},
//		},
//		dataPool,
//		&mock.ForkDetectorMock{},
//		hasher,
//		marshalizer,
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//
//	shardInfo, err := mp.CreateShardInfo(256, 0, func() bool {
//		return true
//	})
//
//	assert.Nil(t, err)
//	assert.NotNil(t, shardInfo)
//	//testing execution
//	assert.Equal(t, tx1Nonce, tx1ExecutionResult)
//	assert.Equal(t, tx2Nonce, tx2ExecutionResult)
//	assert.Equal(t, tx3Nonce, tx3ExecutionResult)
//	//one miniblock output
//	//assert.Equal(t, 1, len(blockBody))
//	//miniblock should have 3 txs
//	//assert.Equal(t, 3, len(blockBody[0].TxHashes))
//	//testing all 3 hashes are present in block body
//	//assert.True(t, isInTxHashes(txHash1, blockBody[0].TxHashes))
//	//assert.True(t, isInTxHashes(txHash2, blockBody[0].TxHashes))
//	//assert.True(t, isInTxHashes(txHash3, blockBody[0].TxHashes))
//}
//
//func TestMetaProcessor_RestoreBlockIntoPoolsShouldErrNilTxBlockBody(t *testing.T) {
//	t.Parallel()
//	tdp := initDataPool()
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		tdp,
//		&mock.ForkDetectorMock{},
//		&mock.HasherStub{},
//		&mock.MarshalizerMock{},
//		initStore(),
//		func(shardID uint32, hdrHash []byte) {},
//	)
//	err := mp.RestoreBlockIntoPools(nil, nil)
//	assert.NotNil(t, err)
//	assert.Equal(t, err, process.ErrNilTxBlockBody)
//}
//
//func TestMetaProcessor_RestoreBlockIntoPoolsShouldWork(t *testing.T) {
//	t.Parallel()
//	dataPool := mock.NewPoolsHolderFake()
//	marshalizerMock := &mock.MarshalizerMock{}
//	hasherMock := &mock.HasherStub{}
//
//	body := make(block.Body, 0)
//	tx := transaction.Transaction{Nonce: 1}
//	buffTx, _ := marshalizerMock.Marshal(tx)
//	txHash := []byte("tx hash 1")
//
//	store := &mock.ChainStorerMock{
//		GetAllCalled: func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
//			m := make(map[string][]byte, 0)
//			m[string(txHash)] = buffTx
//			return m, nil
//		},
//	}
//
//	mp, _ := blproc.NewMetaProcessor(
//		&mock.AccountsStub{},
//		dataPool,
//		&mock.ForkDetectorMock{},
//		hasherMock,
//		marshalizerMock,
//		store,
//		func(shardID uint32, hdrHash []byte) {},
//	)
//
//	txHashes := make([][]byte, 0)
//	txHashes = append(txHashes, txHash)
//	miniblock := block.MiniBlock{
//		ReceiverShardID: 0,
//		SenderShardID:   1,
//		TxHashes:        txHashes,
//	}
//	body = append(body, &miniblock)
//
//	miniblockHash := []byte("mini block hash 1")
//	hasherMock.ComputeCalled = func(s string) []byte {
//		return miniblockHash
//	}
//
//	metablockHash := []byte("meta block hash 1")
//	metablockHeader := createDummyMetaBlock(0, 1, miniblockHash)
//	metablockHeader.SetMiniBlockProcessed(metablockHash, true)
//	dataPool.MetaBlocks().Put(
//		metablockHash,
//		metablockHeader,
//	)
//
//	err := mp.RestoreBlockIntoPools(nil, body)
//
//	miniblockFromPool, _ := dataPool.MiniBlocks().Get(miniblockHash)
//	txFromPool, _ := dataPool.Transactions().SearchFirstData(txHash)
//	metablockFromPool, _ := dataPool.MetaBlocks().Get(metablockHash)
//	metablock := metablockFromPool.(*block.MetaBlock)
//	assert.Nil(t, err)
//	assert.Equal(t, &miniblock, miniblockFromPool)
//	assert.Equal(t, &tx, txFromPool)
//	assert.Equal(t, false, metablock.GetMiniBlockProcessed(miniblockHash))
//}
