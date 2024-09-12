package interceptedBlocks

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const interceptedEquivalentProofType = "intercepted equivalent proof"

// ArgInterceptedEquivalentProof is the argument used in the intercepted equivalent proof constructor
type ArgInterceptedEquivalentProof struct {
	DataBuff          []byte
	Marshaller        marshal.Marshalizer
	ShardCoordinator  sharding.Coordinator
	HeaderSigVerifier consensus.HeaderSigVerifier
	Headers           dataRetriever.HeadersPool
}

type interceptedEquivalentProof struct {
	proof             *block.HeaderProof
	isForCurrentShard bool
	headerSigVerifier consensus.HeaderSigVerifier
	headers           dataRetriever.HeadersPool
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

	return &interceptedEquivalentProof{
		proof:             equivalentProof,
		isForCurrentShard: extractIsForCurrentShard(args.ShardCoordinator, equivalentProof),
		headerSigVerifier: args.HeaderSigVerifier,
		headers:           args.Headers,
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
	if check.IfNil(args.Headers) {
		return process.ErrNilHeadersDataPool
	}

	return nil
}

func createEquivalentProof(marshaller marshal.Marshalizer, buff []byte) (*block.HeaderProof, error) {
	headerProof := &block.HeaderProof{}
	err := marshaller.Unmarshal(headerProof, buff)
	if err != nil {
		return nil, err
	}

	log.Trace("interceptedEquivalentProof successfully created")

	return headerProof, nil
}

func extractIsForCurrentShard(shardCoordinator sharding.Coordinator, equivalentProof *block.HeaderProof) bool {
	proofShardId := equivalentProof.GetHeaderShardId()
	if proofShardId == core.MetachainShardId {
		return true
	}

	return proofShardId == shardCoordinator.SelfId()
}

// CheckValidity checks if the received proof is valid
func (iep *interceptedEquivalentProof) CheckValidity() error {
	err := iep.integrity()
	if err != nil {
		return err
	}

	hdr, err := iep.headers.GetHeaderByHash(iep.proof.HeaderHash)
	if err != nil {
		return err
	}

	return iep.headerSigVerifier.VerifySignatureForHash(hdr, iep.proof.HeaderHash, iep.proof.PubKeysBitmap, iep.proof.AggregatedSignature)
}

func (iep *interceptedEquivalentProof) integrity() error {
	isProofValid := len(iep.proof.AggregatedSignature) > 0 &&
		len(iep.proof.PubKeysBitmap) > 0 &&
		len(iep.proof.HeaderHash) > 0
	if !isProofValid {
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
	return iep.proof.HeaderHash
}

// Type returns the type of this intercepted data
func (iep *interceptedEquivalentProof) Type() string {
	return interceptedEquivalentProofType
}

// Identifiers returns the identifiers used in requests
func (iep *interceptedEquivalentProof) Identifiers() [][]byte {
	return [][]byte{iep.proof.HeaderHash}
}

// String returns the proof's most important fields as string
func (iep *interceptedEquivalentProof) String() string {
	return fmt.Sprintf("bitmap=%s, signature=%s, hash=%s, epoch=%d, shard=%d, nonce=%d",
		logger.DisplayByteSlice(iep.proof.PubKeysBitmap),
		logger.DisplayByteSlice(iep.proof.AggregatedSignature),
		logger.DisplayByteSlice(iep.proof.HeaderHash),
		iep.proof.HeaderEpoch,
		iep.proof.HeaderShardId,
		iep.proof.HeaderNonce,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (iep *interceptedEquivalentProof) IsInterfaceNil() bool {
	return iep == nil
}
