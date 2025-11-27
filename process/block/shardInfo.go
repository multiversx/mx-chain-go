package block

import (
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
)

// ShardInfoCreateData is a component used to create shard info from shard headers
type ShardInfoCreateData struct {
	enableEpochsHandler      common.EnableEpochsHandler
	headersPool              dataRetriever.HeadersPool
	proofsPool               dataRetriever.ProofsPool
	pendingMiniBlocksHandler process.PendingMiniBlocksHandler
	blockTracker             process.BlockTracker
}

// NewShardInfoCreateData creates a new ShardInfoCreateData instance
func NewShardInfoCreateData(
	enableEpochsHandler common.EnableEpochsHandler,
	headersPool dataRetriever.HeadersPool,
	proofsPool dataRetriever.ProofsPool,
	pendingMiniBlocksHandler process.PendingMiniBlocksHandler,
	blockTracker process.BlockTracker,
) (*ShardInfoCreateData, error) {
	if check.IfNil(enableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(headersPool) {
		return nil, process.ErrNilHeadersDataPool
	}
	if check.IfNil(proofsPool) {
		return nil, process.ErrNilProofsPool
	}
	if check.IfNil(pendingMiniBlocksHandler) {
		return nil, process.ErrNilPendingMiniBlocksHandler
	}
	if check.IfNil(blockTracker) {
		return nil, process.ErrNilBlockTracker
	}

	return &ShardInfoCreateData{
		enableEpochsHandler:      enableEpochsHandler,
		headersPool:              headersPool,
		proofsPool:               proofsPool,
		pendingMiniBlocksHandler: pendingMiniBlocksHandler,
		blockTracker:             blockTracker,
	}, nil
}

// CreateShardInfoV3 creates shard info for V3 metablock headers
func (sic *ShardInfoCreateData) CreateShardInfoV3(
	metaHeader data.MetaHeaderHandler,
	shardHeaders []data.HeaderHandler,
	shardHeaderHashes [][]byte,
) ([]data.ShardDataProposalHandler, []data.ShardDataHandler, error) {
	if check.IfNil(metaHeader) {
		return nil, nil, process.ErrNilHeaderHandler
	}
	if !metaHeader.IsHeaderV3() {
		return nil, nil, process.ErrInvalidHeader
	}

	var shardInfo []data.ShardDataHandler
	var shardInfoProposal []data.ShardDataProposalHandler
	if len(shardHeaders) != len(shardHeaderHashes) {
		return nil, nil, process.ErrInconsistentShardHeadersAndHashes
	}

	for i := 0; i < len(shardHeaders); i++ {
		shardDataProposal, shardData, err := sic.createShardInfoFromHeader(shardHeaders[i], shardHeaderHashes[i])
		if err != nil {
			return nil, nil, err
		}
		shardInfo = append(shardInfo, shardData...)
		shardInfoProposal = append(shardInfoProposal, shardDataProposal)
	}

	return shardInfoProposal, shardInfo, nil
}

// CreateShardInfoFromLegacyMeta creates shard info for legacy meta header
func (sic *ShardInfoCreateData) CreateShardInfoFromLegacyMeta(
	metaHeader data.MetaHeaderHandler,
	shardHeaders []data.ShardHeaderHandler,
	shardHeaderHashes [][]byte,
) ([]data.ShardDataHandler, error) {
	if check.IfNil(metaHeader) {
		return nil, process.ErrNilHeaderHandler
	}
	if metaHeader.IsHeaderV3() {
		return nil, process.ErrInvalidHeader
	}

	if len(shardHeaders) != len(shardHeaderHashes) {
		return nil, process.ErrInconsistentShardHeadersAndHashes
	}

	shardInfo := make([]data.ShardDataHandler, 0)
	for i := 0; i < len(shardHeaders); i++ {
		shardData, err := sic.createShardDataFromLegacyHeader(shardHeaders[i], shardHeaderHashes[i])
		if err != nil {
			return nil, err
		}
		shardInfo = append(shardInfo, shardData...)
	}

	return shardInfo, nil
}

func (sic *ShardInfoCreateData) createShardInfoFromHeader(
	shardHeader data.HeaderHandler,
	hdrHash []byte,
) (data.ShardDataProposalHandler, []data.ShardDataHandler, error) {
	if check.IfNil(shardHeader) {
		return nil, nil, process.ErrNilHeaderHandler
	}
	if len(hdrHash) == 0 {
		return nil, nil, process.ErrInvalidHash
	}

	hasMissingShardHdrProof := shardHeader.GetNonce() >= 1 && !sic.proofsPool.HasProof(shardHeader.GetShardID(), hdrHash)
	if hasMissingShardHdrProof {
		return nil, nil, fmt.Errorf("%w for shard header with hash %s", process.ErrMissingHeaderProof, hex.EncodeToString(hdrHash))
	}

	if !shardHeader.IsHeaderV3() {
		shardData, err := sic.createShardDataFromLegacyHeader(shardHeader, hdrHash)
		return sic.createShardDataProposalFromHeader(shardHeader, hdrHash), shardData, err
	}

	return sic.createShardDataFromV3Header(shardHeader, hdrHash)
}

func (sic *ShardInfoCreateData) createShardDataFromLegacyHeader(shardHdr data.HeaderHandler, hdrHash []byte) ([]data.ShardDataHandler, error) {
	shardData := &block.ShardData{}
	shardData.TxCount = shardHdr.GetTxCount()
	shardData.ShardID = shardHdr.GetShardID()
	shardData.HeaderHash = hdrHash
	shardData.Round = shardHdr.GetRound()
	shardData.PrevHash = shardHdr.GetPrevHash()
	shardData.Nonce = shardHdr.GetNonce()
	shardData.PrevRandSeed = shardHdr.GetPrevRandSeed()
	shardData.PubKeysBitmap = shardHdr.GetPubKeysBitmap()
	if sic.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, shardHdr.GetEpoch()) {
		shardData.Epoch = shardHdr.GetEpoch()
	}
	shardData.AccumulatedFees = shardHdr.GetAccumulatedFees()
	shardData.DeveloperFees = shardHdr.GetDeveloperFees()
	shardData.ShardMiniBlockHeaders = createShardMiniBlockHeaderFromHeader(shardHdr, sic.enableEpochsHandler)
	err := sic.updateShardDataWithCrossShardInfo(shardData, shardHdr)
	if err != nil {
		return nil, err
	}

	return []data.ShardDataHandler{shardData}, nil
}

func (sic *ShardInfoCreateData) createShardDataFromV3Header(
	shardHeader data.HeaderHandler,
	hdrHash []byte,
) (data.ShardDataProposalHandler, []data.ShardDataHandler, error) {
	if check.IfNil(shardHeader) {
		return nil, nil, process.ErrNilHeaderHandler
	}
	shardDataProposal := sic.createShardDataProposalFromHeader(shardHeader, hdrHash)
	executionResults := shardHeader.GetExecutionResultsHandlers()
	if len(executionResults) == 0 {
		// return shard data proposal even though the shard header does not hold any execution result
		return shardDataProposal, []data.ShardDataHandler{}, nil
	}

	shardDataHandlers := make([]data.ShardDataHandler, len(executionResults))
	for i, execResult := range executionResults {
		shardData, err := sic.createShardDataFromExecutionResult(execResult)
		if err != nil {
			return nil, nil, err
		}
		shardDataHandlers[i] = shardData
	}

	return shardDataProposal, shardDataHandlers, nil
}

func (sic *ShardInfoCreateData) createShardDataProposalFromHeader(
	header data.HeaderHandler,
	hdrHash []byte,
) data.ShardDataProposalHandler {
	return &block.ShardDataProposal{
		HeaderHash:           hdrHash,
		Round:                header.GetRound(),
		Nonce:                header.GetNonce(),
		ShardID:              header.GetShardID(),
		Epoch:                header.GetEpoch(),
		NumPendingMiniBlocks: uint32(len(sic.pendingMiniBlocksHandler.GetPendingMiniBlocks(header.GetShardID()))),
	}
}

func (sic *ShardInfoCreateData) createShardDataFromExecutionResult(
	execResult data.BaseExecutionResultHandler,
) (data.ShardDataHandler, error) {
	if check.IfNil(execResult) {
		return nil, process.ErrNilExecutionResultHandler
	}

	execResultHandler, ok := execResult.(data.ExecutionResultHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	header, err := sic.headersPool.GetHeaderByHash(execResultHandler.GetHeaderHash())
	if err != nil {
		return nil, err
	}

	shardData := &block.ShardData{}
	shardData.TxCount = uint32(execResultHandler.GetExecutedTxCount())
	shardData.ShardID = header.GetShardID()
	shardData.HeaderHash = execResultHandler.GetHeaderHash()
	shardData.Round = execResultHandler.GetHeaderRound()
	shardData.PrevHash = header.GetPrevHash()
	shardData.Nonce = execResultHandler.GetHeaderNonce()
	shardData.PrevRandSeed = header.GetPrevRandSeed()
	shardData.PubKeysBitmap = header.GetPubKeysBitmap()
	shardData.Epoch = header.GetEpoch()
	shardData.AccumulatedFees = execResultHandler.GetAccumulatedFees()
	shardData.DeveloperFees = execResultHandler.GetDeveloperFees()
	shardData.ShardMiniBlockHeaders = createShardMiniBlockHeaderFromExecutionResultHandler(execResultHandler)
	err = sic.updateShardDataWithCrossShardInfo(shardData, header)
	if err != nil {
		return nil, err
	}

	return shardData, nil
}

func createShardMiniBlockHeaderFromHeader(
	shardHdr data.HeaderHandler,
	enableEpochsHandler common.EnableEpochsHandler,
) []block.MiniBlockHeader {
	shardMiniBlockHeaders := make([]block.MiniBlockHeader, 0)
	for i := 0; i < len(shardHdr.GetMiniBlockHeaderHandlers()); i++ {
		if enableEpochsHandler.IsFlagEnabled(common.ScheduledMiniBlocksFlag) {
			miniBlockHeader := shardHdr.GetMiniBlockHeaderHandlers()[i]
			if !miniBlockHeader.IsFinal() {
				log.Debug("metaProcessor.createShardInfo: do not create shard data with mini block which is not final", "mb hash", miniBlockHeader.GetHash())
				continue
			}
		}

		shardMiniBlockHeader := miniBlockHeaderFromMiniBlockHeaderHandler(shardHdr.GetMiniBlockHeaderHandlers()[i])
		shardMiniBlockHeaders = append(shardMiniBlockHeaders, shardMiniBlockHeader)
	}
	return shardMiniBlockHeaders
}

func createShardMiniBlockHeaderFromExecutionResultHandler(
	execResultHandler data.ExecutionResultHandler,
) []block.MiniBlockHeader {
	mbHeaderHandlers := execResultHandler.GetMiniBlockHeadersHandlers()
	shardMiniBlockHeaders := make([]block.MiniBlockHeader, 0, len(mbHeaderHandlers))
	for i := 0; i < len(mbHeaderHandlers); i++ {
		shardMiniBlockHeader := miniBlockHeaderFromMiniBlockHeaderHandler(mbHeaderHandlers[i])
		shardMiniBlockHeaders = append(shardMiniBlockHeaders, shardMiniBlockHeader)
	}
	return shardMiniBlockHeaders
}

func miniBlockHeaderFromMiniBlockHeaderHandler(miniBlockHeaderHandler data.MiniBlockHeaderHandler) block.MiniBlockHeader {
	shardMiniBlockHeader := block.MiniBlockHeader{}
	shardMiniBlockHeader.SenderShardID = miniBlockHeaderHandler.GetSenderShardID()
	shardMiniBlockHeader.ReceiverShardID = miniBlockHeaderHandler.GetReceiverShardID()
	shardMiniBlockHeader.Hash = miniBlockHeaderHandler.GetHash()
	shardMiniBlockHeader.TxCount = miniBlockHeaderHandler.GetTxCount()
	shardMiniBlockHeader.Type = block.Type(miniBlockHeaderHandler.GetTypeInt32())

	return shardMiniBlockHeader
}

func (sic *ShardInfoCreateData) updateShardDataWithCrossShardInfo(shardData *block.ShardData, header data.HeaderHandler) error {
	if !header.IsHeaderV3() {
		shardData.NumPendingMiniBlocks = uint32(len(sic.pendingMiniBlocksHandler.GetPendingMiniBlocks(header.GetShardID())))
	}

	// TODO: the last self notarized header should be fetched based on the nonce from the execution result
	metaHeader, _, err := sic.blockTracker.GetLastSelfNotarizedHeader(header.GetShardID())
	if err != nil {
		return err
	}
	shardData.LastIncludedMetaNonce = metaHeader.GetNonce()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sic *ShardInfoCreateData) IsInterfaceNil() bool {
	return sic == nil
}
