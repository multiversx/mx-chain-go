package external_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/external"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var pk1 = []byte("pk1")
var defaultShardHeader = block.Header{
	Nonce:     1,
	ShardId:   2,
	TxCount:   3,
	TimeStamp: 4,
}
var testMarshalizer = &mock.MarshalizerFake{}

func createMockProposerResolver() external.ProposerResolver {
	return &mock.ProposerResolverStub{
		ResolveProposerCalled: func(shardId uint32, roundIndex uint32, prevRandomSeed []byte) (bytes []byte, e error) {
			return []byte(pk1), nil
		},
	}
}

func createMockStorer() dataRetriever.StorageService {
	return &mock.ChainStorerMock{
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) (bytes []byte, e error) {
			if unitType == dataRetriever.BlockHeaderUnit {
				hdrBuff, _ := testMarshalizer.Marshal(&defaultShardHeader)
				return hdrBuff, nil
			}

			//metablocks
			//key is something like "hash_0", "hash_1"
			//so we generate a metablock that has prevHash = "hash_"+[nonce - 1]
			nonce, _ := strconv.Atoi(strings.Split(string(key), "_")[1])

			metablock := &block.MetaBlock{
				Nonce:    uint64(nonce),
				PrevHash: []byte(fmt.Sprintf("hash_%d", nonce-1)),
				//one shard info (1 metablock contains one shard header hash)
				ShardInfo: []block.ShardData{{}},
			}

			metaHdrBuff, _ := testMarshalizer.Marshal(metablock)
			return metaHdrBuff, nil
		},
	}
}

func createMockStorerWithMetablockContaining3ShardBlocks() dataRetriever.StorageService {
	return &mock.ChainStorerMock{
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) (bytes []byte, e error) {
			if unitType == dataRetriever.BlockHeaderUnit {
				hdrBuff, _ := testMarshalizer.Marshal(&defaultShardHeader)
				return hdrBuff, nil
			}

			//metablocks
			//key is something like "hash_0", "hash_1"
			//so we generate a metablock that has prevHash = "hash_"+[nonce - 1]
			nonce, _ := strconv.Atoi(strings.Split(string(key), "_")[1])

			metablock := &block.MetaBlock{
				Nonce:    uint64(nonce),
				PrevHash: []byte(fmt.Sprintf("hash_%d", nonce-1)),
				//one shard info (1 metablock contains one shard header hash)
				ShardInfo: []block.ShardData{{}, {}, {}},
			}

			metaHdrBuff, _ := testMarshalizer.Marshal(metablock)
			return metaHdrBuff, nil
		},
	}
}

//------- NewExternalResolver

func TestNewExternalResolver_NilCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	ner, err := external.NewExternalResolver(
		nil,
		&mock.BlockChainMock{},
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.ProposerResolverStub{},
	)

	assert.Nil(t, ner)
	assert.Equal(t, external.ErrNilShardCoordinator, err)
}

func TestNewExternalResolver_NilChainHandlerShouldErr(t *testing.T) {
	t.Parallel()

	ner, err := external.NewExternalResolver(
		&mock.ShardCoordinatorMock{},
		nil,
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.ProposerResolverStub{},
	)

	assert.Nil(t, ner)
	assert.Equal(t, external.ErrNilBlockChain, err)
}

func TestNewExternalResolver_NilChainStorerShouldErr(t *testing.T) {
	t.Parallel()

	ner, err := external.NewExternalResolver(
		&mock.ShardCoordinatorMock{},
		&mock.BlockChainMock{},
		nil,
		&mock.MarshalizerMock{},
		&mock.ProposerResolverStub{},
	)

	assert.Nil(t, ner)
	assert.Equal(t, external.ErrNilStore, err)
}

func TestNewExternalResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	ner, err := external.NewExternalResolver(
		&mock.ShardCoordinatorMock{},
		&mock.BlockChainMock{},
		&mock.ChainStorerMock{},
		nil,
		&mock.ProposerResolverStub{},
	)

	assert.Nil(t, ner)
	assert.Equal(t, external.ErrNilMarshalizer, err)
}

func TestNewExternalResolver_NilValidatorGroupSelectorShouldErr(t *testing.T) {
	t.Parallel()

	ner, err := external.NewExternalResolver(
		&mock.ShardCoordinatorMock{},
		&mock.BlockChainMock{},
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		nil,
	)

	assert.Nil(t, ner)
	assert.Equal(t, external.ErrNilProposerResolver, err)
}

func TestNewExternalResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	ner, err := external.NewExternalResolver(
		&mock.ShardCoordinatorMock{},
		&mock.BlockChainMock{},
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.ProposerResolverStub{},
	)

	assert.NotNil(t, ner)
	assert.Nil(t, err)
}

//------ RecentNotarizedBlocks

func TestExternalResolver_RecentNotarizedBlocksNotMetachainShouldErr(t *testing.T) {
	t.Parallel()

	ner, _ := external.NewExternalResolver(
		&mock.ShardCoordinatorMock{},
		&mock.BlockChainMock{},
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.ProposerResolverStub{},
	)

	recentBlocks, err := ner.RecentNotarizedBlocks(1)
	assert.Nil(t, recentBlocks)
	assert.Equal(t, external.ErrOperationNotSupported, err)
}

func TestExternalResolver_InvalidMaxNumShouldErr(t *testing.T) {
	t.Parallel()

	ner, _ := external.NewExternalResolver(
		&mock.ShardCoordinatorMock{
			SelfShardId: sharding.MetachainShardId,
		},
		&mock.BlockChainMock{},
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.ProposerResolverStub{},
	)

	recentBlocks, err := ner.RecentNotarizedBlocks(0)
	assert.Nil(t, recentBlocks)
	assert.Equal(t, external.ErrInvalidValue, err)
}

func TestExternalResolver_CurrentBlockIsGenesisShouldWork(t *testing.T) {
	t.Parallel()

	ner, _ := external.NewExternalResolver(
		&mock.ShardCoordinatorMock{
			SelfShardId: sharding.MetachainShardId,
		},
		&mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlock{
					Nonce: 0,
				}
			},
		},
		&mock.ChainStorerMock{},
		&mock.MarshalizerMock{},
		&mock.ProposerResolverStub{},
	)

	recentBlocks, err := ner.RecentNotarizedBlocks(1)
	assert.Empty(t, recentBlocks)
	assert.Nil(t, err)
}

func TestExternalResolver_ProposerResolverErrorsShouldErr(t *testing.T) {
	t.Parallel()

	crtNonce := 5
	errExpected := errors.New("expected error")

	ner, _ := external.NewExternalResolver(
		&mock.ShardCoordinatorMock{
			SelfShardId: sharding.MetachainShardId,
		},
		&mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlock{
					Nonce:     uint64(crtNonce),
					PrevHash:  []byte(fmt.Sprintf("hash_%d", crtNonce-1)),
					ShardInfo: []block.ShardData{{}},
				}
			},
		},
		createMockStorer(),
		testMarshalizer,
		&mock.ProposerResolverStub{
			ResolveProposerCalled: func(shardId uint32, roundIndex uint32, prevRandomSeed []byte) (bytes []byte, e error) {
				return nil, errExpected
			},
		},
	)

	recentBlocks, err := ner.RecentNotarizedBlocks(crtNonce)
	assert.Nil(t, recentBlocks)
	assert.Equal(t, errExpected, err)
}

func TestExternalResolver_WithBlocksShouldWork(t *testing.T) {
	t.Parallel()

	crtNonce := 6

	ner, _ := external.NewExternalResolver(
		&mock.ShardCoordinatorMock{
			SelfShardId: sharding.MetachainShardId,
		},
		&mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlock{
					Nonce:     uint64(crtNonce),
					PrevHash:  []byte(fmt.Sprintf("hash_%d", crtNonce-1)),
					ShardInfo: []block.ShardData{{}},
				}
			},
		},
		createMockStorer(),
		testMarshalizer,
		createMockProposerResolver(),
	)

	//need exactly 5 shard blocks, will receive exactly 5 blocks (genesis block with nonce 0 is not considered)
	recentBlocks, err := ner.RecentNotarizedBlocks(crtNonce - 1)
	assert.Equal(t, crtNonce-1, len(recentBlocks))
	assert.Nil(t, err)
}

func TestExternalResolver_WithFewBlocksShouldWork(t *testing.T) {
	t.Parallel()

	crtNonce := 6

	ner, _ := external.NewExternalResolver(
		&mock.ShardCoordinatorMock{
			SelfShardId: sharding.MetachainShardId,
		},
		&mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlock{
					Nonce:     uint64(crtNonce),
					PrevHash:  []byte(fmt.Sprintf("hash_%d", crtNonce-1)),
					ShardInfo: []block.ShardData{{}},
				}
			},
		},
		createMockStorer(),
		testMarshalizer,
		createMockProposerResolver(),
	)

	recentBlocks, err := ner.RecentNotarizedBlocks(crtNonce * 2)
	assert.Equal(t, crtNonce, len(recentBlocks))
	assert.Nil(t, err)
}

func TestExternalResolver_WithMoreBlocksShouldWorkAndReturnMaxBlocksNum(t *testing.T) {
	t.Parallel()

	crtNonce := 50

	ner, _ := external.NewExternalResolver(
		&mock.ShardCoordinatorMock{
			SelfShardId: sharding.MetachainShardId,
		},
		&mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlock{
					Nonce:     uint64(crtNonce),
					PrevHash:  []byte(fmt.Sprintf("hash_%d", crtNonce-1)),
					ShardInfo: []block.ShardData{{}},
				}
			},
		},
		createMockStorer(),
		testMarshalizer,
		createMockProposerResolver(),
	)

	maxBlocks := 10

	recentBlocks, err := ner.RecentNotarizedBlocks(maxBlocks)
	assert.Equal(t, maxBlocks, len(recentBlocks))
	assert.Nil(t, err)
}

func TestExternalResolver_WithMoreBlocksShouldWorkAndReturnMaxBlocksNumButMetablockContainsMoreShardBlocks(t *testing.T) {
	t.Parallel()

	crtNonce := 4

	ner, _ := external.NewExternalResolver(
		&mock.ShardCoordinatorMock{
			SelfShardId: sharding.MetachainShardId,
		},
		&mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlock{
					Nonce:     uint64(crtNonce),
					PrevHash:  []byte(fmt.Sprintf("hash_%d", crtNonce-1)),
					ShardInfo: []block.ShardData{{}, {}, {}},
				}
			},
		},
		createMockStorerWithMetablockContaining3ShardBlocks(),
		testMarshalizer,
		createMockProposerResolver(),
	)

	maxBlocks := 10

	recentBlocks, err := ner.RecentNotarizedBlocks(maxBlocks)
	assert.Equal(t, maxBlocks, len(recentBlocks))
	assert.Nil(t, err)
}

//------- RetrieveShardBlock

func TestExternalResolver_RetrieveShardBlockShouldWork(t *testing.T) {
	t.Parallel()

	crtNonce := 50

	ner, _ := external.NewExternalResolver(
		&mock.ShardCoordinatorMock{
			SelfShardId: sharding.MetachainShardId,
		},
		&mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlock{}
			},
		},
		createMockStorer(),
		testMarshalizer,
		createMockProposerResolver(),
	)

	shardBlock, err := ner.RetrieveShardBlock([]byte(fmt.Sprintf("hash_%d", crtNonce-1)))

	assert.Nil(t, err)
	assert.Equal(t, defaultShardHeader.Nonce, shardBlock.BlockHeader.Nonce)
	assert.Equal(t, defaultShardHeader.ShardId, shardBlock.BlockHeader.ShardId)
	assert.Equal(t, defaultShardHeader.TxCount, shardBlock.BlockHeader.TxCount)
	assert.Equal(t, defaultShardHeader.TimeStamp, shardBlock.BlockHeader.TimeStamp)
}
