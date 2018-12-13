package syncValidators

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"math/big"
	"sync"
)

// RoundsToWaitToBeEligible holds the number of rounds after one node could be moved from wait list to eligible list
const RoundsToWaitToBeEligible = 5

// RoundsToWaitToBeRemoved holds the number of rounds after one node could be removed from eligible list
const RoundsToWaitToBeRemoved = 5

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
	sv.refresh()
	sv.mut.Lock()
	// if the validator is already in wait list its stake will be directly increased with the stake from the new
	// registration request, without needing to wait for more certain rounds than those from the first registration
	// request
	if v, isInWaitList := sv.waitList[nodeId]; isInWaitList {
		stake.Add(&stake, &v.Stake)
		sv.waitList[nodeId] = &validatorData{RoundIndex: v.RoundIndex, Stake: stake}
	} else {
		sv.waitList[nodeId] = &validatorData{RoundIndex: sv.round.Index(), Stake: stake}
	}
	sv.mut.Unlock()
}

// RemoveValidator adds a validator in the unregister list
func (sv *syncValidators) RemoveValidator(nodeId string) {
	sv.refresh()
	sv.mut.Lock()
	sv.unregisterList[nodeId] = &validatorData{RoundIndex: sv.round.Index()}
	sv.mut.Unlock()
}

// refresh executes a refreshing / syncing operation of the validators in the lists
func (sv *syncValidators) refresh() {
	sv.mut.Lock()
	sv.processUnregisterRequests()
	sv.processRegisterRequests()
	sv.mut.Unlock()
}

// processUnregisterRequests remove validators from wait / eligible list after an unregister request
func (sv *syncValidators) processUnregisterRequests() {
	for k, v := range sv.unregisterList {
		// if the validator is in wait list it could be removed directly
		if _, isInWaitList := sv.waitList[k]; isInWaitList {
			delete(sv.waitList, k)
		}

		// if the validator is in eligibile list it could be removed only after some certain rounds
		if _, isInEligibleList := sv.eligibleList[k]; isInEligibleList {
			if v.RoundIndex+RoundsToWaitToBeRemoved < sv.round.Index() {
				delete(sv.eligibleList, k)
			}
		}

		// if the validator does not exist anymore in both, wait and eligible list, its unregister request
		// could be deleted
		if _, isInWaitList := sv.waitList[k]; !isInWaitList {
			if _, isInEligibleList := sv.eligibleList[k]; !isInEligibleList {
				delete(sv.unregisterList, k)
			}
		}
	}
}

// processRegisterRequests move validators from wait to eligible list after a register request and some certain rounds
func (sv *syncValidators) processRegisterRequests() {
	for k, v := range sv.waitList {
		// if the validator already exists in the eligible list, its stake will be directly increased with the stake
		// from the new registration request, without needing to wait for some certain rounds, as it should be
		// already synchronized
		if v2, isInEligibleList := sv.eligibleList[k]; isInEligibleList {
			v.Stake.Add(&v.Stake, &v2.Stake)
			sv.eligibleList[k] = &validatorData{RoundIndex: v.RoundIndex, Stake: v.Stake}
			delete(sv.waitList, k)
		} else { // if the validator is not in the eligible list it should wait for some certain rounds until it
			// would be moved there
			if v.RoundIndex+RoundsToWaitToBeEligible < sv.round.Index() {
				sv.eligibleList[k] = &validatorData{RoundIndex: v.RoundIndex, Stake: v.Stake}
				delete(sv.waitList, k)
			}
		}
	}
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
