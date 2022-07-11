package disabled

// ESDTGlobalSettingsHandler implements the ESDTGlobalSettingsHandler interface but does nothing as it is disabled
type ESDTGlobalSettingsHandler struct {
}

// IsPaused is disabled
func (e *ESDTGlobalSettingsHandler) IsPaused(_ []byte) bool {
	return false
}

// IsLimitedTransfer is disabled
func (e *ESDTGlobalSettingsHandler) IsLimitedTransfer(_ []byte) bool {
	return false
}

// IsInterfaceNil return true if underlying object is nil
func (e *ESDTGlobalSettingsHandler) IsInterfaceNil() bool {
	return e == nil
}
