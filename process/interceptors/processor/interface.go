package processor

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/state"
)

type interceptedDataSizeHandler interface {
	SizeInBytes() int
}

type interceptedHeartbeatMessageHandler interface {
	interceptedDataSizeHandler
	Message() interface{}
}

type interceptedPeerAuthenticationMessageHandler interface {
	interceptedDataSizeHandler
	Message() interface{}
	Payload() []byte
	Pubkey() []byte
}

type interceptedValidatorInfo interface {
	Hash() []byte
	ValidatorInfo() *state.ShardValidatorInfo
}

type interceptedEquivalentProof interface {
	Hash() []byte
	GetProof() data.HeaderProofHandler
}

// EquivalentProofsPool defines the behaviour of a proofs pool components
type EquivalentProofsPool interface {
	AddNotarizedProof(headerProof data.HeaderProofHandler)
	GetNotarizedProof(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error)
	IsInterfaceNil() bool
}
