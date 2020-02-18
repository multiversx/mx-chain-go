package resolvers

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/logger"
)

var log = logger.GetOrCreate("integrationtests/multishard/resolvers")

func createResolverRequester(
	resolverShardID uint32,
	requesterShardID uint32,
) (*integrationTests.TestProcessorNode, *integrationTests.TestProcessorNode) {

	numShards := uint32(2)

	advertiserAddress := ""
	txSignShardId := uint32(0)
	nResolver := integrationTests.NewTestProcessorNode(numShards, resolverShardID, txSignShardId, advertiserAddress)
	nRequester := integrationTests.NewTestProcessorNode(numShards, requesterShardID, txSignShardId, advertiserAddress)

	_ = nRequester.Node.Start()
	_ = nResolver.Node.Start()

	time.Sleep(time.Second)
	err := nRequester.Messenger.ConnectToPeer(integrationTests.GetConnectableAddress(nResolver.Messenger))
	log.LogIfError(err)

	time.Sleep(time.Second)

	return nResolver, nRequester
}

func createShardHeader(nonce uint64, chainID []byte) (data.HeaderHandler, []byte) {
	hdr := &block.Header{
		Nonce:            nonce,
		PubKeysBitmap:    []byte{255, 0},
		Signature:        []byte("signature"),
		PrevHash:         []byte("prev hash"),
		TimeStamp:        uint64(time.Now().Unix()),
		Round:            1,
		Epoch:            2,
		ShardID:          0,
		BlockBodyType:    block.TxBlock,
		RootHash:         []byte{255, 255},
		PrevRandSeed:     make([]byte, 0),
		RandSeed:         make([]byte, 0),
		MiniBlockHeaders: make([]block.MiniBlockHeader, 0),
		ChainID:          chainID,
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, hdr)
	log.LogIfError(err)

	return hdr, hash
}

func createMetaHeader(nonce uint64, chainID []byte) (data.HeaderHandler, []byte) {
	hdr := &block.MetaBlock{
		Nonce:         nonce,
		Epoch:         0,
		ShardInfo:     make([]block.ShardData, 0),
		PeerInfo:      make([]block.PeerData, 0),
		Signature:     []byte("signature"),
		PubKeysBitmap: []byte{1},
		PrevHash:      []byte("prev hash"),
		PrevRandSeed:  []byte("prev rand seed"),
		RandSeed:      []byte("rand seed"),
		RootHash:      []byte("root hash"),
		TxCount:       0,
		ChainID:       chainID,
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, hdr)
	log.LogIfError(err)

	return hdr, hash
}

func createMiniblock(senderShardId uint32, receiverSharId uint32) (*block.MiniBlock, []byte) {
	dummyTxHash := make([]byte, integrationTests.TestHasher.Size())
	miniblock := &block.MiniBlock{
		TxHashes:        [][]byte{dummyTxHash},
		ReceiverShardID: receiverSharId,
		SenderShardID:   senderShardId,
		Type:            0,
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, miniblock)
	log.LogIfError(err)

	return miniblock, hash
}

func createReward(round uint64, shardId uint32) (data.TransactionHandler, []byte) {
	reward := &rewardTx.RewardTx{
		Round:   round,
		Epoch:   0,
		Value:   big.NewInt(1),
		RcvAddr: make([]byte, integrationTests.TestHasher.Size()),
		ShardID: shardId,
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, reward)
	log.LogIfError(err)

	return reward, hash
}
