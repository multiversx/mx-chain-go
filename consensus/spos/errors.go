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

// ErrNilConsensusState is raised when a valid consensus is expected but nil used
var ErrNilConsensusState = errors.New("consensus state is nil")

// ErrNilConsensusCore is raised when a valid ConsensusCore is expected but nil used
var ErrNilConsensusCore = errors.New("consensus core is nil")

// ErrNilBlockChain is raised when a valid blockchain is expected but nil used
var ErrNilBlockChain = errors.New("blockchain is nil")

// ErrNilHasher is raised when a valid hasher is expected but nil used
var ErrNilHasher = errors.New("hasher is nil")

// ErrNilMarshalizer is raised when a valid marshalizer is expected but nil used
var ErrNilMarshalizer = errors.New("marshalizer is nil")

// ErrNilBlockProcessor is raised when a valid block processor is expected but nil used
var ErrNilBlockProcessor = errors.New("block processor is nil")

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

// ErrNilWorker is raised when a valid worker is expected but nil used
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
