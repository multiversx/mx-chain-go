package update_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckDuplicates_ShouldErrDuplicatedMiniBlocksFound(t *testing.T) {
	mapHeaders := map[uint32]data.HeaderHandler{
		0: &block.Header{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{Hash: []byte("hash")},
				{Hash: []byte("hash")},
			}},
		1: &block.Header{},
	}

	err := update.CheckDuplicates(mapHeaders)
	assert.Equal(t, update.ErrDuplicatedMiniBlocksFound, err)
}

func TestCheckDuplicates_ShouldReturnNil(t *testing.T) {
	mapHeaders := map[uint32]data.HeaderHandler{
		0: &block.Header{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{Hash: []byte("hash1")},
				{Hash: []byte("hash2")},
			}},
		1: &block.Header{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{Hash: []byte("hash3")},
				{Hash: []byte("hash4")},
				{Hash: []byte("hash5")},
			}},
	}

	err := update.CheckDuplicates(mapHeaders)
	assert.Nil(t, err)
}

func TestGetLastPostMbs_ShouldWork(t *testing.T) {
	shardIDs := []uint32{0, 1, 2, 3, 4}
	mapPostMbs := map[uint32][]*update.MbInfo{
		0: {
			{MbHash: []byte("hash0_1")},
			{MbHash: []byte("hash0_2")},
		},
		3: {
			{MbHash: []byte("hash1_1")},
			{MbHash: []byte("hash1_2")},
			{MbHash: []byte("hash1_3")},
		},
	}

	mbsInfo := update.GetLastPostMbs(shardIDs, mapPostMbs)

	require.Equal(t, 5, len(mbsInfo))
	assert.Equal(t, []byte("hash0_1"), mbsInfo[0].MbHash)
	assert.Equal(t, []byte("hash0_2"), mbsInfo[1].MbHash)
	assert.Equal(t, []byte("hash1_1"), mbsInfo[2].MbHash)
	assert.Equal(t, []byte("hash1_2"), mbsInfo[3].MbHash)
	assert.Equal(t, []byte("hash1_3"), mbsInfo[4].MbHash)
}

func TestCreateBody_ShouldErrNilHardForkBlockProcessor(t *testing.T) {
	shardIDs := []uint32{0, 1, 2, 3, 4}

	_, err := update.CreateBody(shardIDs, nil, nil)
	assert.Equal(t, update.ErrNilHardForkBlockProcessor, err)
}

func TestCreateBody_ShouldErrWhenCreateBodyFails(t *testing.T) {
	shardIDs := []uint32{0, 1, 2, 3, 4}
	errExpected := errors.New("error")
	hardForkBlockProcessor := &mock.HardForkBlockProcessor{
		CreateBodyCalled: func() (*block.Body, []*update.MbInfo, error) {
			return nil, nil, errExpected
		},
	}
	mapHardForkBlockProcessor := map[uint32]update.HardForkBlockProcessor{
		0: hardForkBlockProcessor,
		1: hardForkBlockProcessor,
		2: hardForkBlockProcessor,
		3: hardForkBlockProcessor,
		4: hardForkBlockProcessor,
	}

	_, err := update.CreateBody(shardIDs, nil, mapHardForkBlockProcessor)
	assert.Equal(t, errExpected, err)
}

func TestCreateBody_ShouldWork(t *testing.T) {
	shardIDs := []uint32{0, 1}
	mapBodies := make(map[uint32]*block.Body)
	body1 := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				SenderShardID:   1,
				ReceiverShardID: 0,
			},
		},
	}
	mbsInfo1 := []*update.MbInfo{
		{
			MbHash:          []byte("hash1"),
			SenderShardID:   0,
			ReceiverShardID: 1,
		},
	}
	body2 := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				SenderShardID:   0,
				ReceiverShardID: 1,
			},
		},
	}
	mbsInfo2 := []*update.MbInfo{
		{
			MbHash:          []byte("hash2"),
			SenderShardID:   1,
			ReceiverShardID: 0,
		},
	}
	hardForkBlockProcessor1 := &mock.HardForkBlockProcessor{
		CreateBodyCalled: func() (*block.Body, []*update.MbInfo, error) {
			return body1, mbsInfo1, nil
		},
	}
	hardForkBlockProcessor2 := &mock.HardForkBlockProcessor{
		CreateBodyCalled: func() (*block.Body, []*update.MbInfo, error) {
			return body2, mbsInfo2, nil
		},
	}
	mapHardForkBlockProcessor := map[uint32]update.HardForkBlockProcessor{
		0: hardForkBlockProcessor1,
		1: hardForkBlockProcessor2,
	}

	postMbs, err := update.CreateBody(shardIDs, mapBodies, mapHardForkBlockProcessor)
	assert.Nil(t, err)
	require.Equal(t, 2, len(mapBodies))
	require.Equal(t, 2, len(postMbs))
	assert.Equal(t, body1, mapBodies[0])
	assert.Equal(t, body2, mapBodies[1])
	assert.Equal(t, mbsInfo1[0], postMbs[0])
	assert.Equal(t, mbsInfo2[0], postMbs[1])
}

func TestCreatePostMiniBlocks_ShouldErrNilHardForkBlockProcessor(t *testing.T) {
	shardIDs := []uint32{0, 1, 2, 3, 4}
	lastPostMbs := []*update.MbInfo{
		{MbHash: []byte("hash")},
	}

	err := update.CreatePostMiniBlocks(shardIDs, lastPostMbs, nil, nil)
	assert.Equal(t, update.ErrNilHardForkBlockProcessor, err)
}

func TestCreatePostMiniBlocks_ShouldErrWhenCreatePostMiniBlocksFails(t *testing.T) {
	shardIDs := []uint32{0, 1, 2, 3, 4}
	lastPostMbs := []*update.MbInfo{
		{MbHash: []byte("hash")},
	}
	errExpected := errors.New("error")
	hardForkBlockProcessor := &mock.HardForkBlockProcessor{
		CreatePostMiniBlocksCalled: func(mbsInfo []*update.MbInfo) (*block.Body, []*update.MbInfo, error) {
			return nil, nil, errExpected
		},
	}
	mapHardForkBlockProcessor := map[uint32]update.HardForkBlockProcessor{
		0: hardForkBlockProcessor,
		1: hardForkBlockProcessor,
		2: hardForkBlockProcessor,
		3: hardForkBlockProcessor,
		4: hardForkBlockProcessor,
	}

	err := update.CreatePostMiniBlocks(shardIDs, lastPostMbs, nil, mapHardForkBlockProcessor)
	assert.Equal(t, errExpected, err)
}

func TestCreatePostMiniBlocks_ShouldErrNilBlockBody(t *testing.T) {
	shardIDs := []uint32{0, 1, 2, 3, 4}
	lastPostMbs := []*update.MbInfo{
		{MbHash: []byte("hash")},
	}
	hardForkBlockProcessor := &mock.HardForkBlockProcessor{
		CreatePostMiniBlocksCalled: func(mbsInfo []*update.MbInfo) (*block.Body, []*update.MbInfo, error) {
			return &block.Body{}, nil, nil
		},
	}
	mapHardForkBlockProcessor := map[uint32]update.HardForkBlockProcessor{
		0: hardForkBlockProcessor,
		1: hardForkBlockProcessor,
		2: hardForkBlockProcessor,
		3: hardForkBlockProcessor,
		4: hardForkBlockProcessor,
	}

	err := update.CreatePostMiniBlocks(shardIDs, lastPostMbs, nil, mapHardForkBlockProcessor)
	assert.Equal(t, update.ErrNilBlockBody, err)
}

func TestCreatePostMiniBlocks_ShouldWork(t *testing.T) {
	shardIDs := []uint32{0, 1}
	lastPostMbs := []*update.MbInfo{
		{MbHash: []byte("hash")},
	}

	mapBodies := map[uint32]*block.Body{
		0: {},
		1: {},
	}

	mb1 := &block.MiniBlock{
		Type:            block.TxBlock,
		SenderShardID:   1,
		ReceiverShardID: 0,
	}
	body1 := &block.Body{
		MiniBlocks: []*block.MiniBlock{mb1},
	}

	mb1post := &block.MiniBlock{
		Type:            block.SmartContractResultBlock,
		SenderShardID:   1,
		ReceiverShardID: 0,
	}
	body1post := &block.Body{
		MiniBlocks: []*block.MiniBlock{mb1post},
	}

	mb2 := &block.MiniBlock{
		Type:            block.TxBlock,
		SenderShardID:   0,
		ReceiverShardID: 1,
	}
	body2 := &block.Body{
		MiniBlocks: []*block.MiniBlock{mb2},
	}

	mb2post := &block.MiniBlock{
		Type:            block.SmartContractResultBlock,
		SenderShardID:   0,
		ReceiverShardID: 1,
	}
	body2post := &block.Body{
		MiniBlocks: []*block.MiniBlock{mb2post},
	}

	mbsInfo1 := []*update.MbInfo{
		{
			MbHash:          []byte("hash1"),
			SenderShardID:   0,
			ReceiverShardID: 1,
			Type:            block.SmartContractResultBlock,
		},
	}

	mbsInfo2 := []*update.MbInfo{
		{
			MbHash:          []byte("hash2"),
			SenderShardID:   1,
			ReceiverShardID: 0,
			Type:            block.SmartContractResultBlock,
		},
	}

	hardForkBlockProcessor1 := &mock.HardForkBlockProcessor{
		CreatePostMiniBlocksCalled: func(mbsInfo []*update.MbInfo) (*block.Body, []*update.MbInfo, error) {
			if bytes.Compare(mbsInfo[0].MbHash, []byte("hash")) == 0 {
				return body1, mbsInfo1, nil
			}
			return body1post, nil, nil
		},
	}

	hardForkBlockProcessor2 := &mock.HardForkBlockProcessor{
		CreatePostMiniBlocksCalled: func(mbsInfo []*update.MbInfo) (*block.Body, []*update.MbInfo, error) {
			if bytes.Compare(mbsInfo[0].MbHash, []byte("hash")) == 0 {
				return body2, mbsInfo2, nil
			}
			return body2post, nil, nil
		},
	}
	mapHardForkBlockProcessor := map[uint32]update.HardForkBlockProcessor{
		0: hardForkBlockProcessor1,
		1: hardForkBlockProcessor2,
	}

	err := update.CreatePostMiniBlocks(shardIDs, lastPostMbs, mapBodies, mapHardForkBlockProcessor)
	assert.Nil(t, err)
	require.Equal(t, 2, len(mapBodies))

	require.Equal(t, 2, len(mapBodies[0].MiniBlocks))
	assert.Equal(t, mb1, mapBodies[0].MiniBlocks[0])
	assert.Equal(t, mb1post, mapBodies[0].MiniBlocks[1])

	require.Equal(t, 2, len(mapBodies[1].MiniBlocks))
	assert.Equal(t, mb2, mapBodies[1].MiniBlocks[0])
	assert.Equal(t, mb2post, mapBodies[1].MiniBlocks[1])
}
