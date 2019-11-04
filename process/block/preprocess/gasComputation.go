package preprocess

type gasComputation struct {
	consumedGas uint64
}

func NewGasComputation() (*gasComputation, error) {
	gc := gasComputation{consumedGas: 0}
	return &gc, nil
}

func (gc *gasComputation) InitConsumedGas() {
	gc.consumedGas = 0
}

func (gc *gasComputation) AddConsumedGas(consumedGas uint64) {
	gc.consumedGas += consumedGas
}

func (gc *gasComputation) GetConsumedGas() uint64 {
	return gc.consumedGas
}

func (gc *gasComputation) IsInterfaceNil() bool {
	if gc == nil {
		return true
	}

	return false
}
