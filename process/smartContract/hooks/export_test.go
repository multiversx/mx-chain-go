package hooks

// CodeMetadataBytesUnchanged -
const CodeMetadataBytesUnchanged = codeMetadataBytesUnchanged

// CodeMetadataValidBytes -
const CodeMetadataValidBytes = codeMetadataValidBytes

// CodeMetadataInvalidBytes -
const CodeMetadataInvalidBytes = codeMetadataInvalidBytes

// SetFlagOptimizeNFTStore -
func (bh *BlockChainHookImpl) SetFlagOptimizeNFTStore(value bool) {
	bh.flagOptimizeNFTStore.SetValue(value)
}
