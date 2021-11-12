package roundActivation

import (
	"reflect"
	"sync"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
)

type roundActivation struct {
	mutex          sync.RWMutex
	roundByNameMap map[string]uint64
}

// NewRoundActivation creates a new round activation handler component
func NewRoundActivation(config config.RoundConfig) (process.RoundActivationHandler, error) {
	configMap, err := getRoundsByNameMap(config)
	if err != nil {
		return nil, err
	}

	return &roundActivation{
		roundByNameMap: configMap,
		mutex:          sync.RWMutex{},
	}, nil
}

// IsEnabled checks if the queried round flag name is enabled in the queried round
func (ra *roundActivation) IsEnabled(name string, round uint64) bool {
	ra.mutex.RLock()
	r, exists := ra.roundByNameMap[name]
	ra.mutex.RUnlock()

	if exists {
		return round >= r
	}

	return false
}

// IsInterfaceNil checks if the underlying pointer receiver is nil
func (ra *roundActivation) IsInterfaceNil() bool {
	return ra == nil
}

func getRoundsByNameMap(roundConfig config.RoundConfig) (map[string]uint64, error) {
	v := reflect.ValueOf(roundConfig)
	ret := make(map[string]uint64, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		tmp := v.Field(i).Interface()

		roundByName, castOk := tmp.(config.ActivationRoundByName)
		if !castOk {
			return nil, process.ErrInvalidRoundActivationConfig
		}
		if len(roundByName.Name) == 0 {
			return nil, process.ErrNilActivationRoundName
		}
		if _, exists := ret[roundByName.Name]; exists {
			return nil, process.ErrDuplicateRoundActivationName
		}

		ret[roundByName.Name] = roundByName.Round
	}

	return ret, nil
}
