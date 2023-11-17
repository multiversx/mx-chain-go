package bls

type extraSignersHolder struct {
	startRoundHolder SubRoundStartExtraSignersHolder
	signRoundHolder  SubRoundSignatureExtraSignersHolder
	endRoundHolder   SubRoundEndExtraSignersHolder
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
