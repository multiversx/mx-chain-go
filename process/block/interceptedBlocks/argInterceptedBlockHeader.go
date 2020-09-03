package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgInterceptedBlockHeader is the argument for the intercepted header
type ArgInterceptedBlockHeader struct {
	HdrBuff                 []byte
	Marshalizer             marshal.Marshalizer
	Hasher                  hashing.Hasher
	ShardCoordinator        sharding.Coordinator
	HeaderSigVerifier       process.InterceptedHeaderSigVerifier
	HeaderIntegrityVerifier process.HeaderIntegrityVerifier
	ValidityAttester        process.ValidityAttester
	EpochStartTrigger       process.EpochStartTriggerHandler
}
