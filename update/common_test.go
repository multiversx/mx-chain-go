package update_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBody_ShouldErrNilHardForkBlockProcessor(t *testing.T) {
	shardIDs := []uint32{0, 1, 2, 3, 4}

	_, err := update.CreateBody(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardIDs,
		nil,
		nil,
	)
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

	_, err := update.CreateBody(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardIDs,
		nil,
		mapHardForkBlockProcessor,
	)
	assert.Equal(t, errExpected, err)
}

func TestCreateBody_ShouldErrWhenCleanDuplicatesFails(t *testing.T) {
	shardIDs := []uint32{0, 1, 2, 3, 4}
	mapBodies := make(map[uint32]*block.Body)
	hardForkBlockProcessor := &mock.HardForkBlockProcessor{
		CreateBodyCalled: func() (*block.Body, []*update.MbInfo, error) {
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

	_, err := update.CreateBody(
		nil,
		&mock.MarshalizerMock{},
		shardIDs,
		mapBodies,
		mapHardForkBlockProcessor,
	)
	assert.Equal(t, update.ErrNilHasher, err)
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

	postMbs, err := update.CreateBody(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardIDs,
		mapBodies,
		mapHardForkBlockProcessor,
	)
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

	err := update.CreatePostMiniBlocks(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardIDs,
		lastPostMbs,
		nil,
		nil,
	)
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

	err := update.CreatePostMiniBlocks(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardIDs,
		lastPostMbs,
		nil,
		mapHardForkBlockProcessor,
	)
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

	err := update.CreatePostMiniBlocks(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardIDs,
		lastPostMbs,
		nil,
		mapHardForkBlockProcessor,
	)
	assert.Equal(t, update.ErrNilBlockBody, err)
}

func TestCreatePostMiniBlocks_ShouldErrWhenCleanDuplicatesFails(t *testing.T) {
	shardIDs := []uint32{0, 1, 2, 3, 4}
	mapBodies := map[uint32]*block.Body{
		0: {},
		1: {},
		2: {},
		3: {},
		4: {},
	}
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

	err := update.CreatePostMiniBlocks(
		nil,
		&mock.MarshalizerMock{},
		shardIDs,
		lastPostMbs,
		mapBodies,
		mapHardForkBlockProcessor,
	)
	assert.Equal(t, update.ErrNilHasher, err)
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

	err := update.CreatePostMiniBlocks(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardIDs,
		lastPostMbs,
		mapBodies,
		mapHardForkBlockProcessor,
	)
	assert.Nil(t, err)
	require.Equal(t, 2, len(mapBodies))

	require.Equal(t, 2, len(mapBodies[0].MiniBlocks))
	assert.Equal(t, mb1, mapBodies[0].MiniBlocks[0])
	assert.Equal(t, mb1post, mapBodies[0].MiniBlocks[1])

	require.Equal(t, 2, len(mapBodies[1].MiniBlocks))
	assert.Equal(t, mb2, mapBodies[1].MiniBlocks[0])
	assert.Equal(t, mb2post, mapBodies[1].MiniBlocks[1])
}

func TestCleanDuplicates_ShouldErrNilHasher(t *testing.T) {
	shardIDs := []uint32{0, 1}
	mapBodies := map[uint32]*block.Body{
		0: {},
		1: {},
	}
	postMbs := []*update.MbInfo{
		{MbHash: []byte("hash1")},
		{MbHash: []byte("hash2")},
	}

	_, err := update.CleanDuplicates(nil, &mock.MarshalizerMock{}, shardIDs, mapBodies, postMbs)
	assert.Equal(t, update.ErrNilHasher, err)
}

func TestCleanDuplicates_ShouldErrNilMarshalizer(t *testing.T) {
	shardIDs := []uint32{0, 1}
	mapBodies := map[uint32]*block.Body{
		0: {},
		1: {},
	}
	postMbs := []*update.MbInfo{
		{MbHash: []byte("hash1")},
		{MbHash: []byte("hash2")},
	}

	_, err := update.CleanDuplicates(&mock.HasherMock{}, nil, shardIDs, mapBodies, postMbs)
	assert.Equal(t, update.ErrNilMarshalizer, err)
}

func TestCleanDuplicates_ShouldErrNilBlockBody(t *testing.T) {
	shardIDs := []uint32{0, 1}
	postMbs := []*update.MbInfo{
		{MbHash: []byte("hash1")},
		{MbHash: []byte("hash2")},
	}

	_, err := update.CleanDuplicates(&mock.HasherMock{}, &mock.MarshalizerMock{}, shardIDs, nil, postMbs)
	assert.Equal(t, update.ErrNilBlockBody, err)
}

func TestCleanDuplicates_ShouldErrWhenCalculateHashFails(t *testing.T) {
	shardIDs := []uint32{0, 1}
	mapBodies := map[uint32]*block.Body{
		0: {MiniBlocks: []*block.MiniBlock{
			{},
		}},
		1: {MiniBlocks: []*block.MiniBlock{
			{},
		}},
	}
	postMbs := []*update.MbInfo{
		{MbHash: []byte("hash1")},
		{MbHash: []byte("hash2")},
	}

	_, err := update.CleanDuplicates(&mock.HasherMock{}, &mock.MarshalizerMock{Fail: true}, shardIDs, mapBodies, postMbs)
	assert.NotNil(t, err)
}

func TestCleanDuplicates_ShouldWork(t *testing.T) {
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	shardIDs := []uint32{0, 1, 2, 3, 4}
	mb10 := &block.MiniBlock{SenderShardID: 1, ReceiverShardID: 0}
	mb20 := &block.MiniBlock{SenderShardID: 2, ReceiverShardID: 0}
	mb01 := &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 1}
	mb21 := &block.MiniBlock{SenderShardID: 2, ReceiverShardID: 1}
	mb31 := &block.MiniBlock{SenderShardID: 3, ReceiverShardID: 1}
	mb10Hash, _ := core.CalculateHash(marshalizer, hasher, mb10)
	mb31Hash, _ := core.CalculateHash(marshalizer, hasher, mb31)
	mapBodies := map[uint32]*block.Body{
		0: {MiniBlocks: []*block.MiniBlock{
			mb10,
			mb20,
		}},
		1: {MiniBlocks: []*block.MiniBlock{
			mb01,
			mb21,
			mb31,
		}},
		2: {},
		3: {},
		4: {},
	}
	postMbs := []*update.MbInfo{
		{MbHash: []byte("hash1")},
		{MbHash: mb10Hash},
		{MbHash: mb31Hash},
		{MbHash: []byte("hash4")},
	}

	cleanedMbs, err := update.CleanDuplicates(&mock.HasherMock{}, &mock.MarshalizerMock{}, shardIDs, mapBodies, postMbs)
	assert.Nil(t, err)
	require.Equal(t, 2, len(cleanedMbs))
	assert.Equal(t, cleanedMbs[0].MbHash, []byte("hash1"))
	assert.Equal(t, cleanedMbs[1].MbHash, []byte("hash4"))
}
