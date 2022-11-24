package converter

type directStringPubkeyConverter struct{}

// NewDirectStringPubkeyConverter creates a new instance of a direct conversion from string to byte converter
func NewDirectStringPubkeyConverter() *directStringPubkeyConverter {
	return &directStringPubkeyConverter{}
}

// Len return zero
func (converter *directStringPubkeyConverter) Len() int {
	return 0
}

// Decode decodes the string as its representation in bytes
func (converter *directStringPubkeyConverter) Decode(humanReadable string) ([]byte, error) {
	return []byte(humanReadable), nil
}

// Encode encodes a byte array in its string representation
func (converter *directStringPubkeyConverter) Encode(pkBytes []byte) string {
	return string(pkBytes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (converter *directStringPubkeyConverter) IsInterfaceNil() bool {
	return converter == nil
}
