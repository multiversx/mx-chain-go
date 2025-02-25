package chaosImpl

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-go/chaos"
)

func (controller *chaosController) doFailPanic(failure string, input chaos.PointInput) chaos.PointOutput {
	log.Info("doFailPanic()", "failure", failure, "point", input.Name)

	panic(fmt.Sprintf("chaos: %s", failure))
}

func (controller *chaosController) doFailReturnError(failure string, input chaos.PointInput) chaos.PointOutput {
	log.Info("doFailReturnError()", "failure", failure, "point", input.Name)
	return chaos.PointOutput{Error: ErrChaoticBehavior}
}

func (controller *chaosController) doFailEarlyReturn(failure string, input chaos.PointInput) chaos.PointOutput {
	log.Info("doFailEarlyReturn()", "failure", failure, "point", input.Name)
	return chaos.PointOutput{Error: ErrEarlyReturn}
}

func (controller *chaosController) doFailCorruptVariables(failure string, input chaos.PointInput) chaos.PointOutput {
	log.Info("doFailCorruptSignature()", "failure", failure, "point", input.Name)

	for index, item := range input.CorruptibleVariables {
		itemAsBytes, ok := item.([]byte)
		if ok {
			before := hex.EncodeToString(itemAsBytes)
			itemAsBytes[0] += 1
			after := hex.EncodeToString(itemAsBytes)

			log.Debug("doFailCorruptSignature(): corrupting bytes", "index", index, "before", before, "after", after)
			continue
		}

		itemAsInt, ok := item.(*int)
		if ok {
			before := *itemAsInt
			*itemAsInt += 1
			after := *itemAsInt

			log.Debug("doFailCorruptSignature(): corrupting int", "index", index, "before", before, "after", after)
			continue
		}
	}

	return chaos.PointOutput{}
}

func (controller *chaosController) doFailSleep(failure string, input chaos.PointInput) chaos.PointOutput {
	log.Info("doFailSleep()", "failure", failure, "point", input.Name)

	duration := controller.profile.getFailureParameterAsFloat64(failure, "duration")
	time.Sleep(time.Duration(duration) * time.Second)

	return chaos.PointOutput{}
}
