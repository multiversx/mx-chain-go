package dblookupext

import (
	"bytes"
	"errors"
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
	miniblockHashByTxHashIndexStorer storage.Storer // static storer
	miniblocksMetadataStorer         storage.Storer
}

func (mh *miniblocksHandler) commitMiniblock(header data.HeaderHandler, headerHash []byte, mb *block.MiniBlock) error {
	miniblockHash, err := core.CalculateHash(mh.marshaller, mh.hasher, mb)
	if err != nil {
		return err
	}

	miniblockHeader := mh.findOrGenerateMiniblockHandler(header, miniblockHash, mb)
	if miniblockHeader.GetIndexOfFirstTxProcessed() == 0 {
		// first time we see this miniblock, we should store the (miniblock, epoch) tuple
		err = mh.epochIndex.saveEpochByHash(miniblockHash, header.GetEpoch())
		if err != nil {
			return err
		}
	}

	err = mh.commitExecutedTransactions(miniblockHeader, miniblockHash, mb)
	if err != nil {
		return err
	}

	return mh.commitMiniblockMetadata(header, headerHash, miniblockHeader, miniblockHash, mb)
}

func (mh *miniblocksHandler) findOrGenerateMiniblockHandler(header data.HeaderHandler, miniblockHash []byte, mb *block.MiniBlock) data.MiniBlockHeaderHandler {
	for _, mbHeader := range header.GetMiniBlockHeaderHandlers() {
		if bytes.Equal(mbHeader.GetHash(), miniblockHash) {
			return mbHeader
		}
	}

	// miniblock header was not found, it should be an internal generated miniblock. Will create a default miniblock header
	return &block.MiniBlockHeader{
		Hash:            miniblockHash,
		SenderShardID:   mb.SenderShardID,
		ReceiverShardID: mb.ReceiverShardID,
		TxCount:         uint32(len(mb.TxHashes)),
		Type:            mb.Type,
	}
}

// commitExecutedTransactions will save only the transactions between the start and end index of the provided miniblockHeader
func (mh *miniblocksHandler) commitExecutedTransactions(miniblockHeader data.MiniBlockHeaderHandler, miniblockHash []byte, mb *block.MiniBlock) error {
	txHashes := mh.getOrderedExecutedTxHashes(miniblockHeader, mb)
	for _, txHash := range txHashes {
		err := mh.miniblockHashByTxHashIndexStorer.Put(txHash, miniblockHash)
		if err != nil {
			return fmt.Errorf("%w for tx hash %x, miniblock hash %x", err, txHash, miniblockHash)
		}
	}

	return nil
}

func (mh *miniblocksHandler) getOrderedExecutedTxHashes(miniblockHeader data.MiniBlockHeaderHandler, mb *block.MiniBlock) [][]byte {
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

	var txHashesWhenPartial [][]byte

	if miniblockHeader.GetConstructionState() == int32(block.PartialExecuted) {
		loadedMiniblockMetadata, err := mh.loadExistingMiniblocksMetadata(miniblockHash, header.GetEpoch())
		if err != nil {
			return err
		}
		if len(loadedMiniblockMetadata.MiniblockHash) != 0 {
			// rewrite, we found an existing info in storage
			miniblockMetaData = loadedMiniblockMetadata
		}
		txHashesWhenPartial = mh.getOrderedExecutedTxHashes(miniblockHeader, mb)
	}

	miniblockOnHeader := &MiniblockMetadataOnBlock{
		Round:                   header.GetRound(),
		HeaderNonce:             header.GetNonce(),
		HeaderHash:              headerHash,
		Epoch:                   header.GetEpoch(),
		TxHashesWhenPartial:     txHashesWhenPartial,
		IndexOfFirstTxProcessed: miniblockHeader.GetIndexOfFirstTxProcessed(),
	}

	miniblockMetaData.MiniblocksInfo = append(miniblockMetaData.MiniblocksInfo, miniblockOnHeader)

	return mh.saveMiniblocksMetadata(miniblockHash, miniblockMetaData, header.GetEpoch())
}

func (mh *miniblocksHandler) loadExistingMiniblocksMetadata(miniblockHash []byte, epoch uint32) (*MiniblockMetadataV2, error) {
	miniblockMetadata := &MiniblockMetadataV2{}
	buff, err := mh.miniblocksMetadataStorer.GetFromEpoch(miniblockHash, epoch)
	if errors.Is(err, storage.ErrKeyNotFound) {
		return miniblockMetadata, nil
	}
	if err != nil {
		return nil, err
	}

	err = mh.marshaller.Unmarshal(miniblockMetadata, buff)
	if err != nil {
		return mh.tryLegacyUnmarshal(buff)
	}

	return miniblockMetadata, err
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

func (mh *miniblocksHandler) saveMiniblocksMetadata(miniblockHash []byte, miniblockMetadata *MiniblockMetadataV2, epoch uint32) error {
	buff, err := mh.marshaller.Marshal(miniblockMetadata)
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
	miniblockHeader := mh.findOrGenerateMiniblockHandler(header, miniblockHash, mb)

	if miniblockHeader.GetIndexOfFirstTxProcessed() == 0 {
		// we either revert a complete miniblock or the first partial executed miniblock header
		err := mh.epochIndex.removeEpochByHash(miniblockHash)
		if err != nil {
			return err
		}
	}

	err := mh.removeExecutedTransactions(miniblockHeader, mb)
	if err != nil {
		return err
	}

	return mh.removeMiniblockMetadata(header, headerHash, miniblockHash)
}

// removeExecutedTransactions will remove only the transactions between the start and end index of the provided miniblockHeader
func (mh *miniblocksHandler) removeExecutedTransactions(miniblockHeader data.MiniBlockHeaderHandler, mb *block.MiniBlock) error {
	txHashes := mh.getOrderedExecutedTxHashes(miniblockHeader, mb)
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
		return mh.miniblocksMetadataStorer.RemoveFromCurrentEpoch(miniblockHash)
	}

	return mh.saveMiniblocksMetadata(miniblockHash, loadedMiniblockMetadata, header.GetEpoch())
}

func (mh *miniblocksHandler) loadAllExistingMiniblocksMetadata(miniblockHash []byte, epoch uint32) (*MiniblockMetadataV2, error) {
	miniblockMetadataFirstEpoch, err := mh.loadExistingMiniblocksMetadata(miniblockHash, epoch)
	if err != nil {
		return nil, err
	}
	numTxHashes := 0
	for _, mbInfo := range miniblockMetadataFirstEpoch.MiniblocksInfo {
		numTxHashes += len(mbInfo.TxHashesWhenPartial)
	}
	if numTxHashes == 0 || numTxHashes == int(miniblockMetadataFirstEpoch.NumTxHashesInMiniblock) {
		return miniblockMetadataFirstEpoch, nil
	}

	// we will try the next epoch, maybe we will find the rest of the miniblockMetadata info
	miniblockMetadataNextEpoch, err := mh.loadExistingMiniblocksMetadata(miniblockHash, epoch+1)
	if err != nil {
		// no, return what we found
		return miniblockMetadataFirstEpoch, nil
	}
	miniblockMetadataFirstEpoch.MiniblocksInfo = append(miniblockMetadataFirstEpoch.MiniblocksInfo, miniblockMetadataNextEpoch.MiniblocksInfo...)

	return miniblockMetadataFirstEpoch, nil
}

func (mh *miniblocksHandler) getMiniblockMetadataByMiniblockHash(miniblockHash []byte) ([]*MiniblockMetadata, error) {
	miniblockMetadata, err := mh.retrieveMiniblockMetadata(miniblockHash)
	if err != nil {
		return nil, err
	}

	return convertMiniblockMetadataV1ToV2(miniblockMetadata), nil
}

func (mh *miniblocksHandler) retrieveMiniblockMetadata(miniblockHash []byte) (*MiniblockMetadataV2, error) {
	epoch, err := mh.epochIndex.getEpochByHash(miniblockHash)
	if err != nil {
		return nil, err
	}

	return mh.loadAllExistingMiniblocksMetadata(miniblockHash, epoch)
}

func (mh *miniblocksHandler) getMiniblockMetadataByTxHash(txHash []byte) (*MiniblockMetadata, error) {
	miniblockHash, err := mh.miniblockHashByTxHashIndexStorer.Get(txHash)
	if err != nil {
		return nil, err
	}

	miniblockMetadata, err := mh.retrieveMiniblockMetadata(miniblockHash)
	if err != nil {
		return nil, err
	}

	for _, mbInfo := range miniblockMetadata.MiniblocksInfo {
		if hasTxHashInMiniblockMetadataOnBlock(mbInfo, txHash) {
			miniblockMetadata.MiniblocksInfo = []*MiniblockMetadataOnBlock{mbInfo}
			mbData := convertMiniblockMetadataV1ToV2(miniblockMetadata)
			return mbData[0], nil
		}
	}

	return nil, fmt.Errorf("programming error, %w for txhash %x and mb hash %x",
		storage.ErrKeyNotFound, txHash, miniblockHash)
}

func convertMiniblockMetadataV1ToV2(miniblockMetadata *MiniblockMetadataV2) []*MiniblockMetadata {
	result := make([]*MiniblockMetadata, 0, len(miniblockMetadata.MiniblocksInfo))
	for _, mbInfo := range miniblockMetadata.MiniblocksInfo {
		mbv1 := &MiniblockMetadata{
			SourceShardID:                     miniblockMetadata.SourceShardID,
			DestinationShardID:                miniblockMetadata.DestinationShardID,
			Round:                             mbInfo.Round,
			HeaderNonce:                       mbInfo.HeaderNonce,
			HeaderHash:                        mbInfo.HeaderHash,
			MiniblockHash:                     miniblockMetadata.MiniblockHash,
			Epoch:                             mbInfo.Epoch,
			NotarizedAtSourceInMetaNonce:      mbInfo.NotarizedAtSourceInMetaNonce,
			NotarizedAtDestinationInMetaNonce: mbInfo.NotarizedAtDestinationInMetaNonce,
			NotarizedAtSourceInMetaHash:       mbInfo.NotarizedAtSourceInMetaHash,
			NotarizedAtDestinationInMetaHash:  mbInfo.NotarizedAtDestinationInMetaHash,
			Type:                              miniblockMetadata.Type,
		}

		result = append(result, mbv1)
	}

	return result
}

func hasTxHashInMiniblockMetadataOnBlock(mbInfo *MiniblockMetadataOnBlock, txHash []byte) bool {
	if len(mbInfo.TxHashesWhenPartial) == 0 {
		return true
	}

	for _, executedTxHash := range mbInfo.TxHashesWhenPartial {
		if bytes.Equal(executedTxHash, txHash) {
			return true
		}
	}

	return false
}

func (mh *miniblocksHandler) updateMiniblockMetadataOnBlock(
	miniblockHash []byte,
	headerHash []byte,
	updateHandler func(mbMetadataOnBlock *MiniblockMetadataOnBlock),
) error {
	epoch, err := mh.epochIndex.getEpochByHash(miniblockHash)
	if err != nil {
		return err
	}

	updated, err := mh.updateMiniblockMetadataOnBlockInEpoch(epoch, miniblockHash, headerHash, updateHandler)
	if err != nil {
		return err
	}
	if updated {
		// found the block hash, no need to search in the next epoch
		return nil
	}

	updated, err = mh.updateMiniblockMetadataOnBlockInEpoch(epoch+1, miniblockHash, headerHash, updateHandler)
	if err != nil {
		return err
	}
	if !updated {
		// programming error: not found in epoch+1, the blockhash should have been written in either epoch `epoch`
		// or epoch `epoch+1`
		return storage.ErrKeyNotFound
	}

	return nil
}

func (mh *miniblocksHandler) updateMiniblockMetadataOnBlockInEpoch(
	epoch uint32,
	miniblockHash []byte,
	headerHash []byte,
	updateHandler func(mbMetadataOnBlock *MiniblockMetadataOnBlock),
) (bool, error) {
	miniblockMetadata, err := mh.loadExistingMiniblocksMetadata(miniblockHash, epoch)
	if err != nil {
		return false, err
	}

	for _, mbInfo := range miniblockMetadata.MiniblocksInfo {
		if bytes.Equal(mbInfo.HeaderHash, headerHash) {
			updateHandler(mbInfo)
			return true, mh.saveMiniblocksMetadata(miniblockHash, miniblockMetadata, epoch)
		}
	}

	return false, nil
}
