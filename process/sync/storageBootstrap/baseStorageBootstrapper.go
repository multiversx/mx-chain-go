package storageBootstrap

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/block/processedMb"
	"github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("process/sync")

// ArgsBaseStorageBootstrapper is structure used to create a new storage bootstrapper
type ArgsBaseStorageBootstrapper struct {
	BootStorer          process.BootStorer
	ForkDetector        process.ForkDetector
	BlockProcessor      process.BlockProcessor
	ChainHandler        data.ChainHandler
	Marshalizer         marshal.Marshalizer
	Store               dataRetriever.StorageService
	Uint64Converter     typeConverters.Uint64ByteSliceConverter
	BootstrapRoundIndex uint64
	ShardCoordinator    sharding.Coordinator
	ResolversFinder     dataRetriever.ResolversFinder
	BlockTracker        process.BlockTracker
}

// ArgsShardStorageBootstrapper is structure used to create a new storage bootstrapper for shard
type ArgsShardStorageBootstrapper struct {
	ArgsBaseStorageBootstrapper
}

// ArgsMetaStorageBootstrapper is structure used to create a new storage bootstrapper for metachain
type ArgsMetaStorageBootstrapper struct {
	ArgsBaseStorageBootstrapper
	PendingMiniBlocksHandler process.PendingMiniBlocksHandler
}

type storageBootstrapper struct {
	bootStorer       process.BootStorer
	forkDetector     process.ForkDetector
	blkExecutor      process.BlockProcessor
	blkc             data.ChainHandler
	marshalizer      marshal.Marshalizer
	store            dataRetriever.StorageService
	uint64Converter  typeConverters.Uint64ByteSliceConverter
	shardCoordinator sharding.Coordinator
	blockTracker     process.BlockTracker

	bootstrapRoundIndex  uint64
	bootstrapper         storageBootstrapperHandler
	headerNonceHashStore storage.Storer
	highestNonce         uint64
}

func (st *storageBootstrapper) loadBlocks() error {
	var err error
	var headerInfo bootstrapStorage.BootstrapData

	round := st.bootStorer.GetHighestRound()
	storageHeadersInfo := make([]bootstrapStorage.BootstrapData, 0)

	log.Debug("Load blocks started...")

	for {
		headerInfo, err = st.bootStorer.Get(round)
		if err != nil {
			break
		}

		if round == headerInfo.LastRound {
			err = sync.ErrCorruptBootstrapFromStorageDb
			break
		}

		storageHeadersInfo = append(storageHeadersInfo, headerInfo)

		if uint64(round) > st.bootstrapRoundIndex {
			round = headerInfo.LastRound
			continue
		}

		err = st.applyHeaderInfo(headerInfo)
		if err != nil {
			round = headerInfo.LastRound
			continue
		}

		var bootInfos []bootstrapStorage.BootstrapData
		bootInfos, err = st.getBootInfos(headerInfo)
		if err != nil {
			round = headerInfo.LastRound
			continue
		}

		err = st.applyBootInfos(bootInfos)
		if err != nil {
			round = headerInfo.LastRound
			continue
		}

		break
	}

	if err != nil {
		log.Warn("bootstrapper", "error", err)
		st.restoreBlockChainToGenesis()
		err = st.bootStorer.SaveLastRound(0)
		log.LogIfError(err, "bootstorer")

		return process.ErrNotEnoughValidBlocksInStorage
	}

	st.bootstrapper.applyNumPendingMiniBlocks(headerInfo.PendingMiniBlocks)

	processedMiniBlocks := processedMb.NewProcessedMiniBlocks()
	processedMiniBlocks.ConvertSliceToProcessedMiniBlocksMap(headerInfo.ProcessedMiniBlocks)
	processedMiniBlocks.DisplayProcessedMiniBlocks()

	st.blkExecutor.ApplyProcessedMiniBlocks(processedMiniBlocks)

	for i := 0; i < len(storageHeadersInfo)-1; i++ {
		st.cleanupStorage(storageHeadersInfo[i].LastHeader)
		st.bootstrapper.cleanupNotarizedStorage(storageHeadersInfo[i].LastHeader.Hash)
	}

	err = st.bootStorer.SaveLastRound(round)
	if err != nil {
		log.Debug("cannot save last round in storage ", "error", err.Error())
	}

	st.highestNonce = headerInfo.LastHeader.Nonce

	return nil
}

// GetHighestBlockNonce will return nonce of last block loaded from storage
func (st *storageBootstrapper) GetHighestBlockNonce() uint64 {
	return st.highestNonce
}

func (st *storageBootstrapper) applyHeaderInfo(hdrInfo bootstrapStorage.BootstrapData) error {
	headerHash := hdrInfo.LastHeader.Hash
	headerFromStorage, err := st.bootstrapper.getHeader(headerHash)
	if err != nil {
		log.Debug("cannot get header ", "nonce", hdrInfo.LastHeader.Nonce,
			"error", err.Error())
		return err
	}

	err = st.blkExecutor.RevertStateToBlock(headerFromStorage)
	if err != nil {
		log.Debug("cannot recreate trie for header with nonce", "nonce", headerFromStorage.GetNonce())
		return err
	}

	err = st.applyBlock(headerFromStorage, headerHash)
	if err != nil {
		log.Debug("cannot apply block for header ", "nonce", headerFromStorage.GetNonce(),
			"error", err.Error())
		return err
	}

	return nil
}

func (st *storageBootstrapper) getBootInfos(hdrInfo bootstrapStorage.BootstrapData) ([]bootstrapStorage.BootstrapData, error) {
	highestFinalBlockNonce := hdrInfo.HighestFinalBlockNonce
	highestBlockNonce := hdrInfo.LastHeader.Nonce

	lastRound := hdrInfo.LastRound
	bootInfos := []bootstrapStorage.BootstrapData{hdrInfo}

	log.Debug("block info from storage",
		"highest block nonce", highestBlockNonce,
		"highest final block nonce", highestFinalBlockNonce,
		"last round", lastRound)

	lowestNonce := core.MaxUint64(highestFinalBlockNonce-1, 1)
	for highestBlockNonce > lowestNonce {
		strHdrI, err := st.bootStorer.Get(lastRound)
		if err != nil {
			log.Debug("cannot load header info from storage ", "error", err.Error())
			return nil, err
		}

		bootInfos = append(bootInfos, strHdrI)
		highestBlockNonce = strHdrI.LastHeader.Nonce

		lastRound = strHdrI.LastRound
		if lastRound == 0 {
			break
		}
	}

	return bootInfos, nil
}

func (st *storageBootstrapper) applyBootInfos(bootInfos []bootstrapStorage.BootstrapData) error {
	var err error

	defer func() {
		if err != nil {
			st.forkDetector.RestoreToGenesis()
			st.blockTracker.RestoreToGenesis()
		}
	}()

	for i := len(bootInfos) - 1; i >= 0; i-- {
		log.Debug("apply header",
			"shard", bootInfos[i].LastHeader.ShardId,
			"nonce", bootInfos[i].LastHeader.Nonce)

		err = st.bootstrapper.applyCrossNotarizedHeaders(bootInfos[i].LastCrossNotarizedHeaders)
		if err != nil {
			log.Debug("cannot apply cross notarized headers", "error", err.Error())
			return err
		}

		selfNotarizedHeadersHashes := make([][]byte, len(bootInfos[i].LastSelfNotarizedHeaders))
		for index, selfNotarizedHeader := range bootInfos[i].LastSelfNotarizedHeaders {
			selfNotarizedHeadersHashes[index] = selfNotarizedHeader.Hash
		}

		var selfNotarizedHeaders []data.HeaderHandler
		selfNotarizedHeaders, err = st.bootstrapper.applySelfNotarizedHeaders(selfNotarizedHeadersHashes)
		if err != nil {
			log.Debug("cannot apply self notarized headers", "error", err.Error())
			return err
		}

		var header data.HeaderHandler
		header, err = st.bootstrapper.getHeader(bootInfos[i].LastHeader.Hash)
		if err != nil {
			return err
		}

		log.Debug("add header to fork detector",
			"shard", header.GetShardID(),
			"round", header.GetRound(),
			"nonce", header.GetNonce(),
			"hash", bootInfos[i].LastHeader.Hash)

		err = st.forkDetector.AddHeader(header, bootInfos[i].LastHeader.Hash, process.BHProcessed, selfNotarizedHeaders, selfNotarizedHeadersHashes)
		if err != nil {
			return err
		}

		if i > 0 {
			log.Debug("added self notarized header in block tracker",
				"shard", header.GetShardID(),
				"round", header.GetRound(),
				"nonce", header.GetNonce(),
				"hash", bootInfos[i].LastHeader.Hash)

			st.blockTracker.AddSelfNotarizedHeader(st.shardCoordinator.SelfId(), header, bootInfos[i].LastHeader.Hash)
		}

		st.blockTracker.AddTrackedHeader(header, bootInfos[i].LastHeader.Hash)
	}

	return nil
}

func (st *storageBootstrapper) cleanupStorage(headerInfo bootstrapStorage.BootstrapHeaderInfo) {
	log.Debug("cleanup storage")

	nonceToByteSlice := st.uint64Converter.ToByteSlice(headerInfo.Nonce)
	err := st.headerNonceHashStore.Remove(nonceToByteSlice)
	if err != nil {
		log.Debug("block was not removed from storage",
			"shradId", headerInfo.ShardId,
			"nonce", headerInfo.Nonce,
			"hash", headerInfo.Hash,
			"error", err.Error())
		return
	}

	log.Debug("block was removed from storage",
		"shradId", headerInfo.ShardId,
		"nonce", headerInfo.Nonce,
		"hash", headerInfo.Hash)
}

func (st *storageBootstrapper) applyBlock(header data.HeaderHandler, headerHash []byte) error {
	blockBody, err := st.bootstrapper.getBlockBody(header)
	if err != nil {
		return err
	}

	err = st.blkc.SetCurrentBlockBody(blockBody)
	if err != nil {
		return err
	}

	err = st.blkc.SetCurrentBlockHeader(header)
	if err != nil {
		return err
	}

	st.blkc.SetCurrentBlockHeaderHash(headerHash)

	return nil
}

func (st *storageBootstrapper) restoreBlockChainToGenesis() {
	genesisHeader := st.blkc.GetGenesisHeader()
	err := st.blkExecutor.RevertStateToBlock(genesisHeader)
	if err != nil {
		log.Debug("cannot recreate trie for header with nonce", "nonce", genesisHeader.GetNonce())
	}

	err = st.blkc.SetCurrentBlockHeader(nil)
	if err != nil {
		log.Debug("cannot set current block header", "error", err.Error())
	}

	err = st.blkc.SetCurrentBlockBody(nil)
	if err != nil {
		log.Debug("cannot set current block body", "error", err.Error())
	}

	st.blkc.SetCurrentBlockHeaderHash(nil)
}
