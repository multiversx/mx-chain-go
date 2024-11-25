package main

type disabledGRPCServer struct {
}

// NewDisabledGRPCServer -
func NewDisabledGRPCServer() *disabledGRPCServer {
	return &disabledGRPCServer{}
}

// Stop -
func (dgs *disabledGRPCServer) Stop() {
}
