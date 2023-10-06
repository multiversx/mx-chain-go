package dblookupext

import (
	"bytes"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/storage"
)

type miniblocksHandler struct {
	marshaller                       marshal.Marshalizer
	hasher                           hashing.Hasher
	epochIndex                       *epochByHashIndex
	miniblockHashByTxHashIndexStorer storage.Storer
	miniblocksMetadataStorer         storage.Storer
}

// blockCommitted will record the miniblocks information
func (mh *miniblocksHandler) blockCommitted(header data.HeaderHandler, blockBody *block.Body) error {
	headerHash, errCalculate := core.CalculateHash(mh.marshaller, mh.hasher, header)
	if errCalculate != nil {
		return fmt.Errorf("%w for header, header round %d", errCalculate, header.GetRound())
	}

	for index, mb := range blockBody.MiniBlocks {
		miniblockHash, err := core.CalculateHash(mh.marshaller, mh.hasher, mb)
		if err != nil {
			return fmt.Errorf("%w for miniblock at index %d, header hash %x", err, index, headerHash)
		}

		err = mh.commitMiniblock(header, headerHash, mb, miniblockHash)
		if err != nil {
			return fmt.Errorf("%w for miniblock at index %d, mbHash %x, header hash %x", err, index, miniblockHash, headerHash)
		}
	}

	return nil
}

func (mh *miniblocksHandler) commitMiniblock(header data.HeaderHandler, headerHash []byte, mb *block.MiniBlock, miniblockHash []byte) error {
	miniblockHeader, err := mh.findMiniblockHandler(header, miniblockHash)
	if err != nil {
		return err
	}

	if miniblockHeader.GetIndexOfFirstTxProcessed() == 0 {
		// first time we see this miniblock, we should store the (miniblock, epoch) tuple
		err = mh.epochIndex.saveEpochByHash(miniblockHash, header.GetEpoch())
		if err != nil {
			return err
		}
	}

	err = mh.commitExecutedTransactions(miniblockHeader, miniblockHash, mb, header.GetEpoch())
	if err != nil {
		return err
	}

	return mh.commitMiniblockMetadata(header, headerHash, miniblockHeader, miniblockHash, mb)
}

func (mh *miniblocksHandler) findMiniblockHandler(header data.HeaderHandler, miniblockHash []byte) (data.MiniBlockHeaderHandler, error) {
	for _, mbHeader := range header.GetMiniBlockHeaderHandlers() {
		if bytes.Equal(mbHeader.GetHash(), miniblockHash) {
			return mbHeader, nil
		}
	}

	return nil, errMiniblockHeaderNotFound
}

// commitExecutedTransactions will save only the transactions between the start and end index of the provided miniblockHeader
func (mh *miniblocksHandler) commitExecutedTransactions(miniblockHeader data.MiniBlockHeaderHandler, miniblockHash []byte, mb *block.MiniBlock, epoch uint32) error {
	txHashes := mh.getExecutedTxHashes(miniblockHeader, mb)
	for _, txHash := range txHashes {
		err := mh.miniblockHashByTxHashIndexStorer.PutInEpoch(txHash, miniblockHash, epoch)
		if err != nil {
			return fmt.Errorf("%w for tx hash %x, miniblock hash %x", err, txHash, miniblockHash)
		}
	}

	return nil
}

func (mh *miniblocksHandler) getExecutedTxHashes(miniblockHeader data.MiniBlockHeaderHandler, mb *block.MiniBlock) [][]byte {
	txHashes := make([][]byte, 0, len(mb.TxHashes))
	for index := miniblockHeader.GetIndexOfFirstTxProcessed(); index <= miniblockHeader.GetIndexOfLastTxProcessed(); index++ {
		txHash := mb.TxHashes[index]
		txHashes = append(txHashes, txHash)
	}

	return txHashes
}

func (mh *miniblocksHandler) commitMiniblockMetadata(
	header data.HeaderHandler,
	headerHash []byte,
	miniblockHeader data.MiniBlockHeaderHandler,
	miniblockHash []byte,
	mb *block.MiniBlock,
) error {
	miniblockMetaData := &MiniblockMetadataV2{
		MiniblockHash:          miniblockHash,
		SourceShardID:          mb.GetSenderShardID(),
		DestinationShardID:     mb.ReceiverShardID,
		Type:                   int32(mb.Type),
		NumTxHashesInMiniblock: int32(len(mb.TxHashes)),
	}

	var txHasheshenPartial [][]byte

	if miniblockHeader.GetConstructionState() == int32(block.PartialExecuted) {
		loadedMiniblockMetadata, err := mh.loadExistingMiniblocksMetadata(miniblockHash, header.GetEpoch())
		if err != nil {
			return err
		}
		if len(loadedMiniblockMetadata.MiniblockHash) != 0 {
			// rewrite, we found an existing info in storage
			miniblockMetaData = loadedMiniblockMetadata
		}
		txHasheshenPartial = mh.getExecutedTxHashes(miniblockHeader, mb)
	}

	miniblockOnHeader := &MiniblockMetadataOnBlock{
		Round:                   header.GetRound(),
		HeaderNonce:             header.GetNonce(),
		HeaderHash:              headerHash,
		Epoch:                   header.GetEpoch(),
		TxHashesWhenPartial:     txHasheshenPartial,
		IndexOfFirstTxProcessed: miniblockHeader.GetIndexOfFirstTxProcessed(),
	}

	miniblockMetaData.MiniblocksInfo = append(miniblockMetaData.MiniblocksInfo, miniblockOnHeader)

	return mh.saveMiniblocksMetadata(miniblockHash, miniblockMetaData, header.GetEpoch())
}

func (mh *miniblocksHandler) loadExistingMiniblocksMetadata(miniblockHash []byte, epoch uint32) (*MiniblockMetadataV2, error) {
	multipleMiniblockMetadata := &MiniblockMetadataV2{}
	buff, err := mh.miniblocksMetadataStorer.GetFromEpoch(miniblockHash, epoch)
	if err == storage.ErrKeyNotFound {
		return multipleMiniblockMetadata, nil
	}
	if err != nil {
		return nil, err
	}

	err = mh.marshaller.Unmarshal(multipleMiniblockMetadata, buff)
	if err != nil {
		return mh.tryLegacyUnmarshal(buff)
	}

	return multipleMiniblockMetadata, err
}

func (mh *miniblocksHandler) tryLegacyUnmarshal(buff []byte) (*MiniblockMetadataV2, error) {
	miniblockMetadata := &MiniblockMetadata{}
	err := mh.marshaller.Unmarshal(miniblockMetadata, buff)
	if err != nil {
		return nil, err
	}

	return &MiniblockMetadataV2{
		MiniblockHash:          miniblockMetadata.MiniblockHash,
		SourceShardID:          miniblockMetadata.SourceShardID,
		DestinationShardID:     miniblockMetadata.DestinationShardID,
		Type:                   miniblockMetadata.Type,
		NumTxHashesInMiniblock: 0,
		MiniblocksInfo: []*MiniblockMetadataOnBlock{
			{
				Round:                             miniblockMetadata.Round,
				HeaderNonce:                       miniblockMetadata.HeaderNonce,
				HeaderHash:                        miniblockMetadata.HeaderHash,
				Epoch:                             miniblockMetadata.Epoch,
				NotarizedAtSourceInMetaNonce:      miniblockMetadata.NotarizedAtSourceInMetaNonce,
				NotarizedAtDestinationInMetaNonce: miniblockMetadata.NotarizedAtDestinationInMetaNonce,
				NotarizedAtSourceInMetaHash:       miniblockMetadata.NotarizedAtSourceInMetaHash,
				NotarizedAtDestinationInMetaHash:  miniblockMetadata.NotarizedAtDestinationInMetaHash,
				TxHashesWhenPartial:               nil,
				IndexOfFirstTxProcessed:           0,
			},
		},
	}, nil
}

func (mh *miniblocksHandler) saveMiniblocksMetadata(miniblockHash []byte, multipleMiniblocksMetadata *MiniblockMetadataV2, epoch uint32) error {
	buff, err := mh.marshaller.Marshal(multipleMiniblocksMetadata)
	if err != nil {
		return err
	}

	return mh.miniblocksMetadataStorer.PutInEpoch(miniblockHash, buff, epoch)
}

func (mh *miniblocksHandler) blockReverted(header data.HeaderHandler, blockBody *block.Body) error {
	headerHash, errCalculate := core.CalculateHash(mh.marshaller, mh.hasher, header)
	if errCalculate != nil {
		return fmt.Errorf("%w for header, header round %d", errCalculate, header.GetRound())
	}

	for index, mb := range blockBody.MiniBlocks {
		miniblockHash, err := core.CalculateHash(mh.marshaller, mh.hasher, mb)
		if err != nil {
			return fmt.Errorf("%w for miniblock at index %d, header hash %x", err, index, headerHash)
		}

		err = mh.revertMiniblock(header, headerHash, mb, miniblockHash)
		if err != nil {
			return fmt.Errorf("%w for miniblock at index %d, mbHash %x, header hash %x", err, index, miniblockHash, headerHash)
		}
	}

	return nil
}

func (mh *miniblocksHandler) revertMiniblock(header data.HeaderHandler, headerHash []byte, mb *block.MiniBlock, miniblockHash []byte) error {
	miniblockHeader, err := mh.findMiniblockHandler(header, miniblockHash)
	if err != nil {
		return err
	}

	if miniblockHeader.GetIndexOfFirstTxProcessed() == 0 {
		// we either revert a complete miniblock or the first partial executed miniblock header
		err = mh.epochIndex.removeEpochByHash(miniblockHash)
		if err != nil {
			return err
		}
	}

	err = mh.removeExecutedTransactions(miniblockHeader, mb)
	if err != nil {
		return err
	}

	return mh.removeMiniblockMetadata(header, headerHash, miniblockHash)
}

// removeExecutedTransactions will remove only the transactions between the start and end index of the provided miniblockHeader
func (mh *miniblocksHandler) removeExecutedTransactions(miniblockHeader data.MiniBlockHeaderHandler, mb *block.MiniBlock) error {
	txHashes := mh.getExecutedTxHashes(miniblockHeader, mb)
	for _, txHash := range txHashes {
		err := mh.miniblockHashByTxHashIndexStorer.Remove(txHash)
		if err != nil {
			return fmt.Errorf("%w for tx hash %x", err, txHash)
		}
	}

	return nil
}

func (mh *miniblocksHandler) removeMiniblockMetadata(
	header data.HeaderHandler,
	headerHash []byte,
	miniblockHash []byte,
) error {
	loadedMiniblockMetadata, err := mh.loadExistingMiniblocksMetadata(miniblockHash, header.GetEpoch())
	if err != nil {
		return err
	}

	newMiniblocksInfo := make([]*MiniblockMetadataOnBlock, 0, len(loadedMiniblockMetadata.MiniblocksInfo))
	for _, mbOnBlock := range loadedMiniblockMetadata.MiniblocksInfo {
		if !bytes.Equal(mbOnBlock.HeaderHash, headerHash) {
			newMiniblocksInfo = append(newMiniblocksInfo, mbOnBlock)
		}
	}
	loadedMiniblockMetadata.MiniblocksInfo = newMiniblocksInfo

	if len(loadedMiniblockMetadata.MiniblocksInfo) == 0 {
		// we assume that the rollback is done in-sync with the node.
		// TODO: refactor pruningStorer and add RemoveFromEpoch functionality
		err = mh.miniblocksMetadataStorer.RemoveFromCurrentEpoch(miniblockHash)
	} else {
		err = mh.saveMiniblocksMetadata(miniblockHash, loadedMiniblockMetadata, header.GetEpoch())
	}

	return err
}
