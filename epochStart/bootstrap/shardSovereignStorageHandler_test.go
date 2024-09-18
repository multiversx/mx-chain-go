package bootstrap

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getStoredBootstrapData(
	t *testing.T,
	marshaller marshal.Marshalizer,
	bootStorer storage.Storer,
	round uint64,
) *bootstrapStorage.BootstrapData {
	roundToUseAsKey := int64(round)
	key := []byte(strconv.FormatInt(roundToUseAsKey, 10))
	bootStrapDataBytes, err := bootStorer.Get(key)
	require.Nil(t, err)

	bootStrapData := &bootstrapStorage.BootstrapData{}
	err = marshaller.Unmarshal(bootStrapData, bootStrapDataBytes)
	require.Nil(t, err)

	return bootStrapData
}

func TestSovereignShardStorageHandler_SaveDataToStorage(t *testing.T) {
	t.Parallel()

	args := createStorageHandlerArgs()
	shardStorage, _ := NewShardStorageHandler(args)
	sovShardStorage := newSovereignShardStorageHandler(shardStorage)

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")
	hdr1 := &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce: 1,
			Round: 1,
			Epoch: 1,
		},
	}
	hdr2 := &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce: 2,
			Round: 2,
			Epoch: 2,
		},
	}
	headers := map[string]data.HeaderHandler{
		string(hash1): hdr1,
		string(hash2): hdr2,
	}

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: hdr1,
		PreviousEpochStart:  hdr2,
		ShardHeader:         hdr1,
		Headers:             headers,
		NodesConfig:         &nodesCoordinator.NodesCoordinatorRegistry{},
	}

	err := sovShardStorage.SaveDataToStorage(components, components.ShardHeader, false, nil)
	assert.Nil(t, err)

	bootStorer, err := sovShardStorage.storageService.GetStorer(dataRetriever.BootstrapUnit)
	require.Nil(t, err)

	hdrHash, err := core.CalculateHash(sovShardStorage.marshalizer, sovShardStorage.hasher, hdr1)
	require.Nil(t, err)

	bootStrapData := getStoredBootstrapData(t, sovShardStorage.marshalizer, bootStorer, hdr1.GetRound())
	require.Equal(t, &bootstrapStorage.BootstrapData{
		LastHeader: bootstrapStorage.BootstrapHeaderInfo{
			Nonce: 1,
			Epoch: 1,
			Hash:  hdrHash,
		},
		LastCrossNotarizedHeaders: []bootstrapStorage.BootstrapHeaderInfo{},
		LastSelfNotarizedHeaders: []bootstrapStorage.BootstrapHeaderInfo{{
			Nonce: 1,
			Epoch: 1,
			Hash:  hdrHash,
		}},
		ProcessedMiniBlocks:        []bootstrapStorage.MiniBlocksInMeta{},
		PendingMiniBlocks:          []bootstrapStorage.PendingMiniBlocksInfo{},
		NodesCoordinatorConfigKey:  nil,
		EpochStartTriggerConfigKey: []byte(fmt.Sprint(1)),
		HighestFinalBlockNonce:     hdr1.GetNonce(),
		LastRound:                  0,
	}, bootStrapData)

	hdr1Bytes, err := bootStorer.Get([]byte("epochStartBlock_1"))
	require.Nil(t, err)

	hdr2Bytes, err := bootStorer.Get([]byte("epochStartBlock_2"))
	require.Nil(t, err)

	hdr1Stored := &block.SovereignChainHeader{}
	err = sovShardStorage.marshalizer.Unmarshal(hdr1Stored, hdr1Bytes)
	require.Nil(t, err)
	require.Equal(t, hdr1, hdr1Stored)

	hdr2Stored := &block.SovereignChainHeader{}
	err = sovShardStorage.marshalizer.Unmarshal(hdr2Stored, hdr2Bytes)
	require.Nil(t, err)
	require.Equal(t, hdr2, hdr2Stored)

	// no meta block unit storage key saved
	metaStorer, err := sovShardStorage.storageService.GetStorer(dataRetriever.MetaBlockUnit)
	require.Nil(t, err)
	_, err = metaStorer.Get([]byte("epochStartBlock_1"))
	require.NotNil(t, err)
	_, err = metaStorer.Get([]byte("epochStartBlock_2"))
	require.NotNil(t, err)
}
