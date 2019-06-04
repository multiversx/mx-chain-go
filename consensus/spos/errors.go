package spos

import (
	"errors"
)

// ErrNilConsensusGroup is raised when an operation is attempted with a nil consensus group
var ErrNilConsensusGroup = errors.New("consensusGroup is null")

// ErrEmptyConsensusGroup is raised when an operation is attempted with an empty consensus group
var ErrEmptyConsensusGroup = errors.New("consensusGroup is empty")

// ErrNotFoundInConsensus is raised when self expected in consensus group but not found
var ErrNotFoundInConsensus = errors.New("self not found in consensus group")

// ErrNilPublicKey is raised when a valid public key was expected but nil was used
var ErrNilPublicKey = errors.New("public key is nil")

// ErrNilPrivateKey is raised when a valid private key was expected but nil was used
var ErrNilPrivateKey = errors.New("private key is nil")

// ErrNilConsensusData is raised when valid consensus data was expected but nil was received
var ErrNilConsensusData = errors.New("consensus data is nil")

// ErrNilSignature is raised when a valid signature was expected but nil was used
var ErrNilSignature = errors.New("signature is nil")

// ErrNilCommitment is raised when a valid commitment was expected but nil was used
var ErrNilCommitment = errors.New("commitment is nil")

// ErrNilKeyGenerator is raised when a valid key generator is expected but nil was used
var ErrNilKeyGenerator = errors.New("key generator is nil")

// ErrNilSingleSigner is raised when a valid singleSigner is expected but nil used
var ErrNilSingleSigner = errors.New("singleSigner is nil")

// ErrNilMultiSigner is raised when a valid multiSigner is expected but nil used
var ErrNilMultiSigner = errors.New("multiSigner is nil")

// ErrInvalidMultiSigner is raised when an invalid multiSigner is used
var ErrInvalidMultiSigner = errors.New("multiSigner is invalid")

// ErrNilConsensusState is raised when a valid consensus is expected but nil used
var ErrNilConsensusState = errors.New("consensus state is nil")

// ErrNilConsensusCore is raised when a valid ConsensusCore is expected but nil used
var ErrNilConsensusCore = errors.New("consensus core is nil")

// ErrNilConsensusService is raised when a valid ConsensusService is expected but nil used
var ErrNilConsensusService = errors.New("consensus service is nil")

// ErrNilBlockChain is raised when a valid blockchain is expected but nil used
var ErrNilBlockChain = errors.New("blockchain is nil")

// ErrNilHasher is raised when a valid hasher is expected but nil used
var ErrNilHasher = errors.New("hasher is nil")

// ErrNilMarshalizer is raised when a valid marshalizer is expected but nil used
var ErrNilMarshalizer = errors.New("marshalizer is nil")

// ErrNilBlockProcessor is raised when a valid block processor is expected but nil used
var ErrNilBlockProcessor = errors.New("block processor is nil")

// ErrNilBlockTracker is raised when a valid block tracker is expected but nil used
var ErrNilBlockTracker = errors.New("block tracker is nil")

// ErrNilBlootstraper is raised when a valid block processor is expected but nil used
var ErrNilBlootstraper = errors.New("boostraper is nil")

// ErrInvalidKey is raised when an invalid key is used with a map
var ErrInvalidKey = errors.New("map key is invalid")

// ErrNilRoundState is raised when a valid round state is expected but nil used
var ErrNilRoundState = errors.New("round state is nil")

// ErrCommitmentHashDoesNotMatch is raised when the commitment hash does not match expected value
var ErrCommitmentHashDoesNotMatch = errors.New("commitment hash does not match")

// ErrNilMessage signals that a nil message has been received
var ErrNilMessage = errors.New("nil message")

// ErrNilDataToProcess signals that nil data was provided
var ErrNilDataToProcess = errors.New("nil data to process")

// ErrNilWorker is raised when a valid Worker is expected but nil used
var ErrNilWorker = errors.New("worker is nil")

// ErrNilShardCoordinator is raised when a valid shard coordinator is expected but nil used
var ErrNilShardCoordinator = errors.New("shard coordinator is nil")

// ErrNilValidatorGroupSelector is raised when a valid validator group selector is expected but nil used
var ErrNilValidatorGroupSelector = errors.New("validator group selector is nil")

// ErrNilChronologyHandler is raised when a valid chronology handler is expected but nil used
var ErrNilChronologyHandler = errors.New("chronology handler is nil")

// ErrNilRounder is raised when a valid rounder is expected but nil used
var ErrNilRounder = errors.New("rounder is nil")

// ErrNilSyncTimer is raised when a valid sync timer is expected but nil used
var ErrNilSyncTimer = errors.New("sync timer is nil")

// ErrNilSubround is raised when a valid subround is expected but nil used
var ErrNilSubround = errors.New("subround is nil")

// ErrNilSendConsensusMessageFunction is raised when a valid send consensus message function is expected but nil used
var ErrNilSendConsensusMessageFunction = errors.New("send consnensus message function is nil")

// ErrNilBroadcastBlockFunction is raised when a valid broadcast block function is expected but nil used
var ErrNilBroadcastBlockFunction = errors.New("broadcast block function is nil")

// ErrNilChannel is raised when a valid channel is expected but nil used
var ErrNilChannel = errors.New("channel is nil")

// ErrRoundCanceled is raised when round is canceled
var ErrRoundCanceled = errors.New("round is canceled")

// ErrSenderNotOk is raised when sender is invalid
var ErrSenderNotOk = errors.New("sender is invalid")

// ErrMessageForPastRound is raised when message is for past round
var ErrMessageForPastRound = errors.New("message is for past round")

// ErrInvalidSignature is raised when signature is invalid
var ErrInvalidSignature = errors.New("signature is invalid")

// ErrMessageFromItself is raised when a message from itself is received
var ErrMessageFromItself = errors.New("message is from itself")

// ErrNilBlsPrivateKey is raised when the bls private key is nil
var ErrNilBlsPrivateKey = errors.New("BLS private key should not be nil")

// ErrNilBlsSingleSigner is raised when a message from itself is received
var ErrNilBlsSingleSigner = errors.New("BLS single signer should not be nil")

// ErrNilHeader is raised when an expected header is nil
var ErrNilHeader = errors.New("header is nil")

// ErrNilBroadcastBlock is raised when a valid broadcastBlock function is expected but nil used
var ErrNilBroadcastBlock = errors.New("broadcastBlock is nil")

// ErrNilBroadcastHeader is raised when a valid broadcastHeader function is expected but nil used
var ErrNilBroadcastHeader = errors.New("broadcastHeader is nil")

// ErrNilBroadcastUnnotarisedBlocks is raised when a valid broadcastUnnotarisedBlocks function is expected but nil used
var ErrNilBroadcastUnnotarisedBlocks = errors.New("broadcastUnnotarisedBlocks is nil")

// ErrNilSendMessage is raised when a valid sendMessage function is expected but nil used
var ErrNilSendMessage = errors.New("sendMessage is nil")

// ErrNilForkDetector is raised when a valid fork detector is expected but nil used
var ErrNilForkDetector = errors.New("fork detector is nil")
