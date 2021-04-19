package resolvers

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
)

// Log -
var Log = logger.GetOrCreate("integrationtests/multishard/resolvers")

// CreateResolverRequester -
func CreateResolverRequester(
	resolverShardID uint32,
	requesterShardID uint32,
) (*integrationTests.TestProcessorNode, *integrationTests.TestProcessorNode) {

	numShards := uint32(2)

	txSignShardId := uint32(0)
	nResolver := integrationTests.NewTestProcessorNode(numShards, resolverShardID, txSignShardId)
	nRequester := integrationTests.NewTestProcessorNode(numShards, requesterShardID, txSignShardId)

	time.Sleep(time.Second)
	err := nRequester.Messenger.ConnectToPeer(integrationTests.GetConnectableAddress(nResolver.Messenger))
	Log.LogIfError(err)

	time.Sleep(time.Second)

	return nResolver, nRequester
}

// CreateShardHeader -
func CreateShardHeader(nonce uint64, chainID []byte) (data.HeaderHandler, []byte) {
	hdr := &block.Header{
		Nonce:        nonce,
		PrevHash:     []byte("prev hash"),
		PrevRandSeed: []byte("prev rand seed"),
		RandSeed:     []byte("rand seed"),
		PubKeysBitmap: []byte{
			255,
			0,
		},
		TimeStamp:       uint64(time.Now().Unix()),
		Round:           1,
		Signature:       []byte("signature"),
		LeaderSignature: nil,
		RootHash: []byte{
			255,
			255,
		},
		MetaBlockHashes:    nil,
		EpochStartMetaHash: nil,
		ReceiptsHash:       nil,
		ChainID:            chainID,
		SoftwareVersion:    []byte("version"),
		MiniBlockHeaders:   make([]block.MiniBlockHeader, 0),
		PeerChanges:        nil,
		Epoch:              2,
		TxCount:            0,
		ShardID:            0,
		BlockBodyType:      block.TxBlock,
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, hdr)
	Log.LogIfError(err)

	return hdr, hash
}

// CreateMetaHeader -
func CreateMetaHeader(nonce uint64, chainID []byte) (data.HeaderHandler, []byte) {
	hdr := &block.MetaBlock{
		Nonce:           nonce,
		Epoch:           0,
		ShardInfo:       make([]block.ShardData, 0),
		Signature:       []byte("signature"),
		PubKeysBitmap:   []byte{1},
		PrevHash:        []byte("prev hash"),
		PrevRandSeed:    []byte("prev rand seed"),
		RandSeed:        []byte("rand seed"),
		RootHash:        []byte("root hash"),
		TxCount:         0,
		ChainID:         chainID,
		SoftwareVersion: []byte("v1.0"),
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, hdr)
	Log.LogIfError(err)

	return hdr, hash
}

// CreateMiniblock -
func CreateMiniblock(senderShardId uint32, receiverSharId uint32) (*block.MiniBlock, []byte) {
	dummyTxHash := make([]byte, integrationTests.TestHasher.Size())
	miniblock := &block.MiniBlock{
		TxHashes:        [][]byte{dummyTxHash},
		ReceiverShardID: receiverSharId,
		SenderShardID:   senderShardId,
		Type:            0,
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, miniblock)
	Log.LogIfError(err)

	return miniblock, hash
}

// CreateReward -
func CreateReward(round uint64) (data.TransactionHandler, []byte) {
	reward := &rewardTx.RewardTx{
		Round:   round,
		Epoch:   0,
		Value:   big.NewInt(1),
		RcvAddr: make([]byte, integrationTests.TestHasher.Size()),
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, reward)
	Log.LogIfError(err)

	return reward, hash
}
