package bn

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

const (
	// SrStartRound defines ID of subround "Start round"
	SrStartRound = iota
	// SrBlock defines ID of subround "block"
	SrBlock
	// SrCommitmentHash defines ID of subround "commitment hash"
	SrCommitmentHash
	// SrBitmap defines ID of subround "bitmap"
	SrBitmap
	// SrCommitment defines ID of subround "commitment"
	SrCommitment
	// SrSignature defines ID of subround "signature"
	SrSignature
	// SrEndRound defines ID of subround "End round"
	SrEndRound
)

// syncThesholdPercent sepcifies the max allocated time to syncronize as a percentage of the total time of the round
const syncThresholdPercent = 50

// processingThresholdPercent specifies the max allocated time for processing the block as a percentage of the total time of the round
const processingThresholdPercent = 65

// maxThresholdPercent specifies the max allocated time percent for doing job as a percentage of the total time of one round
const maxThresholdPercent = 75

// srStartStartTime specifies the start time, from the total time of the round, of subround Start
const srStartStartTime = 0.0

// srEndStartTime specifies the end time, from the total time of the round, of subround Start
const srStartEndTime = 0.05

// srBlockStartTime specifies the start time, from the total time of the round, of subround Block
const srBlockStartTime = 0.05

// srBlockEndTime specifies the end time, from the total time of the round, of subround Block
const srBlockEndTime = 0.25

// srCommitmentHashStartTime specifies the start time, from the total time of the round, of subround CommitmentHash
const srCommitmentHashStartTime = 0.25

// srCommitmentHashEndTime specifies the end time, from the total time of the round, of subround CommitmentHash
const srCommitmentHashEndTime = 0.35

// srBitmapStartTime specifies the start time, from the total time of the round, of subround Bitmap
const srBitmapStartTime = 0.35

// srBitmapEndTime specifies the end time, from the total time of the round, of subround Bitmap
const srBitmapEndTime = 0.45

// srCommitmentStartTime specifies the start time, from the total time of the round, of subround Commitment
const srCommitmentStartTime = 0.45

// srCommitmentEndTime specifies the end time, from the total time of the round, of subround Commitment
const srCommitmentEndTime = 0.55

// srSignatureStartTime specifies the start time, from the total time of the round, of subround Signature
const srSignatureStartTime = 0.55

// srSignatureEndTime specifies the end time, from the total time of the round, of subround Signature
const srSignatureEndTime = 0.65

// srEndStartTime specifies the start time, from the total time of the round, of subround End
const srEndStartTime = 0.65

// srEndEndTime specifies the end time, from the total time of the round, of subround End
const srEndEndTime = 0.75

// MessageType specifies what type of message was received
type MessageType int

func (msgType MessageType) String() string {
	switch msgType {
	case MtBlockBody:
		return "(BLOCK_BODY)"
	case MtBlockHeader:
		return "(BLOCK_HEADER)"
	case MtCommitmentHash:
		return "(COMMITMENT_HASH)"
	case MtBitmap:
		return "(BITMAP)"
	case MtCommitment:
		return "(COMMITMENT)"
	case MtSignature:
		return "(SIGNATURE)"
	case MtUnknown:
		return "(UNKNOWN)"
	default:
		return "Undefined message type"
	}
}

const (
	// MtUnknown defines ID of a message that has unknown Data inside
	MtUnknown MessageType = iota
	// MtBlockBody defines ID of a message that has a block body inside
	MtBlockBody
	// MtBlockHeader defines ID of a message that has a block header inside
	MtBlockHeader
	// MtCommitmentHash defines ID of a message that has a commitment hash inside
	MtCommitmentHash
	// MtBitmap defines ID of a message that has a bitmap inside
	MtBitmap
	// MtCommitment defines ID of a message that has a commitment inside
	MtCommitment
	// MtSignature defines ID of a message that has a Signature inside
	MtSignature
)

// factory defines the data needed by this factory to create all the subrounds and give them their specific
// functionality
type factory struct {
	blockChain             data.ChainHandler
	blockProcessor         process.BlockProcessor
	bootstraper            process.Bootstrapper
	chronologyHandler      consensus.ChronologyHandler
	consensusState         *spos.ConsensusState
	hasher                 hashing.Hasher
	marshalizer            marshal.Marshalizer
	multiSigner            crypto.MultiSigner
	rounder                consensus.Rounder
	shardCoordinator       sharding.Coordinator
	syncTimer              ntp.SyncTimer
	validatorGroupSelector consensus.ValidatorGroupSelector
	worker                 *worker
}

// NewFactory creates a new consensusState object
func NewFactory(
	blockChain data.ChainHandler,
	blockProcessor process.BlockProcessor,
	bootstraper process.Bootstrapper,
	chronologyHandler consensus.ChronologyHandler,
	consensusState *spos.ConsensusState,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	multiSigner crypto.MultiSigner,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	syncTimer ntp.SyncTimer,
	validatorGroupSelector consensus.ValidatorGroupSelector,
	worker *worker,
) (*factory, error) {

	err := checkNewFactoryParams(
		blockChain,
		blockProcessor,
		bootstraper,
		chronologyHandler,
		consensusState,
		hasher,
		marshalizer,
		multiSigner,
		rounder,
		shardCoordinator,
		syncTimer,
		validatorGroupSelector,
		worker,
	)

	if err != nil {
		return nil, err
	}

	fct := factory{
		blockChain:             blockChain,
		blockProcessor:         blockProcessor,
		bootstraper:            bootstraper,
		chronologyHandler:      chronologyHandler,
		consensusState:         consensusState,
		hasher:                 hasher,
		marshalizer:            marshalizer,
		multiSigner:            multiSigner,
		rounder:                rounder,
		shardCoordinator:       shardCoordinator,
		syncTimer:              syncTimer,
		validatorGroupSelector: validatorGroupSelector,
		worker:                 worker,
	}

	return &fct, nil
}

func checkNewFactoryParams(
	blockChain data.ChainHandler,
	blockProcessor process.BlockProcessor,
	bootstraper process.Bootstrapper,
	chronologyHandler consensus.ChronologyHandler,
	consensusState *spos.ConsensusState,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	multiSigner crypto.MultiSigner,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	syncTimer ntp.SyncTimer,
	validatorGroupSelector consensus.ValidatorGroupSelector,
	worker *worker,
) error {
	if blockChain == nil {
		return spos.ErrNilBlockChain
	}

	if blockProcessor == nil {
		return spos.ErrNilBlockProcessor
	}

	if bootstraper == nil {
		return spos.ErrNilBlootstraper
	}

	if chronologyHandler == nil {
		return spos.ErrNilChronologyHandler
	}

	if consensusState == nil {
		return spos.ErrNilConsensusState
	}

	if hasher == nil {
		return spos.ErrNilHasher
	}

	if marshalizer == nil {
		return spos.ErrNilMarshalizer
	}

	if multiSigner == nil {
		return spos.ErrNilMultiSigner
	}

	if rounder == nil {
		return spos.ErrNilRounder
	}

	if shardCoordinator == nil {
		return spos.ErrNilShardCoordinator
	}

	if syncTimer == nil {
		return spos.ErrNilSyncTimer
	}

	if validatorGroupSelector == nil {
		return spos.ErrNilValidatorGroupSelector
	}

	if worker == nil {
		return spos.ErrNilWorker
	}

	return nil
}

// GenerateSubrounds will generate the subrounds used in Belare & Naveen Cns
func (fct *factory) GenerateSubrounds() error {
	fct.initConsensusThreshold()
	fct.chronologyHandler.RemoveAllSubrounds()
	fct.worker.RemoveAllReceivedMessagesCalls()

	err := fct.generateStartRoundSubround()

	if err != nil {
		return err
	}

	err = fct.generateBlockSubround()

	if err != nil {
		return err
	}

	err = fct.generateCommitmentHashSubround()

	if err != nil {
		return err
	}

	err = fct.generateBitmapSubround()

	if err != nil {
		return err
	}

	err = fct.generateCommitmentSubround()

	if err != nil {
		return err
	}

	err = fct.generateSignatureSubround()

	if err != nil {
		return err
	}

	err = fct.generateEndRoundSubround()

	if err != nil {
		return err
	}

	return nil
}

func (fct *factory) generateStartRoundSubround() error {
	subround, err := NewSubround(
		-1,
		SrStartRound,
		SrBlock,
		int64(float64(fct.rounder.TimeDuration())*srStartStartTime),
		int64(float64(fct.rounder.TimeDuration())*srStartEndTime),
		getSubroundName(SrStartRound),
		fct.worker.consensusStateChangedChannels,
	)

	if err != nil {
		return err
	}

	subroundStartRound, err := NewSubroundStartRound(
		subround,
		fct.blockChain,
		fct.bootstraper,
		fct.consensusState,
		fct.multiSigner,
		fct.rounder,
		fct.syncTimer,
		fct.validatorGroupSelector,
		fct.worker.extend,
	)

	if err != nil {
		return err
	}

	fct.chronologyHandler.AddSubround(subroundStartRound)

	return nil
}

func (fct *factory) generateBlockSubround() error {

	subround, err := NewSubround(
		SrStartRound,
		SrBlock,
		SrCommitmentHash,
		int64(float64(fct.rounder.TimeDuration())*srBlockStartTime),
		int64(float64(fct.rounder.TimeDuration())*srBlockEndTime),
		getSubroundName(SrBlock),
		fct.worker.consensusStateChangedChannels,
	)

	if err != nil {
		return err
	}

	subroundBlock, err := NewSubroundBlock(
		subround,
		fct.blockChain,
		fct.blockProcessor,
		fct.consensusState,
		fct.hasher,
		fct.marshalizer,
		fct.multiSigner,
		fct.rounder,
		fct.shardCoordinator,
		fct.syncTimer,
		fct.worker.sendConsensusMessage,
		fct.worker.extend,
	)

	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtBlockBody, subroundBlock.receivedBlockBody)
	fct.worker.AddReceivedMessageCall(MtBlockHeader, subroundBlock.receivedBlockHeader)
	fct.chronologyHandler.AddSubround(subroundBlock)

	return nil
}

func (fct *factory) generateCommitmentHashSubround() error {
	subround, err := NewSubround(
		SrBlock,
		SrCommitmentHash,
		SrBitmap,
		int64(float64(fct.rounder.TimeDuration())*srCommitmentHashStartTime),
		int64(float64(fct.rounder.TimeDuration())*srCommitmentHashEndTime),
		getSubroundName(SrCommitmentHash),
		fct.worker.consensusStateChangedChannels,
	)

	if err != nil {
		return err
	}

	subroundCommitmentHash, err := NewSubroundCommitmentHash(
		subround,
		fct.consensusState,
		fct.hasher,
		fct.multiSigner,
		fct.rounder,
		fct.syncTimer,
		fct.worker.sendConsensusMessage,
		fct.worker.extend,
	)

	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtCommitmentHash, subroundCommitmentHash.receivedCommitmentHash)
	fct.chronologyHandler.AddSubround(subroundCommitmentHash)

	return nil
}

func (fct *factory) generateBitmapSubround() error {
	subround, err := NewSubround(
		SrCommitmentHash,
		SrBitmap,
		SrCommitment,
		int64(float64(fct.rounder.TimeDuration())*srBitmapStartTime),
		int64(float64(fct.rounder.TimeDuration())*srBitmapEndTime),
		getSubroundName(SrBitmap),
		fct.worker.consensusStateChangedChannels,
	)

	if err != nil {
		return err
	}

	subroundBitmap, err := NewSubroundBitmap(
		subround,
		fct.blockProcessor,
		fct.consensusState,
		fct.rounder,
		fct.syncTimer,
		fct.worker.sendConsensusMessage,
		fct.worker.extend,
	)

	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtBitmap, subroundBitmap.receivedBitmap)
	fct.chronologyHandler.AddSubround(subroundBitmap)

	return nil
}

func (fct *factory) generateCommitmentSubround() error {
	subround, err := NewSubround(
		SrBitmap,
		SrCommitment,
		SrSignature,
		int64(float64(fct.rounder.TimeDuration())*srCommitmentStartTime),
		int64(float64(fct.rounder.TimeDuration())*srCommitmentEndTime),
		getSubroundName(SrCommitment),
		fct.worker.consensusStateChangedChannels,
	)

	if err != nil {
		return err
	}

	subroundCommitment, err := NewSubroundCommitment(
		subround,
		fct.consensusState,
		fct.multiSigner,
		fct.rounder,
		fct.syncTimer,
		fct.worker.sendConsensusMessage,
		fct.worker.extend,
	)

	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtCommitment, subroundCommitment.receivedCommitment)
	fct.chronologyHandler.AddSubround(subroundCommitment)

	return nil
}

func (fct *factory) generateSignatureSubround() error {
	subround, err := NewSubround(
		SrCommitment,
		SrSignature,
		SrEndRound,
		int64(float64(fct.rounder.TimeDuration())*srSignatureStartTime),
		int64(float64(fct.rounder.TimeDuration())*srSignatureEndTime),
		getSubroundName(SrSignature),
		fct.worker.consensusStateChangedChannels,
	)

	if err != nil {
		return err
	}

	subroundSignature, err := NewSubroundSignature(
		subround,
		fct.consensusState,
		fct.hasher,
		fct.multiSigner,
		fct.rounder,
		fct.syncTimer,
		fct.worker.sendConsensusMessage,
		fct.worker.extend,
	)

	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtSignature, subroundSignature.receivedSignature)
	fct.chronologyHandler.AddSubround(subroundSignature)

	return nil
}

func (fct *factory) generateEndRoundSubround() error {
	subround, err := NewSubround(
		SrSignature,
		SrEndRound,
		-1,
		int64(float64(fct.rounder.TimeDuration())*srEndStartTime),
		int64(float64(fct.rounder.TimeDuration())*srEndEndTime),
		getSubroundName(SrEndRound),
		fct.worker.consensusStateChangedChannels,
	)

	if err != nil {
		return err
	}

	subroundEndRound, err := NewSubroundEndRound(
		subround,
		fct.blockChain,
		fct.blockProcessor,
		fct.consensusState,
		fct.multiSigner,
		fct.rounder,
		fct.syncTimer,
		fct.worker.BroadcastBlock,
		fct.worker.extend,
	)

	if err != nil {
		return err
	}

	fct.chronologyHandler.AddSubround(subroundEndRound)

	return nil
}

func (fct *factory) initConsensusThreshold() {
	pbftThreshold := fct.consensusState.ConsensusGroupSize()*2/3 + 1

	fct.consensusState.SetThreshold(SrBlock, 1)
	fct.consensusState.SetThreshold(SrCommitmentHash, pbftThreshold)
	fct.consensusState.SetThreshold(SrBitmap, pbftThreshold)
	fct.consensusState.SetThreshold(SrCommitment, pbftThreshold)
	fct.consensusState.SetThreshold(SrSignature, pbftThreshold)
}
