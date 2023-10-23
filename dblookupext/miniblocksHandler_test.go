package dblookupext

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/assert"
)

// we need GogoProtoMarshalizer as it errors if we try to unmarshal incompatible bytes slices
var testMarshaller = &marshal.GogoProtoMarshalizer{}
var testHasher = &hashingMocks.HasherMock{}

func createMockMiniblocksHandler() *miniblocksHandler {
	epochIndexStorer := genericMocks.NewStorerMockWithErrKeyNotFound(0)

	return &miniblocksHandler{
		marshaller:                       testMarshaller,
		hasher:                           testHasher,
		epochIndex:                       newHashToEpochIndex(epochIndexStorer, testMarshaller),
		miniblockHashByTxHashIndexStorer: genericMocks.NewStorerMockWithErrKeyNotFound(0),
		miniblocksMetadataStorer:         genericMocks.NewStorerMockWithErrKeyNotFound(0),
	}
}

func checkTxs(miniblockHashByTxHashIndexStorer storage.Storer, txHashes [][]byte, miniblockHash []byte) error {
	for _, txHash := range txHashes {
		value, err := miniblockHashByTxHashIndexStorer.Get(txHash)
		if err != nil {
			return fmt.Errorf("%w for hash %s", err, txHash)
		}

		if !bytes.Equal(miniblockHash, value) {
			return fmt.Errorf("recorded value mismatch, wanted %x, got %x", miniblockHash, value)
		}
	}

	return nil
}

func checkMiniblockHashInEpoch(indexer *epochByHashIndex, miniblockHash []byte, epoch uint32) error {
	savedEpoch, err := indexer.getEpochByHash(miniblockHash)
	if err != nil {
		return err
	}

	if savedEpoch != epoch {
		return fmt.Errorf("mismatch between saved epoch %d and comparing epoch %d", savedEpoch, epoch)
	}

	return nil
}

func generateTestCompletedMiniblock() (*block.MiniBlock, []byte, *block.MiniBlockHeader) {
	miniblock := &block.MiniBlock{
		TxHashes: [][]byte{
			[]byte("txHashCompleted1"),
			[]byte("txHashCompleted2"),
			[]byte("txHashCompleted3"),
		},
	}
	hash, _ := core.CalculateHash(testMarshaller, testHasher, miniblock)
	miniblockHeader := &block.MiniBlockHeader{
		Hash:     hash,
		TxCount:  uint32(len(miniblock.TxHashes)),
		Type:     0,
		Reserved: nil,
	}
	_ = miniblockHeader.SetConstructionState(int32(block.Final))
	_ = miniblockHeader.SetIndexOfFirstTxProcessed(0)
	_ = miniblockHeader.SetIndexOfLastTxProcessed(int32(len(miniblock.TxHashes) - 1))

	return miniblock, hash, miniblockHeader
}

func generateTestPartialMiniblock() (*block.MiniBlock, []byte, []*block.MiniBlockHeader) {
	miniblock := &block.MiniBlock{
		TxHashes: [][]byte{
			[]byte("txHashPartial1"),
			[]byte("txHashPartial2"),
			[]byte("txHashPartial3"),
			[]byte("txHashPartial4"),
			[]byte("txHashPartial5"),
			[]byte("txHashPartial6"),
		},
	}
	hash, _ := core.CalculateHash(testMarshaller, testHasher, miniblock)
	miniblockHeader1 := &block.MiniBlockHeader{
		Hash:     hash,
		TxCount:  uint32(len(miniblock.TxHashes)),
		Type:     0,
		Reserved: nil,
	}
	_ = miniblockHeader1.SetConstructionState(int32(block.PartialExecuted))
	_ = miniblockHeader1.SetIndexOfFirstTxProcessed(0)
	_ = miniblockHeader1.SetIndexOfLastTxProcessed(1)

	miniblockHeader2 := &block.MiniBlockHeader{
		Hash:     hash,
		TxCount:  uint32(len(miniblock.TxHashes)),
		Type:     0,
		Reserved: nil,
	}
	_ = miniblockHeader2.SetConstructionState(int32(block.PartialExecuted))
	_ = miniblockHeader2.SetIndexOfFirstTxProcessed(2)
	_ = miniblockHeader2.SetIndexOfLastTxProcessed(3)

	miniblockHeader3 := &block.MiniBlockHeader{
		Hash:     hash,
		TxCount:  uint32(len(miniblock.TxHashes)),
		Type:     0,
		Reserved: nil,
	}
	_ = miniblockHeader3.SetConstructionState(int32(block.PartialExecuted))
	_ = miniblockHeader3.SetIndexOfFirstTxProcessed(4)
	_ = miniblockHeader3.SetIndexOfLastTxProcessed(5)

	return miniblock, hash, []*block.MiniBlockHeader{miniblockHeader1, miniblockHeader2, miniblockHeader3}
}

func generateHeader(epoch uint32, nonce uint64, miniblockHeaders ...block.MiniBlockHeader) *block.Header {
	return &block.Header{
		Epoch:            epoch,
		Nonce:            nonce,
		Round:            nonce + 100,
		MiniBlockHeaders: miniblockHeaders,
	}
}

func TestMiniblocksHandler_blockCommitted(t *testing.T) {
	t.Parallel()

	completeMiniblock, completeMiniblockHash, completeMiniblockHeader := generateTestCompletedMiniblock()
	partialMiniblock, partialMiniblockHash, partialMiniblockHeaders := generateTestPartialMiniblock()

	t.Run("should work with completed miniblock", func(t *testing.T) {
		t.Parallel()

		mbHandler := createMockMiniblocksHandler()

		header := generateHeader(37, 22933, *completeMiniblockHeader)
		headerHash, err := core.CalculateHash(testMarshaller, testHasher, header)
		assert.Nil(t, err)

		// commit the completed miniblock and check everything was written
		err = mbHandler.commitMiniblock(header, headerHash, completeMiniblock)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, completeMiniblock.TxHashes, completeMiniblockHash)
		assert.Nil(t, err)
		err = checkMiniblockHashInEpoch(mbHandler.epochIndex, completeMiniblockHash, header.Epoch)
		assert.Nil(t, err)

		expectedMetadata := &MiniblockMetadataV2{
			MiniblockHash:          completeMiniblockHash,
			NumTxHashesInMiniblock: int32(len(completeMiniblock.TxHashes)),
			MiniblocksInfo: []*MiniblockMetadataOnBlock{
				{
					Round:                   header.Round,
					HeaderNonce:             header.Nonce,
					HeaderHash:              headerHash,
					Epoch:                   header.Epoch,
					TxHashesWhenPartial:     nil, // completed, we do not store tx hashes
					IndexOfFirstTxProcessed: 0,
				},
			},
		}

		recoveredMetadata, err := mbHandler.loadExistingMiniblocksMetadata(completeMiniblockHash, header.GetEpoch())
		assert.Nil(t, err)
		assert.Equal(t, expectedMetadata, recoveredMetadata)
	})
	t.Run("should work with completed miniblock and partial miniblock in same epoch", func(t *testing.T) {
		t.Parallel()

		mbHandler := createMockMiniblocksHandler()

		header1 := generateHeader(37, 22933, *completeMiniblockHeader, *partialMiniblockHeaders[0])
		headerHash1, err := core.CalculateHash(testMarshaller, testHasher, header1)
		assert.Nil(t, err)

		header2 := generateHeader(37, 22934, *partialMiniblockHeaders[1])
		headerHash2, err := core.CalculateHash(testMarshaller, testHasher, header2)
		assert.Nil(t, err)

		// commit header1 with completed miniblocks and check all transactions from that block were written
		err = mbHandler.commitMiniblock(header1, headerHash1, completeMiniblock)
		assert.Nil(t, err)
		err = mbHandler.commitMiniblock(header1, headerHash1, partialMiniblock)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, partialMiniblock.TxHashes[0:2], partialMiniblockHash)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, completeMiniblock.TxHashes, completeMiniblockHash)
		assert.Nil(t, err)

		// commit header2 with patial miniblock and check all transactions from header 2 and header 1 were written
		err = mbHandler.commitMiniblock(header2, headerHash2, partialMiniblock)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, completeMiniblock.TxHashes, completeMiniblockHash)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, partialMiniblock.TxHashes[0:4], partialMiniblockHash)
		assert.Nil(t, err)

		// check the miniblock hashes were written correctly
		err = checkMiniblockHashInEpoch(mbHandler.epochIndex, completeMiniblockHash, header1.Epoch)
		assert.Nil(t, err)
		err = checkMiniblockHashInEpoch(mbHandler.epochIndex, partialMiniblockHash, header1.Epoch)
		assert.Nil(t, err)

		// check the miniblocks metadata were written correctly
		expectedMetadataCompleted := &MiniblockMetadataV2{
			MiniblockHash:          completeMiniblockHash,
			NumTxHashesInMiniblock: int32(len(completeMiniblock.TxHashes)),
			MiniblocksInfo: []*MiniblockMetadataOnBlock{
				{
					Round:                   header1.Round,
					HeaderNonce:             header1.Nonce,
					HeaderHash:              headerHash1,
					Epoch:                   header1.Epoch,
					TxHashesWhenPartial:     nil, // completed, we do not store tx hashes
					IndexOfFirstTxProcessed: 0,
				},
			},
		}
		recoveredMetadata, err := mbHandler.loadExistingMiniblocksMetadata(completeMiniblockHash, header1.GetEpoch())
		assert.Nil(t, err)
		assert.Equal(t, expectedMetadataCompleted, recoveredMetadata)

		expectedMetadataPartial := &MiniblockMetadataV2{
			MiniblockHash:          partialMiniblockHash,
			NumTxHashesInMiniblock: int32(len(partialMiniblock.TxHashes)),
			MiniblocksInfo: []*MiniblockMetadataOnBlock{
				{
					// partial 1
					Round:                   header1.Round,
					HeaderNonce:             header1.Nonce,
					HeaderHash:              headerHash1,
					Epoch:                   header1.Epoch,
					TxHashesWhenPartial:     partialMiniblock.TxHashes[0:2],
					IndexOfFirstTxProcessed: 0,
				},
				{
					// partial 2
					Round:                   header2.Round,
					HeaderNonce:             header2.Nonce,
					HeaderHash:              headerHash2,
					Epoch:                   header2.Epoch,
					TxHashesWhenPartial:     partialMiniblock.TxHashes[2:4],
					IndexOfFirstTxProcessed: 2,
				},
			},
		}
		recoveredMetadata, err = mbHandler.loadExistingMiniblocksMetadata(partialMiniblockHash, header1.GetEpoch())
		assert.Nil(t, err)
		assert.Equal(t, expectedMetadataPartial, recoveredMetadata)
	})
	t.Run("should work with completed miniblock and partial miniblocks in different epochs", func(t *testing.T) {
		t.Parallel()

		mbHandler := createMockMiniblocksHandler()

		header1 := generateHeader(37, 22933, *completeMiniblockHeader, *partialMiniblockHeaders[0])
		headerHash1, err := core.CalculateHash(testMarshaller, testHasher, header1)
		assert.Nil(t, err)

		header2 := generateHeader(37, 22934, *partialMiniblockHeaders[1])
		headerHash2, err := core.CalculateHash(testMarshaller, testHasher, header2)
		assert.Nil(t, err)

		header3 := generateHeader(38, 22935, *partialMiniblockHeaders[2])
		headerHash3, err := core.CalculateHash(testMarshaller, testHasher, header3)
		assert.Nil(t, err)

		// commit header1 with completed miniblock and check all transactions from that block were written
		err = mbHandler.commitMiniblock(header1, headerHash1, completeMiniblock)
		assert.Nil(t, err)
		err = mbHandler.commitMiniblock(header1, headerHash1, partialMiniblock)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, partialMiniblock.TxHashes[0:2], partialMiniblockHash)
		assert.Nil(t, err)

		// commit header2 with partial miniblock and check all transactions from header 2 and header 1 were written
		err = mbHandler.commitMiniblock(header2, headerHash2, partialMiniblock)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, completeMiniblock.TxHashes, completeMiniblockHash)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, partialMiniblock.TxHashes[0:4], partialMiniblockHash)
		assert.Nil(t, err)

		// commit header3 with partial miniblock and check all transactions from header 3, header 2 and header 1 were written in
		// the correct epochs
		err = mbHandler.commitMiniblock(header3, headerHash3, partialMiniblock)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, completeMiniblock.TxHashes, completeMiniblockHash)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, partialMiniblock.TxHashes[0:4], partialMiniblockHash)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, partialMiniblock.TxHashes[5:6], partialMiniblockHash)
		assert.Nil(t, err)

		// check the miniblock hashes were written correctly
		err = checkMiniblockHashInEpoch(mbHandler.epochIndex, completeMiniblockHash, header1.Epoch)
		assert.Nil(t, err)
		// partial miniblock was first notarized in epoch 37
		err = checkMiniblockHashInEpoch(mbHandler.epochIndex, partialMiniblockHash, header1.Epoch)
		assert.Nil(t, err)

		// check the miniblocks metadata were written correctly
		expectedMetadataCompleted := &MiniblockMetadataV2{
			MiniblockHash:          completeMiniblockHash,
			NumTxHashesInMiniblock: int32(len(completeMiniblock.TxHashes)),
			MiniblocksInfo: []*MiniblockMetadataOnBlock{
				{
					Round:                   header1.Round,
					HeaderNonce:             header1.Nonce,
					HeaderHash:              headerHash1,
					Epoch:                   header1.Epoch,
					TxHashesWhenPartial:     nil, // completed, we do not store tx hashes
					IndexOfFirstTxProcessed: 0,
				},
			},
		}
		recoveredMetadata, err := mbHandler.loadExistingMiniblocksMetadata(completeMiniblockHash, header1.GetEpoch())
		assert.Nil(t, err)
		assert.Equal(t, expectedMetadataCompleted, recoveredMetadata)

		expectedMetadataPartial1 := &MiniblockMetadataV2{
			MiniblockHash:          partialMiniblockHash,
			NumTxHashesInMiniblock: int32(len(partialMiniblock.TxHashes)),
			MiniblocksInfo: []*MiniblockMetadataOnBlock{
				{
					// partial 1
					Round:                   header1.Round,
					HeaderNonce:             header1.Nonce,
					HeaderHash:              headerHash1,
					Epoch:                   header1.Epoch,
					TxHashesWhenPartial:     partialMiniblock.TxHashes[0:2],
					IndexOfFirstTxProcessed: 0,
				},
				{
					// partial 2
					Round:                   header2.Round,
					HeaderNonce:             header2.Nonce,
					HeaderHash:              headerHash2,
					Epoch:                   header2.Epoch,
					TxHashesWhenPartial:     partialMiniblock.TxHashes[2:4],
					IndexOfFirstTxProcessed: 2,
				},
			},
		}

		expectedMetadataPartial2 := &MiniblockMetadataV2{
			MiniblockHash:          partialMiniblockHash,
			NumTxHashesInMiniblock: int32(len(partialMiniblock.TxHashes)),
			MiniblocksInfo: []*MiniblockMetadataOnBlock{
				{
					// partial 3
					Round:                   header3.Round,
					HeaderNonce:             header3.Nonce,
					HeaderHash:              headerHash3,
					Epoch:                   header3.Epoch,
					TxHashesWhenPartial:     partialMiniblock.TxHashes[4:6],
					IndexOfFirstTxProcessed: 4,
				},
			},
		}
		// only partial1 and partial2 are written in epoch 37
		recoveredMetadata, err = mbHandler.loadExistingMiniblocksMetadata(partialMiniblockHash, header1.GetEpoch())
		assert.Nil(t, err)
		assert.Equal(t, expectedMetadataPartial1, recoveredMetadata)

		// only partial3 is written in epoch 38
		recoveredMetadata, err = mbHandler.loadExistingMiniblocksMetadata(partialMiniblockHash, header3.GetEpoch())
		assert.Nil(t, err)
		assert.Equal(t, expectedMetadataPartial2, recoveredMetadata)
	})
}

func TestMiniblocksHandler_loadExistingMiniblocksMetadataWithLegacy(t *testing.T) {
	t.Parallel()

	t.Run("minimum data", func(t *testing.T) {
		t.Parallel()

		mbHandler := createMockMiniblocksHandler()
		mbHash := []byte("mb hash")
		miniblockMetadata := &MiniblockMetadata{
			Round:         1,
			HeaderNonce:   1,
			HeaderHash:    []byte("header hash1"),
			Epoch:         1,
			MiniblockHash: mbHash,
		}

		expectedMetadata := &MiniblockMetadataV2{
			MiniblockHash: mbHash,
			MiniblocksInfo: []*MiniblockMetadataOnBlock{
				{
					Round:       1,
					HeaderNonce: 1,
					HeaderHash:  []byte("header hash1"),
					Epoch:       1,
				},
			},
		}

		// save miniblockMetadata in legacy format
		buff, _ := testMarshaller.Marshal(miniblockMetadata)
		_ = mbHandler.miniblocksMetadataStorer.PutInEpoch(mbHash, buff, 1)
		recoveredMetadata, err := mbHandler.loadExistingMiniblocksMetadata(mbHash, 1)
		assert.Nil(t, err)
		assert.Equal(t, expectedMetadata, recoveredMetadata)
	})
	t.Run("all data", func(t *testing.T) {
		t.Parallel()

		mbHandler := createMockMiniblocksHandler()
		mbHash := []byte("mb hash")
		miniblockMetadata := &MiniblockMetadata{
			SourceShardID:                     1,
			DestinationShardID:                2,
			Round:                             3,
			HeaderNonce:                       4,
			HeaderHash:                        []byte("header hash1"),
			MiniblockHash:                     mbHash,
			Epoch:                             5,
			NotarizedAtSourceInMetaNonce:      6,
			NotarizedAtDestinationInMetaNonce: 7,
			NotarizedAtSourceInMetaHash:       []byte("src meta"),
			NotarizedAtDestinationInMetaHash:  []byte("src dest"),
			Type:                              8,
		}

		expectedMetadata := &MiniblockMetadataV2{
			MiniblockHash:          mbHash,
			SourceShardID:          1,
			DestinationShardID:     2,
			Type:                   8,
			NumTxHashesInMiniblock: 0,
			MiniblocksInfo: []*MiniblockMetadataOnBlock{
				{
					Round:                             3,
					HeaderNonce:                       4,
					HeaderHash:                        []byte("header hash1"),
					Epoch:                             5,
					NotarizedAtSourceInMetaNonce:      6,
					NotarizedAtDestinationInMetaNonce: 7,
					NotarizedAtSourceInMetaHash:       []byte("src meta"),
					NotarizedAtDestinationInMetaHash:  []byte("src dest"),
					TxHashesWhenPartial:               nil,
					IndexOfFirstTxProcessed:           0,
				},
			},
		}

		// save miniblockMetadata1 in legacy format
		buff, _ := testMarshaller.Marshal(miniblockMetadata)
		_ = mbHandler.miniblocksMetadataStorer.PutInEpoch(mbHash, buff, 1)
		recoveredMetadata, err := mbHandler.loadExistingMiniblocksMetadata(mbHash, 1)
		assert.Nil(t, err)
		assert.Equal(t, expectedMetadata, recoveredMetadata)
	})
}

func TestMiniblocksHandler_blockCommittedAndBlockReverted(t *testing.T) {
	t.Parallel()

	completeMiniblock, completeMiniblockHash, completeMiniblockHeader := generateTestCompletedMiniblock()
	partialMiniblock, partialMiniblockHash, partialMiniblockHeaders := generateTestPartialMiniblock()

	mbHandler := createMockMiniblocksHandler()

	header1 := generateHeader(37, 22933, *completeMiniblockHeader, *partialMiniblockHeaders[0])
	body1 := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			completeMiniblock,
			partialMiniblock,
		},
	}
	headerHash1, err := core.CalculateHash(testMarshaller, testHasher, header1)
	assert.Nil(t, err)

	header2 := generateHeader(37, 22934, *partialMiniblockHeaders[1])
	body2 := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			partialMiniblock,
		},
	}
	headerHash2, err := core.CalculateHash(testMarshaller, testHasher, header2)
	assert.Nil(t, err)

	header3 := generateHeader(38, 22935, *partialMiniblockHeaders[2])
	body3 := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			partialMiniblock,
		},
	}
	headerHash3, err := core.CalculateHash(testMarshaller, testHasher, header3)
	assert.Nil(t, err)

	err = mbHandler.commitMiniblock(header1, headerHash1, completeMiniblock)
	assert.Nil(t, err)
	err = mbHandler.commitMiniblock(header1, headerHash1, partialMiniblock)
	assert.Nil(t, err)
	err = mbHandler.commitMiniblock(header2, headerHash2, partialMiniblock)
	assert.Nil(t, err)
	err = mbHandler.commitMiniblock(header3, headerHash3, partialMiniblock)
	assert.Nil(t, err)

	//we set the current epochs in storers, this will not be necessary when implementing RemoveFromEpoch
	mbHandler.miniblocksMetadataStorer.(*genericMocks.StorerMock).SetCurrentEpoch(header3.Epoch)
	// miniblockHashByTxHashIndexStorer is a static storer, set epoch won't have an efect
	// epochIndex contains a static storer, set epoch won't have an efect

	// ### rollback on header 3
	err = mbHandler.blockReverted(header3, body3)
	assert.Nil(t, err)

	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, completeMiniblock.TxHashes, completeMiniblockHash)
	assert.Nil(t, err)
	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, partialMiniblock.TxHashes[0:4], partialMiniblockHash)
	assert.Nil(t, err)
	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, partialMiniblock.TxHashes[5:6], partialMiniblockHash)
	assert.NotNil(t, err) // these were reverted

	// check the miniblock hashes were written correctly
	err = checkMiniblockHashInEpoch(mbHandler.epochIndex, completeMiniblockHash, header1.Epoch)
	assert.Nil(t, err)
	// partial miniblock was first notarized in epoch 37
	err = checkMiniblockHashInEpoch(mbHandler.epochIndex, partialMiniblockHash, header1.Epoch)
	assert.Nil(t, err) // no revert at this point

	// check the miniblocks metadata were written correctly
	expectedMetadataCompleted := &MiniblockMetadataV2{
		MiniblockHash:          completeMiniblockHash,
		NumTxHashesInMiniblock: int32(len(completeMiniblock.TxHashes)),
		MiniblocksInfo: []*MiniblockMetadataOnBlock{
			{
				Round:                   header1.Round,
				HeaderNonce:             header1.Nonce,
				HeaderHash:              headerHash1,
				Epoch:                   header1.Epoch,
				TxHashesWhenPartial:     nil, // completed, we do not store tx hashes
				IndexOfFirstTxProcessed: 0,
			},
		},
	}
	recoveredMetadata, err := mbHandler.loadExistingMiniblocksMetadata(completeMiniblockHash, header1.GetEpoch())
	assert.Nil(t, err)
	assert.Equal(t, expectedMetadataCompleted, recoveredMetadata)

	expectedMetadataPartial1 := &MiniblockMetadataV2{
		MiniblockHash:          partialMiniblockHash,
		NumTxHashesInMiniblock: int32(len(partialMiniblock.TxHashes)),
		MiniblocksInfo: []*MiniblockMetadataOnBlock{
			{
				// partial 1
				Round:                   header1.Round,
				HeaderNonce:             header1.Nonce,
				HeaderHash:              headerHash1,
				Epoch:                   header1.Epoch,
				TxHashesWhenPartial:     partialMiniblock.TxHashes[0:2],
				IndexOfFirstTxProcessed: 0,
			},
			{
				// partial 2
				Round:                   header2.Round,
				HeaderNonce:             header2.Nonce,
				HeaderHash:              headerHash2,
				Epoch:                   header2.Epoch,
				TxHashesWhenPartial:     partialMiniblock.TxHashes[2:4],
				IndexOfFirstTxProcessed: 2,
			},
		},
	}

	// only partial1 and partial2 are written in epoch 37
	recoveredMetadata, err = mbHandler.loadExistingMiniblocksMetadata(partialMiniblockHash, header1.GetEpoch())
	assert.Nil(t, err)
	assert.Equal(t, expectedMetadataPartial1, recoveredMetadata)

	recoveredMetadata, err = mbHandler.loadExistingMiniblocksMetadata(partialMiniblockHash, header3.GetEpoch())
	assert.Nil(t, err)
	assert.Equal(t, &MiniblockMetadataV2{}, recoveredMetadata) // revert occurred in epoch 38

	//we set the current epochs in storers, this will not be necessary when implementing RemoveFromEpoch
	mbHandler.miniblocksMetadataStorer.(*genericMocks.StorerMock).SetCurrentEpoch(header2.Epoch)
	// miniblockHashByTxHashIndexStorer is a static storer, set epoch won't have an efect
	// epochIndex contains a static storer, set epoch won't have an efect

	// ### rollback on header 2
	err = mbHandler.blockReverted(header2, body2)
	assert.Nil(t, err)

	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, completeMiniblock.TxHashes, completeMiniblockHash)
	assert.Nil(t, err)
	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, partialMiniblock.TxHashes[0:4], partialMiniblockHash)
	assert.NotNil(t, err) // these were reverted
	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, partialMiniblock.TxHashes[5:6], partialMiniblockHash)
	assert.NotNil(t, err) // these were reverted

	// check the miniblock hashes were written correctly
	err = checkMiniblockHashInEpoch(mbHandler.epochIndex, completeMiniblockHash, header1.Epoch)
	assert.Nil(t, err)
	// partial miniblock was first notarized in epoch 37
	err = checkMiniblockHashInEpoch(mbHandler.epochIndex, partialMiniblockHash, header1.Epoch)
	assert.Nil(t, err) // no revert at this point

	// check the miniblocks metadata were written correctly
	recoveredMetadata, err = mbHandler.loadExistingMiniblocksMetadata(completeMiniblockHash, header1.GetEpoch())
	assert.Nil(t, err)
	assert.Equal(t, expectedMetadataCompleted, recoveredMetadata)

	expectedMetadataPartial1 = &MiniblockMetadataV2{
		MiniblockHash:          partialMiniblockHash,
		NumTxHashesInMiniblock: int32(len(partialMiniblock.TxHashes)),
		MiniblocksInfo: []*MiniblockMetadataOnBlock{
			{
				// partial 1
				Round:                   header1.Round,
				HeaderNonce:             header1.Nonce,
				HeaderHash:              headerHash1,
				Epoch:                   header1.Epoch,
				TxHashesWhenPartial:     partialMiniblock.TxHashes[0:2],
				IndexOfFirstTxProcessed: 0,
			},
		},
	}

	// only partial1 remains in epoch 37
	recoveredMetadata, err = mbHandler.loadExistingMiniblocksMetadata(partialMiniblockHash, header1.GetEpoch())
	assert.Nil(t, err)
	assert.Equal(t, expectedMetadataPartial1, recoveredMetadata)

	recoveredMetadata, err = mbHandler.loadExistingMiniblocksMetadata(partialMiniblockHash, header3.GetEpoch())
	assert.Nil(t, err)
	assert.Equal(t, &MiniblockMetadataV2{}, recoveredMetadata) // revert occurred in epoch 38

	// ### rollback on header 1
	err = mbHandler.blockReverted(header1, body1)
	assert.Nil(t, err)

	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, completeMiniblock.TxHashes, completeMiniblockHash)
	assert.NotNil(t, err) // these were reverted
	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, partialMiniblock.TxHashes[0:4], partialMiniblockHash)
	assert.NotNil(t, err) // these were reverted
	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, partialMiniblock.TxHashes[5:6], partialMiniblockHash)
	assert.NotNil(t, err) // these were reverted

	// check the miniblock hashes are not present anymore
	err = checkMiniblockHashInEpoch(mbHandler.epochIndex, completeMiniblockHash, header1.Epoch)
	assert.NotNil(t, err) // reverted
	err = checkMiniblockHashInEpoch(mbHandler.epochIndex, partialMiniblockHash, header1.Epoch)
	assert.NotNil(t, err) // reverted

	// check the miniblocks metadata were completely reverted
	recoveredMetadata, err = mbHandler.loadExistingMiniblocksMetadata(completeMiniblockHash, header1.GetEpoch())
	assert.Nil(t, err)
	assert.Equal(t, &MiniblockMetadataV2{}, recoveredMetadata)

	recoveredMetadata, err = mbHandler.loadExistingMiniblocksMetadata(partialMiniblockHash, header1.GetEpoch())
	assert.Nil(t, err)
	assert.Equal(t, &MiniblockMetadataV2{}, recoveredMetadata)

	recoveredMetadata, err = mbHandler.loadExistingMiniblocksMetadata(partialMiniblockHash, header3.GetEpoch())
	assert.Nil(t, err)
	assert.Equal(t, &MiniblockMetadataV2{}, recoveredMetadata) // revert occurred in epoch 38
}

func TestMiniblocksHandler_getMiniblockMetadataByTxHashAndMbHash(t *testing.T) {
	t.Parallel()

	completeMiniblock, completeMiniblockHash, completeMiniblockHeader := generateTestCompletedMiniblock()
	partialMiniblock, partialMiniblockHash, partialMiniblockHeaders := generateTestPartialMiniblock()

	mbHandler := createMockMiniblocksHandler()

	header1 := generateHeader(37, 22933, *completeMiniblockHeader, *partialMiniblockHeaders[0])
	headerHash1, err := core.CalculateHash(testMarshaller, testHasher, header1)
	assert.Nil(t, err)

	header2 := generateHeader(37, 22934, *partialMiniblockHeaders[1])
	headerHash2, err := core.CalculateHash(testMarshaller, testHasher, header2)
	assert.Nil(t, err)

	header3 := generateHeader(38, 22935, *partialMiniblockHeaders[2])
	headerHash3, err := core.CalculateHash(testMarshaller, testHasher, header3)
	assert.Nil(t, err)

	err = mbHandler.commitMiniblock(header1, headerHash1, completeMiniblock)
	assert.Nil(t, err)

	err = mbHandler.commitMiniblock(header1, headerHash1, partialMiniblock)
	assert.Nil(t, err)

	completedExecutedMiniblock := &MiniblockMetadata{
		Round:         23033,
		HeaderNonce:   22933,
		HeaderHash:    headerHash1,
		MiniblockHash: completeMiniblockHash,
		Epoch:         37,
	}
	//all completed txs should return the completed miniblock
	for _, txHash := range completeMiniblock.TxHashes {
		mb, errGet := mbHandler.getMiniblockMetadataByTxHash(txHash)
		assert.Nil(t, errGet)
		assert.Equal(t, completedExecutedMiniblock, mb)
	}

	partiallyExecutedMiniblock1 := &MiniblockMetadata{
		Round:         23033,
		HeaderNonce:   22933,
		HeaderHash:    headerHash1,
		MiniblockHash: partialMiniblockHash,
		Epoch:         37,
	}
	//the partial executed txs should return the partial miniblock
	for i := partialMiniblockHeaders[0].GetIndexOfFirstTxProcessed(); i <= partialMiniblockHeaders[0].GetIndexOfLastTxProcessed(); i++ {
		txHash := partialMiniblock.TxHashes[i]
		mb, errGet := mbHandler.getMiniblockMetadataByTxHash(txHash)
		assert.Nil(t, errGet)
		assert.Equal(t, partiallyExecutedMiniblock1, mb)
	}
	//a transaction from the second partially executed miniblock should return error
	indexForNotFound := partialMiniblockHeaders[1].GetIndexOfFirstTxProcessed()
	txHashNotFound := partialMiniblock.TxHashes[indexForNotFound]
	mb, errGet := mbHandler.getMiniblockMetadataByTxHash(txHashNotFound)
	assert.NotNil(t, errGet)
	assert.Nil(t, mb)

	err = mbHandler.commitMiniblock(header2, headerHash2, partialMiniblock)
	assert.Nil(t, err)
	err = mbHandler.commitMiniblock(header3, headerHash3, partialMiniblock)
	assert.Nil(t, err)

	//we set the current epochs in storers, this will not be necessary when implementing RemoveFromEpoch
	mbHandler.miniblocksMetadataStorer.(*genericMocks.StorerMock).SetCurrentEpoch(header3.Epoch)

	//all completed txs should return the completed miniblock
	for _, txHash := range completeMiniblock.TxHashes {
		mb, errGet = mbHandler.getMiniblockMetadataByTxHash(txHash)
		assert.Nil(t, errGet)
		assert.Equal(t, completedExecutedMiniblock, mb)
	}
	//the partial executed txs should return the partial miniblock
	for i := partialMiniblockHeaders[0].GetIndexOfFirstTxProcessed(); i <= partialMiniblockHeaders[0].GetIndexOfLastTxProcessed(); i++ {
		txHash := partialMiniblock.TxHashes[i]
		mb, errGet = mbHandler.getMiniblockMetadataByTxHash(txHash)
		assert.Nil(t, errGet)
		assert.Equal(t, partiallyExecutedMiniblock1, mb)
	}

	partiallyExecutedMiniblock2 := &MiniblockMetadata{
		Round:         23034,
		HeaderNonce:   22934,
		HeaderHash:    headerHash2,
		MiniblockHash: partialMiniblockHash,
		Epoch:         37,
	}
	partiallyExecutedMiniblock3 := &MiniblockMetadata{
		Round:         23035,
		HeaderNonce:   22935,
		HeaderHash:    headerHash3,
		MiniblockHash: partialMiniblockHash,
		Epoch:         38,
	}
	//the partial executed txs should return the partial miniblock
	for i := partialMiniblockHeaders[1].GetIndexOfFirstTxProcessed(); i <= partialMiniblockHeaders[1].GetIndexOfLastTxProcessed(); i++ {
		txHash := partialMiniblock.TxHashes[i]
		mb, errGet = mbHandler.getMiniblockMetadataByTxHash(txHash)
		assert.Nil(t, errGet)
		assert.Equal(t, partiallyExecutedMiniblock2, mb)
	}
	for i := partialMiniblockHeaders[2].GetIndexOfFirstTxProcessed(); i <= partialMiniblockHeaders[2].GetIndexOfLastTxProcessed(); i++ {
		txHash := partialMiniblock.TxHashes[i]
		mb, errGet = mbHandler.getMiniblockMetadataByTxHash(txHash)
		assert.Nil(t, errGet)
		assert.Equal(t, partiallyExecutedMiniblock3, mb)
	}

	miniblocksMetadata, err := mbHandler.getMiniblockMetadataByMiniblockHash(completeMiniblockHash)
	assert.Nil(t, err)
	assert.Equal(t, []*MiniblockMetadata{completedExecutedMiniblock}, miniblocksMetadata)

	miniblocksMetadata, err = mbHandler.getMiniblockMetadataByMiniblockHash(partialMiniblockHash)
	assert.Nil(t, err)
	assert.Equal(t, []*MiniblockMetadata{partiallyExecutedMiniblock1, partiallyExecutedMiniblock2, partiallyExecutedMiniblock3}, miniblocksMetadata)
}

func TestMiniblocksHandler_updateMiniblockMetadataOnBlock(t *testing.T) {
	t.Parallel()

	completeMiniblock, completeMiniblockHash, completeMiniblockHeader := generateTestCompletedMiniblock()
	partialMiniblock, partialMiniblockHash, partialMiniblockHeaders := generateTestPartialMiniblock()

	mbHandler := createMockMiniblocksHandler()

	header1 := generateHeader(37, 22933, *completeMiniblockHeader, *partialMiniblockHeaders[0])
	headerHash1, err := core.CalculateHash(testMarshaller, testHasher, header1)
	assert.Nil(t, err)

	header2 := generateHeader(37, 22934, *partialMiniblockHeaders[1])
	headerHash2, err := core.CalculateHash(testMarshaller, testHasher, header2)
	assert.Nil(t, err)

	err = mbHandler.commitMiniblock(header1, headerHash1, completeMiniblock)
	assert.Nil(t, err)

	err = mbHandler.commitMiniblock(header1, headerHash1, partialMiniblock)
	assert.Nil(t, err)

	t.Run("completed miniblock hash with invalid header hash should not call update", func(t *testing.T) {
		err = mbHandler.updateMiniblockMetadataOnBlock(completeMiniblockHash, []byte("missing header hash"), func(mbMetadataOnBlock *MiniblockMetadataOnBlock) {
			assert.Fail(t, "should have not called update handler")
		})
		assert.Equal(t, storage.ErrKeyNotFound, err)
	})
	t.Run("completed miniblock hash with valid header hash should call update", func(t *testing.T) {
		completedExecutedMiniblock := &MiniblockMetadata{
			Round:                             23033,
			HeaderNonce:                       22933,
			HeaderHash:                        headerHash1,
			MiniblockHash:                     completeMiniblockHash,
			Epoch:                             37,
			NotarizedAtDestinationInMetaNonce: 111,
		}

		err = mbHandler.updateMiniblockMetadataOnBlock(completeMiniblockHash, headerHash1, func(mbMetadataOnBlock *MiniblockMetadataOnBlock) {
			mbMetadataOnBlock.NotarizedAtDestinationInMetaNonce = 111
		})
		assert.Nil(t, err)

		mbData, errGet := mbHandler.getMiniblockMetadataByMiniblockHash(completeMiniblockHash)
		assert.Nil(t, errGet)
		assert.Equal(t, []*MiniblockMetadata{completedExecutedMiniblock}, mbData)
	})

	partiallyExecutedMiniblock1 := &MiniblockMetadata{
		Round:         23033,
		HeaderNonce:   22933,
		HeaderHash:    headerHash1,
		MiniblockHash: partialMiniblockHash,
		Epoch:         37,
	}
	t.Run("partial miniblock with valid header1 hash should call update", func(t *testing.T) {
		err = mbHandler.updateMiniblockMetadataOnBlock(partialMiniblockHash, headerHash1, func(mbMetadataOnBlock *MiniblockMetadataOnBlock) {
			mbMetadataOnBlock.NotarizedAtDestinationInMetaNonce = 112
		})
		assert.Nil(t, err)
		partiallyExecutedMiniblock1.NotarizedAtDestinationInMetaNonce = 112

		mbData, errGet := mbHandler.getMiniblockMetadataByMiniblockHash(partialMiniblockHash)
		assert.Nil(t, errGet)
		assert.Equal(t, []*MiniblockMetadata{partiallyExecutedMiniblock1}, mbData)
	})

	err = mbHandler.commitMiniblock(header2, headerHash2, partialMiniblock)
	assert.Nil(t, err)

	partiallyExecutedMiniblock2 := &MiniblockMetadata{
		Round:         23034,
		HeaderNonce:   22934,
		HeaderHash:    headerHash2,
		MiniblockHash: partialMiniblockHash,
		Epoch:         37,
	}

	t.Run("partial miniblock with valid header2 hash should call update", func(t *testing.T) {
		err = mbHandler.updateMiniblockMetadataOnBlock(partialMiniblockHash, headerHash2, func(mbMetadataOnBlock *MiniblockMetadataOnBlock) {
			mbMetadataOnBlock.NotarizedAtDestinationInMetaNonce = 113
		})
		assert.Nil(t, err)
		partiallyExecutedMiniblock2.NotarizedAtDestinationInMetaNonce = 113

		mbData, errGet := mbHandler.getMiniblockMetadataByMiniblockHash(partialMiniblockHash)
		assert.Nil(t, errGet)
		assert.Equal(t, []*MiniblockMetadata{partiallyExecutedMiniblock1, partiallyExecutedMiniblock2}, mbData)
	})
	t.Run("partial miniblock hash with invalid header hash should not call update", func(t *testing.T) {
		err = mbHandler.updateMiniblockMetadataOnBlock(partialMiniblockHash, []byte("missing header hash"), func(mbMetadataOnBlock *MiniblockMetadataOnBlock) {
			assert.Fail(t, "should have not called update handler")
		})
		assert.Equal(t, storage.ErrKeyNotFound, err)
	})
}
