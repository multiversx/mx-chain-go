package resolvers

import (
	"bytes"
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-logger-go"
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
	nResolver := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            numShards,
		NodeShardId:          resolverShardID,
		TxSignPrivKeyShardId: txSignShardId,
	})
	nRequester := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            numShards,
		NodeShardId:          requesterShardID,
		TxSignPrivKeyShardId: txSignShardId,
	})

	time.Sleep(time.Second)
	err := nRequester.MainMessenger.ConnectToPeer(integrationTests.GetConnectableAddress(nResolver.MainMessenger))
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
		AccumulatedFees:    big.NewInt(0),
		DeveloperFees:      big.NewInt(0),
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, hdr)
	Log.LogIfError(err)

	return hdr, hash
}

// CreateMetaHeader -
func CreateMetaHeader(nonce uint64, chainID []byte) (data.HeaderHandler, []byte) {
	hdr := &block.MetaBlock{
		Nonce:                  nonce,
		Epoch:                  0,
		ShardInfo:              make([]block.ShardData, 0),
		Signature:              []byte("signature"),
		PubKeysBitmap:          []byte{1},
		PrevHash:               []byte("prev hash"),
		PrevRandSeed:           []byte("prev rand seed"),
		RandSeed:               []byte("rand seed"),
		RootHash:               []byte("root hash"),
		TxCount:                0,
		ChainID:                chainID,
		SoftwareVersion:        []byte("v1.0"),
		DeveloperFees:          big.NewInt(0),
		AccumulatedFees:        big.NewInt(0),
		ValidatorStatsRootHash: []byte("validator stats root hash"),
		DevFeesInEpoch:         big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(0),
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

// CreateLargeSmartContractResults -
func CreateLargeSmartContractResults() (data.TransactionHandler, []byte) {
	scr := &smartContractResult.SmartContractResult{
		Nonce:      1,
		Value:      big.NewInt(3),
		RcvAddr:    make([]byte, 32),
		SndAddr:    make([]byte, 32),
		Code:       nil,
		Data:       bytes.Repeat([]byte{'A'}, 1100000),
		PrevTxHash: make([]byte, 32),
		GasLimit:   0,
		GasPrice:   0,
		CallType:   0,
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, scr)
	Log.LogIfError(err)

	return scr, hash
}
