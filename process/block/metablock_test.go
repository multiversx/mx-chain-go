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
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/process"
	blproc "github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

func createMockMetaArguments() blproc.ArgMetaProcessor {
	mdp := initMetaDataPool()
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	argsHeaderValidator := blproc.ArgsHeaderValidator{
		Hasher:      &mock.HasherStub{},
		Marshalizer: &mock.MarshalizerMock{},
	}
	headerValidator, _ := blproc.NewHeaderValidator(argsHeaderValidator)

	arguments := blproc.ArgMetaProcessor{
		ArgBaseProcessor: blproc.ArgBaseProcessor{
			Accounts:                     &mock.AccountsStub{},
			ForkDetector:                 &mock.ForkDetectorMock{},
			Hasher:                       &mock.HasherStub{},
			Marshalizer:                  &mock.MarshalizerMock{},
			Store:                        &mock.ChainStorerMock{},
			ShardCoordinator:             shardCoordinator,
			NodesCoordinator:             mock.NewNodesCoordinatorMock(),
			SpecialAddressHandler:        &mock.SpecialAddressHandlerMock{},
			Uint64Converter:              &mock.Uint64ByteSliceConverterMock{},
			StartHeaders:                 createGenesisBlocks(shardCoordinator),
			RequestHandler:               &mock.RequestHandlerMock{},
			Core:                         &mock.ServiceContainerMock{},
			BlockChainHook:               &mock.BlockChainHookHandlerMock{},
			TxCoordinator:                &mock.TransactionCoordinatorMock{},
			ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorMock{},
			EpochStartTrigger:            &mock.EpochStartTriggerStub{},
			HeaderValidator:              headerValidator,
			Rounder:                      &mock.RounderMock{},
			BootStorer: &mock.BoostrapStorerMock{
				PutCalled: func(round int64, bootData bootstrapStorage.BootstrapData) error {
					return nil
				},
			},
		},
		DataPool:           mdp,
		SCDataGetter:       &mock.ScQueryMock{},
		SCToProtocol:       &mock.SCToProtocolStub{},
		PeerChangesHandler: &mock.PeerChangesHandler{},
		PendingMiniBlocks:  &mock.PendingMiniBlocksHandlerStub{},
	}
	return arguments
}

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
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxCount:         1,
	}
	shardMiniBlockHeaders = append(shardMiniBlockHeaders, shardMiniBlockHeader)
	shardData := block.ShardData{
		ShardID:               0,
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
	for ShardID := uint32(0); ShardID < shardCoordinator.NumberOfShards(); ShardID++ {
		genesisBlocks[ShardID] = createGenesisBlock(ShardID)
	}

	genesisBlocks[sharding.MetachainShardId] = createGenesisMetaBlock()

	return genesisBlocks
}

func createGenesisBlock(ShardID uint32) *block.Header {
	rootHash := []byte("roothash")
	return &block.Header{
		Nonce:         0,
		Round:         0,
		Signature:     rootHash,
		RandSeed:      rootHash,
		PrevRandSeed:  rootHash,
		ShardId:       ShardID,
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

func setLastNotarizedHdr(
	noOfShards uint32,
	round uint64,
	nonce uint64,
	randSeed []byte,
	lastNotarizedHdrs map[uint32][]data.HeaderHandler) {
	for i := uint32(0); i < noOfShards; i++ {
		lastHdr := &block.Header{Round: round,
			Nonce:    nonce,
			RandSeed: randSeed,
			ShardId:  i}
		lastNotarizedHdrsCount := len(lastNotarizedHdrs[i])
		if lastNotarizedHdrsCount > 0 {
			lastNotarizedHdrs[i][lastNotarizedHdrsCount-1] = lastHdr
		} else {
			lastNotarizedHdrs[i] = append(lastNotarizedHdrs[i], lastHdr)
		}
	}
}

//------- NewMetaProcessor

func TestNewMetaProcessor_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.Accounts = nil

	be, err := blproc.NewMetaProcessor(arguments)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.DataPool = nil

	be, err := blproc.NewMetaProcessor(arguments)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilForkDetectorShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.ForkDetector = nil

	be, err := blproc.NewMetaProcessor(arguments)
	assert.Equal(t, process.ErrNilForkDetector, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.ShardCoordinator = nil

	be, err := blproc.NewMetaProcessor(arguments)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.Hasher = nil

	be, err := blproc.NewMetaProcessor(arguments)
	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.Marshalizer = nil

	be, err := blproc.NewMetaProcessor(arguments)
	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilChainStorerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.Store = nil

	be, err := blproc.NewMetaProcessor(arguments)
	assert.Equal(t, process.ErrNilStorage, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilRequestHeaderHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.RequestHandler = nil

	be, err := blproc.NewMetaProcessor(arguments)
	assert.Equal(t, process.ErrNilRequestHandler, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilTxCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.TxCoordinator = nil

	be, err := blproc.NewMetaProcessor(arguments)
	assert.Equal(t, process.ErrNilTransactionCoordinator, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilEpochStartShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.EpochStartTrigger = nil

	be, err := blproc.NewMetaProcessor(arguments)
	assert.Equal(t, process.ErrNilEpochStartTrigger, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_NilPendingMiniBlocksShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.PendingMiniBlocks = nil

	be, err := blproc.NewMetaProcessor(arguments)
	assert.Equal(t, process.ErrNilPendingMiniBlocksHandler, err)
	assert.Nil(t, be)
}

func TestNewMetaProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()

	mp, err := blproc.NewMetaProcessor(arguments)
	assert.Nil(t, err)
	assert.NotNil(t, mp)
}

//------- ProcessBlock

func TestMetaProcessor_ProcessBlockWithNilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	mp, _ := blproc.NewMetaProcessor(arguments)
	blk := block.Body{}

	err := mp.ProcessBlock(nil, &block.MetaBlock{}, blk, haveTime)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestMetaProcessor_ProcessBlockWithNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()

	mp, _ := blproc.NewMetaProcessor(arguments)
	blk := block.Body{}

	err := mp.ProcessBlock(&blockchain.MetaChain{}, nil, blk, haveTime)
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestMetaProcessor_ProcessBlockWithNilBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	mp, _ := blproc.NewMetaProcessor(arguments)

	err := mp.ProcessBlock(&blockchain.MetaChain{}, &block.MetaBlock{}, nil, haveTime)
	assert.Equal(t, process.ErrNilBlockBody, err)
}

func TestMetaProcessor_ProcessBlockWithNilHaveTimeFuncShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	mp, _ := blproc.NewMetaProcessor(arguments)
	blk := block.Body{}

	err := mp.ProcessBlock(&blockchain.MetaChain{}, &block.MetaBlock{}, blk, nil)
	assert.Equal(t, process.ErrNilHaveTimeHandler, err)
}

func TestMetaProcessor_ProcessWithDirtyAccountShouldErr(t *testing.T) {
	t.Parallel()

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
	body := block.Body{}
	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revToSnapshot,
	}
	mp, _ := blproc.NewMetaProcessor(arguments)

	// should return err
	err := mp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrAccountStateDirty)
}

func TestMetaProcessor_ProcessWithHeaderNotFirstShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	mp, _ := blproc.NewMetaProcessor(arguments)

	blkc := &blockchain.MetaChain{}
	hdr := &block.MetaBlock{
		Nonce: 2,
	}
	body := block.Body{}
	err := mp.ProcessBlock(blkc, hdr, body, haveTime)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)
}

func TestMetaProcessor_ProcessWithHeaderNotCorrectNonceShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	mp, _ := blproc.NewMetaProcessor(arguments)
	blkc := &blockchain.MetaChain{
		CurrentBlock: &block.MetaBlock{
			Round: 1,
			Nonce: 1,
		},
	}
	hdr := &block.MetaBlock{
		Round: 3,
		Nonce: 3,
	}
	body := block.Body{}

	err := mp.ProcessBlock(blkc, hdr, body, haveTime)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)
}

func TestMetaProcessor_ProcessWithHeaderNotCorrectPrevHashShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	mp, _ := blproc.NewMetaProcessor(arguments)
	blkc := &blockchain.MetaChain{
		CurrentBlock: &block.MetaBlock{
			Round: 1,
			Nonce: 1,
		},
	}
	hdr := &block.MetaBlock{
		Round:    2,
		Nonce:    2,
		PrevHash: []byte("X"),
	}

	body := block.Body{}

	err := mp.ProcessBlock(blkc, hdr, body, haveTime)
	assert.Equal(t, process.ErrBlockHashDoesNotMatch, err)
}

func TestMetaProcessor_ProcessBlockWithErrOnVerifyStateRootCallShouldRevertState(t *testing.T) {
	t.Parallel()

	blkc := &blockchain.MetaChain{
		CurrentBlock: &block.MetaBlock{
			Nonce: 0,
		},
	}
	hdr := createMetaBlockHeader()
	body := block.Body{}
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
	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}
	mp, _ := blproc.NewMetaProcessor(arguments)

	go func() {
		mp.ChRcvAllHdrs() <- true
	}()

	// should return err
	mp.SetShardBlockFinality(0)
	hdr.ShardInfo = make([]block.ShardData, 0)
	err := mp.ProcessBlock(blkc, hdr, body, haveTime)

	assert.Equal(t, process.ErrRootStateDoesNotMatch, err)
	assert.True(t, wasCalled)
}

//------- processBlockHeader

func TestMetaProcessor_ProcessBlockHeaderShouldPass(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	mp, _ := blproc.NewMetaProcessor(arguments)

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

	hash := []byte("aaa")
	ShardID := sharding.MetachainShardId

	mdp := initMetaDataPool()
	accounts := &mock.AccountsStub{}
	accounts.RevertToSnapshotCalled = func(snapshot int) error {
		return nil
	}
	arguments := createMockMetaArguments()
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	arguments.StartHeaders = createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))
	arguments.DataPool = mdp
	mp, _ := blproc.NewMetaProcessor(arguments)
	mdp.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		cs := &mock.Uint64SyncMapCacherStub{}
		cs.GetCalled = func(key uint64) (dataRetriever.ShardIdHashMap, bool) {
			syncMap := &dataPool.ShardIdHashSyncMap{}
			syncMap.Store(ShardID, hash)

			return syncMap, true
		}
		return cs
	}
	mp.AddHdrHashToRequestedList(&block.Header{}, []byte("header_hash"))
	mp.SetHighestHdrNonceForCurrentBlock(0, 1)
	mp.SetHighestHdrNonceForCurrentBlock(1, 2)
	mp.SetHighestHdrNonceForCurrentBlock(2, 3)
	res := mp.RequestMissingFinalityAttestingShardHeaders()
	assert.Equal(t, res, uint32(3))
}

//------- CommitBlock

func TestMetaProcessor_CommitBlockNilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	mp, _ := blproc.NewMetaProcessor(arguments)
	blk := block.Body{}
	err := mp.CommitBlock(nil, &block.MetaBlock{}, blk)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestMetaProcessor_CommitBlockMarshalizerFailForHeaderShouldErr(t *testing.T) {
	t.Parallel()

	accounts := &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	errMarshalizer := errors.New("failure")
	hdr := createMetaBlockHeader()
	body := block.Body{}
	marshalizer := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (i []byte, e error) {
			if reflect.DeepEqual(obj, hdr) {
				return nil, errMarshalizer
			}

			return []byte("obj"), nil
		},
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return nil
		},
	}
	arguments := createMockMetaArguments()
	arguments.Accounts = accounts
	arguments.Marshalizer = marshalizer
	mp, _ := blproc.NewMetaProcessor(arguments)
	blkc := createTestBlockchain()
	err := mp.CommitBlock(blkc, hdr, body)
	assert.Equal(t, errMarshalizer, err)
}

func TestMetaProcessor_CommitBlockStorageFailsForHeaderShouldErr(t *testing.T) {
	t.Parallel()

	wasCalled := false
	errPersister := errors.New("failure")
	marshalizer := &mock.MarshalizerMock{}
	accounts := &mock.AccountsStub{
		CommitCalled: func() (i []byte, e error) {
			return nil, nil
		},
	}
	hdr := createMetaBlockHeader()
	body := block.Body{}
	hdrUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			wasCalled = true
			return errPersister
		},
		GetCalled: func(key []byte) (i []byte, e error) {
			hdr, _ := marshalizer.Marshal(&block.MetaBlock{})
			return hdr, nil
		},
	}
	store := initStore()
	store.AddStorer(dataRetriever.MetaBlockUnit, hdrUnit)

	arguments := createMockMetaArguments()
	arguments.Accounts = accounts
	arguments.Store = store
	arguments.ForkDetector = &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte, isNotarizedShardStuck bool) error {
			return nil
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
	}
	mp, _ := blproc.NewMetaProcessor(arguments)

	blkc, _ := blockchain.NewMetaChain(
		generateTestCache(),
	)
	mp.SetHdrForCurrentBlock([]byte("hdr_hash1"), &block.Header{}, true)
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
	body := block.Body{}
	store := initStore()

	arguments := createMockMetaArguments()
	arguments.Accounts = accounts
	arguments.Store = store
	arguments.DataPool = mdp
	mp, _ := blproc.NewMetaProcessor(arguments)

	mdp.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		return nil
	}
	blkc := createTestBlockchain()
	err := mp.CommitBlock(blkc, hdr, body)

	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
}

func TestMetaProcessor_CommitBlockNoTxInPoolShouldErr(t *testing.T) {
	t.Parallel()

	mdp := initMetaDataPool()
	hdr := createMetaBlockHeader()
	body := block.Body{}
	accounts := &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	fd := &mock.ForkDetectorMock{}
	hasher := &mock.HasherStub{}
	store := initStore()

	arguments := createMockMetaArguments()
	arguments.DataPool = mdp
	arguments.Accounts = accounts
	arguments.ForkDetector = fd
	arguments.Store = store
	arguments.Hasher = hasher
	mp, _ := blproc.NewMetaProcessor(arguments)

	mdp.ShardHeadersCalled = func() storage.Cacher {
		return &mock.CacherStub{
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
			MaxSizeCalled: func() int {
				return 1000
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
	body := block.Body{}
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
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte, isNotarizedShardStuck bool) error {
			if header == hdr {
				forkDetectorAddCalled = true
				return nil
			}

			return errors.New("should have not got here")
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
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

	arguments := createMockMetaArguments()
	arguments.DataPool = mdp
	arguments.Accounts = accounts
	arguments.ForkDetector = fd
	arguments.Store = store
	arguments.Hasher = hasher
	mp, _ := blproc.NewMetaProcessor(arguments)

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
		cs.MaxSizeCalled = func() int {
			return 1000
		}
		cs.KeysCalled = func() [][]byte {
			return nil
		}
		return cs
	}

	blkc := createTestBlockchain()

	mp.SetHdrForCurrentBlock([]byte("hdr_hash1"), &block.Header{}, true)
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

	arguments := createMockMetaArguments()
	arguments.DataPool = mdp
	arguments.Store = initStore()
	mp, _ := blproc.NewMetaProcessor(arguments)

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
		cs.MaxSizeCalled = func() int {
			return 1000
		}
		return cs
	}

	header := createMetaBlockHeader()
	hdrsRequested, _ := mp.RequestBlockHeaders(header)
	assert.Equal(t, uint32(1), hdrsRequested)
}

func TestMetaProcessor_RemoveBlockInfoFromPoolShouldErrNilMetaBlockHeader(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.DataPool = initMetaDataPool()
	arguments.Store = initStore()
	mp, _ := blproc.NewMetaProcessor(arguments)

	err := mp.RemoveBlockInfoFromPool(nil)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilMetaBlockHeader)
}

func TestMetaProcessor_RemoveBlockInfoFromPoolShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.DataPool = initMetaDataPool()
	arguments.Store = initStore()
	mp, _ := blproc.NewMetaProcessor(arguments)

	header := createMetaBlockHeader()
	mp.SetHdrForCurrentBlock([]byte("hdr_hash1"), &block.Header{}, true)
	err := mp.RemoveBlockInfoFromPool(header)
	assert.Nil(t, err)
}

func TestMetaProcessor_ApplyBodyToHeaderShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled: func() int {
			return 0
		},
		RootHashCalled: func() ([]byte, error) {
			return []byte("root"), nil
		},
	}
	arguments.DataPool = initMetaDataPool()
	arguments.Store = initStore()
	mp, _ := blproc.NewMetaProcessor(arguments)

	hdr := &block.MetaBlock{}
	err := mp.ApplyBodyToHeader(hdr, nil)
	assert.Nil(t, err)
}

func TestMetaProcessor_ApplyBodyToHeaderShouldSetEpochStart(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled: func() int {
			return 0
		},
		RootHashCalled: func() ([]byte, error) {
			return []byte("root"), nil
		},
	}
	arguments.DataPool = initMetaDataPool()
	arguments.Store = initStore()
	mp, _ := blproc.NewMetaProcessor(arguments)

	metaBlk := &block.MetaBlock{TimeStamp: 12345}
	bodyHandler := block.Body{&block.MiniBlock{Type: 0}}
	err := mp.ApplyBodyToHeader(metaBlk, bodyHandler)
	assert.Nil(t, err)
}

func TestMetaProcessor_CommitBlockShouldRevertAccountStateWhenErr(t *testing.T) {
	t.Parallel()

	// set accounts dirty
	journalEntries := 3
	revToSnapshot := func(snapshot int) error {
		journalEntries = 0
		return nil
	}

	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		RevertToSnapshotCalled: revToSnapshot,
	}
	arguments.DataPool = initMetaDataPool()
	arguments.Store = initStore()
	mp, _ := blproc.NewMetaProcessor(arguments)

	err := mp.CommitBlock(nil, nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, 0, journalEntries)
}

func TestMetaProcessor_MarshalizedDataToBroadcastShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.Store = initStore()
	mp, _ := blproc.NewMetaProcessor(arguments)

	msh, mstx, err := mp.MarshalizedDataToBroadcast(&block.MetaBlock{}, block.Body{})
	assert.Nil(t, err)
	assert.NotNil(t, msh)
	assert.NotNil(t, mstx)
}

//------- receivedHeader

func TestMetaProcessor_ReceivedHeaderShouldDecreaseMissing(t *testing.T) {
	t.Parallel()

	pool := mock.NewMetaPoolsHolderFake()
	arguments := createMockMetaArguments()
	arguments.DataPool = pool
	arguments.Store = initStore()
	mp, _ := blproc.NewMetaProcessor(arguments)

	//add 3 tx hashes on requested list
	hdrHash1 := []byte("hdr hash 1")
	hdrHash2 := []byte("hdr hash 2")
	hdrHash3 := []byte("hdr hash 3")

	hdr2 := &block.Header{Nonce: 2}

	mp.AddHdrHashToRequestedList(nil, hdrHash1)
	mp.AddHdrHashToRequestedList(nil, hdrHash2)
	mp.AddHdrHashToRequestedList(nil, hdrHash3)

	//received txHash2
	pool.ShardHeaders().Put(hdrHash2, hdr2)

	time.Sleep(100 * time.Millisecond)

	assert.True(t, mp.IsHdrMissing(hdrHash1))
	assert.False(t, mp.IsHdrMissing(hdrHash2))
	assert.True(t, mp.IsHdrMissing(hdrHash3))
}

//------- createShardInfo

func TestMetaProcessor_CreateShardInfoShouldWorkNoHdrAddedNotValid(t *testing.T) {
	t.Parallel()

	pool := mock.NewMetaPoolsHolderFake()
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
	pool.ShardHeaders().Put(hdrHash1, &block.Header{
		Round:            1,
		Nonce:            45,
		ShardId:          0,
		MiniBlockHeaders: miniBlockHeaders1})
	pool.ShardHeaders().Put(hdrHash2, &block.Header{
		Round:            2,
		Nonce:            45,
		ShardId:          1,
		MiniBlockHeaders: miniBlockHeaders2})
	pool.ShardHeaders().Put(hdrHash3, &block.Header{
		Round:            3,
		Nonce:            45,
		ShardId:          2,
		MiniBlockHeaders: miniBlockHeaders3})

	noOfShards := uint32(5)
	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}
	arguments.DataPool = pool
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(noOfShards)
	arguments.StartHeaders = createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(noOfShards))
	arguments.Store = initStore()
	mp, _ := blproc.NewMetaProcessor(arguments)

	round := uint64(10)
	shardInfo, err := mp.CreateShardInfo(round)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(shardInfo))

	metaHdr := &block.MetaBlock{Round: round}
	_, err = mp.CreateBlockBody(metaHdr, func() bool {
		return true
	})
	shardInfo, err = mp.CreateShardInfo(round)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(shardInfo))
}

func TestMetaProcessor_CreateShardInfoShouldWorkNoHdrAddedNotFinal(t *testing.T) {
	t.Parallel()

	pool := mock.NewMetaPoolsHolderFake()
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

	noOfShards := uint32(5)
	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}
	arguments.DataPool = pool
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(noOfShards)
	arguments.StartHeaders = createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(noOfShards))
	arguments.Store = initStore()
	mp, _ := blproc.NewMetaProcessor(arguments)

	haveTime := func() bool { return true }

	prevRandSeed := []byte("prevrand")
	notarizedHdrs := mp.NotarizedHdrs()
	setLastNotarizedHdr(noOfShards, 9, 44, prevRandSeed, notarizedHdrs)

	//put the existing headers inside datapool
	prevHash, _ := mp.ComputeHeaderHash(mp.LastNotarizedHdrForShard(0).(*block.Header))
	pool.ShardHeaders().Put(hdrHash1, &block.Header{
		Round:            10,
		Nonce:            45,
		ShardId:          0,
		PrevRandSeed:     prevRandSeed,
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders1})

	prevHash, _ = mp.ComputeHeaderHash(mp.LastNotarizedHdrForShard(1).(*block.Header))
	pool.ShardHeaders().Put(hdrHash2, &block.Header{
		Round:            20,
		Nonce:            45,
		ShardId:          1,
		PrevRandSeed:     prevRandSeed,
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders2})

	prevHash, _ = mp.ComputeHeaderHash(mp.LastNotarizedHdrForShard(2).(*block.Header))
	pool.ShardHeaders().Put(hdrHash3, &block.Header{
		Round:            30,
		Nonce:            45,
		ShardId:          2,
		PrevRandSeed:     prevRandSeed,
		PrevHash:         prevHash,
		MiniBlockHeaders: miniBlockHeaders3})

	mp.SetShardBlockFinality(0)
	round := uint64(40)
	shardInfo, err := mp.CreateShardInfo(round)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(shardInfo))

	metaHdr := &block.MetaBlock{Round: round}
	_, err = mp.CreateBlockBody(metaHdr, haveTime)
	shardInfo, err = mp.CreateShardInfo(round)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(shardInfo))
}

func TestMetaProcessor_CreateShardInfoShouldWorkHdrsAdded(t *testing.T) {
	t.Parallel()

	pool := mock.NewMetaPoolsHolderFake()
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

	noOfShards := uint32(5)
	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}
	arguments.DataPool = pool
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(noOfShards)
	arguments.StartHeaders = createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(noOfShards))
	arguments.Store = initStore()
	mp, _ := blproc.NewMetaProcessor(arguments)

	haveTime := func() bool { return true }

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := mp.NotarizedHdrs()
	setLastNotarizedHdr(noOfShards, 9, 44, prevRandSeed, notarizedHdrs)

	headers := make([]*block.Header, 0)

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := mp.ComputeHeaderHash(mp.LastNotarizedHdrForShard(0).(*block.Header))
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

	pool.ShardHeaders().Put(hdrHash1, headers[0])
	pool.ShardHeaders().Put(hdrHash11, headers[1])

	// header shard 1
	prevHash, _ = mp.ComputeHeaderHash(mp.LastNotarizedHdrForShard(1).(*block.Header))
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

	pool.ShardHeaders().Put(hdrHash2, headers[2])
	pool.ShardHeaders().Put(hdrHash22, headers[3])

	// header shard 2
	prevHash, _ = mp.ComputeHeaderHash(mp.LastNotarizedHdrForShard(2).(*block.Header))
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

	pool.ShardHeaders().Put(hdrHash3, headers[4])
	pool.ShardHeaders().Put(hdrHash33, headers[5])

	mp.SetShardBlockFinality(1)
	round := uint64(15)
	shardInfo, err := mp.CreateShardInfo(round)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(shardInfo))

	metaHdr := &block.MetaBlock{Round: round}
	_, err = mp.CreateBlockBody(metaHdr, haveTime)
	shardInfo, err = mp.CreateShardInfo(round)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(shardInfo))
}

func TestMetaProcessor_CreateShardInfoEmptyBlockHDRRoundTooHigh(t *testing.T) {
	t.Parallel()

	pool := mock.NewMetaPoolsHolderFake()
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

	noOfShards := uint32(5)
	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}
	arguments.DataPool = pool
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(noOfShards)
	arguments.StartHeaders = createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(noOfShards))
	arguments.Store = initStore()
	mp, _ := blproc.NewMetaProcessor(arguments)

	haveTime := func() bool { return true }

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := mp.NotarizedHdrs()
	setLastNotarizedHdr(noOfShards, 9, 44, prevRandSeed, notarizedHdrs)

	headers := make([]*block.Header, 0)

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := mp.ComputeHeaderHash(mp.LastNotarizedHdrForShard(0).(*block.Header))
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

	pool.ShardHeaders().Put(hdrHash1, headers[0])
	pool.ShardHeaders().Put(hdrHash11, headers[1])

	// header shard 1
	prevHash, _ = mp.ComputeHeaderHash(mp.LastNotarizedHdrForShard(1).(*block.Header))
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

	pool.ShardHeaders().Put(hdrHash2, headers[2])
	pool.ShardHeaders().Put(hdrHash22, headers[3])

	// header shard 2
	prevHash, _ = mp.ComputeHeaderHash(mp.LastNotarizedHdrForShard(2).(*block.Header))
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

	pool.ShardHeaders().Put(hdrHash3, headers[4])
	pool.ShardHeaders().Put(hdrHash33, headers[5])

	mp.SetShardBlockFinality(1)
	round := uint64(20)
	shardInfo, err := mp.CreateShardInfo(round)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(shardInfo))

	metaHdr := &block.MetaBlock{Round: round}
	_, err = mp.CreateBlockBody(metaHdr, haveTime)
	shardInfo, err = mp.CreateShardInfo(round)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(shardInfo))
}

func TestMetaProcessor_RestoreBlockIntoPoolsShouldErrNilMetaBlockHeader(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.Store = initStore()
	mp, _ := blproc.NewMetaProcessor(arguments)

	err := mp.RestoreBlockIntoPools(nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilMetaBlockHeader)
}

func TestMetaProcessor_RestoreBlockIntoPoolsShouldWork(t *testing.T) {
	t.Parallel()

	pool := mock.NewMetaPoolsHolderFake()
	marshalizerMock := &mock.MarshalizerMock{}
	body := block.Body{}
	hdr := block.Header{Nonce: 1}
	buffHdr, _ := marshalizerMock.Marshal(hdr)
	hdrHash := []byte("hdr_hash1")

	store := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				RemoveCalled: func(key []byte) error {
					return nil
				},
				GetCalled: func(key []byte) ([]byte, error) {
					return buffHdr, nil
				},
			}
		},
	}

	arguments := createMockMetaArguments()
	arguments.DataPool = pool
	arguments.Store = store
	mp, _ := blproc.NewMetaProcessor(arguments)

	mhdr := createMetaBlockHeader()

	err := mp.RestoreBlockIntoPools(mhdr, body)

	hdrFromPool, _ := pool.ShardHeaders().Get(hdrHash)
	assert.Nil(t, err)
	assert.Equal(t, &hdr, hdrFromPool)
}

func TestMetaProcessor_CreateLastNotarizedHdrs(t *testing.T) {
	t.Parallel()

	pool := mock.NewMetaPoolsHolderFake()
	noOfShards := uint32(5)
	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}
	arguments.Hasher = &mock.HasherMock{}
	arguments.DataPool = pool
	arguments.Store = initStore()
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(noOfShards)
	arguments.StartHeaders = createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(noOfShards))
	mp, _ := blproc.NewMetaProcessor(arguments)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := mp.NotarizedHdrs()
	firstNonce := uint64(44)
	setLastNotarizedHdr(noOfShards, 9, firstNonce, prevRandSeed, notarizedHdrs)

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := mp.ComputeHeaderHash(mp.LastNotarizedHdrForShard(0).(*block.Header))
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
	shDataCurr := block.ShardData{ShardID: 0, HeaderHash: currHash}
	metaHdr.ShardInfo = make([]block.ShardData, 0)
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataCurr)
	shDataPrev := block.ShardData{ShardID: 0, HeaderHash: prevHash}
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataPrev)

	// test header not in pool and defer called
	err := mp.SaveLastNotarizedHeader(metaHdr)
	assert.Equal(t, process.ErrMissingHeader, err)
	notarizedHdrs = mp.NotarizedHdrs()
	assert.Equal(t, firstNonce, mp.LastNotarizedHdrForShard(currHdr.ShardId).GetNonce())

	// wrong header type in pool and defer called
	pool.ShardHeaders().Put(currHash, metaHdr)
	pool.ShardHeaders().Put(prevHash, prevHdr)
	mp.SetHdrForCurrentBlock(currHash, metaHdr, true)
	mp.SetHdrForCurrentBlock(prevHash, prevHdr, true)

	err = mp.SaveLastNotarizedHeader(metaHdr)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
	notarizedHdrs = mp.NotarizedHdrs()
	assert.Equal(t, firstNonce, mp.LastNotarizedHdrForShard(currHdr.ShardId).GetNonce())

	// put headers in pool
	pool.ShardHeaders().Put(currHash, currHdr)
	pool.ShardHeaders().Put(prevHash, prevHdr)
	mp.CreateBlockStarted()
	mp.SetHdrForCurrentBlock(currHash, currHdr, true)
	mp.SetHdrForCurrentBlock(prevHash, prevHdr, true)

	err = mp.SaveLastNotarizedHeader(metaHdr)
	assert.Nil(t, err)
	notarizedHdrs = mp.NotarizedHdrs()
	assert.Equal(t, currHdr, mp.LastNotarizedHdrForShard(currHdr.ShardId))
}

func TestMetaProcessor_CheckShardHeadersValidity(t *testing.T) {
	t.Parallel()

	pool := mock.NewMetaPoolsHolderFake()
	noOfShards := uint32(5)
	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}
	arguments.Hasher = &mock.HasherMock{}
	arguments.DataPool = pool
	arguments.Store = initStore()
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(noOfShards)
	arguments.StartHeaders = createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(noOfShards))

	argsHeaderValidator := blproc.ArgsHeaderValidator{
		Hasher:      arguments.Hasher,
		Marshalizer: arguments.Marshalizer,
	}
	arguments.HeaderValidator, _ = blproc.NewHeaderValidator(argsHeaderValidator)

	mp, _ := blproc.NewMetaProcessor(arguments)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := mp.NotarizedHdrs()
	setLastNotarizedHdr(noOfShards, 9, 44, prevRandSeed, notarizedHdrs)

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := mp.ComputeHeaderHash(mp.LastNotarizedHdrForShard(0).(*block.Header))
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
	pool.ShardHeaders().Put(currHash, currHdr)
	prevHash, _ = mp.ComputeHeaderHash(prevHdr)
	pool.ShardHeaders().Put(prevHash, prevHdr)
	wrongCurrHdr := &block.Header{
		Round:        11,
		Nonce:        48,
		ShardId:      0,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash")}
	wrongCurrHash, _ := mp.ComputeHeaderHash(wrongCurrHdr)
	pool.ShardHeaders().Put(wrongCurrHash, wrongCurrHdr)

	metaHdr := &block.MetaBlock{Round: 20}
	shDataCurr := block.ShardData{ShardID: 0, HeaderHash: wrongCurrHash}
	metaHdr.ShardInfo = make([]block.ShardData, 0)
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataCurr)
	shDataPrev := block.ShardData{ShardID: 0, HeaderHash: prevHash}
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataPrev)

	mp.SetHdrForCurrentBlock(wrongCurrHash, wrongCurrHdr, true)
	mp.SetHdrForCurrentBlock(prevHash, prevHdr, true)

	_, err := mp.CheckShardHeadersValidity(metaHdr)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)

	shDataCurr = block.ShardData{ShardID: 0, HeaderHash: currHash}
	metaHdr.ShardInfo = make([]block.ShardData, 0)
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataCurr)
	shDataPrev = block.ShardData{ShardID: 0, HeaderHash: prevHash}
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataPrev)

	mp.CreateBlockStarted()
	mp.SetHdrForCurrentBlock(currHash, currHdr, true)
	mp.SetHdrForCurrentBlock(prevHash, prevHdr, true)

	highestNonceHdrs, err := mp.CheckShardHeadersValidity(metaHdr)
	assert.Nil(t, err)
	assert.NotNil(t, highestNonceHdrs)
	assert.Equal(t, currHdr.Nonce, highestNonceHdrs[currHdr.ShardId].GetNonce())
}

func TestMetaProcessor_CheckShardHeadersValidityWrongNonceFromLastNoted(t *testing.T) {
	t.Parallel()

	pool := mock.NewMetaPoolsHolderFake()
	noOfShards := uint32(5)
	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}
	arguments.DataPool = pool
	arguments.Store = initStore()
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(noOfShards)
	arguments.StartHeaders = createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(noOfShards))
	mp, _ := blproc.NewMetaProcessor(arguments)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := mp.NotarizedHdrs()
	setLastNotarizedHdr(noOfShards, 9, 44, prevRandSeed, notarizedHdrs)

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
	pool.ShardHeaders().Put(currHash, currHdr)
	metaHdr := &block.MetaBlock{Round: 20}

	shDataCurr := block.ShardData{ShardID: 0, HeaderHash: currHash}
	metaHdr.ShardInfo = make([]block.ShardData, 0)
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataCurr)

	mp.SetHdrForCurrentBlock(currHash, currHdr, true)

	highestNonceHdrs, err := mp.CheckShardHeadersValidity(metaHdr)
	assert.Nil(t, highestNonceHdrs)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)
}

func TestMetaProcessor_CheckShardHeadersValidityRoundZeroLastNoted(t *testing.T) {
	t.Parallel()

	pool := mock.NewMetaPoolsHolderFake()

	noOfShards := uint32(5)
	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}
	arguments.DataPool = pool
	arguments.Store = initStore()
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(noOfShards)
	arguments.StartHeaders = createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(noOfShards))
	mp, _ := blproc.NewMetaProcessor(arguments)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := mp.NotarizedHdrs()
	setLastNotarizedHdr(noOfShards, 0, 0, prevRandSeed, notarizedHdrs)

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

	shDataCurr := block.ShardData{ShardID: 0, HeaderHash: currHash}
	metaHdr.ShardInfo = make([]block.ShardData, 0)
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shDataCurr)

	highestNonceHdrs, err := mp.CheckShardHeadersValidity(metaHdr)
	assert.Equal(t, 0, len(highestNonceHdrs))

	pool.ShardHeaders().Put(currHash, currHdr)
	mp.SetHdrForCurrentBlock(currHash, currHdr, true)
	highestNonceHdrs, err = mp.CheckShardHeadersValidity(metaHdr)
	assert.NotNil(t, highestNonceHdrs)
	assert.Nil(t, err)
	assert.Equal(t, currHdr.Nonce, highestNonceHdrs[currHdr.ShardId].GetNonce())
}

func TestMetaProcessor_CheckShardHeadersFinality(t *testing.T) {
	t.Parallel()

	pool := mock.NewMetaPoolsHolderFake()
	noOfShards := uint32(5)
	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}
	arguments.DataPool = pool
	arguments.Store = initStore()
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(noOfShards)
	arguments.StartHeaders = createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(noOfShards))
	mp, _ := blproc.NewMetaProcessor(arguments)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := mp.NotarizedHdrs()
	setLastNotarizedHdr(noOfShards, 9, 44, prevRandSeed, notarizedHdrs)

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := mp.ComputeHeaderHash(mp.LastNotarizedHdrForShard(0).(*block.Header))
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
	pool.ShardHeaders().Put(prevHash, nextWrongHdr)

	mp.SetShardBlockFinality(0)
	metaHdr := &block.MetaBlock{Round: 1}

	highestNonceHdrs := make(map[uint32]data.HeaderHandler)
	for i := uint32(0); i < noOfShards; i++ {
		highestNonceHdrs[i] = nil
	}

	err := mp.CheckShardHeadersFinality(highestNonceHdrs)
	assert.Equal(t, process.ErrNilBlockHeader, err)

	for i := uint32(0); i < noOfShards; i++ {
		highestNonceHdrs[i] = mp.LastNotarizedHdrForShard(i)
	}

	// should work for empty highest nonce hdrs - no hdrs added this round to metablock
	err = mp.CheckShardHeadersFinality(nil)
	assert.Nil(t, err)

	mp.SetShardBlockFinality(0)
	highestNonceHdrs = make(map[uint32]data.HeaderHandler, 0)
	highestNonceHdrs[0] = currHdr
	err = mp.CheckShardHeadersFinality(highestNonceHdrs)
	assert.Nil(t, err)

	mp.SetShardBlockFinality(1)
	err = mp.CheckShardHeadersFinality(highestNonceHdrs)
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

	nextHash, _ := mp.ComputeHeaderHash(nextHdr)
	pool.ShardHeaders().Put(nextHash, nextHdr)
	mp.SetHdrForCurrentBlock(nextHash, nextHdr, false)

	metaHdr.Round = 20
	err = mp.CheckShardHeadersFinality(highestNonceHdrs)
	assert.Nil(t, err)
}

func TestMetaProcessor_IsHdrConstructionValid(t *testing.T) {
	t.Parallel()

	pool := mock.NewMetaPoolsHolderFake()
	noOfShards := uint32(5)
	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}
	arguments.DataPool = pool
	arguments.Store = initStore()
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(noOfShards)
	arguments.StartHeaders = createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(noOfShards))
	mp, _ := blproc.NewMetaProcessor(arguments)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := mp.NotarizedHdrs()
	setLastNotarizedHdr(noOfShards, 9, 44, prevRandSeed, notarizedHdrs)

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := mp.ComputeHeaderHash(mp.LastNotarizedHdrForShard(0).(*block.Header))
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
	assert.Equal(t, err, process.ErrRootStateDoesNotMatch)

	currHdr.Nonce = 0
	prevHdr.Nonce = 0
	prevHdr.RootHash = nil
	err = mp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Nil(t, err)

	currHdr.Nonce = 46
	prevHdr.Nonce = 45
	prevHdr.Round = currHdr.Round + 1
	err = mp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrLowerRoundInBlock)

	prevHdr.Round = currHdr.Round - 1
	currHdr.Nonce = prevHdr.Nonce + 2
	err = mp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrWrongNonceInBlock)

	currHdr.Nonce = prevHdr.Nonce + 1
	currHdr.PrevHash = []byte("wronghash")
	err = mp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrBlockHashDoesNotMatch)

	prevHdr.RandSeed = []byte("randomwrong")
	currHdr.PrevHash, _ = mp.ComputeHeaderHash(prevHdr)
	err = mp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrRandSeedDoesNotMatch)

	currHdr.PrevHash = prevHash
	prevHdr.RandSeed = currRandSeed
	prevHdr.RootHash = []byte("prevRootHash")
	err = mp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Nil(t, err)
}

func TestMetaProcessor_IsShardHeaderValidFinal(t *testing.T) {
	t.Parallel()

	pool := mock.NewMetaPoolsHolderFake()
	noOfShards := uint32(5)

	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}
	arguments.DataPool = pool
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(noOfShards)
	arguments.StartHeaders = createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(noOfShards))
	arguments.Store = initStore()
	mp, _ := blproc.NewMetaProcessor(arguments)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := mp.NotarizedHdrs()
	setLastNotarizedHdr(noOfShards, 9, 44, prevRandSeed, notarizedHdrs)

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := mp.ComputeHeaderHash(mp.LastNotarizedHdrForShard(0).(*block.Header))
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

	mp.SetShardBlockFinality(0)
	valid, hdrIds = mp.IsShardHeaderValidFinal(currHdr, prevHdr, srtShardHdrs)
	assert.True(t, valid)
	assert.NotNil(t, hdrIds)

	mp.SetShardBlockFinality(1)
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

	marshalizerMock := &mock.MarshalizerMock{}
	arguments := createMockMetaArguments()
	mp, _ := blproc.NewMetaProcessor(arguments)
	body := block.Body{}
	message, err := marshalizerMock.Marshal(body)
	assert.Nil(t, err)

	dcdBlk := mp.DecodeBlockBody(nil)
	assert.Nil(t, dcdBlk)

	dcdBlk = mp.DecodeBlockBody(message)
	assert.Equal(t, body, dcdBlk)
}

func TestMetaProcessor_DecodeBlockHeader(t *testing.T) {
	t.Parallel()

	marshalizerMock := &mock.MarshalizerMock{}
	arguments := createMockMetaArguments()
	mp, _ := blproc.NewMetaProcessor(arguments)
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

func TestMetaProcessor_UpdateShardsHeadersNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	mp, _ := blproc.NewMetaProcessor(arguments)

	numberOfShards := uint32(4)
	type DataForMap struct {
		ShardID     uint32
		HeaderNonce uint64
	}
	testData := []DataForMap{
		{uint32(0), uint64(100)},
		{uint32(1), uint64(200)},
		{uint32(2), uint64(300)},
		{uint32(3), uint64(400)},
		{uint32(0), uint64(400)},
	}

	for i := range testData {
		mp.UpdateShardsHeadersNonce(testData[i].ShardID, testData[i].HeaderNonce)
	}

	shardsHeadersNonce := mp.GetShardsHeadersNonce()

	mapDates := make([]uint64, 0)

	//Get all data from map and put then in a slice
	for i := uint32(0); i < numberOfShards; i++ {
		mapDataI, _ := shardsHeadersNonce.Load(i)
		mapDates = append(mapDates, mapDataI.(uint64))

	}

	//Check data from map is stored correctly
	expectedData := []uint64{400, 200, 300, 400}
	for i := 0; i < int(numberOfShards); i++ {
		assert.Equal(t, expectedData[i], mapDates[i])
	}
}

func TestMetaProcessor_CreateMiniBlocksJournalLenNotZeroShouldErr(t *testing.T) {
	t.Parallel()

	accntAdapter := &mock.AccountsStub{
		JournalLenCalled: func() int {
			return 1
		},
	}
	arguments := createMockMetaArguments()
	arguments.Accounts = accntAdapter
	mp, _ := blproc.NewMetaProcessor(arguments)
	round := uint64(10)
	metaHdr := &block.MetaBlock{Round: round}

	bodyHandler, err := mp.CreateBlockBody(metaHdr, func() bool { return true })
	assert.Nil(t, bodyHandler)
	assert.Equal(t, process.ErrAccountStateDirty, err)
}

func TestMetaProcessor_CreateMiniBlocksNoTimeShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	mp, _ := blproc.NewMetaProcessor(arguments)
	round := uint64(10)
	metaHdr := &block.MetaBlock{Round: round}

	bodyHandler, err := mp.CreateBlockBody(metaHdr, func() bool { return false })
	assert.Nil(t, bodyHandler)
	assert.Equal(t, process.ErrTimeIsOut, err)
}

func TestMetaProcessor_CreateMiniBlocksNilTxPoolShouldErr(t *testing.T) {
	t.Parallel()

	dPool := initMetaDataPool()
	dPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return nil
	}

	arguments := createMockMetaArguments()
	arguments.DataPool = dPool
	mp, _ := blproc.NewMetaProcessor(arguments)
	round := uint64(10)

	metaHdr := &block.MetaBlock{Round: round}
	bodyHandler, err := mp.CreateBlockBody(metaHdr, func() bool { return true })
	assert.Nil(t, bodyHandler)
	assert.Equal(t, process.ErrNilTransactionPool, err)
}

func TestMetaProcessor_CreateMiniBlocksDestMe(t *testing.T) {
	t.Parallel()

	hash1 := []byte("hash1")
	hdrHash1Bytes := []byte("hdr_hash1")
	hdrHash2Bytes := []byte("hdr_hash2")
	expectedMiniBlock1 := &block.MiniBlock{TxHashes: [][]byte{hash1}}
	expectedMiniBlock2 := &block.MiniBlock{TxHashes: [][]byte{[]byte("hash2")}}
	dPool := initMetaDataPool()
	dPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dPool.ShardHeadersCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(hdrHash1Bytes, key) {
				return &block.Header{
					Nonce:            1,
					Round:            1,
					PrevRandSeed:     []byte("roothash"),
					MiniBlockHeaders: []block.MiniBlockHeader{{Hash: hash1, SenderShardID: 1}},
				}, true
			}
			if bytes.Equal(hdrHash2Bytes, key) {
				return &block.Header{Nonce: 2, Round: 2}, true
			}
			return nil, false
		}
		cs.LenCalled = func() int {
			return 0
		}
		cs.RemoveCalled = func(key []byte) {}
		cs.KeysCalled = func() [][]byte {
			return [][]byte{hdrHash1Bytes, hdrHash2Bytes}
		}
		cs.MaxSizeCalled = func() int {
			return 1000
		}
		return cs
	}

	txCoordinator := &mock.TransactionCoordinatorMock{
		CreateMbsAndProcessCrossShardTransactionsDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksHashes map[string]struct{}, maxTxRemaining uint32, maxMbRemaining uint32, haveTime func() bool) (slices block.MiniBlockSlice, u uint32, b bool) {
			return block.MiniBlockSlice{expectedMiniBlock1}, 0, true
		},
		CreateMbsAndProcessTransactionsFromMeCalled: func(maxTxRemaining uint32, maxMbRemaining uint32, haveTime func() bool) block.MiniBlockSlice {
			return block.MiniBlockSlice{expectedMiniBlock2}
		},
	}

	arguments := createMockMetaArguments()
	arguments.DataPool = dPool
	arguments.TxCoordinator = txCoordinator

	mp, _ := blproc.NewMetaProcessor(arguments)
	round := uint64(10)

	metaHdr := &block.MetaBlock{Round: round}
	bodyHandler, err := mp.CreateBlockBody(metaHdr, func() bool { return true })
	miniBlocks, _ := bodyHandler.(block.Body)

	assert.Equal(t, expectedMiniBlock1, miniBlocks[0])
	assert.Equal(t, expectedMiniBlock2, miniBlocks[1])
	assert.Nil(t, err)
}

func TestMetaProcessor_ProcessBlockWrongHeaderShouldErr(t *testing.T) {
	t.Parallel()

	journalLen := func() int { return 0 }
	revToSnapshot := func(snapshot int) error { return nil }
	blkc := &blockchain.MetaChain{}
	hdr := block.MetaBlock{
		Nonce:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
		RootHash:      []byte("roothash"),
		ShardInfo: []block.ShardData{
			{
				ShardID:               0,
				ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{{Hash: []byte("hash1")}},
			},
			{
				ShardID:               1,
				ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{{Hash: []byte("hash2")}},
			},
		},
	}
	body := block.Body{
		&block.MiniBlock{
			TxHashes:        [][]byte{[]byte("hashTx")},
			ReceiverShardID: 0,
		},
	}
	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revToSnapshot,
	}
	mp, _ := blproc.NewMetaProcessor(arguments)

	// should return err
	err := mp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestMetaProcessor_ProcessBlockNoShardHeadersReceivedShouldErr(t *testing.T) {
	t.Parallel()

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")
	hasher := &mock.HasherStub{}
	hasher.ComputeCalled = func(s string) []byte {
		return []byte("hash1")
	}
	journalLen := func() int { return 0 }
	revToSnapshot := func(snapshot int) error { return nil }
	blkc := &blockchain.MetaChain{}
	hdr := block.MetaBlock{
		Nonce:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
		RootHash:      []byte("roothash"),
		ShardInfo: []block.ShardData{
			{
				ShardID:               0,
				ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{{Hash: []byte("hashTx"), TxCount: 1}},
				TxCount:               1,
			},
			{
				ShardID:               0,
				ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{{Hash: hash2, TxCount: 1}},
				TxCount:               1,
			},
		},
		MiniBlockHeaders: []block.MiniBlockHeader{{Hash: hash1, SenderShardID: 0, TxCount: 1}},
	}
	body := block.Body{
		&block.MiniBlock{
			TxHashes:        [][]byte{hash1},
			ReceiverShardID: 0,
		},
	}
	arguments := createMockMetaArguments()
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revToSnapshot,
	}
	arguments.Hasher = hasher
	mp, _ := blproc.NewMetaProcessor(arguments)

	err := mp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrTimeIsOut, err)
}

func TestMetaProcessor_VerifyCrossShardMiniBlocksDstMe(t *testing.T) {
	t.Parallel()

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")
	hdrHash1Bytes := []byte("hdr_hash1")
	hdrHash2Bytes := []byte("hdr_hash2")
	miniBlock1 := &block.MiniBlock{TxHashes: [][]byte{hash1}}
	miniBlock2 := &block.MiniBlock{TxHashes: [][]byte{hash2}}
	dPool := initMetaDataPool()
	dPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dPool.ShardHeadersCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(hdrHash1Bytes, key) {
				return &block.Header{
					Nonce:            1,
					Round:            1,
					PrevRandSeed:     []byte("roothash"),
					MiniBlockHeaders: []block.MiniBlockHeader{{Hash: hash1, SenderShardID: 1}},
				}, true
			}
			if bytes.Equal(hdrHash2Bytes, key) {
				return &block.Header{Nonce: 2, Round: 2}, true
			}
			return nil, false
		}
		cs.LenCalled = func() int {
			return 0
		}
		cs.RemoveCalled = func(key []byte) {}
		cs.KeysCalled = func() [][]byte {
			return [][]byte{hdrHash1Bytes, hdrHash2Bytes}
		}
		cs.MaxSizeCalled = func() int {
			return 1000
		}
		return cs
	}

	txCoordinator := &mock.TransactionCoordinatorMock{
		CreateMbsAndProcessCrossShardTransactionsDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksHashes map[string]struct{}, maxTxRemaining uint32, maxMbRemaining uint32, haveTime func() bool) (slices block.MiniBlockSlice, u uint32, b bool) {
			return block.MiniBlockSlice{miniBlock1}, 0, true
		},
		CreateMbsAndProcessTransactionsFromMeCalled: func(maxTxRemaining uint32, maxMbRemaining uint32, haveTime func() bool) block.MiniBlockSlice {
			return block.MiniBlockSlice{miniBlock2}
		},
	}

	hdr := &block.MetaBlock{
		Nonce:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
		Round:         2,
		RootHash:      []byte("roothash"),
		ShardInfo: []block.ShardData{
			{
				HeaderHash:            hdrHash1Bytes,
				ShardID:               0,
				ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{{Hash: hash1, TxCount: 1}},
				TxCount:               1,
			},
			{
				ShardID:               0,
				ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{{Hash: hash2, TxCount: 1}},
				TxCount:               1,
			},
		},
		MiniBlockHeaders: []block.MiniBlockHeader{{Hash: hash1, SenderShardID: 0, TxCount: 1}},
	}

	arguments := createMockMetaArguments()
	arguments.DataPool = dPool
	arguments.TxCoordinator = txCoordinator

	mp, _ := blproc.NewMetaProcessor(arguments)
	round := uint64(10)

	metaHdr := &block.MetaBlock{Round: round}
	_, err := mp.CreateBlockBody(metaHdr, func() bool { return true })
	assert.Nil(t, err)

	err = mp.VerifyCrossShardMiniBlockDstMe(hdr)
	assert.Nil(t, err)
}

func TestMetaProcessor_CreateBlockCreateHeaderProcessBlock(t *testing.T) {
	t.Parallel()

	hash := []byte("hash1")
	hdrHash1Bytes := []byte("hdr_hash1")
	hrdHash2Bytes := []byte("hdr_hash2")
	hasher := &mock.HasherStub{}
	hasher.ComputeCalled = func(s string) []byte {
		return hash
	}
	miniBlock1 := &block.MiniBlock{TxHashes: [][]byte{hash}}
	dPool := initMetaDataPool()
	dPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dPool.ShardHeadersCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(hdrHash1Bytes, key) {
				return &block.Header{
					PrevHash:         []byte("hash1"),
					Nonce:            1,
					Round:            1,
					PrevRandSeed:     []byte("roothash"),
					MiniBlockHeaders: []block.MiniBlockHeader{{Hash: []byte("hash1"), SenderShardID: 1}},
				}, true
			}
			if bytes.Equal(hrdHash2Bytes, key) {
				return &block.Header{Nonce: 2, Round: 2}, true
			}
			return nil, false
		}
		cs.LenCalled = func() int {
			return 0
		}
		cs.RemoveCalled = func(key []byte) {}
		cs.KeysCalled = func() [][]byte {
			return [][]byte{hdrHash1Bytes, hrdHash2Bytes}
		}
		cs.MaxSizeCalled = func() int {
			return 1000
		}
		return cs
	}

	txCoordinator := &mock.TransactionCoordinatorMock{
		CreateMbsAndProcessCrossShardTransactionsDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksHashes map[string]struct{}, maxTxRemaining uint32, maxMbRemaining uint32, haveTime func() bool) (slices block.MiniBlockSlice, u uint32, b bool) {
			return block.MiniBlockSlice{miniBlock1}, 0, true
		},
	}

	arguments := createMockMetaArguments()
	arguments.DataPool = dPool
	arguments.TxCoordinator = txCoordinator
	arguments.Hasher = hasher

	mp, _ := blproc.NewMetaProcessor(arguments)
	round := uint64(10)

	metaHdr := &block.MetaBlock{Round: round}
	bodyHandler, err := mp.CreateBlockBody(metaHdr, func() bool { return true })
	assert.Nil(t, err)

	headerHandler := mp.CreateNewHeader()
	headerHandler.SetRound(uint64(1))
	headerHandler.SetNonce(1)
	headerHandler.SetPrevHash(hash)

	blkc := &mock.BlockChainMock{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.MetaBlock{Nonce: 0}
		},
	}
	err = mp.ProcessBlock(blkc, headerHandler, bodyHandler, func() time.Duration { return time.Second })
	assert.Nil(t, err)
}

func TestMetaProcessor_CreateEpochStartFromMetaBlockEpochIsNotStarted(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		IsEpochStartCalled: func() bool {
			return false
		},
	}

	mp, _ := blproc.NewMetaProcessor(arguments)

	epStart, err := mp.CreateEpochStartForMetablock()
	assert.Nil(t, err)

	emptyEpochStart := block.EpochStart{}
	assert.Equal(t, emptyEpochStart, *epStart)
}

func TestMetaProcessor_CreateEpochStartFromMetaBlockHashComputeIssueShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("err computing hash")

	arguments := createMockMetaArguments()
	arguments.Marshalizer = &mock.MarshalizerStub{
		// trigger an error on the Marshal method called from core's ComputeHash
		MarshalCalled: func(obj interface{}) (i []byte, e error) {
			return nil, expectedErr
		},
	}
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		IsEpochStartCalled: func() bool {
			return true
		},
	}

	mp, err := blproc.NewMetaProcessor(arguments)
	assert.Nil(t, err)

	epStart, err := mp.CreateEpochStartForMetablock()
	assert.Nil(t, epStart)
	assert.Equal(t, expectedErr, err)
}

func TestMetaProcessor_CreateEpochStartFromMetaBlockShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		IsEpochStartCalled: func() bool {
			return true
		},
	}

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")
	hdr := arguments.StartHeaders[0].(*block.Header)
	hdr.MetaBlockHashes = [][]byte{hash1, hash2}
	hdr.Nonce = 1
	arguments.StartHeaders[0] = hdr

	dPool := initMetaDataPool()
	dPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dPool.ShardHeadersCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			return &block.Header{
				PrevHash:         []byte("hash1"),
				Nonce:            1,
				Round:            1,
				PrevRandSeed:     []byte("roothash"),
				MiniBlockHeaders: []block.MiniBlockHeader{{Hash: []byte("hash1"), SenderShardID: 1}},
				MetaBlockHashes:  [][]byte{[]byte("hash1"), []byte("hash2")},
			}, true
		}
		cs.LenCalled = func() int {
			return 0
		}
		cs.RemoveCalled = func(key []byte) {}
		cs.KeysCalled = func() [][]byte {
			return [][]byte{[]byte("hdr1"), []byte("hdr2")}
		}
		cs.MaxSizeCalled = func() int {
			return 1000
		}
		return cs
	}
	arguments.DataPool = dPool
	arguments.PendingMiniBlocks = &mock.PendingMiniBlocksHandlerStub{
		PendingMiniBlockHeadersCalled: func(lastNotarizedHeaders []data.HeaderHandler) ([]block.ShardMiniBlockHeader, error) {
			var hdrs []block.ShardMiniBlockHeader
			hdrs = append(hdrs, block.ShardMiniBlockHeader{
				Hash:            hash1,
				ReceiverShardID: 0,
				SenderShardID:   1,
				TxCount:         2,
			})

			return hdrs, nil
		},
	}
	mp, _ := blproc.NewMetaProcessor(arguments)

	epStart, err := mp.CreateEpochStartForMetablock()
	assert.Nil(t, err)
	assert.NotNil(t, epStart)
	assert.Equal(t, hash1, epStart.LastFinalizedHeaders[0].LastFinishedMetaBlock)
	assert.Equal(t, hash2, epStart.LastFinalizedHeaders[0].FirstPendingMetaBlock)
}

func TestShardProcessor_getLastFinalizedMetaHashForShardMetaHashNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		IsEpochStartCalled: func() bool {
			return false
		},
	}

	mp, _ := blproc.NewMetaProcessor(arguments)
	round := uint64(10)

	shardHdr := &block.Header{Round: round}
	last, lastFinal, err := mp.GetLastFinalizedMetaHashForShard(shardHdr)
	assert.Nil(t, last)
	assert.Nil(t, lastFinal)
	assert.Equal(t, nil, err)
}

func TestShardProcessor_getLastFinalizedMetaHashForShardShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockMetaArguments()
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		IsEpochStartCalled: func() bool {
			return false
		},
	}

	dPool := initMetaDataPool()
	dPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dPool.ShardHeadersCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			return &block.Header{
				PrevHash:         []byte("hash1"),
				Nonce:            2,
				Round:            2,
				PrevRandSeed:     []byte("roothash"),
				MiniBlockHeaders: []block.MiniBlockHeader{{Hash: []byte("hash1"), SenderShardID: 1}},
				MetaBlockHashes:  [][]byte{[]byte("hash1"), []byte("hash2")},
			}, true
		}
		cs.LenCalled = func() int {
			return 0
		}
		cs.RemoveCalled = func(key []byte) {}
		cs.KeysCalled = func() [][]byte {
			return [][]byte{[]byte("hdr1"), []byte("hdr2")}
		}
		cs.MaxSizeCalled = func() int {
			return 1000
		}
		return cs
	}

	arguments.DataPool = dPool

	mp, _ := blproc.NewMetaProcessor(arguments)
	round := uint64(10)
	nonce := uint64(1)

	shardHdr := &block.Header{
		Round:           round,
		Nonce:           nonce,
		MetaBlockHashes: [][]byte{[]byte("mb_hash1")},
	}
	last, lastFinal, err := mp.GetLastFinalizedMetaHashForShard(shardHdr)
	assert.NotNil(t, last)
	assert.NotNil(t, lastFinal)
	assert.Nil(t, err)
}
