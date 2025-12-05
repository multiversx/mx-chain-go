package headerForBlock

import "github.com/multiversx/mx-chain-core-go/data"

// HeaderInfo holds the information about a header
type HeaderInfo interface {
	GetHeader() data.HeaderHandler
	UsedInBlock() bool
	HasProof() bool
	HasProofRequested() bool
	SetUsedInBlock(bool)
	SetHeader(data.HeaderHandler)
	SetHasProof(bool)
	SetHasProofRequested(bool)
	IsInterfaceNil() bool
}

// LastNotarizedHeaderInfoHandler is an interface that has the methods for the last notarized header info
type LastNotarizedHeaderInfoHandler interface {
	GetHeader() data.HeaderHandler
	GetHash() []byte
	SetHeader(hdr data.HeaderHandler)
	SetHash(hash []byte)
	HasProof() bool
	SetHasProof(hasProof bool)
	NotarizedBasedOnProof() bool
	SetNotarizedBasedOnProof(notarizedBasedOnProof bool)
	IsInterfaceNil() bool
}

// NonceAndHashInfo defines a component that holds nonce and hash
type NonceAndHashInfo interface {
	GetNonce() uint64
	GetHash() []byte
	IsInterfaceNil() bool
}
