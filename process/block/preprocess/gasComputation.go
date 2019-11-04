package preprocess

type gasComputation struct {
	gasConsumed uint64
}

func NewGasComputation() (*gasComputation, error) {
	gc := gasComputation{gasConsumed: 0}
	return &gc, nil
}

func (gc *gasComputation) InitGasConsumed() {
	gc.gasConsumed = 0
}

func (gc *gasComputation) AddGasConsumed(gasConsumed uint64) {
	gc.gasConsumed += gasConsumed
}

func (gc *gasComputation) SetGasConsumed(gasConsumed uint64) {
	gc.gasConsumed = gasConsumed
}

func (gc *gasComputation) GetGasConsumed() uint64 {
	return gc.gasConsumed
}

func (gc *gasComputation) IsInterfaceNil() bool {
	if gc == nil {
		return true
	}

	return false
}
