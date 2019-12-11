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

// ArgsStorageBootstrapper is structure used to create a new storage bootstrapper
type ArgsStorageBootstrapper struct {
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

		bootInfos, err := st.getBootInfos(headerInfo)
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
		_ = st.bootStorer.SaveLastRound(0)
		return process.ErrNotEnoughValidBlocksInStorage
	}

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
	highestFinalNonce := hdrInfo.HighestFinalNonce
	highestNonce := hdrInfo.LastHeader.Nonce

	lastRound := hdrInfo.LastRound
	bootInfos := make([]bootstrapStorage.BootstrapData, 0)
	bootInfos = append(bootInfos, hdrInfo)

	log.Debug("block info from storage",
		"highest nonce", highestNonce, "lastFinalNone", highestFinalNonce, "last round", lastRound)

	lowestNonce := core.MaxUint64(highestFinalNonce-1, 1)
	for highestNonce > lowestNonce {
		strHdrI, err := st.bootStorer.Get(lastRound)
		if err != nil {
			log.Debug("cannot load header info from storage ", "error", err.Error())
			return nil, err
		}

		bootInfos = append(bootInfos, strHdrI)
		highestNonce = strHdrI.LastHeader.Nonce

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
			st.blkExecutor.RestoreLastNotarizedHrdsToGenesis()
			st.forkDetector.RestoreFinalCheckPointToGenesis()
		}
	}()

	for i := len(bootInfos) - 1; i >= 0; i-- {
		log.Debug("apply block",
			"nonce", bootInfos[i].LastHeader.Nonce,
			"shardId", bootInfos[i].LastHeader.ShardId)

		lastNotarized := make(map[uint32]*sync.HdrInfo)
		for _, lastNotarizedHeader := range bootInfos[i].LastNotarizedHeaders {
			log.Debug("added notarized header",
				"nonce", lastNotarizedHeader.Nonce,
				"shardId", lastNotarizedHeader.ShardId)

			lastNotarized[lastNotarizedHeader.ShardId] = &sync.HdrInfo{
				Nonce: lastNotarizedHeader.Nonce,
				Hash:  lastNotarizedHeader.Hash,
			}
		}

		err = st.bootstrapper.applyNotarizedBlocks(lastNotarized)
		if err != nil {
			log.Debug("cannot apply notarized block", "error", err.Error())

			return err
		}

		lastFinalHashes := make([][]byte, 0)
		for _, lastFinal := range bootInfos[i].LastFinals {
			lastFinalHashes = append(lastFinalHashes, lastFinal.Hash)
		}

		err = st.addHeaderToForkDetector(bootInfos[i].LastHeader.Hash, lastFinalHashes)
		if err != nil {
			log.Debug("cannot apply final block", "error", err.Error())
			return err
		}
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

func (st *storageBootstrapper) addHeaderToForkDetector(headerHash []byte, finalHeadersHashes [][]byte) error {
	header, err := st.bootstrapper.getHeader(headerHash)
	if err != nil {
		return err
	}

	log.Debug("added header to fork detector",
		"nonce", header.GetNonce(),
		"shardId", header.GetShardID())

	finalHeaders := make([]data.HeaderHandler, 0)
	for _, hash := range finalHeadersHashes {
		finalHeader, err := st.bootstrapper.getHeader(hash)
		if err != nil {
			return err
		}
		finalHeaders = append(finalHeaders, finalHeader)
		log.Debug("added final header", "nonce", finalHeader.GetNonce())
	}

	err = st.forkDetector.AddHeader(header, headerHash, process.BHProcessed, finalHeaders, finalHeadersHashes, false)
	if err != nil {
		return err
	}

	return nil
}
