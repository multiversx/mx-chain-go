package block_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	blproc "github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

func createMetaBlockHeader() *block.MetaBlock {
	hdr := block.MetaBlock{
		Nonce:         1,
		Round:         1,
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

func createGenesisBlocks(shardCoordinator sharding.Coordinator) map[uint32]data.HeaderHandler {
	genesisBlocks := make(map[uint32]data.HeaderHandler)
	for shardId := uint32(0); shardId < shardCoordinator.NumberOfShards(); shardId++ {
		genesisBlocks[shardId] = createGenesisBlock(shardId)
	}

	genesisBlocks[sharding.MetachainShardId] = createGenesisMetaBlock()

	return genesisBlocks
}

func createGenesisBlock(shardId uint32) *block.Header {
	rootHash := []byte("roothash")
	return &block.Header{
		Nonce:         0,
		Round:         0,
		Signature:     rootHash,
		RandSeed:      rootHash,
		PrevRandSeed:  rootHash,
		ShardId:       shardId,
		PubKeysBitmap: rootHash,
		RootHash:      rootHash,
		PrevHash:      rootHash,
	}
}

func createGenesisMetaBlock() *block.MetaBlock {
	rootHash := []byte("roothash")
	return &block.MetaBlock{
		Nonce:         0,
		Round:         0,
		Signature:     rootHash,
		RandSeed:      rootHash,
		PrevRandSeed:  rootHash,
		PubKeysBitmap: rootHash,
		RootHash:      rootHash,
		PrevHash:      rootHash,
	}
}

//------- NewMetaProcessor

func TestNewMetaProcessor_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		nil,
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	be, err := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		nil,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilForkDetectorShouldErr(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		mdp,
		nil,
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)
	assert.Equal(t, process.ErrNilForkDetector, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		nil,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		nil,
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)
	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		nil,
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)
	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilChainStorerShouldErr(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		nil,
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)
	assert.Equal(t, process.ErrNilStorage, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilRequestHeaderHandlerShouldErr(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	be, err := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		nil,
	)
	assert.Equal(t, process.ErrNilRequestHandler, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	genesisBlocks := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	mp, err := blproc.NewMetaProcessorBasicSingleShard(mdp, genesisBlocks)
	assert.Nil(t, err)
	assert.NotNil(t, mp)
}

//------- ProcessBlock

func TestMetaProcessor_ProcessBlockWithNilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	genesisBlocks := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	mp, _ := blproc.NewMetaProcessorBasicSingleShard(mdp, genesisBlocks)
	blk := &block.MetaBlockBody{}

	err := mp.ProcessBlock(nil, &block.MetaBlock{}, blk, haveTime)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestMetaProcessor_ProcessBlockWithNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	genesisBlocks := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	mp, _ := blproc.NewMetaProcessorBasicSingleShard(mdp, genesisBlocks)
	blk := &block.MetaBlockBody{}

	err := mp.ProcessBlock(&blockchain.MetaChain{}, nil, blk, haveTime)
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestMetaProcessor_ProcessBlockWithNilBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	genesisBlocks := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	mp, _ := blproc.NewMetaProcessorBasicSingleShard(mdp, genesisBlocks)

	err := mp.ProcessBlock(&blockchain.MetaChain{}, &block.MetaBlock{}, nil, haveTime)
	assert.Equal(t, process.ErrNilBlockBody, err)
}

func TestMetaProcessor_ProcessBlockWithNilHaveTimeFuncShouldErr(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	genesisBlocks := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	mp, _ := blproc.NewMetaProcessorBasicSingleShard(mdp, genesisBlocks)
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
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revToSnapshot,
		},
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)
	// should return err
	err := mp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrAccountStateDirty)
}

func TestMetaProcessor_ProcessWithHeaderNotFirstShouldErr(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	genesisBlocks := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	mp, _ := blproc.NewMetaProcessorBasicSingleShard(mdp, genesisBlocks)

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
	genesisBlocks := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	mp, _ := blproc.NewMetaProcessorBasicSingleShard(mdp, genesisBlocks)
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
	genesisBlocks := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	mp, _ := blproc.NewMetaProcessorBasicSingleShard(mdp, genesisBlocks)
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
	rootHashCalled := func() ([]byte, error) {
		return []byte("rootHashX"), nil
	}
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revertToSnapshot,
			RootHashCalled:         rootHashCalled,
		},
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)

	go func() {
		mp.ChRcvAllHdrs() <- true
	}()

	// should return err
	mp.SetNextKValidity(0)
	hdr.ShardInfo = make([]block.ShardData, 0)
	err := mp.ProcessBlock(blkc, hdr, body, haveTime)

	assert.Equal(t, process.ErrRootStateMissmatch, err)
	assert.True(t, wasCalled)
}

//------- processBlockHeader

func TestMetaProcessor_ProcessBlockHeaderShouldPass(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	accounts := &mock.AccountsStub{}
	accounts.RevertToSnapshotCalled = func(snapshot int) error {
		return nil
	}
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		accounts,
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)

	txHash := []byte("txhash")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock1 := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	miniblock2 := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   2,
		TxHashes:        txHashes,
	}

	miniBlocks := make([]block.MiniBlock, 0)
	miniBlocks = append(miniBlocks, miniblock1, miniblock2)

	metaBlock := &block.MetaBlock{
		Round:     10,
		Nonce:     45,
		RootHash:  []byte("prevRootHash"),
		ShardInfo: createShardData(mock.HasherMock{}, &mock.MarshalizerMock{}, miniBlocks),
	}

	err := mp.ProcessBlockHeaders(metaBlock, 1, haveTime)
	assert.Nil(t, err)
}

//------- requestFinalMissingHeader
func TestMetaProcessor_RequestFinalMissingHeaderShouldPass(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	accounts := &mock.AccountsStub{}
	accounts.RevertToSnapshotCalled = func(snapshot int) error {
		return nil
	}
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		accounts,
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		&mock.RequestHandlerMock{},
	)
	mdp.ShardHeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
		cs := &mock.Uint64CacherStub{}
		cs.PeekCalled = func(key uint64) (value interface{}, ok bool) {
			return &block.Header{Nonce: 1}, true
		}
		cs.RemoveCalled = func(u uint64) {

		}
		return cs
	}
	mp.AddHdrHashToRequestedList([]byte("header_hash"))
	mp.SetCurrHighestShardHdrsNonces(0, 1)
	mp.SetCurrHighestShardHdrsNonces(1, 2)
	mp.SetCurrHighestShardHdrsNonces(2, 3)
	res := mp.RequestFinalMissingHeaders()
	assert.Equal(t, res, uint32(3))
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
		&mock.ServiceContainerMock{},
		accounts,
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
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
		&mock.ServiceContainerMock{},
		accounts,
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		marshalizer,
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)
	blkc := createTestBlockchain()
	err := mp.CommitBlock(blkc, hdr, body)
	assert.Equal(t, errMarshalizer, err)
}

func TestMetaProcessor_CommitBlockStorageFailsForHeaderShouldErr(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	wasCalled := false
	errPersister := errors.New("failure")
	accounts := &mock.AccountsStub{
		CommitCalled: func() (i []byte, e error) {
			return nil, nil
		},
	}
	hdr := createMetaBlockHeader()
	body := &block.MetaBlockBody{}
	hdrUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			wasCalled = true
			return errPersister
		},
	}
	store := initStore()
	store.AddStorer(dataRetriever.MetaBlockUnit, hdrUnit)

	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		accounts,
		mdp,
		&mock.ForkDetectorMock{
			AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
				return nil
			},
		},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		store,
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)

	blkc, _ := blockchain.NewMetaChain(
		generateTestCache(),
	)
	err := mp.CommitBlock(blkc, hdr, body)
	assert.True(t, wasCalled)
	assert.Nil(t, err)
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
	store := initStore()

	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		accounts,
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		store,
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
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
	store := initStore()

	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		accounts,
		mdp,
		fd,
		mock.NewOneShardCoordinatorMock(),
		hasher,
		&mock.MarshalizerMock{},
		store,
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
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
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
	}
	forkDetectorAddCalled := false
	fd := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
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
	store := initStore()
	store.AddStorer(dataRetriever.BlockHeaderUnit, blockHeaderUnit)

	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		accounts,
		mdp,
		fd,
		mock.NewOneShardCoordinatorMock(),
		hasher,
		&mock.MarshalizerMock{},
		store,
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
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

func TestBlockProc_RequestTransactionFromNetwork(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
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
	hdrsRequested, _ := mp.RequestBlockHeaders(header)
	assert.Equal(t, uint32(1), hdrsRequested)
}

func TestMetaProcessor_RemoveBlockInfoFromPoolShouldErrNilMetaBlockHeader(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)
	err := mp.RemoveBlockInfoFromPool(nil)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilMetaBlockHeader)
}

func TestMetaProcessor_RemoveBlockInfoFromPoolShouldWork(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
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
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)
	hdr.PrevHash = hasher.Compute("prev hash")
	mp.DisplayMetaBlock(hdr)
}

func TestMetaProcessor_CreateBlockHeaderShouldNotReturnNilWhenCreateShardInfoFail(t *testing.T) {
	t.Parallel()

	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{
			JournalLenCalled: func() int {
				return 1
			},
		},
		initMetaDataPool(),
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)
	haveTime := func() bool { return true }
	hdr, err := mp.CreateBlockHeader(nil, 0, haveTime)
	assert.NotNil(t, err)
	assert.Nil(t, hdr)
}

func TestMetaProcessor_CreateBlockHeaderShouldWork(t *testing.T) {
	t.Parallel()

	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{
			JournalLenCalled: func() int {
				return 0
			},
			RootHashCalled: func() ([]byte, error) {
				return []byte("root"), nil
			},
		},
		initMetaDataPool(),
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
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
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{
			RevertToSnapshotCalled: revToSnapshot,
		},
		initMetaDataPool(),
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)
	err := mp.CommitBlock(nil, nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, 0, journalEntries)
}

func TestMetaProcessor_MarshalizedDataToBroadcastShouldWork(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)

	msh, mstx, err := mp.MarshalizedDataToBroadcast(&block.MetaBlock{}, &block.MetaBlockBody{})
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
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		dataPool,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		hasher,
		marshalizer,
		initStore(),
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)

	//add 3 tx hashes on requested list
	hdrHash1 := []byte("hdr hash 1")
	hdrHash2 := []byte("hdr hash 2")
	hdrHash3 := []byte("hdr hash 3")

	mp.AddHdrHashToRequestedList(hdrHash1)
	mp.AddHdrHashToRequestedList(hdrHash2)
	mp.AddHdrHashToRequestedList(hdrHash3)

	//received txHash2
	hdr := &block.Header{Nonce: 1}
	dataPool.ShardHeaders().Put(hdrHash2, hdr)
	mp.ReceivedHeader(hdrHash2)

	assert.True(t, mp.IsHdrHashRequested(hdrHash1))
	assert.False(t, mp.IsHdrHashRequested(hdrHash2))
	assert.True(t, mp.IsHdrHashRequested(hdrHash3))
}

//------- createShardInfo

func TestMetaProcessor_CreateShardInfoShouldWorkNoHdrAddedNotValid(t *testing.T) {
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

	//put the existing headers inside datapool
	dataPool.ShardHeaders().Put(hdrHash1, &block.Header{
		Round:            1,
		Nonce:            45,
		ShardId:          0,
		MiniBlockHeaders: miniBlockHeaders1})
	dataPool.ShardHeaders().Put(hdrHash2, &block.Header{
		Round:            2,
		Nonce:            45,
		ShardId:          1,
		MiniBlockHeaders: miniBlockHeaders2})
	dataPool.ShardHeaders().Put(hdrHash3, &block.Header{
		Round:            3,
		Nonce:            45,
		ShardId:          2,
		MiniBlockHeaders: miniBlockHeaders3})

	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
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
		mock.NewMultiShardsCoordinatorMock(5),
		hasher,
		marshalizer,
		initStore(),
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(5)),
		&mock.RequestHandlerMock{},
	)

	haveTime := func() bool { return true }
	round := uint32(10)
	shardInfo, err := mp.CreateShardInfo(2, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(3, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(4, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(5, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(6, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(shardInfo))
}

func TestMetaProcessor_CreateShardInfoShouldWorkNoHdrAddedNotFinal(t *testing.T) {
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

	shardNr := uint32(5)
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
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
		mock.NewMultiShardsCoordinatorMock(shardNr),
		hasher,
		marshalizer,
		initStore(),
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(5)),
		&mock.RequestHandlerMock{},
	)

	haveTime := func() bool { return true }

	prevRandSeed := []byte("prevrand")

	lastNodesHdrs := mp.LastNotarizedHdrs()
	for i := uint32(0); i < shardNr; i++ {
		lastHdr := &block.Header{Round: 9,
			Nonce:    44,
			RandSeed: prevRandSeed,
			ShardId:  i}
		lastNodesHdrs[i] = lastHdr
	}

	//put the existing headers inside datapool
	prevHash, _ := mp.ComputeHeaderHash(lastNodesHdrs[0].(*block.Header))
	dataPool.ShardHeaders().Put(hdrHash1, &block.Header{
		Round:            10,
		Nonce:            45,
		ShardId:          0,
		PrevRandSeed:     prevRandSeed,
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders1})

	prevHash, _ = mp.ComputeHeaderHash(lastNodesHdrs[1].(*block.Header))
	dataPool.ShardHeaders().Put(hdrHash2, &block.Header{
		Round:            20,
		Nonce:            45,
		ShardId:          1,
		PrevRandSeed:     prevRandSeed,
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders2})

	prevHash, _ = mp.ComputeHeaderHash(lastNodesHdrs[2].(*block.Header))
	dataPool.ShardHeaders().Put(hdrHash3, &block.Header{
		Round:            30,
		Nonce:            45,
		ShardId:          2,
		PrevRandSeed:     prevRandSeed,
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders3})

	mp.SetNextKValidity(0)
	round := uint32(40)
	shardInfo, err := mp.CreateShardInfo(2, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(3, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(4, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(5, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(6, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(shardInfo))
}

func TestMetaProcessor_CreateShardInfoShouldWorkHdrsAdded(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewMetaPoolsHolderFake()

	//we will have a 3 hdrs in pool

	hdrHash1 := []byte("hdr hash 1")
	hdrHash2 := []byte("hdr hash 2")
	hdrHash3 := []byte("hdr hash 3")

	hdrHash11 := []byte("hdr hash 11")
	hdrHash22 := []byte("hdr hash 22")
	hdrHash33 := []byte("hdr hash 33")

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

	shardNr := uint32(5)
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
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
		mock.NewMultiShardsCoordinatorMock(shardNr),
		hasher,
		marshalizer,
		initStore(),
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(shardNr)),
		&mock.RequestHandlerMock{},
	)

	haveTime := func() bool { return true }

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	lastNodesHdrs := mp.LastNotarizedHdrs()
	for i := uint32(0); i < shardNr; i++ {
		lastHdr := &block.Header{Round: 9,
			Nonce:    44,
			RandSeed: prevRandSeed,
			ShardId:  i}
		lastNodesHdrs[i] = lastHdr
	}

	headers := make([]*block.Header, 0)

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := mp.ComputeHeaderHash(lastNodesHdrs[0].(*block.Header))
	headers = append(headers, &block.Header{
		Round:            10,
		Nonce:            45,
		ShardId:          0,
		PrevRandSeed:     prevRandSeed,
		RandSeed:         currRandSeed,
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders1})

	prevHash, _ = mp.ComputeHeaderHash(headers[0])
	headers = append(headers, &block.Header{
		Round:            11,
		Nonce:            46,
		ShardId:          0,
		PrevRandSeed:     currRandSeed,
		RandSeed:         []byte("nextrand"),
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders1})

	dataPool.ShardHeaders().Put(hdrHash1, headers[0])
	dataPool.ShardHeaders().Put(hdrHash11, headers[1])

	// header shard 1
	prevHash, _ = mp.ComputeHeaderHash(lastNodesHdrs[1].(*block.Header))
	headers = append(headers, &block.Header{
		Round:            10,
		Nonce:            45,
		ShardId:          1,
		PrevRandSeed:     prevRandSeed,
		RandSeed:         currRandSeed,
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders2})

	prevHash, _ = mp.ComputeHeaderHash(headers[2])
	headers = append(headers, &block.Header{
		Round:            11,
		Nonce:            46,
		ShardId:          1,
		PrevRandSeed:     currRandSeed,
		RandSeed:         []byte("nextrand"),
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders2})

	dataPool.ShardHeaders().Put(hdrHash2, headers[2])
	dataPool.ShardHeaders().Put(hdrHash22, headers[3])

	// header shard 2
	prevHash, _ = mp.ComputeHeaderHash(lastNodesHdrs[2].(*block.Header))
	headers = append(headers, &block.Header{
		Round:            10,
		Nonce:            45,
		ShardId:          2,
		PrevRandSeed:     prevRandSeed,
		RandSeed:         currRandSeed,
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders3})

	prevHash, _ = mp.ComputeHeaderHash(headers[4])
	headers = append(headers, &block.Header{
		Round:            11,
		Nonce:            46,
		ShardId:          2,
		PrevRandSeed:     currRandSeed,
		RandSeed:         []byte("nextrand"),
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders3})

	dataPool.ShardHeaders().Put(hdrHash3, headers[4])
	dataPool.ShardHeaders().Put(hdrHash33, headers[5])

	mp.SetNextKValidity(1)
	round := uint32(15)
	shardInfo, err := mp.CreateShardInfo(2, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(3, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(4, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(5, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(6, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(shardInfo))
}

func TestMetaProcessor_CreateShardInfoEmptyBlockHDRRoundTooHigh(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewMetaPoolsHolderFake()

	//we will have a 3 hdrs in pool

	hdrHash1 := []byte("hdr hash 1")
	hdrHash2 := []byte("hdr hash 2")
	hdrHash3 := []byte("hdr hash 3")

	hdrHash11 := []byte("hdr hash 11")
	hdrHash22 := []byte("hdr hash 22")
	hdrHash33 := []byte("hdr hash 33")

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

	shardNr := uint32(5)
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
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
		mock.NewMultiShardsCoordinatorMock(shardNr),
		hasher,
		marshalizer,
		initStore(),
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(shardNr)),
		&mock.RequestHandlerMock{},
	)

	haveTime := func() bool { return true }

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	lastNodesHdrs := mp.LastNotarizedHdrs()
	for i := uint32(0); i < shardNr; i++ {
		lastHdr := &block.Header{Round: 9,
			Nonce:    44,
			RandSeed: prevRandSeed,
			ShardId:  i}
		lastNodesHdrs[i] = lastHdr
	}

	headers := make([]*block.Header, 0)

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := mp.ComputeHeaderHash(lastNodesHdrs[0].(*block.Header))
	headers = append(headers, &block.Header{
		Round:            10,
		Nonce:            45,
		ShardId:          0,
		PrevRandSeed:     prevRandSeed,
		RandSeed:         currRandSeed,
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders1})

	prevHash, _ = mp.ComputeHeaderHash(headers[0])
	headers = append(headers, &block.Header{
		Round:            11,
		Nonce:            46,
		ShardId:          0,
		PrevRandSeed:     currRandSeed,
		RandSeed:         []byte("nextrand"),
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders1})

	dataPool.ShardHeaders().Put(hdrHash1, headers[0])
	dataPool.ShardHeaders().Put(hdrHash11, headers[1])

	// header shard 1
	prevHash, _ = mp.ComputeHeaderHash(lastNodesHdrs[1].(*block.Header))
	headers = append(headers, &block.Header{
		Round:            10,
		Nonce:            45,
		ShardId:          1,
		PrevRandSeed:     prevRandSeed,
		RandSeed:         currRandSeed,
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders2})

	prevHash, _ = mp.ComputeHeaderHash(headers[2])
	headers = append(headers, &block.Header{
		Round:            11,
		Nonce:            46,
		ShardId:          1,
		PrevRandSeed:     currRandSeed,
		RandSeed:         []byte("nextrand"),
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders2})

	dataPool.ShardHeaders().Put(hdrHash2, headers[2])
	dataPool.ShardHeaders().Put(hdrHash22, headers[3])

	// header shard 2
	prevHash, _ = mp.ComputeHeaderHash(lastNodesHdrs[2].(*block.Header))
	headers = append(headers, &block.Header{
		Round:            10,
		Nonce:            45,
		ShardId:          2,
		PrevRandSeed:     prevRandSeed,
		RandSeed:         currRandSeed,
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders3})

	prevHash, _ = mp.ComputeHeaderHash(headers[4])
	headers = append(headers, &block.Header{
		Round:            11,
		Nonce:            46,
		ShardId:          2,
		PrevRandSeed:     currRandSeed,
		RandSeed:         []byte("nextrand"),
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders3})

	dataPool.ShardHeaders().Put(hdrHash3, headers[4])
	dataPool.ShardHeaders().Put(hdrHash33, headers[5])

	mp.SetNextKValidity(1)
	round := uint32(20)
	shardInfo, err := mp.CreateShardInfo(2, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(3, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(4, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(5, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(shardInfo))

	shardInfo, err = mp.CreateShardInfo(6, round, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(shardInfo))
}

func TestMetaProcessor_RestoreBlockIntoPoolsShouldErrNilMetaBlockHeader(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initStore(),
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
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
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
			return buffHdr, nil
		},
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				RemoveCalled: func(key []byte) error {
					return nil
				},
			}
		},
	}

	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		dataPool,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		hasherMock,
		marshalizerMock,
		store,
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)

	mhdr := createMetaBlockHeader()

	err := mp.RestoreBlockIntoPools(mhdr, body)

	hdrFromPool, _ := dataPool.ShardHeaders().Get(hdrHash)
	assert.Nil(t, err)
	assert.Equal(t, &hdr, hdrFromPool)
}

func TestMetaProcessor_CreateLastNotarizedHdrs(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewMetaPoolsHolderFake()

	shardNr := uint32(5)
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
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
		mock.NewMultiShardsCoordinatorMock(shardNr),
		hasher,
		marshalizer,
		initStore(),
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(shardNr)),
		&mock.RequestHandlerMock{},
	)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	lastNodesHdrs := mp.LastNotarizedHdrs()
	firstNonce := uint64(44)
	for i := uint32(0); i < shardNr; i++ {
		lastHdr := &block.Header{Round: 9,
			Nonce:    firstNonce,
			RandSeed: prevRandSeed,
			ShardId:  i}
		lastNodesHdrs[i] = lastHdr
	}

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := mp.ComputeHeaderHash(lastNodesHdrs[0].(*block.Header))
	prevHdr := &block.Header{
		Round:        10,
		Nonce:        45,
		ShardId:      0,
		PrevRandSeed: prevRandSeed,
		RandSeed:     currRandSeed,
		PrevHash:     prevHash,
		RootHash:     []byte("prevRootHash")}

	prevHash, _ = mp.ComputeHeaderHash(prevHdr)
	currHdr := &block.Header{
		Round:        11,
		Nonce:        46,
		ShardId:      0,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash")}
	currHash, _ := mp.ComputeHeaderHash(currHdr)
	prevHash, _ = mp.ComputeHeaderHash(prevHdr)

	metaHdr := &block.MetaBlock{Round: 15}
	shDataCurr := block.ShardData{ShardId: 0, HeaderHash: currHash}
	metaHdr.ShardInfo = make([]block.ShardData, 0)
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataCurr)
	shDataPrev := block.ShardData{ShardId: 0, HeaderHash: prevHash}
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataPrev)

	// test header not in pool and defer called
	err := mp.CreateLastNotarizedHdrs(metaHdr)
	assert.Equal(t, process.ErrMissingHeader, err)
	lastNodesHdrs = mp.LastNotarizedHdrs()
	assert.Equal(t, firstNonce, lastNodesHdrs[currHdr.ShardId].GetNonce())

	// wrong header type in pool and defer called
	dataPool.ShardHeaders().Put(currHash, metaHdr)
	dataPool.ShardHeaders().Put(prevHash, prevHdr)

	err = mp.CreateLastNotarizedHdrs(metaHdr)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
	lastNodesHdrs = mp.LastNotarizedHdrs()
	assert.Equal(t, firstNonce, lastNodesHdrs[currHdr.ShardId].GetNonce())

	// put headers in pool
	dataPool.ShardHeaders().Put(currHash, currHdr)
	dataPool.ShardHeaders().Put(prevHash, prevHdr)

	err = mp.CreateLastNotarizedHdrs(metaHdr)
	assert.Nil(t, err)
	lastNodesHdrs = mp.LastNotarizedHdrs()
	assert.Equal(t, currHdr, lastNodesHdrs[currHdr.ShardId])
}

func TestMetaProcessor_CheckShardHeadersValidity(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewMetaPoolsHolderFake()

	shardNr := uint32(5)
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
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
		mock.NewMultiShardsCoordinatorMock(shardNr),
		hasher,
		marshalizer,
		initStore(),
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(shardNr)),
		&mock.RequestHandlerMock{},
	)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	lastNodesHdrs := mp.LastNotarizedHdrs()
	for i := uint32(0); i < shardNr; i++ {
		lastHdr := &block.Header{Round: 9,
			Nonce:    44,
			RandSeed: prevRandSeed,
			ShardId:  i}
		lastNodesHdrs[i] = lastHdr
	}

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := mp.ComputeHeaderHash(lastNodesHdrs[0].(*block.Header))
	prevHdr := &block.Header{
		Round:        10,
		Nonce:        45,
		ShardId:      0,
		PrevRandSeed: prevRandSeed,
		RandSeed:     currRandSeed,
		PrevHash:     prevHash,
		RootHash:     []byte("prevRootHash")}

	prevHash, _ = mp.ComputeHeaderHash(prevHdr)
	currHdr := &block.Header{
		Round:        11,
		Nonce:        46,
		ShardId:      0,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash")}
	currHash, _ := mp.ComputeHeaderHash(currHdr)
	dataPool.ShardHeaders().Put(currHash, currHdr)
	prevHash, _ = mp.ComputeHeaderHash(prevHdr)
	dataPool.ShardHeaders().Put(prevHash, prevHdr)
	wrongCurrHdr := &block.Header{
		Round:        11,
		Nonce:        48,
		ShardId:      0,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash")}
	wrongCurrHash, _ := mp.ComputeHeaderHash(wrongCurrHdr)
	dataPool.ShardHeaders().Put(wrongCurrHash, wrongCurrHdr)

	metaHdr := &block.MetaBlock{Round: 20}
	shDataCurr := block.ShardData{ShardId: 0, HeaderHash: wrongCurrHash}
	metaHdr.ShardInfo = make([]block.ShardData, 0)
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataCurr)
	shDataPrev := block.ShardData{ShardId: 0, HeaderHash: prevHash}
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataPrev)

	_, err := mp.CheckShardHeadersValidity(metaHdr)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)

	shDataCurr = block.ShardData{ShardId: 0, HeaderHash: currHash}
	metaHdr.ShardInfo = make([]block.ShardData, 0)
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataCurr)
	shDataPrev = block.ShardData{ShardId: 0, HeaderHash: prevHash}
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataPrev)

	highestNonceHdrs, err := mp.CheckShardHeadersValidity(metaHdr)
	assert.Nil(t, err)
	assert.NotNil(t, highestNonceHdrs)
	assert.Equal(t, currHdr.Nonce, highestNonceHdrs[currHdr.ShardId].GetNonce())
}

func TestMetaProcessor_CheckShardHeadersValidityWrongNonceFromLastNoted(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewMetaPoolsHolderFake()

	shardNr := uint32(5)
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
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
		mock.NewMultiShardsCoordinatorMock(shardNr),
		hasher,
		marshalizer,
		initStore(),
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(shardNr)),
		&mock.RequestHandlerMock{},
	)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	lastNodesHdrs := mp.LastNotarizedHdrs()
	for i := uint32(0); i < shardNr; i++ {
		lastHdr := &block.Header{Round: 9,
			Nonce:    44,
			RandSeed: prevRandSeed,
			ShardId:  i}
		lastNodesHdrs[i] = lastHdr
	}

	//put the existing headers inside datapool
	currHdr := &block.Header{
		Round:        11,
		Nonce:        46,
		ShardId:      0,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     []byte("prevhash"),
		RootHash:     []byte("currRootHash")}
	currHash, _ := mp.ComputeHeaderHash(currHdr)
	dataPool.ShardHeaders().Put(currHash, currHdr)
	metaHdr := &block.MetaBlock{Round: 20}

	shDataCurr := block.ShardData{ShardId: 0, HeaderHash: currHash}
	metaHdr.ShardInfo = make([]block.ShardData, 0)
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataCurr)

	highestNonceHdrs, err := mp.CheckShardHeadersValidity(metaHdr)
	assert.Nil(t, highestNonceHdrs)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)
}

func TestMetaProcessor_CheckShardHeadersValidityRoundZeroLastNoted(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewMetaPoolsHolderFake()

	shardNr := uint32(5)
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
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
		mock.NewMultiShardsCoordinatorMock(shardNr),
		hasher,
		marshalizer,
		initStore(),
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(shardNr)),
		&mock.RequestHandlerMock{},
	)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	lastNodesHdrs := mp.LastNotarizedHdrs()
	for i := uint32(0); i < shardNr; i++ {
		lastHdr := &block.Header{Round: 0,
			Nonce:    0,
			RandSeed: prevRandSeed,
			ShardId:  i}
		lastNodesHdrs[i] = lastHdr
	}

	//put the existing headers inside datapool
	currHdr := &block.Header{
		Round:        0,
		Nonce:        0,
		ShardId:      0,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     []byte("prevhash"),
		RootHash:     []byte("currRootHash")}
	currHash, _ := mp.ComputeHeaderHash(currHdr)

	metaHdr := &block.MetaBlock{Round: 20}

	shDataCurr := block.ShardData{ShardId: 0, HeaderHash: currHash}
	metaHdr.ShardInfo = make([]block.ShardData, 0)
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataCurr)

	highestNonceHdrs, err := mp.CheckShardHeadersValidity(metaHdr)
	assert.Nil(t, highestNonceHdrs)
	assert.Equal(t, process.ErrMissingHeader, err)

	dataPool.ShardHeaders().Put(currHash, currHdr)
	highestNonceHdrs, err = mp.CheckShardHeadersValidity(metaHdr)
	assert.NotNil(t, highestNonceHdrs)
	assert.Nil(t, err)
	assert.Equal(t, currHdr.Nonce, highestNonceHdrs[currHdr.ShardId].GetNonce())
}

func TestMetaProcessor_CheckShardHeadersFinality(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewMetaPoolsHolderFake()

	shardNr := uint32(5)
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
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
		mock.NewMultiShardsCoordinatorMock(shardNr),
		hasher,
		marshalizer,
		initStore(),
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(shardNr)),
		&mock.RequestHandlerMock{},
	)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	lastNodesHdrs := mp.LastNotarizedHdrs()
	for i := uint32(0); i < shardNr; i++ {
		lastHdr := &block.Header{Round: 9,
			Nonce:    44,
			RandSeed: prevRandSeed,
			ShardId:  i}
		lastNodesHdrs[i] = lastHdr
	}

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := mp.ComputeHeaderHash(lastNodesHdrs[0].(*block.Header))
	prevHdr := &block.Header{
		Round:        10,
		Nonce:        45,
		ShardId:      0,
		PrevRandSeed: prevRandSeed,
		RandSeed:     currRandSeed,
		PrevHash:     prevHash,
		RootHash:     []byte("prevRootHash")}

	prevHash, _ = mp.ComputeHeaderHash(prevHdr)
	currHdr := &block.Header{
		Round:        11,
		Nonce:        46,
		ShardId:      0,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash")}

	nextWrongHdr := &block.Header{
		Round:        11,
		Nonce:        44,
		ShardId:      0,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash")}
	prevHash, _ = mp.ComputeHeaderHash(nextWrongHdr)
	dataPool.ShardHeaders().Put(prevHash, nextWrongHdr)

	mp.SetNextKValidity(0)
	metaHdr := &block.MetaBlock{Round: 1}

	err := mp.CheckShardHeadersFinality(nil, lastNodesHdrs)
	assert.Equal(t, process.ErrNilBlockHeader, err)

	// should work for empty highest nonce hdrs - no hdrs added this round to metablock
	err = mp.CheckShardHeadersFinality(metaHdr, nil)
	assert.Nil(t, err)

	mp.SetNextKValidity(0)
	highestNonceHdrs := make(map[uint32]data.HeaderHandler, 0)
	highestNonceHdrs[0] = currHdr
	err = mp.CheckShardHeadersFinality(metaHdr, highestNonceHdrs)
	assert.Nil(t, err)

	mp.SetNextKValidity(1)
	err = mp.CheckShardHeadersFinality(metaHdr, highestNonceHdrs)
	assert.Equal(t, process.ErrHeaderNotFinal, err)

	prevHash, _ = mp.ComputeHeaderHash(currHdr)
	nextHdr := &block.Header{
		Round:        12,
		Nonce:        47,
		ShardId:      0,
		PrevRandSeed: []byte("nextrand"),
		RandSeed:     []byte("nextnextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash")}

	prevHash, _ = mp.ComputeHeaderHash(nextHdr)
	dataPool.ShardHeaders().Put(prevHash, nextHdr)

	metaHdr.Round = 20
	err = mp.CheckShardHeadersFinality(metaHdr, highestNonceHdrs)
	assert.Nil(t, err)
}

func TestMetaProcessor_IsHdrConstructionValid(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewMetaPoolsHolderFake()

	shardNr := uint32(5)
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
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
		mock.NewMultiShardsCoordinatorMock(shardNr),
		hasher,
		marshalizer,
		initStore(),
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(shardNr)),
		&mock.RequestHandlerMock{},
	)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	lastNodesHdrs := mp.LastNotarizedHdrs()
	for i := uint32(0); i < shardNr; i++ {
		lastHdr := &block.Header{Round: 9,
			Nonce:    44,
			RandSeed: prevRandSeed,
			ShardId:  i}
		lastNodesHdrs[i] = lastHdr
	}

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := mp.ComputeHeaderHash(lastNodesHdrs[0].(*block.Header))
	prevHdr := &block.Header{
		Round:        10,
		Nonce:        45,
		ShardId:      0,
		PrevRandSeed: prevRandSeed,
		RandSeed:     currRandSeed,
		PrevHash:     prevHash,
		RootHash:     []byte("prevRootHash")}

	prevHash, _ = mp.ComputeHeaderHash(prevHdr)
	currHdr := &block.Header{
		Round:        11,
		Nonce:        46,
		ShardId:      0,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash")}

	err := mp.IsHdrConstructionValid(nil, prevHdr)
	assert.Equal(t, err, process.ErrNilBlockHeader)

	err = mp.IsHdrConstructionValid(currHdr, nil)
	assert.Equal(t, err, process.ErrNilBlockHeader)

	currHdr.Nonce = 0
	err = mp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrWrongNonceInBlock)

	currHdr.Nonce = 0
	prevHdr.Nonce = 0
	err = mp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrRootStateMissmatch)

	currHdr.Nonce = 0
	prevHdr.Nonce = 0
	prevHdr.RootHash = nil
	err = mp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Nil(t, err)

	currHdr.Nonce = 46
	prevHdr.Nonce = 45
	prevHdr.Round = currHdr.Round + 1
	err = mp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrLowShardHeaderRound)

	prevHdr.Round = currHdr.Round - 1
	currHdr.Nonce = prevHdr.Nonce + 2
	err = mp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrWrongNonceInBlock)

	currHdr.Nonce = prevHdr.Nonce + 1
	prevHdr.RandSeed = []byte("randomwrong")
	err = mp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrRandSeedMismatch)

	prevHdr.RandSeed = currRandSeed
	currHdr.PrevHash = []byte("wronghash")
	err = mp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrInvalidBlockHash)

	currHdr.PrevHash = prevHash
	prevHdr.RootHash = []byte("prevRootHash")
	err = mp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Nil(t, err)
}

func TestMetaProcessor_IsShardHeaderValidFinal(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewMetaPoolsHolderFake()

	shardNr := uint32(5)
	mp, _ := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
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
		mock.NewMultiShardsCoordinatorMock(shardNr),
		hasher,
		marshalizer,
		initStore(),
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(shardNr)),
		&mock.RequestHandlerMock{},
	)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	lastNodesHdrs := mp.LastNotarizedHdrs()
	for i := uint32(0); i < shardNr; i++ {
		lastHdr := &block.Header{Round: 9,
			Nonce:    44,
			RandSeed: prevRandSeed,
			ShardId:  i}
		lastNodesHdrs[i] = lastHdr
	}

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := mp.ComputeHeaderHash(lastNodesHdrs[0].(*block.Header))
	prevHdr := &block.Header{
		Round:        10,
		Nonce:        45,
		ShardId:      0,
		PrevRandSeed: prevRandSeed,
		RandSeed:     currRandSeed,
		PrevHash:     prevHash,
		RootHash:     []byte("prevRootHash")}

	wrongPrevHdr := &block.Header{
		Round:        10,
		Nonce:        50,
		ShardId:      0,
		PrevRandSeed: prevRandSeed,
		RandSeed:     currRandSeed,
		PrevHash:     prevHash,
		RootHash:     []byte("prevRootHash")}

	prevHash, _ = mp.ComputeHeaderHash(prevHdr)
	currHdr := &block.Header{
		Round:        11,
		Nonce:        46,
		ShardId:      0,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash")}

	srtShardHdrs := make([]*block.Header, 0)

	valid, hdrIds := mp.IsShardHeaderValidFinal(currHdr, prevHdr, nil)
	assert.False(t, valid)
	assert.Nil(t, hdrIds)

	valid, hdrIds = mp.IsShardHeaderValidFinal(nil, prevHdr, srtShardHdrs)
	assert.False(t, valid)
	assert.Nil(t, hdrIds)

	valid, hdrIds = mp.IsShardHeaderValidFinal(currHdr, nil, srtShardHdrs)
	assert.False(t, valid)
	assert.Nil(t, hdrIds)

	valid, hdrIds = mp.IsShardHeaderValidFinal(currHdr, wrongPrevHdr, srtShardHdrs)
	assert.False(t, valid)
	assert.Nil(t, hdrIds)

	mp.SetNextKValidity(0)
	valid, hdrIds = mp.IsShardHeaderValidFinal(currHdr, prevHdr, srtShardHdrs)
	assert.True(t, valid)
	assert.NotNil(t, hdrIds)

	mp.SetNextKValidity(1)
	nextWrongHdr := &block.Header{
		Round:        12,
		Nonce:        44,
		ShardId:      0,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash")}

	srtShardHdrs = append(srtShardHdrs, nextWrongHdr)
	valid, hdrIds = mp.IsShardHeaderValidFinal(currHdr, prevHdr, srtShardHdrs)
	assert.False(t, valid)
	assert.Nil(t, hdrIds)

	prevHash, _ = mp.ComputeHeaderHash(currHdr)
	nextHdr := &block.Header{
		Round:        12,
		Nonce:        47,
		ShardId:      0,
		PrevRandSeed: []byte("nextrand"),
		RandSeed:     []byte("nextnextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash")}

	srtShardHdrs = append(srtShardHdrs, nextHdr)
	valid, hdrIds = mp.IsShardHeaderValidFinal(currHdr, prevHdr, srtShardHdrs)
	assert.True(t, valid)
	assert.NotNil(t, hdrIds)
}

func TestMetaProcessor_DecodeBlockBody(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	marshalizerMock := &mock.MarshalizerMock{}
	mp, err := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		marshalizerMock,
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)
	body := &block.MetaBlockBody{}
	message, err := marshalizerMock.Marshal(body)
	assert.Nil(t, err)

	dcdBlk := mp.DecodeBlockBody(nil)
	assert.Nil(t, dcdBlk)

	dcdBlk = mp.DecodeBlockBody(message)
	assert.Equal(t, body, dcdBlk)
}

func TestMetaProcessor_DecodeBlockHeader(t *testing.T) {
	t.Parallel()
	mdp := initMetaDataPool()
	marshalizerMock := &mock.MarshalizerMock{}
	mp, err := blproc.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		marshalizerMock,
		&mock.ChainStorerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
	)
	hdr := &block.MetaBlock{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(0)
	hdr.Signature = []byte("A")
	message, err := marshalizerMock.Marshal(hdr)
	assert.Nil(t, err)

	message, err = marshalizerMock.Marshal(hdr)
	assert.Nil(t, err)

	dcdHdr := mp.DecodeBlockHeader(nil)
	assert.Nil(t, dcdHdr)

	dcdHdr = mp.DecodeBlockHeader(message)
	assert.Equal(t, hdr, dcdHdr)
	assert.Equal(t, []byte("A"), dcdHdr.GetSignature())
}
