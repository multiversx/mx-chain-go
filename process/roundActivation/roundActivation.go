package roundActivation

import (
	"reflect"
	"sync"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
)

type roundActivation struct {
	roundByNameMap map[string]uint64
	round          uint64
	mutex          sync.RWMutex
}

// NewRoundActivation creates a new round activation handler component
func NewRoundActivation(config config.RoundConfig) (*roundActivation, error) {
	configMap, err := getRoundsByNameMap(config)
	if err != nil {
		return nil, err
	}

	return &roundActivation{
		roundByNameMap: configMap,
		mutex:          sync.RWMutex{},
	}, nil
}

// IsEnabledInRound checks if the queried round flag name is enabled in the queried round
func (ra *roundActivation) IsEnabledInRound(name string, round uint64) bool {
	ra.mutex.RLock()
	defer ra.mutex.RUnlock()

	return ra.isEnabledInRound(name, round)
}

func (ra *roundActivation) isEnabledInRound(name string, round uint64) bool {
	r, exists := ra.roundByNameMap[name]
	if exists {
		return round >= r
	}

	return false
}

// IsEnabled checks if the queried round flag name is enabled in the current processed round
func (ra *roundActivation) IsEnabled(name string) bool {
	ra.mutex.RLock()
	defer ra.mutex.RUnlock()

	currRound := ra.round
	return ra.isEnabledInRound(name, currRound)
}

// RoundConfirmed resets the current stored round
func (ra *roundActivation) RoundConfirmed(round uint64) {
	ra.mutex.Lock()
	ra.round = round
	ra.mutex.Unlock()
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
