package bls

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
)

type extraSignersHolder struct {
	startRoundHolder SubRoundStartExtraSignersHolder
	signRoundHolder  SubRoundSignatureExtraSignersHolder
	endRoundHolder   SubRoundEndExtraSignersHolder
}

// NewExtraSignersHolder creates a holder for all extra signer holders
func NewExtraSignersHolder(
	startRoundHolder SubRoundStartExtraSignersHolder,
	signRoundHolder SubRoundSignatureExtraSignersHolder,
	endRoundHolder SubRoundEndExtraSignersHolder,
) (*extraSignersHolder, error) {
	if check.IfNil(startRoundHolder) {
		return nil, errors.ErrNilStartRoundExtraSignersHolder
	}
	if check.IfNil(signRoundHolder) {
		return nil, errors.ErrNilSignatureRoundExtraSignersHolder
	}
	if check.IfNil(endRoundHolder) {
		return nil, errors.ErrNilEndRoundExtraSignersHolder
	}

	return &extraSignersHolder{
		startRoundHolder: startRoundHolder,
		signRoundHolder:  signRoundHolder,
		endRoundHolder:   endRoundHolder,
	}, nil
}

// NewEmptyExtraSignersHolder creates an empty holder
func NewEmptyExtraSignersHolder() *extraSignersHolder {
	return &extraSignersHolder{
		startRoundHolder: NewSubRoundStartExtraSignersHolder(),
		signRoundHolder:  NewSubRoundSignatureExtraSignersHolder(),
		endRoundHolder:   NewSubRoundEndExtraSignersHolder(),
	}
}

// GetSubRoundStartExtraSignersHolder returns internal start round extra signers holder
func (holder *extraSignersHolder) GetSubRoundStartExtraSignersHolder() SubRoundStartExtraSignersHolder {
	return holder.startRoundHolder
}

// GetSubRoundSignatureExtraSignersHolder returns internal sign round extra signers holder
func (holder *extraSignersHolder) GetSubRoundSignatureExtraSignersHolder() SubRoundSignatureExtraSignersHolder {
	return holder.signRoundHolder
}

// GetSubRoundEndExtraSignersHolder returns internal end round extra signers holder
func (holder *extraSignersHolder) GetSubRoundEndExtraSignersHolder() SubRoundEndExtraSignersHolder {
	return holder.endRoundHolder
}

// IsInterfaceNil checks if the underlying pointer is nil
func (holder *extraSignersHolder) IsInterfaceNil() bool {
	return holder == nil
}
