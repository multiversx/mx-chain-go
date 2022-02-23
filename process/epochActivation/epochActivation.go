package epochActivation

import (
	"errors"
	"reflect"
	"sync"

	"github.com/ElrondNetwork/elrond-go/config"
)

type epochActivation struct {
	epochByName config.EnableEpochs
	epoch       uint64
	mutex       sync.RWMutex
}

// NewEpochActivation creates a new epoch activation handler component
func NewEpochActivation(config config.EnableEpochs) (*epochActivation, error) {
	epochActivationConfig, err := getEpochsByNameMap(config)
	if err != nil {
		return nil, err
	}

	return &epochActivation{
		epochByName: epochActivationConfig,
		mutex:       sync.RWMutex{},
	}, nil
}

func (ea *epochActivation) IsEnabledInEpoch(name string, epoch uint64) bool {
	//TODO: add implementation
	return false
}

func (ea *epochActivation) IsEnabled(name string) bool {
	//TODO: add implementation
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (ea *epochActivation) IsInterfaceNil() bool {
	return ea == nil
}

func getEpochsByNameMap(enableEpochsConfig config.EnableEpochs) (config.EnableEpochs, error) {
	values := reflect.ValueOf(enableEpochsConfig)
	types := reflect.TypeOf(enableEpochsConfig)
	var epochByName config.EnableEpochs

	for i := 0; i < values.NumField(); i++ {
		name := types.Field(i).Name

		if name == "MaxNodesChangeEnableEpoch" {
			tmp := values.Field(i).Interface()
			maxNodesChangeConfigElem, castOk := tmp.([]config.MaxNodesChangeConfig)
			if !castOk {
				return config.EnableEpochs{}, errors.New(name)

			}
			if len(maxNodesChangeConfigElem) == 0 {
				return config.EnableEpochs{}, errors.New(name)
			}
			epochByName.MaxNodesChangeEnableEpoch = maxNodesChangeConfigElem
		} else {
			value := values.Field(i).Uint() //.(uint64)
			if value == 0 {
				return config.EnableEpochs{}, errors.New(name)
			}

			if len(name) == 0 {
				return config.EnableEpochs{}, errors.New(name)
			}

			var tmp interface{} = &epochByName
			valueToBeSet := reflect.ValueOf(tmp).Elem().FieldByName(name)

			if valueToBeSet.IsValid() {
				valueToBeSet.SetUint(value)
			}

		}

	}

	return epochByName, nil
}
