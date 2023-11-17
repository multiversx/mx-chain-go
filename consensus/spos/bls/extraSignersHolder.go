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

func NewEmptyExtraSignersHolder() *extraSignersHolder {
	return &extraSignersHolder{
		startRoundHolder: NewSubRoundStartExtraSignersHolder(),
		signRoundHolder:  NewSubRoundSignatureExtraSignersHolder(),
		endRoundHolder:   NewSubRoundEndExtraSignersHolder(),
	}
}

func (holder *extraSignersHolder) GetSubRoundStartExtraSignersHolder() SubRoundStartExtraSignersHolder {
	return holder.startRoundHolder
}

func (holder *extraSignersHolder) GetSubRoundSignatureExtraSignersHolder() SubRoundSignatureExtraSignersHolder {
	return holder.signRoundHolder
}

func (holder *extraSignersHolder) GetSubRoundEndExtraSignersHolder() SubRoundEndExtraSignersHolder {
	return holder.endRoundHolder
}

func (holder *extraSignersHolder) IsInterfaceNil() bool {
	return holder == nil
}
