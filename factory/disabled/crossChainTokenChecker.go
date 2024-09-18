package disabled

type crossChainTokenChecker struct {
}

// NewCrossChainTokenChecker creates a new cross chain token checker
func NewCrossChainTokenChecker() *crossChainTokenChecker {
	return &crossChainTokenChecker{}
}

// IsCrossChainOperation returns false
func (ctc *crossChainTokenChecker) IsCrossChainOperation(_ []byte) bool {
	return false
}

// IsCrossChainOperationAllowed returns false
func (ctc *crossChainTokenChecker) IsCrossChainOperationAllowed(_ []byte, _ []byte) bool {
	return false
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ctc *crossChainTokenChecker) IsInterfaceNil() bool {
	return ctc == nil
}
