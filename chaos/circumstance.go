package chaos

import (
	"encoding/hex"
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-logger-go/proto"
)

type failureCircumstance struct {
	// Always available:
	point           string
	failure         string
	randomNumber    uint64
	now             int64
	uptime          int64
	nodeDisplayName string
	shard           uint32
	epoch           uint32
	round           uint64

	// Not always available:
	nodeIndex           int
	nodePublicKey       string
	consensusSize       int
	amILeader           bool
	blockNonce          uint64
	blockIsStartOfEpoch bool
}

func newFailureCircumstance() *failureCircumstance {
	randomNumber := rand.Uint64()
	now := time.Now().Unix()
	uptime := time.Since(startTime).Seconds()

	return &failureCircumstance{
		point:           "",
		nodeDisplayName: "",
		randomNumber:    randomNumber,
		now:             now,
		uptime:          int64(uptime),
		shard:           math.MaxUint32,
		epoch:           math.MaxUint32,
		round:           math.MaxUint64,

		nodeIndex:           -1,
		nodePublicKey:       "",
		consensusSize:       -1,
		amILeader:           false,
		blockNonce:          0,
		blockIsStartOfEpoch: false,
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
	if check.IfNil(consensusState) {
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

	circumstance.nodePublicKey = hex.EncodeToString([]byte(nodePublicKey))
	circumstance.consensusSize = consensusState.ConsensusGroupSize()
	circumstance.amILeader = consensusState.Leader() == nodePublicKey
	circumstance.enrichWithBlockHeader(consensusState.GetHeader())
}

func (circumstance *failureCircumstance) enrichWithBlockHeader(header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	circumstance.blockNonce = header.GetNonce()
	circumstance.blockIsStartOfEpoch = header.IsStartOfEpochBlock()
}

func (circumstance *failureCircumstance) anyExpression(triggerExpressions []string) bool {
	fileSet := token.NewFileSet()
	pack := circumstance.createGoPackage()

	defer func() {
		if r := recover(); r != nil {
			log.Error("panic occurred during evaluation", "panic", r)
		}
	}()

	for _, expression := range triggerExpressions {
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

		log.Trace("evaluated expression",
			"result", resultAsString,
			"expression", fmt.Sprintf("[ %s ]", expression),
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
	scope.Insert(createFailureExpressionStringParameter(pack, parameterPoint, circumstance.point))
	scope.Insert(createFailureExpressionNumericParameter(pack, parameterRandomNumber, circumstance.randomNumber))
	scope.Insert(createFailureExpressionNumericParameter(pack, parameterNow, uint64(circumstance.now)))
	scope.Insert(createFailureExpressionNumericParameter(pack, parameterUptime, uint64(circumstance.uptime)))
	scope.Insert(createFailureExpressionStringParameter(pack, parameterNodeDisplayName, circumstance.nodeDisplayName))
	scope.Insert(createFailureExpressionNumericParameter(pack, parameterShard, uint64(circumstance.shard)))
	scope.Insert(createFailureExpressionNumericParameter(pack, parameterEpoch, uint64(circumstance.epoch)))
	scope.Insert(createFailureExpressionNumericParameter(pack, parameterRound, uint64(circumstance.round)))

	// Not always available (set):

	if circumstance.nodeIndex >= 0 {
		scope.Insert(createFailureExpressionNumericParameter(pack, parameterNodeIndex, uint64(circumstance.nodeIndex)))
		scope.Insert(createFailureExpressionStringParameter(pack, parameterNodePublicKey, circumstance.nodePublicKey))
		scope.Insert(createFailureExpressionNumericParameter(pack, parameterConsensusSize, uint64(circumstance.consensusSize)))
		scope.Insert(createFailureExpressionBoolParameter(pack, parameterAmILeader, circumstance.amILeader))
	}

	if circumstance.blockNonce > 0 {
		scope.Insert(createFailureExpressionNumericParameter(pack, parameterBlockNonce, circumstance.blockNonce))
		scope.Insert(createFailureExpressionBoolParameter(pack, parameterBlockIsStartOfEpoch, circumstance.blockIsStartOfEpoch))
	}

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
