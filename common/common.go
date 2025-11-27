package common

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"math/bits"
	"reflect"
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/errors"
)

const (
	keySeparator   = "-"
	expectedKeyLen = 2
	hashIndex      = 0
	shardIndex     = 1
	nonceIndex     = 0
)

type executionResultHandler interface {
	GetMiniBlockHeadersHandlers() []data.MiniBlockHeaderHandler
}

type chainParametersHandler interface {
	CurrentChainParameters() config.ChainParametersByEpochConfig
	ChainParametersForEpoch(epoch uint32) (config.ChainParametersByEpochConfig, error)
	IsInterfaceNil() bool
}

// PrepareLogEventsKey will prepare logs key for cacher
func PrepareLogEventsKey(headerHash []byte) []byte {
	return append([]byte("logs"), headerHash...)
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

// IsAsyncExecutionEnabledForEpochAndRound returns true if both Supernova epochs and Supernova rounds are enabled for the provided epoch and round
func IsAsyncExecutionEnabledForEpochAndRound(
	enableEpochsHandler EnableEpochsHandler,
	enableRoundsHandler EnableRoundsHandler,
	epoch uint32,
	round uint64,
) bool {
	return enableEpochsHandler.IsFlagEnabledInEpoch(SupernovaFlag, epoch) &&
		enableRoundsHandler.IsFlagEnabledInRound(SupernovaRoundFlag, round)
}

// IsAsyncExecutionEnabled returns true if both Supernova epochs and Supernova rounds are enabled
func IsAsyncExecutionEnabled(enableEpochsHandler EnableEpochsHandler, enableRoundsHandler EnableRoundsHandler) bool {
	return enableEpochsHandler.IsFlagEnabled(SupernovaFlag) &&
		enableRoundsHandler.IsFlagEnabled(SupernovaRoundFlag)
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

// ConvertTimeStampSecToMs will convert unix timestamp from seconds to milliseconds
func ConvertTimeStampSecToMs(timeStamp uint64) uint64 {
	return timeStamp * 1000
}

func convertTimeStampMsToSec(timeStamp uint64) uint64 {
	return timeStamp / 1000
}

// GetHeaderTimestamps will return timestamps as seconds and milliseconds based on supernova round activation
func GetHeaderTimestamps(
	header data.HeaderHandler,
	enableEpochsHandler EnableEpochsHandler,
) (uint64, uint64, error) {
	if check.IfNil(header) {
		return 0, 0, ErrNilHeaderHandler
	}
	if check.IfNil(enableEpochsHandler) {
		return 0, 0, errors.ErrNilEnableEpochsHandler
	}

	headerTimestamp := header.GetTimeStamp()

	timestampSec := headerTimestamp
	timestampMs := headerTimestamp

	if !enableEpochsHandler.IsFlagEnabledInEpoch(SupernovaFlag, header.GetEpoch()) {
		timestampMs = ConvertTimeStampSecToMs(headerTimestamp)
		return timestampSec, timestampMs, nil
	}

	// reduce block timestamp (which now comes as milliseconds) to seconds to keep backwards compatibility
	// from now on timestampMs will be used for milliseconds granularity
	timestampSec = convertTimeStampMsToSec(headerTimestamp)

	return timestampSec, timestampMs, nil
}

// PrettifyStruct returns a JSON string representation of a struct, converting byte slices to hex
// and formatting big number values into readable strings. Useful for logging or debugging.
func PrettifyStruct(x interface{}) (string, error) {
	if x == nil {
		return "nil", nil
	}

	val := reflect.ValueOf(x)
	result := prettifyValue(val, val.Type())

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// prettifyValue recursively converts a reflect.Value into a representation suitable for JSON serialization,
// handling pointers, slices, structs, and special formatting for big numeric types.
func prettifyValue(val reflect.Value, typ reflect.Type) interface{} {
	if bigValue, isBig := prettifyBigNumbers(val); isBig {
		return bigValue
	}

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
		typ = val.Type()
	}

	switch val.Kind() {
	case reflect.Struct:
		return prettifyStructFields(val, typ)
	case reflect.Slice, reflect.Array:
		return prettifySliceOrArray(val)
	default:
		return val.Interface()
	}
}

func prettifyStructFields(val reflect.Value, typ reflect.Type) map[string]interface{} {
	out := make(map[string]interface{})
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		name := fieldType.Tag.Get("json")
		if name == "" {
			name = fieldType.Name
		} else {
			name = strings.Split(name, ",")[0]
		}

		if fieldType.PkgPath != "" {
			out[name] = "<unexported>"
			continue
		}

		if field.Kind() == reflect.Slice && field.Type() == reflect.TypeOf([]byte{}) {
			out[name] = fmt.Sprintf("%x", field.Bytes())
		} else {
			out[name] = prettifyValue(field, field.Type())
		}
	}
	return out
}

func prettifySliceOrArray(val reflect.Value) interface{} {
	if val.Type().Elem().Kind() == reflect.Uint8 {
		b := make([]byte, val.Len())
		for i := 0; i < val.Len(); i++ {
			b[i] = byte(val.Index(i).Uint())
		}
		return fmt.Sprintf("%x", b)
	}

	out := make([]interface{}, val.Len())
	for i := 0; i < val.Len(); i++ {
		out[i] = prettifyValue(val.Index(i), val.Index(i).Type())
	}
	return out
}

func prettifyBigNumbers(val reflect.Value) (string, bool) {
	if val.CanInterface() {
		switch v := val.Interface().(type) {
		case *big.Int:
			if v != nil {
				return v.String(), true
			}
		case big.Int:
			return v.String(), true
		case *big.Float:
			if v != nil {
				return v.Text('g', -1), true
			}
		case big.Float:
			return v.Text('g', -1), true
		case *big.Rat:
			if v != nil {
				return v.RatString(), true
			}
		case big.Rat:
			return v.RatString(), true
		}
	}
	return "", false
}

// GetLastBaseExecutionResultHandler extracts the BaseExecutionResultHandler from the provided header, based on its type
func GetLastBaseExecutionResultHandler(header data.HeaderHandler) (data.BaseExecutionResultHandler, error) {
	if check.IfNil(header) {
		return nil, ErrNilHeaderHandler
	}

	lastExecResultsHandler := header.GetLastExecutionResultHandler()
	return ExtractBaseExecutionResultHandler(lastExecResultsHandler)
}

// GetOrCreateLastExecutionResultForPrevHeader extracts base execution result from
// header if header v3. Otherwise, it will create last execution result based
// on the provided header
func GetOrCreateLastExecutionResultForPrevHeader(
	prevHeader data.HeaderHandler,
	prevHeaderHash []byte,
) (data.BaseExecutionResultHandler, error) {
	if prevHeader.IsHeaderV3() {
		return ExtractBaseExecutionResultHandler(prevHeader.GetLastExecutionResultHandler())
	}

	lastExecResult, err := CreateLastExecutionResultFromPrevHeader(prevHeader, prevHeaderHash)
	if err != nil {
		return nil, err
	}

	return ExtractBaseExecutionResultHandler(lastExecResult)
}

func isValidHeaderBeforeV3(header data.HeaderHandler) error {
	_, isHeaderV2 := header.(*block.HeaderV2)
	if isHeaderV2 {
		return nil
	}

	_, isHeaderV1 := header.(*block.Header)
	if !isHeaderV1 {
		return ErrWrongTypeAssertion
	}

	return nil
}

// CreateLastExecutionResultFromPrevHeader creates a LastExecutionResultInfo object from the given previous header
func CreateLastExecutionResultFromPrevHeader(prevHeader data.HeaderHandler, prevHeaderHash []byte) (data.LastExecutionResultHandler, error) {
	if check.IfNil(prevHeader) {
		return nil, ErrNilHeaderHandler
	}
	if len(prevHeaderHash) == 0 {
		return nil, ErrInvalidHeaderHash
	}

	if prevHeader.GetShardID() != core.MetachainShardId {
		err := isValidHeaderBeforeV3(prevHeader)
		if err != nil {
			return nil, err
		}

		return &block.ExecutionResultInfo{
			NotarizedInRound: prevHeader.GetRound(),
			ExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  prevHeaderHash,
				HeaderNonce: prevHeader.GetNonce(),
				HeaderRound: prevHeader.GetRound(),
				RootHash:    prevHeader.GetRootHash(),
				GasUsed:     0, // we don't have this information in previous header
			},
		}, nil
	}

	prevMetaHeader, ok := prevHeader.(*block.MetaBlock)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	return &block.MetaExecutionResultInfo{
		NotarizedInRound: prevHeader.GetRound(),
		ExecutionResult: &block.BaseMetaExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  prevHeaderHash,
				HeaderNonce: prevMetaHeader.GetNonce(),
				HeaderRound: prevMetaHeader.GetRound(),
				RootHash:    prevMetaHeader.GetRootHash(),
				GasUsed:     0, // we don't have this information in previous header
			},
			ValidatorStatsRootHash: prevMetaHeader.GetValidatorStatsRootHash(),
			AccumulatedFeesInEpoch: prevMetaHeader.GetAccumulatedFeesInEpoch(),
			DevFeesInEpoch:         prevMetaHeader.GetDevFeesInEpoch(),
		},
	}, nil
}

// ExtractBaseExecutionResultHandler extracts the base execution result handler from a last execution result handler
func ExtractBaseExecutionResultHandler(lastExecResultsHandler data.LastExecutionResultHandler) (data.BaseExecutionResultHandler, error) {
	if check.IfNil(lastExecResultsHandler) {
		log.Error("ExtractBaseExecutionResultHandler: nil exec")
		return nil, ErrNilLastExecutionResultHandler
	}

	var baseExecutionResultsHandler data.BaseExecutionResultHandler
	var ok bool
	switch executionResultsHandlerType := lastExecResultsHandler.(type) {
	case data.LastMetaExecutionResultHandler:
		metaBaseExecutionResults := executionResultsHandlerType.GetExecutionResultHandler()
		if check.IfNil(metaBaseExecutionResults) {
			return nil, ErrNilBaseExecutionResult
		}
		baseExecutionResultsHandler, ok = metaBaseExecutionResults.(data.BaseExecutionResultHandler)
		if !ok {
			return nil, ErrWrongTypeAssertion
		}
	case data.LastShardExecutionResultHandler:
		baseExecutionResultsHandler = executionResultsHandlerType.GetExecutionResultHandler()
	default:
		return nil, fmt.Errorf("%w: unsupported execution result handler type", ErrWrongTypeAssertion)
	}

	if check.IfNil(baseExecutionResultsHandler) {
		return nil, ErrNilBaseExecutionResult
	}

	return baseExecutionResultsHandler, nil
}

// GetMiniBlocksHeaderHandlersFromExecResult returns miniblock handlers based on execution result
func GetMiniBlocksHeaderHandlersFromExecResult(
	baseExecResult data.BaseExecutionResultHandler,
) ([]data.MiniBlockHeaderHandler, error) {
	if check.IfNil(baseExecResult) {
		return nil, ErrNilBaseExecutionResult
	}

	execResult, ok := baseExecResult.(executionResultHandler)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	return execResult.GetMiniBlockHeadersHandlers(), nil
}

// GetLastExecutionResultNonce returns last execution result nonce if header v3 enabled, otherwise it returns provided header nonce
func GetLastExecutionResultNonce(
	header data.HeaderHandler,
) uint64 {
	nonce := header.GetNonce()

	if !header.IsHeaderV3() {
		return nonce
	}

	lastExecutionResult, err := GetLastBaseExecutionResultHandler(header)
	if err != nil {
		return nonce
	}

	return lastExecutionResult.GetHeaderNonce()
}

// GetFirstExecutionResultNonce returns first execution result nonce if header v3 enabled, otherwise it returns provided header nonce.
// For header v3, it returns first execution result if there are any, otherwise it returns last execution results on the header
func GetFirstExecutionResultNonce(
	header data.HeaderHandler,
) uint64 {
	nonce := header.GetNonce()

	if !header.IsHeaderV3() {
		return nonce
	}

	if len(header.GetExecutionResultsHandlers()) > 0 {
		firstExecResult := header.GetExecutionResultsHandlers()[0]
		return firstExecResult.GetHeaderNonce()
	}

	return GetLastExecutionResultNonce(header)
}

// GetMiniBlockHeadersFromExecResult returns mb headers from meta header if v3, otherwise, returns mini block headers
func GetMiniBlockHeadersFromExecResult(metaBlock data.HeaderHandler) ([]data.MiniBlockHeaderHandler, error) {
	if !metaBlock.IsHeaderV3() {
		return metaBlock.GetMiniBlockHeaderHandlers(), nil
	}

	mbHeaderHandlers := make([]data.MiniBlockHeaderHandler, 0)
	for _, execResult := range metaBlock.GetExecutionResultsHandlers() {
		mbHeaders, err := GetMiniBlocksHeaderHandlersFromExecResult(execResult)
		if err != nil {
			return nil, fmt.Errorf("%w in GetMiniBlockHeadersFromExecResult.GetMiniBlocksHeaderHandlersFromExecResult", err)
		}

		mbHeaderHandlers = append(mbHeaderHandlers, mbHeaders...)
	}

	return mbHeaderHandlers, nil
}
