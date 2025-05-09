package interceptedBlocks

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/sync"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-vm-v1_2-go/ipc/marshaling"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

const interceptedEquivalentProofType = "intercepted equivalent proof"

// ArgInterceptedEquivalentProof is the argument used in the intercepted equivalent proof constructor
type ArgInterceptedEquivalentProof struct {
	DataBuff          []byte
	Marshaller        marshal.Marshalizer
	Hasher            hashing.Hasher
	ShardCoordinator  sharding.Coordinator
	HeaderSigVerifier consensus.HeaderSigVerifier
	Proofs            dataRetriever.ProofsPool
	ProofSizeChecker  common.FieldsSizeChecker
	KeyRWMutexHandler sync.KeyRWMutexHandler
}

type interceptedEquivalentProof struct {
	proof             *block.HeaderProof
	isForCurrentShard bool
	headerSigVerifier consensus.HeaderSigVerifier
	proofsPool        dataRetriever.ProofsPool
	marshaller        marshaling.Marshalizer
	hasher            hashing.Hasher
	hash              []byte
	proofSizeChecker  common.FieldsSizeChecker
	km                sync.KeyRWMutexHandler
}

// NewInterceptedEquivalentProof returns a new instance of interceptedEquivalentProof
func NewInterceptedEquivalentProof(args ArgInterceptedEquivalentProof) (*interceptedEquivalentProof, error) {
	err := checkArgInterceptedEquivalentProof(args)
	if err != nil {
		return nil, err
	}

	equivalentProof, err := createEquivalentProof(args.Marshaller, args.DataBuff)
	if err != nil {
		return nil, err
	}

	hash := args.Hasher.Compute(string(args.DataBuff))

	return &interceptedEquivalentProof{
		proof:             equivalentProof,
		isForCurrentShard: extractIsForCurrentShard(args.ShardCoordinator, equivalentProof),
		headerSigVerifier: args.HeaderSigVerifier,
		proofsPool:        args.Proofs,
		marshaller:        args.Marshaller,
		hasher:            args.Hasher,
		proofSizeChecker:  args.ProofSizeChecker,
		hash:              hash,
		km:                args.KeyRWMutexHandler,
	}, nil
}

func checkArgInterceptedEquivalentProof(args ArgInterceptedEquivalentProof) error {
	if len(args.DataBuff) == 0 {
		return process.ErrNilBuffer
	}
	if check.IfNil(args.Marshaller) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(args.HeaderSigVerifier) {
		return process.ErrNilHeaderSigVerifier
	}
	if check.IfNil(args.Proofs) {
		return process.ErrNilProofsPool
	}
	if check.IfNil(args.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(args.ProofSizeChecker) {
		return errors.ErrNilFieldsSizeChecker
	}
	if check.IfNil(args.KeyRWMutexHandler) {
		return process.ErrNilKeyRWMutexHandler
	}

	return nil
}

func createEquivalentProof(marshaller marshal.Marshalizer, buff []byte) (*block.HeaderProof, error) {
	headerProof := &block.HeaderProof{}
	err := marshaller.Unmarshal(headerProof, buff)
	if err != nil {
		return nil, err
	}

	log.Trace("interceptedEquivalentProof successfully created",
		"header hash", logger.DisplayByteSlice(headerProof.HeaderHash),
		"header shard", headerProof.HeaderShardId,
		"header epoch", headerProof.HeaderEpoch,
		"header nonce", headerProof.HeaderNonce,
		"header round", headerProof.HeaderRound,
		"bitmap", logger.DisplayByteSlice(headerProof.PubKeysBitmap),
		"signature", logger.DisplayByteSlice(headerProof.AggregatedSignature),
		"isEpochStart", headerProof.IsStartOfEpoch,
	)

	return headerProof, nil
}

func extractIsForCurrentShard(shardCoordinator sharding.Coordinator, equivalentProof *block.HeaderProof) bool {
	proofShardId := equivalentProof.GetHeaderShardId()
	if shardCoordinator.SelfId() == core.MetachainShardId {
		return true
	}

	if proofShardId == core.MetachainShardId {
		return true
	}

	return proofShardId == shardCoordinator.SelfId()
}

// CheckValidity checks if the received proof is valid
func (iep *interceptedEquivalentProof) CheckValidity() error {
	log.Debug("Checking intercepted equivalent proof validity", "proof header hash", iep.proof.HeaderHash)
	err := iep.integrity()
	if err != nil {
		return err
	}

	headerHash := string(iep.proof.GetHeaderHash())
	iep.km.Lock(headerHash)
	defer iep.km.Unlock(headerHash)

	ok := iep.proofsPool.HasProof(iep.proof.GetHeaderShardId(), iep.proof.GetHeaderHash())
	if ok {
		return common.ErrAlreadyExistingEquivalentProof
	}

	err = iep.headerSigVerifier.VerifyHeaderProof(iep.proof)
	if err != nil {
		return err
	}

	// also save the proof here in order to complete the flow under mutex lock
	wasAdded := iep.proofsPool.AddProof(iep.proof)
	if !wasAdded {
		// with the current implementation, this should never happen
		return common.ErrAlreadyExistingEquivalentProof
	}

	return nil
}

func (iep *interceptedEquivalentProof) integrity() error {
	if !iep.proofSizeChecker.IsProofSizeValid(iep.proof) {
		return ErrInvalidProof
	}

	return nil
}

// GetProof returns the underlying intercepted header proof
func (iep *interceptedEquivalentProof) GetProof() data.HeaderProofHandler {
	return iep.proof
}

// IsForCurrentShard returns true if the equivalent proof should be processed by the current shard
func (iep *interceptedEquivalentProof) IsForCurrentShard() bool {
	return iep.isForCurrentShard
}

// Hash returns the header hash the proof belongs to
func (iep *interceptedEquivalentProof) Hash() []byte {
	return iep.hash
}

// Type returns the type of this intercepted data
func (iep *interceptedEquivalentProof) Type() string {
	return interceptedEquivalentProofType
}

// Identifiers returns the identifiers used in requests
func (iep *interceptedEquivalentProof) Identifiers() [][]byte {
	return [][]byte{
		iep.proof.HeaderHash,
		// needed for the interceptor, when data is requested by nonce
		[]byte(common.GetEquivalentProofNonceShardKey(iep.proof.HeaderNonce, iep.proof.HeaderShardId)),
	}
}

// String returns the proof's most important fields as string
func (iep *interceptedEquivalentProof) String() string {
	return fmt.Sprintf("bitmap=%s, signature=%s, hash=%s, epoch=%d, shard=%d, nonce=%d, round=%d, isEpochStart=%t",
		logger.DisplayByteSlice(iep.proof.PubKeysBitmap),
		logger.DisplayByteSlice(iep.proof.AggregatedSignature),
		logger.DisplayByteSlice(iep.proof.HeaderHash),
		iep.proof.HeaderEpoch,
		iep.proof.HeaderShardId,
		iep.proof.HeaderNonce,
		iep.proof.HeaderRound,
		iep.proof.IsStartOfEpoch,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (iep *interceptedEquivalentProof) IsInterfaceNil() bool {
	return iep == nil
}
