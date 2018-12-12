package syncValidators

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// GetUnregisterList returns a list containing nodes from unregister list after a refresh action
func (sv *syncValidators) GetUnregisterList() map[string]*validatorData {
	sv.refresh()

	unregisterList := make(map[string]*validatorData, 0)

	sv.mut.Lock()

	for k, v := range sv.unregisterList {
		unregisterList[k] = &validatorData{RoundIndex: v.RoundIndex, Stake: v.Stake}
	}

	sv.mut.Unlock()

	return unregisterList
}

// GetWaitList returns a list containing nodes from wait list after a refresh action
func (sv *syncValidators) GetWaitList() map[string]*validatorData {
	sv.refresh()

	waitList := make(map[string]*validatorData, 0)

	sv.mut.Lock()

	for k, v := range sv.waitList {
		waitList[k] = &validatorData{RoundIndex: v.RoundIndex, Stake: v.Stake}
	}

	sv.mut.Unlock()

	return waitList
}

func (sv *syncValidators) Round() *chronology.Round {
	return sv.round
}
