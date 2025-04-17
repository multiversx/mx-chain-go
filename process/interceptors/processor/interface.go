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

// EquivalentProofsPool defines the behaviour of a proofs pool components
type EquivalentProofsPool interface {
	AddProof(headerProof data.HeaderProofHandler) bool
	CleanupProofsBehindNonce(shardID uint32, nonce uint64) error
	GetProof(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error)
	GetProofByNonce(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error)
	HasProof(shardID uint32, headerHash []byte) bool
	IsInterfaceNil() bool
}
