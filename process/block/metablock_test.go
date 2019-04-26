package block_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	blproc "github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/stretchr/testify/assert"
)

func createMetaBlockHeader() *block.MetaBlock {
	hdr := block.MetaBlock{
		Nonce:         1,
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
		PubKeysBitmap: []byte("pubKeysBitmap"),
		RootHash:      []byte("rootHash"),
		ShardInfo:     make([]block.ShardData, 0),
		PeerInfo:      make([]block.PeerData, 0),
		TxCount:       1,
		PrevRandSeed:  make([]byte, 0),
		RandSeed:      make([]byte, 0),
	}

	shardMiniBlockHeaders := make([]block.ShardMiniBlockHeader, 0)
	shardMiniBlockHeader := block.ShardMiniBlockHeader{
		Hash:            []byte("mb_hash1"),
		ReceiverShardId: 0,
		SenderShardId:   0,
		TxCount:         1,
	}
	shardMiniBlockHeaders = append(shardMiniBlockHeaders, shardMiniBlockHeader)
	shardData := block.ShardData{
		ShardId:               0,
		HeaderHash:            []byte("hdr_hash1"),
		TxCount:               1,
		ShardMiniBlockHeaders: shardMiniBlockHeaders,
	}
	hdr.ShardInfo = append(hdr.ShardInfo, shardData)

	peerData := block.PeerData{
		PublicKey: []byte("public_key1"),
	}
	hdr.PeerInfo = append(hdr.PeerInfo, peerData)

	return &hdr
}

//------- NewMetaProcessor

func TestNewMetaProcessor_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	be, err := blproc.NewMetaProcessor(
		nil,
		mdp,
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
	mdp := initMetaDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
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
	mdp := initMetaDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
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
	mdp := initMetaDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
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
	mdp := initMetaDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
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
	mdp := initMetaDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
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
	mdp := initMetaDataPool()
	mp, err := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	assert.Nil(t, err)
	assert.NotNil(t, mp)
}

//------- ProcessBlock

func TestMetaProcessor_ProcessBlockWithNilBlockchainShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	blk := &block.MetaBlockBody{}
	err := mp.ProcessBlock(nil, &block.MetaBlock{}, blk, haveTime)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestMetaProcessor_ProcessBlockWithNilHeaderShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	blk := &block.MetaBlockBody{}
	err := mp.ProcessBlock(&blockchain.MetaChain{}, nil, blk, haveTime)
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestMetaProcessor_ProcessBlockWithNilBlockBodyShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	err := mp.ProcessBlock(&blockchain.MetaChain{}, &block.MetaBlock{}, nil, haveTime)
	assert.Equal(t, process.ErrNilMiniBlocks, err)
}

func TestMetaProcessor_ProcessBlockWithNilHaveTimeFuncShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	blk := &block.MetaBlockBody{}
	err := mp.ProcessBlock(&blockchain.MetaChain{}, &block.MetaBlock{}, blk, nil)
	assert.Equal(t, process.ErrNilHaveTimeHandler, err)
}

func TestMetaProcessor_ProcessWithDirtyAccountShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	// set accounts dirty
	journalLen := func() int { return 3 }
	revToSnapshot := func(snapshot int) error { return nil }
	blkc := &blockchain.MetaChain{}
	hdr := block.MetaBlock{
		Nonce:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
		RootHash:      []byte("roothash"),
	}
	body := &block.MetaBlockBody{}
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revToSnapshot,
		},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	// should return err
	err := mp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrAccountStateDirty)
}

func TestMetaProcessor_ProcessWithHeaderNotFirstShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)

	blkc := &blockchain.MetaChain{}
	hdr := &block.MetaBlock{
		Nonce: 2,
	}
	body := &block.MetaBlockBody{}
	err := mp.ProcessBlock(blkc, hdr, body, haveTime)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)
}

func TestMetaProcessor_ProcessWithHeaderNotCorrectNonceShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	blkc := &blockchain.MetaChain{
		CurrentBlock: &block.MetaBlock{
			Nonce: 1,
		},
	}
	hdr := &block.MetaBlock{
		Nonce: 3,
	}
	body := &block.MetaBlockBody{}
	err := mp.ProcessBlock(blkc, hdr, body, haveTime)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)
}

func TestMetaProcessor_ProcessWithHeaderNotCorrectPrevHashShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	blkc := &blockchain.MetaChain{
		CurrentBlock: &block.MetaBlock{
			Nonce: 1,
		},
	}
	hdr := &block.MetaBlock{
		Nonce:    2,
		PrevHash: []byte("X"),
	}

	body := &block.MetaBlockBody{}
	err := mp.ProcessBlock(blkc, hdr, body, haveTime)
	assert.Equal(t, process.ErrInvalidBlockHash, err)
}

func TestMetaProcessor_ProcessBlockWithErrOnVerifyStateRootCallShouldRevertState(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	blkc := &blockchain.MetaChain{
		CurrentBlock: &block.MetaBlock{
			Nonce: 0,
		},
	}
	hdr := createMetaBlockHeader()
	body := &block.MetaBlockBody{}
	// set accounts not dirty
	journalLen := func() int { return 0 }
	wasCalled := false
	revertToSnapshot := func(snapshot int) error {
		wasCalled = true
		return nil
	}
	rootHashCalled := func() []byte {
		return []byte("rootHashX")
	}
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revertToSnapshot,
			RootHashCalled:         rootHashCalled,
		},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	go func() {
		mp.ChRcvAllHdrs <- true
	}()
	// should return err
	err := mp.ProcessBlock(blkc, hdr, body, haveTime)
	assert.Equal(t, process.ErrRootStateMissmatch, err)
	assert.True(t, wasCalled)
}

//------- CommitBlock

func TestMetaProcessor_CommitBlockNilBlockchainShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	accounts := &mock.AccountsStub{}
	accounts.RevertToSnapshotCalled = func(snapshot int) error {
		return nil
	}
	mp, _ := blproc.NewMetaProcessor(
		accounts,
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	blk := &block.MetaBlockBody{}
	err := mp.CommitBlock(nil, &block.MetaBlock{}, blk)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestMetaProcessor_CommitBlockMarshalizerFailForHeaderShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	accounts := &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	errMarshalizer := errors.New("failure")
	hdr := createMetaBlockHeader()
	body := &block.MetaBlockBody{}
	marshalizer := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (i []byte, e error) {
			if reflect.DeepEqual(obj, hdr) {
				return nil, errMarshalizer
			}

			return []byte("obj"), nil
		},
	}
	mp, _ := blproc.NewMetaProcessor(
		accounts,
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		marshalizer,
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	blkc := createTestBlockchain()
	err := mp.CommitBlock(blkc, hdr, body)
	assert.Equal(t, errMarshalizer, err)
}

func TestMetaProcessor_CommitBlockStorageFailsForHeaderShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	errPersister := errors.New("failure")
	accounts := &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	hdr := createMetaBlockHeader()
	body := &block.MetaBlockBody{}
	hdrUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return errPersister
		},
	}
	store := initStore()
	store.AddStorer(dataRetriever.MetaBlockUnit, hdrUnit)

	mp, _ := blproc.NewMetaProcessor(
		accounts,
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		store,
		func(shardID uint32, hdrHash []byte) {},
	)

	blkc, _ := blockchain.NewMetaChain(
		generateTestCache(),
	)
	err := mp.CommitBlock(blkc, hdr, body)
	assert.Equal(t, errPersister, err)
}

func TestMetaProcessor_CommitBlockStorageFailsForShardDataShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	errPersister := errors.New("failure")
	accounts := &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	hdr := createMetaBlockHeader()
	body := &block.MetaBlockBody{}

	shardDataUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return errPersister
		},
	}
	store := initStore()
	store.AddStorer(dataRetriever.MetaShardDataUnit, shardDataUnit)

	mp, _ := blproc.NewMetaProcessor(
		accounts,
		mdp,
		&mock.ForkDetectorMock{
			AddHeaderCalled: func(header data.HeaderHandler, hash []byte, isProcessed bool) error {
				return nil
			},
		},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		store,
		func(shardID uint32, hdrHash []byte) {},
	)
	blkc, _ := blockchain.NewMetaChain(
		generateTestCache(),
	)
	err := mp.CommitBlock(blkc, hdr, body)

	assert.Equal(t, errPersister, err)
}

func TestMetaProcessor_CommitBlockStorageFailsForPeerDataShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	errPersister := errors.New("failure")
	accounts := &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	hdr := createMetaBlockHeader()
	body := &block.MetaBlockBody{}

	shardDataUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return nil
		},
	}
	peerDataUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return errPersister
		},
	}
	store := initStore()
	store.AddStorer(dataRetriever.MetaShardDataUnit, shardDataUnit)
	store.AddStorer(dataRetriever.MetaPeerDataUnit, peerDataUnit)

	mp, _ := blproc.NewMetaProcessor(
		accounts,
		mdp,
		&mock.ForkDetectorMock{
			AddHeaderCalled: func(header data.HeaderHandler, hash []byte, isProcessed bool) error {
				return nil
			},
		},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		store,
		func(shardID uint32, hdrHash []byte) {},
	)
	blkc, _ := blockchain.NewMetaChain(
		generateTestCache(),
	)
	err := mp.CommitBlock(blkc, hdr, body)

	assert.Equal(t, errPersister, err)
}

func TestMetaProcessor_CommitBlockNilNoncesDataPoolShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	accounts := &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	hdr := createMetaBlockHeader()
	body := &block.MetaBlockBody{}
	shardDataUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return nil
		},
	}
	peerDataUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return nil
		},
	}
	store := initStore()
	store.AddStorer(dataRetriever.MetaShardDataUnit, shardDataUnit)
	store.AddStorer(dataRetriever.MetaPeerDataUnit, peerDataUnit)

	mp, _ := blproc.NewMetaProcessor(
		accounts,
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		store,
		func(shardID uint32, hdrHash []byte) {},
	)
	mdp.MetaBlockNoncesCalled = func() dataRetriever.Uint64Cacher {
		return nil
	}
	blkc := createTestBlockchain()
	err := mp.CommitBlock(blkc, hdr, body)

	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestMetaProcessor_CommitBlockNoTxInPoolShouldErr(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	hdr := createMetaBlockHeader()
	body := &block.MetaBlockBody{}
	accounts := &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	fd := &mock.ForkDetectorMock{}
	hasher := &mock.HasherStub{}
	shardDataUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return nil
		},
	}
	peerDataUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return nil
		},
	}
	store := initStore()
	store.AddStorer(dataRetriever.MetaShardDataUnit, shardDataUnit)
	store.AddStorer(dataRetriever.MetaPeerDataUnit, peerDataUnit)

	mp, _ := blproc.NewMetaProcessor(
		accounts,
		mdp,
		fd,
		hasher,
		&mock.MarshalizerMock{},
		store,
		func(shardID uint32, hdrHash []byte) {},
	)

	mdp.ShardHeadersCalled = func() storage.Cacher {
		return &mock.CacherStub{
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}
	}

	blkc := createTestBlockchain()
	err := mp.CommitBlock(blkc, hdr, body)
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestMetaProcessor_CommitBlockOkValsShouldWork(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	rootHash := []byte("rootHash")
	hdr := createMetaBlockHeader()
	body := &block.MetaBlockBody{}
	accounts := &mock.AccountsStub{
		CommitCalled: func() (i []byte, e error) {
			return rootHash, nil
		},
		RootHashCalled: func() []byte {
			return rootHash
		},
	}
	forkDetectorAddCalled := false
	fd := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, isProcessed bool) error {
			if header == hdr {
				forkDetectorAddCalled = true
				return nil
			}

			return errors.New("should have not got here")
		},
	}
	hasher := &mock.HasherStub{}
	blockHeaderUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return nil
		},
	}
	shardDataUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return nil
		},
	}
	peerDataUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return nil
		},
	}
	store := initStore()
	store.AddStorer(dataRetriever.BlockHeaderUnit, blockHeaderUnit)
	store.AddStorer(dataRetriever.MetaShardDataUnit, shardDataUnit)
	store.AddStorer(dataRetriever.MetaPeerDataUnit, peerDataUnit)

	mp, _ := blproc.NewMetaProcessor(
		accounts,
		mdp,
		fd,
		hasher,
		&mock.MarshalizerMock{},
		store,
		func(shardID uint32, hdrHash []byte) {},
	)

	removeHdrWasCalled := false
	mdp.ShardHeadersCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			return &block.Header{}, true
		}
		cs.RemoveCalled = func(key []byte) {
			if bytes.Equal(key, []byte("hdr_hash1")) {
				removeHdrWasCalled = true
			}
		}
		cs.LenCalled = func() int {
			return 0
		}
		return cs
	}

	blkc := createTestBlockchain()

	err := mp.CommitBlock(blkc, hdr, body)
	assert.Nil(t, err)
	assert.True(t, removeHdrWasCalled)
	assert.True(t, forkDetectorAddCalled)
	//this should sleep as there is an async call to display current header and block in CommitBlock
	time.Sleep(time.Second)
}

func TestMetaProcessor_GetHeaderFromPool(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	hdrHash := []byte("hdr_hash1")
	hdr := mp.GetHeaderFromPool(0, hdrHash)
	assert.NotNil(t, hdr)
	assert.Equal(t, uint64(1), hdr.GetNonce())
}

func TestBlockProc_RequestTransactionFromNetwork(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	mdp.ShardHeadersCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			return nil, false
		}
		cs.RemoveCalled = func(key []byte) {
		}
		cs.LenCalled = func() int {
			return 0
		}
		return cs
	}

	header := createMetaBlockHeader()
	//TODO refactor the test
	if mp.RequestBlockHeaders(header) > 0 {
		mp.WaitForBlockHeaders(haveTime())
	}
}

func TestMetaProcessor_CreateBlockBodyShouldWork(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	haveTime := func() bool {
		return true
	}
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		func(shardID uint32, hdrHash []byte) {},
	)
	blk, err := mp.CreateBlockBody(0, haveTime)
	assert.NotNil(t, blk)
	assert.Nil(t, err)
}

func TestMetaProcessor_CreateGenesisBlockBodyOK(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		func(shardID uint32, hdrHash []byte) {},
	)
	rootHash, err := mp.CreateGenesisBlock(nil)
	assert.Nil(t, err)
	assert.NotNil(t, rootHash)
	assert.Equal(t, rootHash, []byte("metachain genesis block root hash"))
}

func TestMetaProcessor_RemoveBlockInfoFromPoolShouldErrNilMetaBlockHeader(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		func(shardID uint32, hdrHash []byte) {},
	)
	err := mp.RemoveBlockInfoFromPool(nil)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilMetaBlockHeader)
}

func TestMetaProcessor_RemoveBlockInfoFromPoolShouldWork(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		func(shardID uint32, hdrHash []byte) {},
	)
	header := createMetaBlockHeader()
	err := mp.RemoveBlockInfoFromPool(header)
	assert.Nil(t, err)
}

//------- ComputeNewNoncePrevHash

func TestMetaProcessor_DisplayLogInfo(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	hasher := mock.HasherMock{}
	hdr := createMetaBlockHeader()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		func(shardID uint32, hdrHash []byte) {},
	)
	hdr.PrevHash = hasher.Compute("prev hash")
	mp.DisplayMetaBlock(hdr)
}

func TestMetaProcessor_CreateBlockHeaderShouldNotReturnNilWhenCreateShardInfoFail(t *testing.T) {
	t.Parallel()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{
			JournalLenCalled: func() int {
				return 1
			},
		},
		initMetaDataPool(),
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		func(shardID uint32, hdrHash []byte) {},
	)
	haveTime := func() bool { return true }
	hdr, err := mp.CreateBlockHeader(nil, 0, haveTime)
	assert.NotNil(t, err)
	assert.Nil(t, hdr)
}

func TestMetaProcessor_CreateBlockHeaderShouldWork(t *testing.T) {
	t.Parallel()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{
			JournalLenCalled: func() int {
				return 0
			},
		},
		initMetaDataPool(),
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		func(shardID uint32, hdrHash []byte) {},
	)
	haveTime := func() bool { return true }
	hdr, err := mp.CreateBlockHeader(nil, 0, haveTime)
	assert.Nil(t, err)
	assert.NotNil(t, hdr)
}

func TestMetaProcessor_CommitBlockShouldRevertAccountStateWhenErr(t *testing.T) {
	t.Parallel()
	// set accounts dirty
	journalEntries := 3
	revToSnapshot := func(snapshot int) error {
		journalEntries = 0
		return nil
	}
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{
			RevertToSnapshotCalled: revToSnapshot,
		},
		initMetaDataPool(),
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		func(shardID uint32, hdrHash []byte) {},
	)
	err := mp.CommitBlock(nil, nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, 0, journalEntries)
}

func TestMetaProcessor_MarshalizedDataForCrossShardShouldWork(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		func(shardID uint32, hdrHash []byte) {},
	)

	msh, mstx, err := mp.MarshalizedDataForCrossShard(&block.MetaBlockBody{})
	assert.Nil(t, err)
	assert.NotNil(t, msh)
	assert.NotNil(t, mstx)
}

//------- receivedHeader

func TestMetaProcessor_ReceivedHeaderShouldEraseRequested(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewMetaPoolsHolderFake()

	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		dataPool,
		&mock.ForkDetectorMock{},
		hasher,
		marshalizer,
		initStore(),
		func(shardID uint32, hdrHash []byte) {},
	)

	//add 3 tx hashes on requested list
	hdrHash1 := []byte("hdr hash 1")
	hdrHash2 := []byte("hdr hash 2")
	hdrHash3 := []byte("hdr hash 3")

	mp.AddHdrHashToRequestedList(hdrHash1)
	mp.AddHdrHashToRequestedList(hdrHash2)
	mp.AddHdrHashToRequestedList(hdrHash3)

	//received txHash2
	mp.ReceivedHeader(hdrHash2)

	assert.True(t, mp.IsHdrHashRequested(hdrHash1))
	assert.False(t, mp.IsHdrHashRequested(hdrHash2))
	assert.True(t, mp.IsHdrHashRequested(hdrHash3))
}

//------- createShardInfo

func TestMetaProcessor_CreateShardInfoShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewMetaPoolsHolderFake()

	//we will have a 3 hdrs in pool

	hdrHash1 := []byte("hdr hash 1")
	hdrHash2 := []byte("hdr hash 2")
	hdrHash3 := []byte("hdr hash 3")

	mbHash1 := []byte("mb hash 1")
	mbHash2 := []byte("mb hash 2")
	mbHash3 := []byte("mb hash 3")

	miniBlockHeader1 := block.MiniBlockHeader{Hash: mbHash1}
	miniBlockHeader2 := block.MiniBlockHeader{Hash: mbHash2}
	miniBlockHeader3 := block.MiniBlockHeader{Hash: mbHash3}

	miniBlockHeaders1 := make([]block.MiniBlockHeader, 0)
	miniBlockHeaders1 = append(miniBlockHeaders1, miniBlockHeader1)
	miniBlockHeaders1 = append(miniBlockHeaders1, miniBlockHeader2)
	miniBlockHeaders1 = append(miniBlockHeaders1, miniBlockHeader3)

	miniBlockHeaders2 := make([]block.MiniBlockHeader, 0)
	miniBlockHeaders2 = append(miniBlockHeaders2, miniBlockHeader1)
	miniBlockHeaders2 = append(miniBlockHeaders2, miniBlockHeader2)

	miniBlockHeaders3 := make([]block.MiniBlockHeader, 0)
	miniBlockHeaders3 = append(miniBlockHeaders3, miniBlockHeader1)

	//put the existing tx inside datapool
	dataPool.ShardHeaders().Put(hdrHash1, &block.Header{
		Nonce:            45,
		ShardId:          0,
		MiniBlockHeaders: miniBlockHeaders1})
	dataPool.ShardHeaders().Put(hdrHash2, &block.Header{
		Nonce:            45,
		ShardId:          1,
		MiniBlockHeaders: miniBlockHeaders2})
	dataPool.ShardHeaders().Put(hdrHash3, &block.Header{
		Nonce:            45,
		ShardId:          2,
		MiniBlockHeaders: miniBlockHeaders3})

	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{
			RevertToSnapshotCalled: func(snapshot int) error {
				assert.Fail(t, "revert should have not been called")
				return nil
			},
			JournalLenCalled: func() int {
				return 0
			},
		},
		dataPool,
		&mock.ForkDetectorMock{},
		hasher,
		marshalizer,
		initStore(),
		func(shardID uint32, hdrHash []byte) {},
	)

	haveTime := func() bool { return true }

	shardInfo, err := mp.CreateShardInfo(2, 0, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(3, 0, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(4, 0, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(5, 0, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(6, 0, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(shardInfo))
}

func TestMetaProcessor_RestoreBlockIntoPoolsShouldErrNilMetaBlockHeader(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		func(shardID uint32, hdrHash []byte) {},
	)
	err := mp.RestoreBlockIntoPools(nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilMetaBlockHeader)
}

func TestMetaProcessor_RestoreBlockIntoPoolsShouldWork(t *testing.T) {
	t.Parallel()
	dataPool := mock.NewMetaPoolsHolderFake()
	marshalizerMock := &mock.MarshalizerMock{}
	hasherMock := &mock.HasherStub{}

	body := &block.MetaBlockBody{}
	hdr := block.Header{Nonce: 1}
	buffHdr, _ := marshalizerMock.Marshal(hdr)
	hdrHash := []byte("hdr_hash1")

	store := &mock.ChainStorerMock{
		GetAllCalled: func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
			m := make(map[string][]byte, 0)
			m[string(hdrHash)] = buffHdr
			return m, nil
		},
	}

	mp, _ := blproc.NewMetaProcessor(
		&mock.AccountsStub{},
		dataPool,
		&mock.ForkDetectorMock{},
		hasherMock,
		marshalizerMock,
		store,
		func(shardID uint32, hdrHash []byte) {},
	)

	mhdr := createMetaBlockHeader()

	err := mp.RestoreBlockIntoPools(mhdr, body)

	hdrFromPool, _ := dataPool.ShardHeaders().Get(hdrHash)
	assert.Nil(t, err)
	assert.Equal(t, &hdr, hdrFromPool)
}
