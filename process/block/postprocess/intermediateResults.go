package postprocess

import (
	"bytes"
	"sort"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/davecgh/go-spew/spew"
)

var _ process.IntermediateTransactionHandler = (*intermediateResultsProcessor)(nil)

type intermediateResultsProcessor struct {
	pubkeyConv core.PubkeyConverter
	blockType  block.Type
	currTxs    dataRetriever.TransactionCacher

	*basePostProcessor
}

// NewIntermediateResultsProcessor creates a new intermediate results processor
func NewIntermediateResultsProcessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	coordinator sharding.Coordinator,
	pubkeyConv core.PubkeyConverter,
	store dataRetriever.StorageService,
	blockType block.Type,
	currTxs dataRetriever.TransactionCacher,
) (*intermediateResultsProcessor, error) {
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(coordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(pubkeyConv) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(store) {
		return nil, process.ErrNilStorage
	}
	if check.IfNil(currTxs) {
		return nil, process.ErrNilTxForCurrentBlockHandler
	}

	base := &basePostProcessor{
		hasher:           hasher,
		marshalizer:      marshalizer,
		shardCoordinator: coordinator,
		store:            store,
		storageType:      dataRetriever.UnsignedTransactionUnit,
		mapTxToResult:    make(map[string][]string),
	}

	irp := &intermediateResultsProcessor{
		basePostProcessor: base,
		pubkeyConv:        pubkeyConv,
		blockType:         blockType,
		currTxs:           currTxs,
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

			sort.Slice(miniblock.TxHashes, func(a, b int) bool {
				return bytes.Compare(miniblock.TxHashes[a], miniblock.TxHashes[b]) < 0
			})

			log.Trace("intermediateResultsProcessor.CreateAllInterMiniBlocks",
				"type", miniblock.Type,
				"senderShardID", miniblock.SenderShardID,
				"receiverShardID", miniblock.ReceiverShardID,
				"numTxs", len(miniblock.TxHashes),
			)

			finalMBs = append(finalMBs, miniblock)

			if miniblock.ReceiverShardID == irp.shardCoordinator.SelfId() {
				irp.intraShardMiniBlock = miniblock.Clone()
			}
		}
	}

	sort.Slice(finalMBs, func(i, j int) bool {
		return finalMBs[i].ReceiverShardID < finalMBs[j].ReceiverShardID
	})

	irp.mutInterResultsForBlock.Unlock()

	return finalMBs
}

// VerifyInterMiniBlocks verifies if the smart contract results added to the block are valid
func (irp *intermediateResultsProcessor) VerifyInterMiniBlocks(body *block.Body) error {
	scrMbs := irp.CreateAllInterMiniBlocks()
	createdMapMbs := make(map[uint32]*block.MiniBlock)
	for _, mb := range scrMbs {
		createdMapMbs[mb.ReceiverShardID] = mb
	}

	countedCrossShard := 0
	for i := 0; i < len(body.MiniBlocks); i++ {
		mb := body.MiniBlocks[i]
		if mb.Type != irp.blockType {
			continue
		}
		if mb.ReceiverShardID == irp.shardCoordinator.SelfId() {
			continue
		}

		countedCrossShard++
		err := irp.verifyMiniBlock(createdMapMbs, mb)
		if err != nil {
			return err
		}
	}

	createCrossShard := 0
	for _, mb := range scrMbs {
		if mb.Type != irp.blockType {
			continue
		}
		if mb.ReceiverShardID == irp.shardCoordinator.SelfId() {
			continue
		}

		createCrossShard++
	}

	if createCrossShard != countedCrossShard {
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

		addScrShardInfo := &txShardInfo{receiverShardID: dstShId, senderShardID: sndShId}
		scrInfo := &txInfo{tx: addScr, txShardInfo: addScrShardInfo}
		irp.interResultsForBlock[string(scrHash)] = scrInfo
		irp.mapTxToResult[string(addScr.PrevTxHash)] = append(irp.mapTxToResult[string(addScr.PrevTxHash)], string(scrHash))
	}

	return nil
}

func (irp *intermediateResultsProcessor) checkSmartContractResultIntegrity(scr *smartContractResult.SmartContractResult) error {
	err := scr.CheckIntegrity()
	if err != nil {
		return err
	}

	sndShId, dstShId := irp.getShardIdsFromAddresses(scr.SndAddr, scr.RcvAddr)
	isInShardSCR := dstShId == irp.shardCoordinator.SelfId()
	if isInShardSCR {
		return nil
	}

	if sndShId != irp.shardCoordinator.SelfId() && dstShId != irp.shardCoordinator.SelfId() {
		return process.ErrShardIdMissmatch
	}

	return nil
}

func (irp *intermediateResultsProcessor) getShardIdsFromAddresses(sndAddr []byte, rcvAddr []byte) (uint32, uint32) {
	shardForSrc := irp.shardCoordinator.ComputeId(sndAddr)
	shardForDst := irp.shardCoordinator.ComputeId(rcvAddr)

	isEmptyAddress := bytes.Equal(sndAddr, make([]byte, irp.pubkeyConv.Len()))
	if isEmptyAddress {
		shardForSrc = irp.shardCoordinator.SelfId()
	}

	isEmptyAddress = bytes.Equal(rcvAddr, make([]byte, irp.pubkeyConv.Len()))
	if isEmptyAddress {
		shardForDst = irp.shardCoordinator.SelfId()
	}

	return shardForSrc, shardForDst
}

// IsInterfaceNil returns true if there is no value under the interface
func (irp *intermediateResultsProcessor) IsInterfaceNil() bool {
	return irp == nil
}
