package grpcdriver

import "errors"

var (
	// ErrNilGRPCClient signals that no gRPC client was provided.
	ErrNilGRPCClient = errors.New("nil grpc client")
	// ErrNilMarshaller signals that no marshaller was provided.
	ErrNilMarshaller = errors.New("nil grpc driver marshaller")
)
