package chaos

import (
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
	"strconv"
)

type failureCircumstance struct {
	// Always available:
	randomNumber    uint64
	now             int64
	nodeDisplayName string
	shard           uint32
	epoch           uint32
	round           uint64

	// Always available (counters):
	counterProcessTransaction uint64

	// Not always available:
	blockNonce      uint64
	nodePublicKey   []byte
	transactionHash []byte
}

func (circumstance *failureCircumstance) evalExpression(expression string) (resultAsBool bool, err error) {
	fileSet := token.NewFileSet()
	pack := circumstance.createGoPackage()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred during evaluation: %v", r)
		}
	}()

	result, err := types.Eval(fileSet, pack, token.NoPos, expression)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate expression: %v", err)
	}

	resultAsBool, err = strconv.ParseBool(result.Value.String())
	return resultAsBool, err
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

	// Always available (counters):
	scope.Insert(createFailureExpressionNumericParameter(pack, parameterCounterProcessTransaction, circumstance.counterProcessTransaction))

	// Not always available:
	if circumstance.blockNonce > 0 {
		scope.Insert(createFailureExpressionNumericParameter(pack, parameterBlockNonce, circumstance.blockNonce))
	}

	if len(circumstance.nodePublicKey) > 0 {
		nodePublicKeyLastByte := circumstance.nodePublicKey[len(circumstance.nodePublicKey)-1]
		scope.Insert(createFailureExpressionNumericParameter(pack, parameterNodePublicKeyLastByte, uint64(nodePublicKeyLastByte)))
	}

	if len(circumstance.transactionHash) > 0 {
		transactionHashLastByte := circumstance.transactionHash[len(circumstance.transactionHash)-1]
		scope.Insert(createFailureExpressionNumericParameter(pack, parameterTransactionHashLastByte, uint64(transactionHashLastByte)))
	}

	scope.Insert(types.NewConst(token.NoPos, pack, "cucu", types.Typ[types.String], constant.MakeString("hello")))

	return pack
}

func createFailureExpressionNumericParameter(pack *types.Package, name failureExpressionParameterName, value uint64) *types.Const {
	return types.NewConst(token.NoPos, pack, string(name), types.Typ[types.Uint64], constant.MakeUint64(value))
}

func createFailureExpressionStringParameter(pack *types.Package, name failureExpressionParameterName, value string) *types.Const {
	return types.NewConst(token.NoPos, pack, string(name), types.Typ[types.String], constant.MakeString(value))
}
