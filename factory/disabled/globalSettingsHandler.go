package disabled

type disabledGlobalSettingsHandler struct{}

// NewDisabledGlobalSettingHandler returns a new instance of disabledGlobalSettingsHandler
func NewDisabledGlobalSettingHandler() *disabledGlobalSettingsHandler {
	return &disabledGlobalSettingsHandler{}
}

// IsBurnForAll returns false as this is a disabled component
func (d *disabledGlobalSettingsHandler) IsBurnForAll(_ []byte) bool {
	return false
}

// IsSenderOrDestinationWithTransferRole returns false as this is a disabled component
func (d *disabledGlobalSettingsHandler) IsSenderOrDestinationWithTransferRole(_, _, _ []byte) bool {
	return false
}

// GetTokenType returns 0 as this is a disabled component
func (d *disabledGlobalSettingsHandler) GetTokenType(_ []byte) (uint32, error) {
	return 0, nil
}

// SetTokenType does nothing as this is a disabled component
func (d *disabledGlobalSettingsHandler) SetTokenType(_ []byte, _ uint32) error {
	return nil
}

// IsPaused returns false as this is a disabled component
func (d *disabledGlobalSettingsHandler) IsPaused(_ []byte) bool {
	return false
}

// IsLimitedTransfer returns false as this is a disabled components
func (d *disabledGlobalSettingsHandler) IsLimitedTransfer(_ []byte) bool {
	return false
}

// IsInterfaceNil returns true if the value under interface is nil
func (d *disabledGlobalSettingsHandler) IsInterfaceNil() bool {
	return d == nil
}
