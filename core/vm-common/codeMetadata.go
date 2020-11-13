package vmcommon

const lengthOfCodeMetadata = 2

const (
	METADATA_UPGRADEABLE = 1
	METADATA_PAYABLE     = 2
	METADATA_READABLE    = 4
)

// CodeMetadata represents smart contract code metadata
type CodeMetadata struct {
	Payable     bool
	Upgradeable bool
	Readable    bool
}

// CodeMetadataFromBytes creates a metadata object from bytes
func CodeMetadataFromBytes(bytes []byte) CodeMetadata {
	if len(bytes) != lengthOfCodeMetadata {
		return CodeMetadata{}
	}

	return CodeMetadata{
		Upgradeable: (bytes[0] & METADATA_UPGRADEABLE) != 0,
		Readable:    (bytes[0] & METADATA_READABLE) != 0,
		Payable:     (bytes[1] & METADATA_PAYABLE) != 0,
	}
}

// ToBytes converts the metadata to bytes
func (metadata *CodeMetadata) ToBytes() []byte {
	bytes := make([]byte, lengthOfCodeMetadata)

	if metadata.Upgradeable {
		bytes[0] |= METADATA_UPGRADEABLE
	}
	if metadata.Readable {
		bytes[0] |= METADATA_READABLE
	}
	if metadata.Payable {
		bytes[1] |= METADATA_PAYABLE
	}

	return bytes
}
