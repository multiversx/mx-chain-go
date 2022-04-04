package staking

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/require"
)

func createMetaBlockHeader(epoch uint32, round uint64, prevHash []byte) *block.MetaBlock {
	hdr := block.MetaBlock{
		Epoch:                  epoch,
		Nonce:                  round,
		Round:                  round,
		PrevHash:               prevHash,
		Signature:              []byte("signature"),
		PubKeysBitmap:          []byte("pubKeysBitmap"),
		RootHash:               []byte("roothash"),
		ShardInfo:              make([]block.ShardData, 0),
		TxCount:                1,
		PrevRandSeed:           []byte("roothash"),
		RandSeed:               []byte("roothash" + strconv.Itoa(int(round))),
		AccumulatedFeesInEpoch: big.NewInt(0),
		AccumulatedFees:        big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
	}

	shardMiniBlockHeaders := make([]block.MiniBlockHeader, 0)
	shardMiniBlockHeader := block.MiniBlockHeader{
		Hash:            []byte("mb_hash" + strconv.Itoa(int(round))),
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxCount:         1,
	}
	shardMiniBlockHeaders = append(shardMiniBlockHeaders, shardMiniBlockHeader)
	shardData := block.ShardData{
		Nonce:                 round,
		ShardID:               0,
		HeaderHash:            []byte("hdr_hash" + strconv.Itoa(int(round))),
		TxCount:               1,
		ShardMiniBlockHeaders: shardMiniBlockHeaders,
		DeveloperFees:         big.NewInt(0),
		AccumulatedFees:       big.NewInt(0),
	}
	hdr.ShardInfo = append(hdr.ShardInfo, shardData)

	return &hdr
}

func TestNewTestMetaProcessor(t *testing.T) {
	node := NewTestMetaProcessor(3, 3, 3, 2, 2)
	node.DisplayNodesConfig(0, 4)

	node.EpochStartTrigger.SetRoundsPerEpoch(4)

	newHdr := createMetaBlockHeader(0, 1, node.GenesisHeader.Hash)
	_, _ = node.MetaBlockProcessor.CreateNewHeader(1, 1)
	newHdr2, newBodyHandler2, err := node.MetaBlockProcessor.CreateBlock(newHdr, func() bool { return true })
	require.Nil(t, err)
	err = node.MetaBlockProcessor.CommitBlock(newHdr2, newBodyHandler2)
	require.Nil(t, err)

	node.DisplayNodesConfig(0, 4)

	marshaller := &mock.MarshalizerMock{}
	hasher := sha256.NewSha256()

	prevBlockBytes, _ := marshaller.Marshal(newHdr2)
	prevBlockBytes = hasher.Compute(string(prevBlockBytes))
	prevRandomness := node.BlockChain.GetCurrentBlockHeader().GetRandSeed()
	newHdr = createMetaBlockHeader(0, 2, prevBlockBytes)
	newHdr.PrevRandSeed = prevRandomness

	_, _ = node.MetaBlockProcessor.CreateNewHeader(2, 2)
	newHdr2, newBodyHandler2, err = node.MetaBlockProcessor.CreateBlock(newHdr, func() bool { return true })
	require.Nil(t, err)
	err = node.MetaBlockProcessor.CommitBlock(newHdr2, newBodyHandler2)
	require.Nil(t, err)
	node.DisplayNodesConfig(0, 4)

	prevBlockBytes, _ = marshaller.Marshal(newHdr2)
	prevBlockBytes = hasher.Compute(string(prevBlockBytes))
	prevRandomness = node.BlockChain.GetCurrentBlockHeader().GetRandSeed()
	newHdr = createMetaBlockHeader(0, 3, prevBlockBytes)
	newHdr.PrevRandSeed = prevRandomness

	_, _ = node.MetaBlockProcessor.CreateNewHeader(3, 3)
	newHdr2, newBodyHandler2, err = node.MetaBlockProcessor.CreateBlock(newHdr, func() bool { return true })
	require.Nil(t, err)
	err = node.MetaBlockProcessor.CommitBlock(newHdr2, newBodyHandler2)
	require.Nil(t, err)
	node.DisplayNodesConfig(0, 4)

	prevBlockBytes, _ = marshaller.Marshal(newHdr2)
	prevBlockBytes = hasher.Compute(string(prevBlockBytes))
	prevRandomness = node.BlockChain.GetCurrentBlockHeader().GetRandSeed()
	newHdr = createMetaBlockHeader(1, 4, prevBlockBytes)
	newHdr.PrevRandSeed = prevRandomness

	_, _ = node.MetaBlockProcessor.CreateNewHeader(4, 4)
	newHdr2, newBodyHandler2, err = node.MetaBlockProcessor.CreateBlock(newHdr, func() bool { return true })
	require.Nil(t, err)
	err = node.MetaBlockProcessor.CommitBlock(newHdr2, newBodyHandler2)
	require.Nil(t, err)
	node.DisplayNodesConfig(0, 4)

	prevBlockBytes, _ = marshaller.Marshal(newHdr2)
	prevBlockBytes = hasher.Compute(string(prevBlockBytes))
	prevRandomness = node.BlockChain.GetCurrentBlockHeader().GetRandSeed()
	newHdr = createMetaBlockHeader(1, 5, prevBlockBytes)
	newHdr.PrevRandSeed = prevRandomness
	newHdr.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{}}
	newHdr.EpochStart.Economics = block.Economics{RewardsForProtocolSustainability: big.NewInt(0)}

	_, _ = node.MetaBlockProcessor.CreateNewHeader(5, 5)
	newHdr2, newBodyHandler2, err = node.MetaBlockProcessor.CreateBlock(newHdr, func() bool { return true })
	//node.CoreComponents.EpochStartNotifierWithConfirm().NotifyAllPrepare(newHdr2,newBodyHandler2)
	require.Nil(t, err)
	err = node.MetaBlockProcessor.CommitBlock(newHdr2, newBodyHandler2)
	require.Nil(t, err)
	node.DisplayNodesConfig(1, 4)

	// epoch start
	prevBlockBytes, _ = marshaller.Marshal(newHdr2)
	prevBlockBytes = hasher.Compute(string(prevBlockBytes))
	prevRandomness = node.BlockChain.GetCurrentBlockHeader().GetRandSeed()
	newHdr = createMetaBlockHeader(1, 6, prevBlockBytes)
	newHdr.PrevRandSeed = prevRandomness

	_, _ = node.MetaBlockProcessor.CreateNewHeader(6, 6)
	newHdr2, newBodyHandler2, err = node.MetaBlockProcessor.CreateBlock(newHdr, func() bool { return true })
	require.Nil(t, err)
	err = node.MetaBlockProcessor.CommitBlock(newHdr2, newBodyHandler2)
	require.Nil(t, err)
	node.DisplayNodesConfig(1, 4)

}
