package common

import (
	"encoding/hex"
	"fmt"
	"math/bits"
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const (
	keySeparator   = "-"
	expectedKeyLen = 2
	hashIndex      = 0
	shardIndex     = 1
	nonceIndex     = 0
)

type chainParametersHandler interface {
	CurrentChainParameters() config.ChainParametersByEpochConfig
	ChainParametersForEpoch(epoch uint32) (config.ChainParametersByEpochConfig, error)
	IsInterfaceNil() bool
}

// IsValidRelayedTxV3 returns true if the provided transaction is a valid transaction of type relayed v3
func IsValidRelayedTxV3(tx data.TransactionHandler) bool {
	relayedTx, isRelayedV3 := tx.(data.RelayedTransactionHandler)
	if !isRelayedV3 {
		return false
	}
	hasValidRelayer := len(relayedTx.GetRelayerAddr()) == len(tx.GetSndAddr()) && len(relayedTx.GetRelayerAddr()) > 0
	hasValidRelayerSignature := len(relayedTx.GetRelayerSignature()) == len(relayedTx.GetSignature()) && len(relayedTx.GetRelayerSignature()) > 0
	return hasValidRelayer && hasValidRelayerSignature
}

// IsRelayedTxV3 returns true if the provided transaction is a transaction of type relayed v3, without any further checks
func IsRelayedTxV3(tx data.TransactionHandler) bool {
	relayedTx, isRelayedV3 := tx.(data.RelayedTransactionHandler)
	if !isRelayedV3 {
		return false
	}

	hasRelayer := len(relayedTx.GetRelayerAddr()) > 0
	hasRelayerSignature := len(relayedTx.GetRelayerSignature()) > 0
	return hasRelayer || hasRelayerSignature
}

// IsEpochChangeBlockForFlagActivation returns true if the provided header is the first one after the specified flag's activation
func IsEpochChangeBlockForFlagActivation(header data.HeaderHandler, enableEpochsHandler EnableEpochsHandler, flag core.EnableEpochFlag) bool {
	isStartOfEpochBlock := header.IsStartOfEpochBlock()
	isBlockInActivationEpoch := header.GetEpoch() == enableEpochsHandler.GetActivationEpoch(flag)

	return isStartOfEpochBlock && isBlockInActivationEpoch
}

// IsFlagEnabledAfterEpochsStartBlock returns true if the flag is enabled for the header, but it is not the epoch start block
func IsFlagEnabledAfterEpochsStartBlock(header data.HeaderHandler, enableEpochsHandler EnableEpochsHandler, flag core.EnableEpochFlag) bool {
	isFlagEnabled := enableEpochsHandler.IsFlagEnabledInEpoch(flag, header.GetEpoch())
	isEpochStartBlock := IsEpochChangeBlockForFlagActivation(header, enableEpochsHandler, flag)
	return isFlagEnabled && !isEpochStartBlock
}

// GetShardIDs returns a map of shard IDs based on the number of shards
func GetShardIDs(numShards uint32) map[uint32]struct{} {
	shardIdentifiers := make(map[uint32]struct{})
	for i := uint32(0); i < numShards; i++ {
		shardIdentifiers[i] = struct{}{}
	}
	shardIdentifiers[core.MetachainShardId] = struct{}{}

	return shardIdentifiers
}

// GetBitmapSize will return expected bitmap size based on provided consensus size
func GetBitmapSize(
	consensusSize int,
) int {
	expectedBitmapSize := consensusSize / 8
	if consensusSize%8 != 0 {
		expectedBitmapSize++
	}

	return expectedBitmapSize
}

// IsConsensusBitmapValid checks if the provided keys and bitmap match the consensus requirements
func IsConsensusBitmapValid(
	log logger.Logger,
	consensusPubKeys []string,
	bitmap []byte,
	shouldApplyFallbackValidation bool,
) error {
	consensusSize := len(consensusPubKeys)

	expectedBitmapSize := GetBitmapSize(consensusSize)
	if len(bitmap) != expectedBitmapSize {
		log.Debug("wrong size bitmap",
			"expected number of bytes", expectedBitmapSize,
			"actual", len(bitmap))
		return ErrWrongSizeBitmap
	}

	numOfOnesInBitmap := 0
	for index := range bitmap {
		numOfOnesInBitmap += bits.OnesCount8(bitmap[index])
	}

	minNumRequiredSignatures := core.GetPBFTThreshold(consensusSize)
	if shouldApplyFallbackValidation {
		minNumRequiredSignatures = core.GetPBFTFallbackThreshold(consensusSize)
		log.Warn("IsConsensusBitmapValid: fallback validation has been applied",
			"minimum number of signatures required", minNumRequiredSignatures,
			"actual number of signatures in bitmap", numOfOnesInBitmap,
		)
	}

	if numOfOnesInBitmap >= minNumRequiredSignatures {
		return nil
	}

	log.Debug("not enough signatures",
		"minimum expected", minNumRequiredSignatures,
		"actual", numOfOnesInBitmap)

	return ErrNotEnoughSignatures
}

// ConsensusGroupSizeForShardAndEpoch returns the consensus group size for a specific shard in a given epoch
func ConsensusGroupSizeForShardAndEpoch(
	log logger.Logger,
	chainParametersHandler chainParametersHandler,
	shardID uint32,
	epoch uint32,
) int {
	currentChainParameters, err := chainParametersHandler.ChainParametersForEpoch(epoch)
	if err != nil {
		log.Warn("ConsensusGroupSizeForShardAndEpoch: could not compute chain params for epoch. "+
			"Will use the current chain parameters", "epoch", epoch, "error", err)
		currentChainParameters = chainParametersHandler.CurrentChainParameters()
	}

	if shardID == core.MetachainShardId {
		return int(currentChainParameters.MetachainConsensusGroupSize)
	}

	return int(currentChainParameters.ShardConsensusGroupSize)
}

// GetEquivalentProofNonceShardKey returns a string key nonce-shardID
func GetEquivalentProofNonceShardKey(nonce uint64, shardID uint32) string {
	return fmt.Sprintf("%d%s%d", nonce, keySeparator, shardID)
}

// GetEquivalentProofHashShardKey returns a string key hash-shardID
func GetEquivalentProofHashShardKey(hash []byte, shardID uint32) string {
	return fmt.Sprintf("%s%s%d", hex.EncodeToString(hash), keySeparator, shardID)
}

// GetHashAndShardFromKey returns the hash and shard from the provided key
func GetHashAndShardFromKey(hashShardKey []byte) ([]byte, uint32, error) {
	hashShardKeyStr := string(hashShardKey)
	result := strings.Split(hashShardKeyStr, keySeparator)
	if len(result) != expectedKeyLen {
		return nil, 0, ErrInvalidHashShardKey
	}

	hash, err := hex.DecodeString(result[hashIndex])
	if err != nil {
		return nil, 0, err
	}

	shard, err := strconv.Atoi(result[shardIndex])
	if err != nil {
		return nil, 0, err
	}

	return hash, uint32(shard), nil
}

// GetNonceAndShardFromKey returns the nonce and shard from the provided key
func GetNonceAndShardFromKey(nonceShardKey []byte) (uint64, uint32, error) {
	nonceShardKeyStr := string(nonceShardKey)
	result := strings.Split(nonceShardKeyStr, keySeparator)
	if len(result) != expectedKeyLen {
		return 0, 0, ErrInvalidNonceShardKey
	}

	nonce, err := strconv.Atoi(result[nonceIndex])
	if err != nil {
		return 0, 0, err
	}

	shard, err := strconv.Atoi(result[shardIndex])
	if err != nil {
		return 0, 0, err
	}

	return uint64(nonce), uint32(shard), nil
}
