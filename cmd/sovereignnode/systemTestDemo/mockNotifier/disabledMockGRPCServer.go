package main

type disabledMockServer struct {
}

// NewDisabledMockServer -
func NewDisabledMockServer() *disabledMockServer {
	return &disabledMockServer{}
}

// ExtractRandomBridgeTopicsForConfirmation -
func (dms *disabledMockServer) ExtractRandomBridgeTopicsForConfirmation() ([]*ConfirmedBridgeOp, error) {
	return make([]*ConfirmedBridgeOp, 0), nil
}

// IsInterfaceNil -
func (dms *disabledMockServer) IsInterfaceNil() bool {
	return dms == nil
}
