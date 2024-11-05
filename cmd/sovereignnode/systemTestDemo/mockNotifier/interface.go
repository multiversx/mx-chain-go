package main

// MockServer holds the mock server actions
type MockServer interface {
	ExtractRandomBridgeTopicsForConfirmation() ([]*ConfirmedBridgeOp, error)
}

// GRPCServerMock holds the grpc server actions
type GRPCServerMock interface {
	Stop()
}
