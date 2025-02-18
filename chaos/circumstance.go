package chaos

import (
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-logger-go/proto"
)

type failureCircumstance struct {
	// Always available:
	randomNumber    uint64
	now             int64
	nodeDisplayName string
	shard           uint32
	epoch           uint32
	round           uint64

	// Not always available:
	nodeIndex     int
	nodePublicKey string
	consensusSize int
	amILeader     bool
	blockNonce    uint64
}

func newFailureCircumstance() *failureCircumstance {
	randomNumber := rand.Uint64()
	now := time.Now().Unix()

	return &failureCircumstance{
		nodeDisplayName: "",
		randomNumber:    randomNumber,
		now:             now,
		shard:           math.MaxUint32,
		epoch:           math.MaxUint32,
		round:           math.MaxUint64,

		nodeIndex:     -1,
		nodePublicKey: "",
		consensusSize: -1,
		amILeader:     false,
		blockNonce:    0,
	}
}

// For simplicity, we get the current shard, epoch and round from the logger correlation facility.
func (circumstance *failureCircumstance) enrichWithLoggerCorrelation(correlation proto.LogCorrelationMessage) {
	shard, err := core.ConvertShardIDToUint32(correlation.Shard)
	if err != nil {
		shard = math.MaxInt16
	}

	circumstance.shard = shard
	circumstance.epoch = correlation.Epoch
	circumstance.round = uint64(correlation.Round)
}

func (circumstance *failureCircumstance) enrichWithConsensusState(consensusState spos.ConsensusStateHandler, nodePublicKey string) {
	if consensusState == nil {
		return
	}

	if nodePublicKey == "" {
		nodePublicKey = consensusState.SelfPubKey()
	}

	nodeIndex, err := consensusState.ConsensusGroupIndex(nodePublicKey)
	if err != nil {
		log.Warn("failureCircumstance.enrichWithConsensusState(): error getting node index", "error", err)
	} else {
		circumstance.nodeIndex = nodeIndex
	}

	circumstance.nodePublicKey = nodePublicKey
	circumstance.consensusSize = consensusState.ConsensusGroupSize()
	circumstance.amILeader = consensusState.Leader() == nodePublicKey
	circumstance.blockNonce = consensusState.GetHeader().GetNonce()
}

func (circumstance *failureCircumstance) anyExpression(expressions []string) bool {
	fileSet := token.NewFileSet()
	pack := circumstance.createGoPackage()

	defer func() {
		if r := recover(); r != nil {
			log.Error("panic occurred during evaluation", "panic", r)
		}
	}()

	for _, expression := range expressions {
		result, err := types.Eval(fileSet, pack, token.NoPos, expression)
		if err != nil {
			log.Error("failed to evaluate expression", "error", err, "expression", expression)
			continue
		}

		resultAsString := result.Value.String()
		resultAsBool, err := strconv.ParseBool(resultAsString)
		if err != nil {
			log.Error("failed to parse result as bool", "error", err, "expression", expression, "result", result.Value.String())
			continue
		}

		log.Trace("failureCircumstance.anyExpression()",
			"result", resultAsString,
			"expression", fmt.Sprintf("[ %s ]", expression),
			"nodeDisplayName", circumstance.nodeDisplayName,
			"randomNumber", circumstance.randomNumber,
			"now", circumstance.now,
			"shard", circumstance.shard,
			"epoch", circumstance.epoch,
			"round", circumstance.round,
			"nodeIndex", circumstance.nodeIndex,
			"nodePublicKey", circumstance.nodePublicKey,
			"consensusSize", circumstance.consensusSize,
			"amILeader", circumstance.amILeader,
			"blockNonce", circumstance.blockNonce,
		)

		if resultAsBool {
			return true
		}
	}

	return false
}

func (circumstance *failureCircumstance) createGoPackage() *types.Package {
	pack := types.NewPackage("chaosCircumstanceEvaluation", "chaosCircumstanceEvaluation")
	scope := pack.Scope()

	// Always available:
	scope.Insert(createFailureExpressionNumericParameter(pack, parameterRandomNumber, circumstance.randomNumber))
	scope.Insert(createFailureExpressionNumericParameter(pack, parameterNow, uint64(circumstance.now)))
	scope.Insert(createFailureExpressionStringParameter(pack, parameterNodeDisplayName, circumstance.nodeDisplayName))
	scope.Insert(createFailureExpressionNumericParameter(pack, parameterShard, uint64(circumstance.shard)))
	scope.Insert(createFailureExpressionNumericParameter(pack, parameterEpoch, uint64(circumstance.epoch)))
	scope.Insert(createFailureExpressionNumericParameter(pack, parameterRound, uint64(circumstance.round)))

	// Not always available (set):
	scope.Insert(createFailureExpressionNumericParameter(pack, parameterNodeIndex, uint64(circumstance.nodeIndex)))
	scope.Insert(createFailureExpressionStringParameter(pack, parameterNodePublicKey, circumstance.nodePublicKey))
	scope.Insert(createFailureExpressionNumericParameter(pack, parameterConsensusSize, uint64(circumstance.consensusSize)))
	scope.Insert(createFailureExpressionBoolParameter(pack, parameterAmILeader, circumstance.amILeader))
	scope.Insert(createFailureExpressionNumericParameter(pack, parameterBlockNonce, circumstance.blockNonce))

	return pack
}

func createFailureExpressionNumericParameter(pack *types.Package, name failureExpressionParameterName, value uint64) *types.Const {
	return types.NewConst(token.NoPos, pack, string(name), types.Typ[types.Uint64], constant.MakeUint64(value))
}

func createFailureExpressionStringParameter(pack *types.Package, name failureExpressionParameterName, value string) *types.Const {
	return types.NewConst(token.NoPos, pack, string(name), types.Typ[types.String], constant.MakeString(value))
}

func createFailureExpressionBoolParameter(pack *types.Package, name failureExpressionParameterName, value bool) *types.Const {
	return types.NewConst(token.NoPos, pack, string(name), types.Typ[types.Bool], constant.MakeBool(value))
}
