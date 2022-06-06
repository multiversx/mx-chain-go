package disabled

// EnableRoundsHandler is the disabled variant of the EnableRoundsHandler interface
type EnableRoundsHandler struct {
}

// CheckRound does nothing
func (handler *EnableRoundsHandler) CheckRound(_ uint64) {
}

// IsCheckValueOnExecByCallerEnabled returns false
func (handler *EnableRoundsHandler) IsCheckValueOnExecByCallerEnabled() bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *EnableRoundsHandler) IsInterfaceNil() bool {
	return handler == nil
}
