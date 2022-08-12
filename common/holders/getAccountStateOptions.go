package holders

type getAccountStateOptions struct {
	blockRootHash []byte
}

// NewGetAccountStateOptions creates a getAccountStateOptions
func NewGetAccountStateOptions(blockRootHash []byte) *getAccountStateOptions {
	return &getAccountStateOptions{blockRootHash: blockRootHash}
}

// NewGetAccountStateOptionsAsEmpty creates an empty getAccountStateOptions
func NewGetAccountStateOptionsAsEmpty() *getAccountStateOptions {
	return &getAccountStateOptions{blockRootHash: nil}
}

// GetBlockRootHash returns the contained blockRootHash
func (options *getAccountStateOptions) GetBlockRootHash() []byte {
	return options.GetBlockRootHash()
}

// IsInterfaceNil returns true if there is no value under the interface
func (options *getAccountStateOptions) IsInterfaceNil() bool {
	return options == nil
}
