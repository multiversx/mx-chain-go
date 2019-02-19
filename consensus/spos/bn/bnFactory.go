package bn

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/validators/groupSelectors"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
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

//TODO: maximum transactions in one block (this should be injected, and this const should be removed later)
const maxTransactionsInBlock = 15000

// consensusSubrounds specifies how many subrounds of consensus are in this implementation
const consensusSubrounds = 6

// maxBlockProcessingTimePercent specifies which is the max allocated time percent,
// for processing block, from the total time of one round
const maxBlockProcessingTimePercent = float64(0.85)

// MessageType specifies what type of message was received
type MessageType int

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

// Factory defines the data needed by this factory to create all the subrounds and give them their specific
// functionality
type Factory struct {
	blockChain             *blockchain.BlockChain
	blockProcessor         process.BlockProcessor
	bootstraper            process.Bootstraper
	chronologyHandler      chronology.ChronologyHandler
	consensusState         *spos.ConsensusState
	hasher                 hashing.Hasher
	marshalizer            marshal.Marshalizer
	multiSigner            crypto.MultiSigner
	rounder                round.Rounder
	shardCoordinator       sharding.ShardCoordinator
	syncTimer              ntp.SyncTimer
	validatorGroupSelector groupSelectors.ValidatorGroupSelector
	worker                 *Worker
}

// NewFactory creates a new ConsensusState object
func NewFactory(
	blockChain *blockchain.BlockChain,
	blockProcessor process.BlockProcessor,
	bootstraper process.Bootstraper,
	chronologyHandler chronology.ChronologyHandler,
	consensusState *spos.ConsensusState,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	multiSigner crypto.MultiSigner,
	rounder round.Rounder,
	shardCoordinator sharding.ShardCoordinator,
	syncTimer ntp.SyncTimer,
	validatorGroupSelector groupSelectors.ValidatorGroupSelector,
	worker *Worker,
) (*Factory, error) {

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

	fct := Factory{
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
	blockChain *blockchain.BlockChain,
	blockProcessor process.BlockProcessor,
	bootstraper process.Bootstraper,
	chronologyHandler chronology.ChronologyHandler,
	consensusState *spos.ConsensusState,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	multiSigner crypto.MultiSigner,
	rounder round.Rounder,
	shardCoordinator sharding.ShardCoordinator,
	syncTimer ntp.SyncTimer,
	validatorGroupSelector groupSelectors.ValidatorGroupSelector,
	worker *Worker,
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
func (fct *Factory) GenerateSubrounds() error {
	fct.initConsensusThreshold()
	fct.chronologyHandler.RemoveAllSubrounds()
	fct.worker.RemoveAllReceivedMessagesCalls()

	fct.generateStartRoundSubround()
	fct.generateBlockSubround()
	fct.generateCommitmentHashSubround()
	fct.generateBitmapSubround()
	fct.generateCommitmentSubround()
	fct.generateSignatureSubround()
	fct.generateEndRoundSubround()

	return nil
}

func (fct *Factory) generateStartRoundSubround() bool {
	subround, err := NewSubround(
		-1,
		SrStartRound,
		SrBlock,
		int64(fct.rounder.TimeDuration()*0/100),
		int64(fct.rounder.TimeDuration()*5/100),
		getSubroundName(SrStartRound),
		fct.worker.ConsensusStateChangedChannel,
	)

	if err != nil {
		log.Error(err.Error())
		return false
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
		log.Error(err.Error())
		return false
	}

	fct.chronologyHandler.AddSubround(subroundStartRound)

	return true
}

func (fct *Factory) generateBlockSubround() bool {

	subround, err := NewSubround(
		SrStartRound,
		SrBlock,
		SrCommitmentHash,
		int64(fct.rounder.TimeDuration()*5/100),
		int64(fct.rounder.TimeDuration()*25/100),
		getSubroundName(SrBlock),
		fct.worker.ConsensusStateChangedChannel,
	)

	if err != nil {
		log.Error(err.Error())
		return false
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
		log.Error(err.Error())
		return false
	}

	fct.worker.AddReceivedMessageCall(MtBlockBody, subroundBlock.receivedBlockBody)
	fct.worker.AddReceivedMessageCall(MtBlockHeader, subroundBlock.receivedBlockHeader)
	fct.chronologyHandler.AddSubround(subroundBlock)

	return true
}

func (fct *Factory) generateCommitmentHashSubround() bool {
	subround, err := NewSubround(
		SrBlock,
		SrCommitmentHash,
		SrBitmap,
		int64(fct.rounder.TimeDuration()*25/100),
		int64(fct.rounder.TimeDuration()*40/100),
		getSubroundName(SrCommitmentHash),
		fct.worker.ConsensusStateChangedChannel,
	)

	if err != nil {
		log.Error(err.Error())
		return false
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
		log.Error(err.Error())
		return false
	}

	fct.worker.AddReceivedMessageCall(MtCommitmentHash, subroundCommitmentHash.receivedCommitmentHash)
	fct.chronologyHandler.AddSubround(subroundCommitmentHash)

	return true
}

func (fct *Factory) generateBitmapSubround() bool {
	subround, err := NewSubround(
		SrCommitmentHash,
		SrBitmap,
		SrCommitment,
		int64(fct.rounder.TimeDuration()*40/100),
		int64(fct.rounder.TimeDuration()*55/100),
		getSubroundName(SrBitmap),
		fct.worker.ConsensusStateChangedChannel,
	)

	if err != nil {
		log.Error(err.Error())
		return false
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
		log.Error(err.Error())
		return false
	}

	fct.worker.AddReceivedMessageCall(MtBitmap, subroundBitmap.receivedBitmap)
	fct.chronologyHandler.AddSubround(subroundBitmap)

	return true
}

func (fct *Factory) generateCommitmentSubround() bool {
	subround, err := NewSubround(
		SrBitmap,
		SrCommitment,
		SrSignature,
		int64(fct.rounder.TimeDuration()*55/100),
		int64(fct.rounder.TimeDuration()*70/100),
		getSubroundName(SrCommitment),
		fct.worker.ConsensusStateChangedChannel,
	)

	if err != nil {
		log.Error(err.Error())
		return false
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
		log.Error(err.Error())
		return false
	}

	fct.worker.AddReceivedMessageCall(MtCommitment, subroundCommitment.receivedCommitment)
	fct.chronologyHandler.AddSubround(subroundCommitment)

	return true
}

func (fct *Factory) generateSignatureSubround() bool {
	subround, err := NewSubround(
		SrCommitment,
		SrSignature,
		SrEndRound,
		int64(fct.rounder.TimeDuration()*70/100),
		int64(fct.rounder.TimeDuration()*85/100),
		getSubroundName(SrSignature),
		fct.worker.ConsensusStateChangedChannel,
	)

	if err != nil {
		log.Error(err.Error())
		return false
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
		log.Error(err.Error())
		return false
	}

	fct.worker.AddReceivedMessageCall(MtSignature, subroundSignature.receivedSignature)
	fct.chronologyHandler.AddSubround(subroundSignature)

	return true
}

func (fct *Factory) generateEndRoundSubround() bool {
	subround, err := NewSubround(
		SrSignature,
		SrEndRound,
		-1,
		int64(fct.rounder.TimeDuration()*85/100),
		int64(fct.rounder.TimeDuration()*95/100),
		getSubroundName(SrEndRound),
		fct.worker.ConsensusStateChangedChannel,
	)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	subroundEndRound, err := NewSubroundEndRound(
		subround,
		fct.blockChain,
		fct.blockProcessor,
		fct.consensusState,
		fct.multiSigner,
		fct.rounder,
		fct.syncTimer,
		fct.worker.broadcastTxBlockBody,
		fct.worker.broadcastHeader,
		fct.worker.extend,
	)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	fct.chronologyHandler.AddSubround(subroundEndRound)

	return true
}

func (fct *Factory) initConsensusThreshold() {
	pbftThreshold := fct.consensusState.ConsensusGroupSize()*2/3 + 1

	fct.consensusState.SetThreshold(SrBlock, 1)
	fct.consensusState.SetThreshold(SrCommitmentHash, pbftThreshold)
	fct.consensusState.SetThreshold(SrBitmap, pbftThreshold)
	fct.consensusState.SetThreshold(SrCommitment, pbftThreshold)
	fct.consensusState.SetThreshold(SrSignature, pbftThreshold)
}
