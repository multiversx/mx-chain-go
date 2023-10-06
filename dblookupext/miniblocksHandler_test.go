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

func checkTxs(miniblockHashByTxHashIndexStorer storage.Storer, epoch uint32, txHashes [][]byte, miniblockHash []byte) error {
	for _, txHash := range txHashes {
		value, err := miniblockHashByTxHashIndexStorer.GetFromEpoch(txHash, epoch)
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

func TestMiniblocksHandler_blockCommitted(t *testing.T) {
	t.Parallel()

	completeMiniblock := &block.MiniBlock{
		TxHashes: [][]byte{
			[]byte("txHashCompleted1"),
			[]byte("txHashCompleted2"),
			[]byte("txHashCompleted3"),
		},
	}
	completeMiniblockHash, _ := core.CalculateHash(testMarshaller, testHasher, completeMiniblock)
	completeMiniblockHeader := &block.MiniBlockHeader{
		Hash:     completeMiniblockHash,
		TxCount:  uint32(len(completeMiniblock.TxHashes)),
		Type:     0,
		Reserved: nil,
	}
	_ = completeMiniblockHeader.SetConstructionState(int32(block.Final))
	_ = completeMiniblockHeader.SetIndexOfFirstTxProcessed(0)
	_ = completeMiniblockHeader.SetIndexOfLastTxProcessed(int32(len(completeMiniblock.TxHashes) - 1))

	partialMiniblock := &block.MiniBlock{
		TxHashes: [][]byte{
			[]byte("txHashPartial1"),
			[]byte("txHashPartial2"),
			[]byte("txHashPartial3"),
			[]byte("txHashPartial4"),
			[]byte("txHashPartial5"),
			[]byte("txHashPartial6"),
		},
	}
	partialMiniblockHash, _ := core.CalculateHash(testMarshaller, testHasher, partialMiniblock)
	partialMiniblockHeader1 := &block.MiniBlockHeader{
		Hash:     partialMiniblockHash,
		TxCount:  uint32(len(completeMiniblock.TxHashes)),
		Type:     0,
		Reserved: nil,
	}
	_ = partialMiniblockHeader1.SetConstructionState(int32(block.PartialExecuted))
	_ = partialMiniblockHeader1.SetIndexOfFirstTxProcessed(0)
	_ = partialMiniblockHeader1.SetIndexOfLastTxProcessed(1)

	partialMiniblockHeader2 := &block.MiniBlockHeader{
		Hash:     partialMiniblockHash,
		TxCount:  uint32(len(completeMiniblock.TxHashes)),
		Type:     0,
		Reserved: nil,
	}
	_ = partialMiniblockHeader2.SetConstructionState(int32(block.PartialExecuted))
	_ = partialMiniblockHeader2.SetIndexOfFirstTxProcessed(2)
	_ = partialMiniblockHeader2.SetIndexOfLastTxProcessed(3)

	partialMiniblockHeader3 := &block.MiniBlockHeader{
		Hash:     partialMiniblockHash,
		TxCount:  uint32(len(completeMiniblock.TxHashes)),
		Type:     0,
		Reserved: nil,
	}
	_ = partialMiniblockHeader3.SetConstructionState(int32(block.PartialExecuted))
	_ = partialMiniblockHeader3.SetIndexOfFirstTxProcessed(4)
	_ = partialMiniblockHeader3.SetIndexOfLastTxProcessed(5)

	t.Run("should work with completed miniblock", func(t *testing.T) {
		t.Parallel()

		mbHandler := createMockMiniblocksHandler()

		header := &block.Header{
			Epoch: 37,
			Nonce: 22933,
			Round: 28927,
			MiniBlockHeaders: []block.MiniBlockHeader{
				*completeMiniblockHeader,
			},
		}
		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{completeMiniblock},
		}
		headerHash, err := core.CalculateHash(testMarshaller, testHasher, header)
		assert.Nil(t, err)

		// commit header with completed miniblock and check everything war written
		err = mbHandler.blockCommitted(header, body)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header.GetEpoch(), completeMiniblock.TxHashes, completeMiniblockHash)
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

		header1 := &block.Header{
			Epoch: 37,
			Nonce: 22933,
			Round: 28927,
			MiniBlockHeaders: []block.MiniBlockHeader{
				*completeMiniblockHeader,
				*partialMiniblockHeader1,
			},
		}
		body1 := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				completeMiniblock,
				partialMiniblock,
			},
		}
		headerHash1, err := core.CalculateHash(testMarshaller, testHasher, header1)
		assert.Nil(t, err)

		header2 := &block.Header{
			Epoch: 37,
			Nonce: 22934,
			Round: 28928,
			MiniBlockHeaders: []block.MiniBlockHeader{
				*partialMiniblockHeader2,
			},
		}
		body2 := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				partialMiniblock,
			},
		}
		headerHash2, err := core.CalculateHash(testMarshaller, testHasher, header2)
		assert.Nil(t, err)

		// commit header1 with completed miniblock and check all transactions from that block were written
		err = mbHandler.blockCommitted(header1, body1)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header1.GetEpoch(), partialMiniblock.TxHashes[0:2], partialMiniblockHash)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header1.GetEpoch(), completeMiniblock.TxHashes, completeMiniblockHash)
		assert.Nil(t, err)

		// commit header2 with completed miniblock and check all transactions from header 2 and header 1 were written
		err = mbHandler.blockCommitted(header2, body2)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header1.GetEpoch(), completeMiniblock.TxHashes, completeMiniblockHash)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header1.GetEpoch(), partialMiniblock.TxHashes[0:4], partialMiniblockHash)
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

		header1 := &block.Header{
			Epoch: 37,
			Nonce: 22933,
			Round: 28927,
			MiniBlockHeaders: []block.MiniBlockHeader{
				*completeMiniblockHeader,
				*partialMiniblockHeader1,
			},
		}
		body1 := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				completeMiniblock,
				partialMiniblock,
			},
		}
		headerHash1, err := core.CalculateHash(testMarshaller, testHasher, header1)
		assert.Nil(t, err)

		header2 := &block.Header{
			Epoch: 37,
			Nonce: 22934,
			Round: 28928,
			MiniBlockHeaders: []block.MiniBlockHeader{
				*partialMiniblockHeader2,
			},
		}
		body2 := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				partialMiniblock,
			},
		}
		headerHash2, err := core.CalculateHash(testMarshaller, testHasher, header2)
		assert.Nil(t, err)

		header3 := &block.Header{
			Epoch: 38,
			Nonce: 22935,
			Round: 28929,
			MiniBlockHeaders: []block.MiniBlockHeader{
				*partialMiniblockHeader3,
			},
		}
		body3 := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				partialMiniblock,
			},
		}
		headerHash3, err := core.CalculateHash(testMarshaller, testHasher, header3)
		assert.Nil(t, err)

		// commit header1 with completed miniblock and check all transactions from that block were written
		err = mbHandler.blockCommitted(header1, body1)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header1.GetEpoch(), partialMiniblock.TxHashes[0:2], partialMiniblockHash)
		assert.Nil(t, err)

		// commit header2 with completed miniblock and check all transactions from header 2 and header 1 were written
		err = mbHandler.blockCommitted(header2, body2)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header1.GetEpoch(), completeMiniblock.TxHashes, completeMiniblockHash)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header1.GetEpoch(), partialMiniblock.TxHashes[0:4], partialMiniblockHash)
		assert.Nil(t, err)

		// commit header3 with completed miniblock and check all transactions from header 3, header 2 and header 1 were written in
		// the correct epochs
		err = mbHandler.blockCommitted(header3, body3)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header1.GetEpoch(), completeMiniblock.TxHashes, completeMiniblockHash)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header1.GetEpoch(), partialMiniblock.TxHashes[0:4], partialMiniblockHash)
		assert.Nil(t, err)
		err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header3.GetEpoch(), partialMiniblock.TxHashes[5:6], partialMiniblockHash)
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

		// save miniblockMetadata1 in legacy format
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

	completeMiniblock := &block.MiniBlock{
		TxHashes: [][]byte{
			[]byte("txHashCompleted1"),
			[]byte("txHashCompleted2"),
			[]byte("txHashCompleted3"),
		},
	}
	completeMiniblockHash, _ := core.CalculateHash(testMarshaller, testHasher, completeMiniblock)
	completeMiniblockHeader := &block.MiniBlockHeader{
		Hash:     completeMiniblockHash,
		TxCount:  uint32(len(completeMiniblock.TxHashes)),
		Type:     0,
		Reserved: nil,
	}
	_ = completeMiniblockHeader.SetConstructionState(int32(block.Final))
	_ = completeMiniblockHeader.SetIndexOfFirstTxProcessed(0)
	_ = completeMiniblockHeader.SetIndexOfLastTxProcessed(int32(len(completeMiniblock.TxHashes) - 1))

	partialMiniblock := &block.MiniBlock{
		TxHashes: [][]byte{
			[]byte("txHashPartial1"),
			[]byte("txHashPartial2"),
			[]byte("txHashPartial3"),
			[]byte("txHashPartial4"),
			[]byte("txHashPartial5"),
			[]byte("txHashPartial6"),
		},
	}
	partialMiniblockHash, _ := core.CalculateHash(testMarshaller, testHasher, partialMiniblock)
	partialMiniblockHeader1 := &block.MiniBlockHeader{
		Hash:     partialMiniblockHash,
		TxCount:  uint32(len(completeMiniblock.TxHashes)),
		Type:     0,
		Reserved: nil,
	}
	_ = partialMiniblockHeader1.SetConstructionState(int32(block.PartialExecuted))
	_ = partialMiniblockHeader1.SetIndexOfFirstTxProcessed(0)
	_ = partialMiniblockHeader1.SetIndexOfLastTxProcessed(1)

	partialMiniblockHeader2 := &block.MiniBlockHeader{
		Hash:     partialMiniblockHash,
		TxCount:  uint32(len(completeMiniblock.TxHashes)),
		Type:     0,
		Reserved: nil,
	}
	_ = partialMiniblockHeader2.SetConstructionState(int32(block.PartialExecuted))
	_ = partialMiniblockHeader2.SetIndexOfFirstTxProcessed(2)
	_ = partialMiniblockHeader2.SetIndexOfLastTxProcessed(3)

	partialMiniblockHeader3 := &block.MiniBlockHeader{
		Hash:     partialMiniblockHash,
		TxCount:  uint32(len(completeMiniblock.TxHashes)),
		Type:     0,
		Reserved: nil,
	}
	_ = partialMiniblockHeader3.SetConstructionState(int32(block.PartialExecuted))
	_ = partialMiniblockHeader3.SetIndexOfFirstTxProcessed(4)
	_ = partialMiniblockHeader3.SetIndexOfLastTxProcessed(5)

	mbHandler := createMockMiniblocksHandler()

	header1 := &block.Header{
		Epoch: 37,
		Nonce: 22933,
		Round: 28927,
		MiniBlockHeaders: []block.MiniBlockHeader{
			*completeMiniblockHeader,
			*partialMiniblockHeader1,
		},
	}
	body1 := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			completeMiniblock,
			partialMiniblock,
		},
	}
	headerHash1, err := core.CalculateHash(testMarshaller, testHasher, header1)
	assert.Nil(t, err)

	header2 := &block.Header{
		Epoch: 37,
		Nonce: 22934,
		Round: 28928,
		MiniBlockHeaders: []block.MiniBlockHeader{
			*partialMiniblockHeader2,
		},
	}
	body2 := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			partialMiniblock,
		},
	}
	headerHash2, err := core.CalculateHash(testMarshaller, testHasher, header2)
	assert.Nil(t, err)

	header3 := &block.Header{
		Epoch: 38,
		Nonce: 22935,
		Round: 28929,
		MiniBlockHeaders: []block.MiniBlockHeader{
			*partialMiniblockHeader3,
		},
	}
	body3 := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			partialMiniblock,
		},
	}

	err = mbHandler.blockCommitted(header1, body1)
	assert.Nil(t, err)
	err = mbHandler.blockCommitted(header2, body2)
	assert.Nil(t, err)
	err = mbHandler.blockCommitted(header3, body3)
	assert.Nil(t, err)

	//we set the current epochs in storers, this will not be necessary when implementing RemoveFromEpoch
	mbHandler.miniblocksMetadataStorer.(*genericMocks.StorerMock).SetCurrentEpoch(header3.Epoch)
	// miniblockHashByTxHashIndexStorer is a static storer, set epoch won't have an efect
	// epochIndex contains a static storer, set epoch won't have an efect

	// ### rollback on header 3
	err = mbHandler.blockReverted(header3, body3)
	assert.Nil(t, err)

	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header1.GetEpoch(), completeMiniblock.TxHashes, completeMiniblockHash)
	assert.Nil(t, err)
	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header1.GetEpoch(), partialMiniblock.TxHashes[0:4], partialMiniblockHash)
	assert.Nil(t, err)
	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header3.GetEpoch(), partialMiniblock.TxHashes[5:6], partialMiniblockHash)
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

	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header1.GetEpoch(), completeMiniblock.TxHashes, completeMiniblockHash)
	assert.Nil(t, err)
	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header1.GetEpoch(), partialMiniblock.TxHashes[0:4], partialMiniblockHash)
	assert.NotNil(t, err) // these were reverted
	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header3.GetEpoch(), partialMiniblock.TxHashes[5:6], partialMiniblockHash)
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

	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header1.GetEpoch(), completeMiniblock.TxHashes, completeMiniblockHash)
	assert.NotNil(t, err) // these were reverted
	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header1.GetEpoch(), partialMiniblock.TxHashes[0:4], partialMiniblockHash)
	assert.NotNil(t, err) // these were reverted
	err = checkTxs(mbHandler.miniblockHashByTxHashIndexStorer, header3.GetEpoch(), partialMiniblock.TxHashes[5:6], partialMiniblockHash)
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
