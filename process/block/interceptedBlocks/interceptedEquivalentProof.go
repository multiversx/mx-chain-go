package interceptedBlocks

import (
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-vm-v1_2-go/ipc/marshaling"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	proofscache "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/proofsCache"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
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
	Headers           dataRetriever.HeadersPool
	Storage           dataRetriever.StorageService
}

type interceptedEquivalentProof struct {
	proof             *block.HeaderProof
	isForCurrentShard bool
	headerSigVerifier consensus.HeaderSigVerifier
	proofsPool        dataRetriever.ProofsPool
	headersPool       dataRetriever.HeadersPool
	storage           dataRetriever.StorageService
	marshaller        marshaling.Marshalizer
	hasher            hashing.Hasher
	hash              []byte
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
		headersPool:       args.Headers,
		marshaller:        args.Marshaller,
		storage:           args.Storage,
		hasher:            args.Hasher,
		hash:              hash,
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
	if check.IfNil(args.Headers) {
		return process.ErrNilHeadersDataPool
	}
	if check.IfNil(args.Storage) {
		return process.ErrNilStore
	}
	if check.IfNil(args.Hasher) {
		return process.ErrNilHasher
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
	err := iep.integrity()
	if err != nil {
		return err
	}

	ok := iep.proofsPool.HasProof(iep.proof.GetHeaderShardId(), iep.proof.GetHeaderHash())
	if ok {
		return proofscache.ErrAlreadyExistingEquivalentProof
	}

	err = iep.checkHeaderParamsFromProof()
	if err != nil {
		return err
	}

	return iep.headerSigVerifier.VerifyHeaderProof(iep.proof)
}

func (iep *interceptedEquivalentProof) checkHeaderParamsFromProof() error {
	headersStorer, err := iep.getHeadersStorer(iep.proof.GetHeaderShardId())
	if err != nil {
		return err
	}

	header, err := common.GetHeader(iep.proof.GetHeaderHash(), iep.headersPool, headersStorer, iep.marshaller)
	if err != nil {
		return fmt.Errorf("%w while getting header for proof hash %s", err, hex.EncodeToString(iep.proof.GetHeaderHash()))
	}

	return common.VerifyProofAgainstHeader(iep.proof, header)
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

func (iep *interceptedEquivalentProof) getHeadersStorer(shardID uint32) (storage.Storer, error) {
	if shardID == core.MetachainShardId {
		return iep.storage.GetStorer(dataRetriever.MetaBlockUnit)
	}

	return iep.storage.GetStorer(dataRetriever.BlockHeaderUnit)
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
