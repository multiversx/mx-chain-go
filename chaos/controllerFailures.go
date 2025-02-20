package chaos

import (
	"fmt"
	"time"
)

func (controller *chaosController) doFailPanic(failure string, input PointInput) error {
	log.Info("doFailPanic()", "failure", failure, "point", input.Name)

	panic(fmt.Sprintf("chaos: %s", failure))
}

func (controller *chaosController) doFailReturnError(failure string, input PointInput) error {
	log.Info("doFailReturnError()", "failure", failure, "point", input.Name)
	return ErrChaoticBehavior
}

func (controller *chaosController) doFailEarlyReturn(failure string, input PointInput) error {
	log.Info("doFailEarlyReturn()", "failure", failure, "point", input.Name)
	return ErrEarlyReturn
}

func (controller *chaosController) doFailCorruptSignature(failure string, input PointInput) error {
	log.Info("doFailCorruptSignature()", "failure", failure, "point", input.Name)

	input.Signature[0] += 1
	return nil
}

func (controller *chaosController) doFailSleep(failure string, input PointInput) error {
	log.Info("doFailSleep()", "failure", failure, "point", input.Name)

	duration := controller.profile.getFailureParameterAsFloat64(failure, "duration")
	time.Sleep(time.Duration(duration) * time.Second)

	return nil
}
