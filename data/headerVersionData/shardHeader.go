package headerVersionData

// HeaderAdditionalData holds getters for the additional version related header data for new versions
// for future versions this interface can grow
type HeaderAdditionalData interface {
	GetScheduledRootHash() []byte
	IsInterfaceNil() bool
}

// AdditionalData holds the additional version related header data
// for future header versions this structure can grow
type AdditionalData struct {
	ScheduledRootHash []byte
}

// GetScheduledRootHash returns the scheduled RootHash
func (ad *AdditionalData) GetScheduledRootHash() []byte {
	if ad == nil {
		return nil
	}

	return ad.ScheduledRootHash
}

// IsInterfaceNil returns true if there is no value under the interface
func (ad *AdditionalData) IsInterfaceNil() bool {
	return ad == nil
}
