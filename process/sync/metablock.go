package sync

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// MetaBootstrap implements the bootstrap mechanism
type MetaBootstrap struct {
	*baseBootstrap

	resolversFinder dataRetriever.ResolversFinder
	hdrRes          dataRetriever.HeaderResolver
}

// NewMetaBootstrap creates a new Bootstrap object
func NewMetaBootstrap(
	poolsHolder dataRetriever.MetaPoolsHolder,
	store dataRetriever.StorageService,
	blkc data.ChainHandler,
	rounder consensus.Rounder,
	blkExecutor process.BlockProcessor,
	waitTime time.Duration,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	forkDetector process.ForkDetector,
	resolversFinder dataRetriever.ResolversFinder,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	bootstrapRoundIndex uint32,
) (*MetaBootstrap, error) {

	if poolsHolder == nil {
		return nil, process.ErrNilPoolsHolder
	}
	if poolsHolder.MetaBlockNonces() == nil {
		return nil, process.ErrNilHeadersNoncesDataPool
	}
	if poolsHolder.MetaChainBlocks() == nil {
		return nil, process.ErrNilMetaBlockPool
	}

	err := checkBootstrapNilParameters(
		blkc,
		rounder,
		blkExecutor,
		hasher,
		marshalizer,
		forkDetector,
		resolversFinder,
		shardCoordinator,
		accounts,
		store,
	)
	if err != nil {
		return nil, err
	}

	base := &baseBootstrap{
		blkc:                blkc,
		blkExecutor:         blkExecutor,
		store:               store,
		headers:             poolsHolder.MetaChainBlocks(),
		headersNonces:       poolsHolder.MetaBlockNonces(),
		rounder:             rounder,
		waitTime:            waitTime,
		hasher:              hasher,
		marshalizer:         marshalizer,
		forkDetector:        forkDetector,
		shardCoordinator:    shardCoordinator,
		accounts:            accounts,
		bootstrapRoundIndex: bootstrapRoundIndex,
	}

	boot := MetaBootstrap{
		baseBootstrap: base,
	}

	//there is one header topic so it is ok to save it
	hdrResolver, err := resolversFinder.MetaChainResolver(factory.MetachainBlocksTopic)
	if err != nil {
		return nil, err
	}

	//placed in struct fields for performance reasons
	hdrRes, ok := hdrResolver.(dataRetriever.HeaderResolver)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}
	boot.hdrRes = hdrRes

	boot.chRcvHdr = make(chan bool)

	boot.setRequestedHeaderNonce(nil)
	boot.headersNonces.RegisterHandler(boot.receivedHeaderNonce)
	boot.headers.RegisterHandler(boot.receivedHeader)

	boot.chStopSync = make(chan bool)

	boot.syncStateListeners = make([]func(bool), 0)
	boot.requestedHashes = process.RequiredDataPool{}

	//TODO: This should be injected when BlockProcessor will be refactored
	boot.uint64Converter = uint64ByteSlice.NewBigEndianConverter()

	return &boot, nil
}

func (boot *MetaBootstrap) syncFromStorer(
	blockFinality uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
	notarizedBlockFinality uint64,
	notarizedHdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	err := boot.loadBlocks(blockFinality,
		blockUnit,
		hdrNonceHashDataUnit,
		boot.getHeaderFromStorage,
		boot.getBlockBody,
		boot.removeBlockBody)
	if err != nil {
		return err
	}

	for i := uint32(0); i < boot.shardCoordinator.NumberOfShards(); i++ {
		err = boot.loadNotarizedBlocks(notarizedBlockFinality,
			notarizedHdrNonceHashDataUnit+dataRetriever.UnitType(i),
			boot.applyNotarizedBlock)
		if err != nil {
			return err
		}
	}

	return nil
}

func (boot *MetaBootstrap) removeBlockBody(
	nonce uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	return nil
}

func (boot *MetaBootstrap) applyNotarizedBlock(
	nonce uint64,
	notarizedHdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	nonceToByteSlice := boot.uint64Converter.ToByteSlice(nonce)
	headerHash, err := boot.store.Get(notarizedHdrNonceHashDataUnit, nonceToByteSlice)
	if err != nil {
		return err
	}

	header, err := process.GetShardHeaderFromStorage(headerHash, boot.marshalizer, boot.store)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("apply notarized block with nonce %d and round %d\n", header.Nonce, header.Round))

	if header.GetRound() > boot.bootstrapRoundIndex {
		return ErrHigherBootstrapRound
	}

	boot.blkExecutor.SetLastNotarizedHdr(header.ShardId, header)

	return nil
}

func (boot *MetaBootstrap) getHeaderFromStorage(nonce uint64) (data.HeaderHandler, []byte, error) {
	nonceToByteSlice := boot.uint64Converter.ToByteSlice(nonce)
	headerHash, err := boot.store.Get(dataRetriever.MetaHdrNonceHashDataUnit, nonceToByteSlice)
	if err != nil {
		return nil, nil, err
	}

	header, err := process.GetMetaHeaderFromStorage(headerHash, boot.marshalizer, boot.store)

	return header, headerHash, err
}

func (boot *MetaBootstrap) getBlockBody(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
	return &block.MetaBlockBody{}, nil
}

func (boot *MetaBootstrap) receivedHeader(headerHash []byte) {
	header, err := process.GetMetaHeader(headerHash, boot.headers, boot.marshalizer, boot.store)
	if err != nil {
		log.Debug(err.Error())
		return
	}

	boot.processReceivedHeader(header, headerHash)
}

// StartSync method will start SyncBlocks as a go routine
func (boot *MetaBootstrap) StartSync() {
	// when a node starts it first tries to bootstrap from storage, if there already exist a database saved
	err := boot.syncFromStorer(process.MetaBlockFinality,
		dataRetriever.MetaBlockUnit,
		dataRetriever.MetaHdrNonceHashDataUnit,
		process.ShardBlockFinality,
		dataRetriever.ShardHdrNonceHashDataUnit)
	if err != nil {
		log.Info(err.Error())
	}

	go boot.syncBlocks()
}

// StopSync method will stop SyncBlocks
func (boot *MetaBootstrap) StopSync() {
	boot.chStopSync <- true
}

// syncBlocks method calls repeatedly synchronization method SyncBlock
func (boot *MetaBootstrap) syncBlocks() {
	for {
		time.Sleep(sleepTime)
		select {
		case <-boot.chStopSync:
			return
		default:
			err := boot.SyncBlock()
			if err != nil {
				log.Info(err.Error())
			}
		}
	}
}

// SyncBlock method actually does the synchronization. It requests the next block header from the pool
// and if it is not found there it will be requested from the network. After the header is received,
// it requests the block body in the same way(pool and than, if it is not found in the pool, from network).
// If either header and body are received the ProcessAndCommit method will be called. This method will execute
// the block and its transactions. Finally if everything works, the block will be committed in the blockchain,
// and all this mechanism will be reiterated for the next block.
func (boot *MetaBootstrap) SyncBlock() error {
	if !boot.ShouldSync() {
		return nil
	}

	if boot.isForkDetected {
		log.Info(fmt.Sprintf("fork detected at nonce %d\n", boot.forkNonce))
		return boot.forkChoice()
	}

	boot.setRequestedHeaderNonce(nil)

	nonce := boot.getNonceForNextBlock()

	hdr, err := boot.getHeaderRequestingIfMissing(nonce)
	if err != nil {
		return err
	}

	haveTime := func() time.Duration {
		return boot.rounder.TimeDuration()
	}

	blockBody := &block.MetaBlockBody{}
	err = boot.blkExecutor.ProcessBlock(boot.blkc, hdr, blockBody, haveTime)
	if err != nil {
		isForkDetected := err == process.ErrInvalidBlockHash || err == process.ErrRootStateMissmatch
		if isForkDetected {
			log.Info(err.Error())
			boot.removeHeaderFromPools(hdr)
			err = boot.forkChoice()
		}

		return err
	}

	timeBefore := time.Now()
	err = boot.blkExecutor.CommitBlock(boot.blkc, hdr, blockBody)
	if err != nil {
		return err
	}
	timeAfter := time.Now()
	log.Info(fmt.Sprintf("time elapsed to commit block: %v sec\n", timeAfter.Sub(timeBefore).Seconds()))

	log.Info(fmt.Sprintf("block with nonce %d has been synced successfully\n", hdr.Nonce))
	return nil
}

func (boot *MetaBootstrap) getHeaderWithNonce(nonce uint64) (*block.MetaBlock, error) {
	var hash []byte
	hdr, err := boot.getHeaderFromPoolWithNonce(nonce)
	if err != nil {
		hash, err = boot.getHeaderHashFromStorage(nonce)
		if err != nil {
			return nil, err
		}

		hdr, err = process.GetMetaHeaderFromStorage(hash, boot.marshalizer, boot.store)
		if err != nil {
			return nil, err
		}
	}

	return hdr, nil
}

// getHeaderFromPoolWithNonce method returns the block header from a given nonce
func (boot *MetaBootstrap) getHeaderFromPoolWithNonce(nonce uint64) (*block.MetaBlock, error) {
	value, _ := boot.headersNonces.Get(nonce)
	hash, ok := value.([]byte)

	if hash == nil || !ok {
		return nil, process.ErrMissingHashForHeaderNonce
	}

	obj, ok := boot.headers.Peek(hash)
	if !ok {
		return nil, process.ErrMissingHeader
	}

	hdr, ok := obj.(*block.MetaBlock)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return hdr, nil
}

// getHeaderHashFromStorage method returns the block header hash from a given nonce
func (boot *MetaBootstrap) getHeaderHashFromStorage(nonce uint64) ([]byte, error) {
	headerStore := boot.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	if headerStore == nil {
		return nil, process.ErrNilHeadersStorage
	}

	nonceToByteSlice := boot.uint64Converter.ToByteSlice(nonce)
	headerHash, err := headerStore.Get(nonceToByteSlice)
	if err != nil {
		return nil, process.ErrMissingHashForHeaderNonce
	}

	return headerHash, nil
}

// requestHeader method requests a block header from network when it is not found in the pool
func (boot *MetaBootstrap) requestHeader(nonce uint64) {
	boot.setRequestedHeaderNonce(&nonce)
	err := boot.hdrRes.RequestDataFromNonce(nonce)

	log.Info(fmt.Sprintf("requested header with nonce %d from network\n", nonce))

	if err != nil {
		log.Error(err.Error())
	}
}

// getHeaderWithNonce method gets the header with given nonce from pool, if it exist there,
// and if not it will be requested from network
func (boot *MetaBootstrap) getHeaderRequestingIfMissing(nonce uint64) (*block.MetaBlock, error) {
	hdr, err := boot.getHeaderFromPoolWithNonce(nonce)
	if err != nil {
		process.EmptyChannel(boot.chRcvHdr)
		boot.requestHeader(nonce)
		err := boot.waitForHeaderNonce()
		if err != nil {
			return nil, err
		}

		hdr, err = boot.getHeaderFromPoolWithNonce(nonce)
		if err != nil {
			return nil, err
		}
	}

	return hdr, nil
}

// forkChoice decides if rollback must be called
func (boot *MetaBootstrap) forkChoice() error {
	log.Info("starting fork choice\n")
	isForkResolved := false
	for !isForkResolved {
		header, err := boot.getCurrentHeader()
		if err != nil {
			return err
		}

		msg := fmt.Sprintf("roll back to header with nonce %d and hash %s",
			header.GetNonce()-1, core.ToB64(header.GetPrevHash()))

		isSigned := isSigned(header)
		if isSigned {
			msg = fmt.Sprintf("%s from a signed block, as the highest final block nonce is %d",
				msg,
				boot.forkDetector.GetHighestFinalBlockNonce())
			canRevertBlock := header.GetNonce() > boot.forkDetector.GetHighestFinalBlockNonce()
			if !canRevertBlock {
				return &ErrSignedBlock{CurrentNonce: header.GetNonce()}
			}
		}

		log.Info(msg + "\n")

		err = boot.rollback(header)
		if err != nil {
			return err
		}

		if header.GetNonce() <= boot.forkNonce {
			isForkResolved = true
		}
	}

	log.Info("ending fork choice\n")
	return nil
}

func (boot *MetaBootstrap) rollback(header *block.MetaBlock) error {
	if header.GetNonce() == 0 {
		return process.ErrRollbackFromGenesis
	}

	headerStore := boot.store.GetStorer(dataRetriever.MetaBlockUnit)
	if headerStore == nil {
		return process.ErrNilHeadersStorage
	}

	headerNonceHashStore := boot.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	if headerNonceHashStore == nil {
		return process.ErrNilHeadersNonceHashStorage
	}

	var err error
	var newHeader *block.MetaBlock
	var newHeaderHash []byte
	var newRootHash []byte

	if header.GetNonce() > 1 {
		newHeader, err = boot.getPrevHeader(headerStore, header)
		if err != nil {
			return err
		}

		newHeaderHash = header.GetPrevHash()
		newRootHash = newHeader.GetRootHash()
	} else { // rollback to genesis block
		newRootHash = boot.blkc.GetGenesisHeader().GetRootHash()
	}

	err = boot.blkc.SetCurrentBlockHeader(newHeader)
	if err != nil {
		return err
	}

	boot.blkc.SetCurrentBlockHeaderHash(newHeaderHash)

	err = boot.accounts.RecreateTrie(newRootHash)
	if err != nil {
		return err
	}

	boot.cleanCachesOnRollback(header, headerStore, headerNonceHashStore)
	errNotCritical := boot.blkExecutor.RestoreBlockIntoPools(header, nil)
	if errNotCritical != nil {
		log.Info(errNotCritical.Error())
	}

	return nil
}

func (boot *MetaBootstrap) getPrevHeader(headerStore storage.Storer, header *block.MetaBlock) (*block.MetaBlock, error) {
	prevHash := header.GetPrevHash()
	buffHeader, err := headerStore.Get(prevHash)
	if err != nil {
		return nil, err
	}

	newHeader := &block.MetaBlock{}
	err = boot.marshalizer.Unmarshal(newHeader, buffHeader)
	if err != nil {
		return nil, err
	}

	return newHeader, nil
}

func (boot *MetaBootstrap) getCurrentHeader() (*block.MetaBlock, error) {
	blockHeader := boot.blkc.GetCurrentBlockHeader()
	if blockHeader == nil {
		return nil, process.ErrNilBlockHeader
	}

	header, ok := blockHeader.(*block.MetaBlock)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return header, nil
}
