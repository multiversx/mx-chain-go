package vmcommon

const lengthOfCodeMetadata = 2

const (
	// MetadataUpgradeable is the bit for upgradable flag
	MetadataUpgradeable = 1
	// MetadataPayable is the bit for payable flag
	MetadataPayable = 2
	// MetadataReadable is the bit for readable flag
	MetadataReadable = 4
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
		Upgradeable: (bytes[0] & MetadataUpgradeable) != 0,
		Readable:    (bytes[0] & MetadataReadable) != 0,
		Payable:     (bytes[1] & MetadataPayable) != 0,
	}
}

// ToBytes converts the metadata to bytes
func (metadata *CodeMetadata) ToBytes() []byte {
	bytes := make([]byte, lengthOfCodeMetadata)

	if metadata.Upgradeable {
		bytes[0] |= MetadataUpgradeable
	}
	if metadata.Readable {
		bytes[0] |= MetadataReadable
	}
	if metadata.Payable {
		bytes[1] |= MetadataPayable
	}

	return bytes
}
