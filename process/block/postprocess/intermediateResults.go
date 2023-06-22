package postprocess

import (
	"bytes"
	"sort"

	"github.com/davecgh/go-spew/spew"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ process.IntermediateTransactionHandler = (*intermediateResultsProcessor)(nil)

type intermediateResultsProcessor struct {
	pubkeyConv          core.PubkeyConverter
	blockType           block.Type
	currTxs             dataRetriever.TransactionCacher
	enableEpochsHandler common.EnableEpochsHandler

	*basePostProcessor
}

// ArgsNewIntermediateResultsProcessor defines the arguments needed for new smart contract processor
type ArgsNewIntermediateResultsProcessor struct {
	Hasher              hashing.Hasher
	Marshalizer         marshal.Marshalizer
	Coordinator         sharding.Coordinator
	PubkeyConv          core.PubkeyConverter
	Store               dataRetriever.StorageService
	BlockType           block.Type
	CurrTxs             dataRetriever.TransactionCacher
	EconomicsFee        process.FeeHandler
	EnableEpochsHandler common.EnableEpochsHandler
}

// NewIntermediateResultsProcessor creates a new intermediate results processor
func NewIntermediateResultsProcessor(
	args ArgsNewIntermediateResultsProcessor,
) (*intermediateResultsProcessor, error) {
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.Coordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.PubkeyConv) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.Store) {
		return nil, process.ErrNilStorage
	}
	if check.IfNil(args.CurrTxs) {
		return nil, process.ErrNilTxForCurrentBlockHandler
	}
	if check.IfNil(args.EconomicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}

	base := &basePostProcessor{
		hasher:             args.Hasher,
		marshalizer:        args.Marshalizer,
		shardCoordinator:   args.Coordinator,
		store:              args.Store,
		storageType:        dataRetriever.UnsignedTransactionUnit,
		mapProcessedResult: make(map[string][][]byte),
		economicsFee:       args.EconomicsFee,
	}

	irp := &intermediateResultsProcessor{
		basePostProcessor:   base,
		pubkeyConv:          args.PubkeyConv,
		blockType:           args.BlockType,
		currTxs:             args.CurrTxs,
		enableEpochsHandler: args.EnableEpochsHandler,
	}

	irp.interResultsForBlock = make(map[string]*txInfo)

	return irp, nil
}

// GetNumOfCrossInterMbsAndTxs returns the number of cross shard miniblocks and transactions for the current round,
// created from the smart contract results
func (irp *intermediateResultsProcessor) GetNumOfCrossInterMbsAndTxs() (int, int) {
	miniBlocks := make(map[uint32]int)

	irp.mutInterResultsForBlock.Lock()
	for _, tx := range irp.interResultsForBlock {
		if tx.receiverShardID == irp.shardCoordinator.SelfId() {
			continue
		}
		miniBlocks[tx.receiverShardID]++
	}
	irp.mutInterResultsForBlock.Unlock()

	numMbs := len(miniBlocks)
	numTxs := 0
	for _, numTxsInMb := range miniBlocks {
		numTxs += numTxsInMb
	}

	return numMbs, numTxs
}

// CreateAllInterMiniBlocks returns the miniblocks for the current round created from the smart contract results
func (irp *intermediateResultsProcessor) CreateAllInterMiniBlocks() []*block.MiniBlock {
	miniBlocks := make(map[uint32]*block.MiniBlock, int(irp.shardCoordinator.NumberOfShards())+1)
	for i := uint32(0); i < irp.shardCoordinator.NumberOfShards(); i++ {
		miniBlocks[i] = &block.MiniBlock{}
	}
	miniBlocks[core.MetachainShardId] = &block.MiniBlock{}

	irp.currTxs.Clean()
	irp.mutInterResultsForBlock.Lock()

	for key, value := range irp.interResultsForBlock {
		recvShId := value.receiverShardID
		miniBlocks[recvShId].TxHashes = append(miniBlocks[recvShId].TxHashes, []byte(key))
		irp.currTxs.AddTx([]byte(key), value.tx)
	}

	finalMBs := make([]*block.MiniBlock, 0)
	for shId, miniblock := range miniBlocks {
		if len(miniblock.TxHashes) > 0 {
			miniblock.SenderShardID = irp.shardCoordinator.SelfId()
			miniblock.ReceiverShardID = shId
			miniblock.Type = irp.blockType

			if irp.enableEpochsHandler.IsKeepExecOrderOnCreatedSCRsEnabled() {
				sort.Slice(miniblock.TxHashes, func(a, b int) bool {
					scrInfoA := irp.interResultsForBlock[string(miniblock.TxHashes[a])]
					scrInfoB := irp.interResultsForBlock[string(miniblock.TxHashes[b])]
					return scrInfoA.index < scrInfoB.index
				})
			} else {
				sort.Slice(miniblock.TxHashes, func(a, b int) bool {
					return bytes.Compare(miniblock.TxHashes[a], miniblock.TxHashes[b]) < 0
				})
			}

			log.Debug("intermediateResultsProcessor.CreateAllInterMiniBlocks",
				"type", miniblock.Type,
				"senderShardID", miniblock.SenderShardID,
				"receiverShardID", miniblock.ReceiverShardID,
				"numTxs", len(miniblock.TxHashes),
			)

			finalMBs = append(finalMBs, miniblock)

			if miniblock.ReceiverShardID == irp.shardCoordinator.SelfId() {
				irp.intraShardMiniBlock = miniblock.DeepClone()
			}
		}
	}

	sort.Slice(finalMBs, func(i, j int) bool {
		return finalMBs[i].ReceiverShardID < finalMBs[j].ReceiverShardID
	})

	splitMBs := irp.splitMiniBlocksIfNeeded(finalMBs)

	irp.mutInterResultsForBlock.Unlock()

	return splitMBs
}

// VerifyInterMiniBlocks verifies if the smart contract results added to the block are valid
func (irp *intermediateResultsProcessor) VerifyInterMiniBlocks(body *block.Body) error {
	scrMbs := irp.CreateAllInterMiniBlocks()
	createdMapMbs := createMiniBlocksMap(scrMbs)

	receivedCrossShard := 0
	for i := 0; i < len(body.MiniBlocks); i++ {
		mb := body.MiniBlocks[i]
		if mb.Type != irp.blockType {
			continue
		}
		if mb.ReceiverShardID == irp.shardCoordinator.SelfId() {
			continue
		}
		//if mb.ReceiverShardID == core.SovereignChainShardId && mb.SenderShardID == core.MainChainShardId {
		//	continue
		//}

		receivedCrossShard++
		err := irp.verifyMiniBlock(createdMapMbs, mb)
		if err != nil {
			return err
		}
	}

	createdCrossShard := 0
	for _, mb := range scrMbs {
		if mb.Type != irp.blockType {
			continue
		}
		if mb.ReceiverShardID == irp.shardCoordinator.SelfId() {
			continue
		}

		createdCrossShard++
	}

	if createdCrossShard != receivedCrossShard {
		log.Debug("intermediateResultsProcessor.VerifyInterMiniBlocks",
			"num created cross shard", createdCrossShard,
			"num received cross shard", receivedCrossShard)
		return process.ErrMiniBlockNumMissMatch
	}

	return nil
}

// AddIntermediateTransactions adds smart contract results from smart contract processing for cross-shard calls
func (irp *intermediateResultsProcessor) AddIntermediateTransactions(txs []data.TransactionHandler) error {
	irp.mutInterResultsForBlock.Lock()
	defer irp.mutInterResultsForBlock.Unlock()

	log.Trace("intermediateResultsProcessor.AddIntermediateTransactions()", "txs", len(txs))

	for i := 0; i < len(txs); i++ {
		addScr, ok := txs[i].(*smartContractResult.SmartContractResult)
		if !ok {
			return process.ErrWrongTypeAssertion
		}

		err := irp.checkSmartContractResultIntegrity(addScr)
		if err != nil {
			log.Error("AddIntermediateTransaction", "error", err, "dump", spew.Sdump(addScr))
			return err
		}

		scrHash, err := core.CalculateHash(irp.marshalizer, irp.hasher, addScr)
		if err != nil {
			return err
		}

		if log.GetLevel() == logger.LogTrace {
			//spew.Sdump is very useful when debugging errors like `receipts hash mismatch`
			log.Trace("scr added", "txHash", addScr.PrevTxHash, "hash", scrHash, "nonce", addScr.Nonce, "gasLimit", addScr.GasLimit, "value", addScr.Value,
				"dump", spew.Sdump(addScr))
		}

		sndShId, dstShId := irp.getShardIdsFromAddresses(addScr.SndAddr, addScr.RcvAddr)
		irp.addIntermediateTxToResultsForBlock(addScr, scrHash, sndShId, dstShId)
	}

	return nil
}

func (irp *intermediateResultsProcessor) checkSmartContractResultIntegrity(scr *smartContractResult.SmartContractResult) error {
	if len(scr.RcvAddr) == 0 {
		return process.ErrNilRcvAddr
	}
	if len(scr.SndAddr) == 0 {
		return process.ErrNilSndAddr
	}

	sndShId, dstShId := irp.getShardIdsFromAddresses(scr.SndAddr, scr.RcvAddr)
	isInShardSCR := dstShId == irp.shardCoordinator.SelfId()
	if isInShardSCR {
		return nil
	}

	err := scr.CheckIntegrity()
	if err != nil {
		return err
	}

	if sndShId != irp.shardCoordinator.SelfId() && dstShId != irp.shardCoordinator.SelfId() {
		return process.ErrShardIdMissmatch
	}

	return nil
}

func (irp *intermediateResultsProcessor) getShardIdsFromAddresses(sndAddr []byte, rcvAddr []byte) (uint32, uint32) {
	shardForSrc := irp.shardCoordinator.ComputeId(sndAddr)
	shardForDst := irp.shardCoordinator.ComputeId(rcvAddr)

	return shardForSrc, shardForDst
}

// IsInterfaceNil returns true if there is no value under the interface
func (irp *intermediateResultsProcessor) IsInterfaceNil() bool {
	return irp == nil
}
