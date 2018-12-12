package syncValidators

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"math/big"
	"sync"
)

// RoundsToWait holds numer of rounds after one node could be transfered from wait list to eligible list,
// or after it can be removed from eligible list after an unregister request
const RoundsToWait = 5

type validatorData struct {
	RoundIndex int
	Stake      big.Int
}

// syncValidators implements a validators sync mechanism
type syncValidators struct {
	eligibleList   map[string]*validatorData
	waitList       map[string]*validatorData
	unregisterList map[string]*validatorData

	mut sync.RWMutex

	round *chronology.Round
}

// NewSyncValidators creates a new syncValidators object
func NewSyncValidators(
	round *chronology.Round,
) (*syncValidators, error) {
	err := checkSyncValidatorsNilParameters(round)

	if err != nil {
		return nil, err
	}

	sv := syncValidators{}

	sv.eligibleList = make(map[string]*validatorData, 0)
	sv.waitList = make(map[string]*validatorData, 0)
	sv.unregisterList = make(map[string]*validatorData, 0)

	sv.round = round

	return &sv, nil
}

// checkSyncValidatorsNilParameters will check the imput parameters for nil values
func checkSyncValidatorsNilParameters(
	round *chronology.Round,
) error {
	if round == nil {
		return execution.ErrNilRound
	}

	return nil
}

// AddValidator adds a validator in the wait list
func (sv *syncValidators) AddValidator(nodeId string, stake big.Int) {
	sv.mut.Lock()
	if v, ok := sv.waitList[nodeId]; ok {
		stake.Add(&stake, &v.Stake)
	}
	sv.waitList[nodeId] = &validatorData{RoundIndex: sv.round.Index(), Stake: stake}
	sv.mut.Unlock()
}

// RemoveValidator adds a validator in the unregister list
func (sv *syncValidators) RemoveValidator(nodeId string) {
	sv.mut.Lock()
	sv.unregisterList[nodeId] = &validatorData{RoundIndex: sv.round.Index()}
	sv.mut.Unlock()
}

// refresh executes a refreshing / syncing operation of the validators in the lists
func (sv *syncValidators) refresh() {
	sv.mut.Lock()

	for k, v := range sv.unregisterList { // remove validators from wait or eligible list after an unregister request
		// if the validator is in wait list it could be removed directly
		if _, ok := sv.waitList[k]; ok {
			delete(sv.waitList, k)
			delete(sv.unregisterList, k)
			continue
		}

		// if the validator is in eligibile list it could be removed only after some rounds
		if _, ok := sv.eligibleList[k]; ok {
			if v.RoundIndex+RoundsToWait < sv.round.Index() {
				delete(sv.eligibleList, k)
				delete(sv.unregisterList, k)
				continue
			}
		}
	}

	for k, v := range sv.waitList { // move validators from wait list to eligible list only after some rounds
		if v.RoundIndex+RoundsToWait < sv.round.Index() {
			// if validator already exists in the eligible list its Stake will be increased with the new register value
			if v2, ok := sv.eligibleList[k]; ok {
				v.Stake.Add(&v.Stake, &v2.Stake)
			}

			sv.eligibleList[k] = &validatorData{RoundIndex: v.RoundIndex, Stake: v.Stake}
			delete(sv.waitList, k)
		}
	}

	sv.mut.Unlock()
}

// GetEligibleList returns a list containing nodes from eligible list after a refresh action
func (sv *syncValidators) GetEligibleList() map[string]*validatorData {
	sv.refresh()

	eligibleList := make(map[string]*validatorData, 0)

	sv.mut.Lock()

	for k, v := range sv.eligibleList {
		eligibleList[k] = &validatorData{RoundIndex: v.RoundIndex, Stake: v.Stake}
	}

	sv.mut.Unlock()

	return eligibleList
}
