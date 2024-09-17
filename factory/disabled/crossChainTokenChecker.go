package disabled

type crossChainTokenChecker struct {
}

func NewCrossChainTokenChecker() *crossChainTokenChecker {
	return &crossChainTokenChecker{}
}

// IsCrossChainOperation does nothing
func (ctc *crossChainTokenChecker) IsCrossChainOperation(tokenID []byte) bool {
	return false
}

// IsCrossChainOperationAllowed does nothing
func (ctc *crossChainTokenChecker) IsCrossChainOperationAllowed(address []byte, tokenID []byte) bool {
	return false
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ctc *crossChainTokenChecker) IsInterfaceNil() bool {
	return ctc == nil
}
