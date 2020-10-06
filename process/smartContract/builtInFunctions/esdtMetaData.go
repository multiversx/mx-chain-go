package builtInFunctions

const lengthOfESDTMetadata = 2

const (
	// MetadataPaused is the location of paused flag in the esdt global meta data
	MetadataPaused = 1
)

const (
	// MetadataFrozen is the location of frozen flag in the esdt user meta data
	MetadataFrozen = 1
)

// ESDTGlobalMetadata represents esdt global metadata saved on system account
type ESDTGlobalMetadata struct {
	Paused bool
}

// ESDTGlobalMetadataFromBytes creates a metadata object from bytes
func ESDTGlobalMetadataFromBytes(bytes []byte) ESDTGlobalMetadata {
	if len(bytes) != lengthOfESDTMetadata {
		return ESDTGlobalMetadata{}
	}

	return ESDTGlobalMetadata{
		Paused: (bytes[0] & MetadataPaused) != 0,
	}
}

// ToBytes converts the metadata to bytes
func (metadata *ESDTGlobalMetadata) ToBytes() []byte {
	bytes := make([]byte, lengthOfESDTMetadata)

	if metadata.Paused {
		bytes[0] |= MetadataPaused
	}

	return bytes
}

// ESDTUserMetadata represents esdt user metadata saved on every account
type ESDTUserMetadata struct {
	Frozen bool
}

// ESDTUserMetadataFromBytes creates a metadata object from bytes
func ESDTUserMetadataFromBytes(bytes []byte) ESDTUserMetadata {
	if len(bytes) != lengthOfESDTMetadata {
		return ESDTUserMetadata{}
	}

	return ESDTUserMetadata{
		Frozen: (bytes[0] & MetadataFrozen) != 0,
	}
}

// ToBytes converts the metadata to bytes
func (metadata *ESDTUserMetadata) ToBytes() []byte {
	bytes := make([]byte, lengthOfESDTMetadata)

	if metadata.Frozen {
		bytes[0] |= MetadataFrozen
	}

	return bytes
}
