package preprocess

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type validatorInfoPreprocessor struct {
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

// NewValidatorInfoPreprocessor creates a new validatorInfo preprocessor object
func NewValidatorInfoPreprocessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) (*validatorInfoPreprocessor, error) {
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}

	rtp := &validatorInfoPreprocessor{
		hasher:      hasher,
		marshalizer: marshalizer,
	}
	return rtp, nil
}

// IsDataPrepared does nothing
func (rtp *validatorInfoPreprocessor) IsDataPrepared(_ int, _ func() time.Duration) error {
	return nil
}

// RemoveTxBlockFromPools does nothing
func (rtp *validatorInfoPreprocessor) RemoveTxBlockFromPools(_ block.Body, _ storage.Cacher) error {
	return nil
}

// RestoreTxBlockIntoPools does nothing
func (rtp *validatorInfoPreprocessor) RestoreTxBlockIntoPools(
	_ block.Body,
	_ storage.Cacher,
) (int, error) {
	return 0, nil
}

// ProcessBlockTransactions does nothing
func (rtp *validatorInfoPreprocessor) ProcessBlockTransactions(
	_ block.Body,
	_ func() bool,
) error {
	return nil
}

// SaveTxBlockToStorage does nothing
func (rtp *validatorInfoPreprocessor) SaveTxBlockToStorage(_ block.Body) error {
	return nil
}

// CreateBlockStarted does nothing
func (rtp *validatorInfoPreprocessor) CreateBlockStarted() {
}

// RequestBlockTransactions does nothing
func (rtp *validatorInfoPreprocessor) RequestBlockTransactions(_ block.Body) int {
	return 0
}

// RequestTransactionsForMiniBlock does nothing
func (rtp *validatorInfoPreprocessor) RequestTransactionsForMiniBlock(_ *block.MiniBlock) int {
	return 0
}

// CreateAndProcessMiniBlocks does nothing
func (rtp *validatorInfoPreprocessor) CreateAndProcessMiniBlocks(
	_ uint32,
	_ uint32,
	_ func() bool,
) (block.MiniBlockSlice, error) {
	// validatorInfos are created only by meta
	return make(block.MiniBlockSlice, 0), nil
}

// ProcessMiniBlock does nothing
func (rtp *validatorInfoPreprocessor) ProcessMiniBlock(miniBlock *block.MiniBlock, _ func() bool) error {
	if miniBlock.Type != block.PeerBlock {
		return process.ErrWrongTypeInMiniBlock
	}
	if miniBlock.SenderShardID != core.MetachainShardId {
		return process.ErrValidatorInfoMiniBlockNotFromMeta
	}

	return nil
}

// CreateMarshalizedData does nothing
func (rtp *validatorInfoPreprocessor) CreateMarshalizedData(_ [][]byte) ([][]byte, error) {
	marshalized := make([][]byte, 0)
	return marshalized, nil
}

// GetAllCurrentUsedTxs does nothing
func (rtp *validatorInfoPreprocessor) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	validatorInfoTxPool := make(map[string]data.TransactionHandler)

	return validatorInfoTxPool
}

// IsInterfaceNil does nothing
func (rtp *validatorInfoPreprocessor) IsInterfaceNil() bool {
	return rtp == nil
}
