package main

type MockServer interface {
	ExtractRandomBridgeTopicsForConfirmation() ([]*ConfirmedBridgeOp, error)
}

type GRPCServerMock interface {
	Stop()
}
