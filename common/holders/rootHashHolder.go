package holders

type rootHashHolder struct {
	rootHash []byte
}

// NewRootHashHolder creates a rootHashHolder
func NewRootHashHolder(rootHash []byte) *rootHashHolder {
	return &rootHashHolder{
		rootHash: rootHash,
	}
}

// NewRootHashHolderAsEmpty creates an empty rootHashHolder
func NewRootHashHolderAsEmpty() *rootHashHolder {
	return &rootHashHolder{
		rootHash: nil,
	}
}

// GetRootHash returns the contained rootHash
func (holder *rootHashHolder) GetRootHash() []byte {
	return holder.rootHash
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *rootHashHolder) IsInterfaceNil() bool {
	return holder == nil
}
