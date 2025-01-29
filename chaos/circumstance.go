package chaos

import (
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
	"strconv"
)

func evalExpression(expression string, variables map[string]int) (bool, error) {
	fileSet := token.NewFileSet()

	myPackage := types.NewPackage("chaosCircumstance", "chaosCircumstance")

	for name, value := range variables {
		// We'll only use integer values
		valueAsConstant := types.NewConst(
			token.NoPos,
			nil,
			name,
			types.Typ[types.Int],
			constant.MakeInt64(int64(value)),
		)

		myPackage.Scope().Insert(valueAsConstant)
	}

	// Evaluate the expression
	tv, err := types.Eval(fileSet, myPackage, token.NoPos, expression)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate expression: %v", err)
	}

	// Convert result to boolean
	return strconv.ParseBool(tv.Value.String())
}
