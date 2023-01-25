package fallback_test

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/fallback"
	"github.com/multiversx/mx-chain-go/fallback/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewFallbackHeaderValidator_ShouldErrNilHeadersDataPool(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerStub{}
	storageService := &storage.ChainStorerStub{}

	fhv, err := fallback.NewFallbackHeaderValidator(nil, marshalizer, storageService)
	assert.Nil(t, fhv)
	assert.Equal(t, process.ErrNilHeadersDataPool, err)
}

func TestNewFallbackHeaderValidator_ShouldErrNilMarshalizer(t *testing.T) {
	t.Parallel()

	headersPool := &mock.HeadersCacherStub{}
	storageService := &storage.ChainStorerStub{}

	fhv, err := fallback.NewFallbackHeaderValidator(headersPool, nil, storageService)
	assert.Nil(t, fhv)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewFallbackHeaderValidator_ShouldErrNilStorage(t *testing.T) {
	t.Parallel()

	headersPool := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerStub{}

	fhv, err := fallback.NewFallbackHeaderValidator(headersPool, marshalizer, nil)
	assert.Nil(t, fhv)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewFallbackHeaderValidator_ShouldWork(t *testing.T) {
	t.Parallel()

	headersPool := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerStub{}
	storageService := &storage.ChainStorerStub{}

	fhv, err := fallback.NewFallbackHeaderValidator(headersPool, marshalizer, storageService)
	assert.False(t, check.IfNil(fhv))
	assert.Nil(t, err)
}

func TestShouldApplyFallbackConsensus_ShouldReturnFalseWhenHeaderIsNil(t *testing.T) {
	t.Parallel()

	headersPool := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerStub{}
	storageService := &storage.ChainStorerStub{}

	fhv, _ := fallback.NewFallbackHeaderValidator(headersPool, marshalizer, storageService)
	assert.False(t, fhv.ShouldApplyFallbackValidation(nil))
}

func TestShouldApplyFallbackConsensus_ShouldReturnFalseWhenIsNotMetachainBlock(t *testing.T) {
	t.Parallel()

	headersPool := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerStub{}
	storageService := &storage.ChainStorerStub{}
	header := &block.Header{}

	fhv, _ := fallback.NewFallbackHeaderValidator(headersPool, marshalizer, storageService)
	assert.False(t, fhv.ShouldApplyFallbackValidation(header))
}

func TestShouldApplyFallbackConsensus_ShouldReturnFalseWhenIsNotStartOfEpochMetachainBlock(t *testing.T) {
	t.Parallel()

	headersPool := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerStub{}
	storageService := &storage.ChainStorerStub{}
	metaBlock := &block.MetaBlock{}

	fhv, _ := fallback.NewFallbackHeaderValidator(headersPool, marshalizer, storageService)
	assert.False(t, fhv.ShouldApplyFallbackValidation(metaBlock))
}

func TestShouldApplyFallbackConsensus_ShouldReturnFalseWhenPreviousHeaderIsNotFound(t *testing.T) {
	t.Parallel()

	headersPool := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerStub{}
	storageService := &storage.ChainStorerStub{}
	epochStartShardData := block.EpochStartShardData{}
	metaBlock := &block.MetaBlock{
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				epochStartShardData,
			},
		},
	}

	fhv, _ := fallback.NewFallbackHeaderValidator(headersPool, marshalizer, storageService)
	assert.False(t, fhv.ShouldApplyFallbackValidation(metaBlock))
}

func TestShouldApplyFallbackConsensus_ShouldReturnFalseWhenRoundIsNotTooOld(t *testing.T) {
	t.Parallel()

	prevHash := []byte("prev_hash")
	headersPool := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			if bytes.Equal(hash, prevHash) {
				return &block.MetaBlock{}, nil
			}
			return nil, errors.New("error")
		},
	}
	marshalizer := &mock.MarshalizerStub{}
	storageService := &storage.ChainStorerStub{}
	epochStartShardData := block.EpochStartShardData{}
	metaBlock := &block.MetaBlock{
		Round: common.MaxRoundsWithoutCommittedStartInEpochBlock - 1,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				epochStartShardData,
			},
		},
		PrevHash: prevHash,
	}

	fhv, _ := fallback.NewFallbackHeaderValidator(headersPool, marshalizer, storageService)
	assert.False(t, fhv.ShouldApplyFallbackValidation(metaBlock))
}

func TestShouldApplyFallbackConsensus_ShouldReturnTrue(t *testing.T) {
	t.Parallel()

	prevHash := []byte("prev_hash")
	headersPool := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			if bytes.Equal(hash, prevHash) {
				return &block.MetaBlock{}, nil
			}
			return nil, errors.New("error")
		},
	}
	marshalizer := &mock.MarshalizerStub{}
	storageService := &storage.ChainStorerStub{}
	epochStartShardData := block.EpochStartShardData{}
	metaBlock := &block.MetaBlock{
		Round: common.MaxRoundsWithoutCommittedStartInEpochBlock,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				epochStartShardData,
			},
		},
		PrevHash: prevHash,
	}

	fhv, _ := fallback.NewFallbackHeaderValidator(headersPool, marshalizer, storageService)
	assert.True(t, fhv.ShouldApplyFallbackValidation(metaBlock))
}
