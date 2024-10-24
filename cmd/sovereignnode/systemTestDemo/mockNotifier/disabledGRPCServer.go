package main

type disabledGRPCServer struct {
}

// NewDisabledGRPCServer -
func NewDisabledGRPCServer() *disabledGRPCServer {
	return &disabledGRPCServer{}
}

// Stop does nothing
func (dgs *disabledGRPCServer) Stop() {
}
