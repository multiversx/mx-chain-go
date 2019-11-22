package peer

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

type shardMetaMediatedMock struct {
	loadPreviousShardHeadersCalled func(header, previousHeader *block.MetaBlock) error
	loadPreviousShardHeadersMetaCalled func(header *block.MetaBlock) error
}

func (smm *shardMetaMediatedMock) loadPreviousShardHeaders(header, previousHeader *block.MetaBlock) error {
	if smm.loadPreviousShardHeadersCalled != nil {
		return smm.loadPreviousShardHeadersCalled(header, previousHeader)
	}
	return nil
}

func (smm *shardMetaMediatedMock) loadPreviousShardHeadersMeta(header *block.MetaBlock) error {
	if smm.loadPreviousShardHeadersMetaCalled != nil {
		return smm.loadPreviousShardHeadersMetaCalled(header)
	}
	return nil
}

func TestShardMediator(t *testing.T) {
	loadShardCalled := false
	loadMetaCalled := false
	sm := shardMediator{
		&shardMetaMediatedMock{
			loadPreviousShardHeadersCalled: func(header, previousHeader *block.MetaBlock) error {
				loadShardCalled = true
				return nil
			},
			loadPreviousShardHeadersMetaCalled: func(header *block.MetaBlock) error {
				loadMetaCalled = true
				return nil
			},
		},
	}

	_ = sm.loadPreviousShardHeaders(nil, nil)

	assert.True(t, loadShardCalled)
	assert.False(t, loadMetaCalled)
}

func TestShardMediatorNilMediated(t *testing.T) {
	sm := shardMediator{}
	err := sm.loadPreviousShardHeaders(nil, nil)

	assert.Equal(t, process.ErrNilMediator, err)
}

func TestMetaMediator(t *testing.T) {
	loadShardCalled := false
	loadMetaCalled := false
	sm := metaMediator{
		&shardMetaMediatedMock{
			loadPreviousShardHeadersCalled: func(header, previousHeader *block.MetaBlock) error {
				loadShardCalled = true
				return nil
			},
			loadPreviousShardHeadersMetaCalled: func(header *block.MetaBlock) error {
				loadMetaCalled = true
				return nil
			},
		},
	}

	sm.loadPreviousShardHeaders(nil, nil)

	assert.False(t, loadShardCalled)
	assert.True(t, loadMetaCalled)
}

func TestMetadMediatorNilMediated(t *testing.T) {
	sm := metaMediator{}
	err := sm.loadPreviousShardHeaders(nil, nil)

	assert.Equal(t, process.ErrNilMediator, err)
}
