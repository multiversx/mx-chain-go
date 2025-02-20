package chaos

import (
	"fmt"
	"time"
)

func (controller *chaosController) doFailPanic(failureName string, _ PointInput) error {
	panic(fmt.Sprintf("chaos: %s", failureName))
}

func (controller *chaosController) doFailReturnError(_ string, _ PointInput) error {
	return ErrChaoticBehavior
}

func (controller *chaosController) doFailEarlyReturn(_ string, _ PointInput) error {
	return ErrEarlyReturn
}

func (controller *chaosController) doFailCorruptSignature(_ string, input PointInput) error {
	input.Signature[0] += 1
	return nil
}

func (controller *chaosController) doFailSleep(failureName string, _ PointInput) error {
	duration := controller.profile.getFailureParameterAsFloat64(failureName, "duration")
	time.Sleep(time.Duration(duration) * time.Second)

	return nil
}
