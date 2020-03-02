package preprocess

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewValidatorInfoPreprocessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	rtp, err := NewValidatorInfoPreprocessor(
		nil,
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewValidatorInfoPreprocessor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	rtp, err := NewValidatorInfoPreprocessor(
		&mock.HasherMock{},
		nil,
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewValidatorInfoPreprocessorPreprocessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	rtp, err := NewValidatorInfoPreprocessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, rtp)
}

func TestNewValidatorInfoPreprocessorPreprocessor_CreateMarshalizedDataShouldWork(t *testing.T) {
	t.Parallel()

	rtp, err := NewValidatorInfoPreprocessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	hash := make([][]byte, 0)
	res, err := rtp.CreateMarshalizedData(hash)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(res))
}

func TestNewValidatorInfoPreprocessorPreprocessor_ProcessMiniBlockInvalidMiniBlockTypeShouldErr(t *testing.T) {
	t.Parallel()

	rtp, err := NewValidatorInfoPreprocessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	txHashes := make([][]byte, 0)
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            0,
	}

	err = rtp.ProcessMiniBlock(&mb1, haveTimeTrue)
	assert.Equal(t, process.ErrWrongTypeInMiniBlock, err)
}

func TestNewValidatorInfoPreprocessorPreprocessor_ProcessMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	rtp, err := NewValidatorInfoPreprocessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	txHashes := make([][]byte, 0)
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   core.MetachainShardId,
		Type:            block.PeerBlock,
	}

	err = rtp.ProcessMiniBlock(&mb1, haveTimeTrue)
	assert.Nil(t, err)
}

func TestNewValidatorInfoPreprocessorPreprocessor_ProcessMiniBlockNotFromMeta(t *testing.T) {
	t.Parallel()

	rtp, err := NewValidatorInfoPreprocessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	txHashes := make([][]byte, 0)
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.PeerBlock,
	}

	err = rtp.ProcessMiniBlock(&mb1, haveTimeTrue)
	assert.Equal(t, process.ErrRewardMiniBlockNotFromMeta, err)
}
