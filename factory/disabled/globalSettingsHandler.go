package disabled

type disabledGlobalSettingsHandler struct{}

// NewDisabledGlobalSettingHandler returns a new instance of disabledGlobalSettingsHandler
func NewDisabledGlobalSettingHandler() *disabledGlobalSettingsHandler {
	return &disabledGlobalSettingsHandler{}
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
