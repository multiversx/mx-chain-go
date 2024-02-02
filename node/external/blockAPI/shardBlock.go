package blockAPI

import (
	"encoding/hex"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/node/filters"
	"github.com/multiversx/mx-chain-go/process"
)

type shardAPIBlockProcessor struct {
	*baseAPIBlockProcessor
}

// newShardApiBlockProcessor will create a new instance of shard api block processor
func newShardApiBlockProcessor(arg *ArgAPIBlockProcessor, emptyReceiptsHash []byte) *shardAPIBlockProcessor {
	hasDbLookupExtensions := arg.HistoryRepo.IsEnabled()

	return &shardAPIBlockProcessor{
		baseAPIBlockProcessor: &baseAPIBlockProcessor{
			hasDbLookupExtensions:        hasDbLookupExtensions,
			selfShardID:                  arg.SelfShardID,
			store:                        arg.Store,
			marshalizer:                  arg.Marshalizer,
			uint64ByteSliceConverter:     arg.Uint64ByteSliceConverter,
			historyRepo:                  arg.HistoryRepo,
			apiTransactionHandler:        arg.APITransactionHandler,
			txStatusComputer:             arg.StatusComputer,
			hasher:                       arg.Hasher,
			addressPubKeyConverter:       arg.AddressPubkeyConverter,
			emptyReceiptsHash:            emptyReceiptsHash,
			logsFacade:                   arg.LogsFacade,
			receiptsRepository:           arg.ReceiptsRepository,
			alteredAccountsProvider:      arg.AlteredAccountsProvider,
			accountsRepository:           arg.AccountsRepository,
			scheduledTxsExecutionHandler: arg.ScheduledTxsExecutionHandler,
			enableEpochsHandler:          arg.EnableEpochsHandler,
		},
	}
}

// GetBlockByNonce will return a shard APIBlock by nonce
func (sbp *shardAPIBlockProcessor) GetBlockByNonce(nonce uint64, options api.BlockQueryOptions) (*api.Block, error) {
	storerUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(sbp.selfShardID)

	nonceToByteSlice := sbp.uint64ByteSliceConverter.ToByteSlice(nonce)
	headerHash, err := sbp.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, err
	}

	// if genesis block, get the nonce key corresponding to the altered block
	if nonce == 0 {
		nonceToByteSlice = append(nonceToByteSlice, []byte(common.GenesisStorageSuffix)...)
	}

	alteredHeaderHash, err := sbp.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, err
	}

	blockBytes, err := sbp.getFromStorer(dataRetriever.BlockHeaderUnit, alteredHeaderHash)
	if err != nil {
		return nil, err
	}

	return sbp.convertShardBlockBytesToAPIBlock(headerHash, blockBytes, options)
}

// GetBlockByHash will return a shard APIBlock by hash
func (sbp *shardAPIBlockProcessor) GetBlockByHash(hash []byte, options api.BlockQueryOptions) (*api.Block, error) {
	blockBytes, err := sbp.getFromStorer(dataRetriever.BlockHeaderUnit, hash)
	if err != nil {
		return nil, err
	}

	blockHeader, err := process.UnmarshalShardHeader(sbp.marshalizer, blockBytes)
	if err != nil {
		return nil, err
	}

	// if genesis block, get the altered block bytes
	if blockHeader.GetNonce() == 0 {
		alteredHash := createAlteredBlockHash(hash)
		blockBytes, err = sbp.getFromStorer(dataRetriever.BlockHeaderUnit, alteredHash)
		if err != nil {
			return nil, err
		}
	}

	blockAPI, err := sbp.convertShardBlockBytesToAPIBlock(hash, blockBytes, options)
	if err != nil {
		return nil, err
	}

	storerUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(sbp.selfShardID)

	return sbp.computeStatusAndPutInBlock(blockAPI, storerUnit)
}

// GetBlockByRound will return a shard APIBlock by round
func (sbp *shardAPIBlockProcessor) GetBlockByRound(round uint64, options api.BlockQueryOptions) (*api.Block, error) {
	headerHash, blockBytes, err := sbp.getBlockHeaderHashAndBytesByRound(round, dataRetriever.BlockHeaderUnit)
	if err != nil {
		return nil, err
	}

	return sbp.convertShardBlockBytesToAPIBlock(headerHash, blockBytes, options)
}

// GetAlteredAccountsForBlock will return the altered accounts for the desired shard block
func (sbp *shardAPIBlockProcessor) GetAlteredAccountsForBlock(options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error) {
	headerHash, blockBytes, err := sbp.getHashAndBlockBytesFromStorer(options.GetBlockParameters)
	if err != nil {
		return nil, err
	}

	apiBlock, err := sbp.convertShardBlockBytesToAPIBlock(headerHash, blockBytes, api.BlockQueryOptions{WithTransactions: true, WithLogs: true})
	if err != nil {
		return nil, err
	}

	return sbp.apiBlockToAlteredAccounts(apiBlock, options)
}

func (sbp *shardAPIBlockProcessor) getHashAndBlockBytesFromStorer(params api.GetBlockParameters) ([]byte, []byte, error) {
	switch params.RequestType {
	case api.BlockFetchTypeByHash:
		return sbp.getHashAndBlockBytesFromStorerByHash(params)
	case api.BlockFetchTypeByNonce:
		return sbp.getHashAndBlockBytesFromStorerByNonce(params)
	default:
		return nil, nil, errUnknownBlockRequestType
	}
}

func (sbp *shardAPIBlockProcessor) getHashAndBlockBytesFromStorerByHash(params api.GetBlockParameters) ([]byte, []byte, error) {
	headerBytes, err := sbp.getFromStorer(dataRetriever.BlockHeaderUnit, params.Hash)
	return params.Hash, headerBytes, err
}

func (sbp *shardAPIBlockProcessor) getHashAndBlockBytesFromStorerByNonce(params api.GetBlockParameters) ([]byte, []byte, error) {
	storerUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(sbp.selfShardID)

	nonceToByteSlice := sbp.uint64ByteSliceConverter.ToByteSlice(params.Nonce)
	headerHash, err := sbp.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, nil, err
	}

	headerBytes, err := sbp.getFromStorer(dataRetriever.BlockHeaderUnit, headerHash)
	return headerHash, headerBytes, err
}

func (sbp *shardAPIBlockProcessor) convertShardBlockBytesToAPIBlock(hash []byte, blockBytes []byte, options api.BlockQueryOptions) (*api.Block, error) {
	blockHeader, err := process.UnmarshalShardHeader(sbp.marshalizer, blockBytes)
	if err != nil {
		return nil, err
	}

	numOfTxs := uint32(0)
	miniblocks := make([]*api.MiniBlock, 0)

	for _, mb := range blockHeader.GetMiniBlockHeaderHandlers() {
		if block.Type(mb.GetTypeInt32()) == block.PeerBlock {
			continue
		}

		numOfTxs += mb.GetTxCount()

		miniblockAPI := &api.MiniBlock{
			Hash:                    hex.EncodeToString(mb.GetHash()),
			Type:                    block.Type(mb.GetTypeInt32()).String(),
			SourceShard:             mb.GetSenderShardID(),
			DestinationShard:        mb.GetReceiverShardID(),
			ProcessingType:          block.ProcessingType(mb.GetProcessingType()).String(),
			ConstructionState:       block.MiniBlockState(mb.GetConstructionState()).String(),
			IndexOfFirstTxProcessed: mb.GetIndexOfFirstTxProcessed(),
			IndexOfLastTxProcessed:  mb.GetIndexOfLastTxProcessed(),
		}
		if options.WithTransactions {
			miniBlockCopy := mb
			err = sbp.getAndAttachTxsToMb(miniBlockCopy, blockHeader, miniblockAPI, options)
			if err != nil {
				return nil, err
			}
		}

		miniblocks = append(miniblocks, miniblockAPI)
	}

	intraMb, err := sbp.getIntrashardMiniblocksFromReceiptsStorage(blockHeader, hash, options)
	if err != nil {
		return nil, err
	}

	miniblocks = append(miniblocks, intraMb...)
	miniblocks = filterOutDuplicatedMiniblocks(miniblocks)

	statusFilters := filters.NewStatusFilters(sbp.selfShardID)
	statusFilters.ApplyStatusFilters(miniblocks)

	apiBlock := &api.Block{
		Nonce:           blockHeader.GetNonce(),
		Round:           blockHeader.GetRound(),
		Epoch:           blockHeader.GetEpoch(),
		Shard:           blockHeader.GetShardID(),
		Hash:            hex.EncodeToString(hash),
		PrevBlockHash:   hex.EncodeToString(blockHeader.GetPrevHash()),
		NumTxs:          numOfTxs,
		MiniBlocks:      miniblocks,
		AccumulatedFees: blockHeader.GetAccumulatedFees().String(),
		DeveloperFees:   blockHeader.GetDeveloperFees().String(),
		Timestamp:       time.Duration(blockHeader.GetTimeStamp()),
		Status:          BlockStatusOnChain,
		StateRootHash:   hex.EncodeToString(blockHeader.GetRootHash()),
		PubKeyBitmap:    hex.EncodeToString(blockHeader.GetPubKeysBitmap()),
		Signature:       hex.EncodeToString(blockHeader.GetSignature()),
		LeaderSignature: hex.EncodeToString(blockHeader.GetLeaderSignature()),
		ChainID:         string(blockHeader.GetChainID()),
		SoftwareVersion: hex.EncodeToString(blockHeader.GetSoftwareVersion()),
		ReceiptsHash:    hex.EncodeToString(blockHeader.GetReceiptsHash()),
		Reserved:        blockHeader.GetReserved(),
		RandSeed:        hex.EncodeToString(blockHeader.GetRandSeed()),
		PrevRandSeed:    hex.EncodeToString(blockHeader.GetPrevRandSeed()),
	}

	addScheduledInfoInBlock(blockHeader, apiBlock)

	return apiBlock, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (sbp *shardAPIBlockProcessor) IsInterfaceNil() bool {
	return sbp == nil
}
