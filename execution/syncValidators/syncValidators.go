package syncValidators

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
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

	rounder chronology.Rounder
}

// NewSyncValidators creates a new syncValidators object
func NewSyncValidators(
	rounder chronology.Rounder,
) (*syncValidators, error) {
	err := checkSyncValidatorsNilParameters(rounder)

	if err != nil {
		return nil, err
	}

	sv := syncValidators{}

	sv.eligibleList = make(map[string]*validatorData, 0)
	sv.waitList = make(map[string]*validatorData, 0)
	sv.unregisterList = make(map[string]*validatorData, 0)

	sv.rounder = rounder

	return &sv, nil
}

// checkSyncValidatorsNilParameters will check the input parameters for nil values
func checkSyncValidatorsNilParameters(
	rounder chronology.Rounder,
) error {
	if rounder == nil {
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
		sv.waitList[nodeId] = &validatorData{RoundIndex: sv.rounder.Index(), Stake: stake}
	}

	sv.mut.Unlock()
}

// RemoveValidator adds a validator in the unregister list
func (sv *syncValidators) RemoveValidator(nodeId string) {
	sv.refresh()

	sv.mut.Lock()
	sv.unregisterList[nodeId] = &validatorData{RoundIndex: sv.rounder.Index()}
	sv.mut.Unlock()
}

// refresh executes a refreshing / syncing operation of the validators in the lists
func (sv *syncValidators) refresh() {
	sv.processUnregisterRequests()
	sv.processRegisterRequests()
}

// processUnregisterRequests remove validators from wait / eligible list after an unregister request
func (sv *syncValidators) processUnregisterRequests() {
	sv.mut.Lock()

	for k, v := range sv.unregisterList {
		// if the validator is in wait list it could be removed directly
		if _, isInWaitList := sv.waitList[k]; isInWaitList {
			delete(sv.waitList, k)
		}

		// if the validator is in eligible list it could be removed only after some certain rounds
		if _, isInEligibleList := sv.eligibleList[k]; isInEligibleList {
			if v.RoundIndex+RoundsToWaitToBeRemoved < sv.rounder.Index() {
				delete(sv.eligibleList, k)
			}
		}

		// if the validator does not exist anymore in both wait and eligible list, its unregister request
		// could be deleted
		if _, isInWaitList := sv.waitList[k]; !isInWaitList {
			if _, isInEligibleList := sv.eligibleList[k]; !isInEligibleList {
				delete(sv.unregisterList, k)
			}
		}
	}

	sv.mut.Unlock()
}

// processRegisterRequests move validators from wait to eligible list after a register request and a certain number
// of rounds
func (sv *syncValidators) processRegisterRequests() {
	sv.mut.Lock()

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
			if v.RoundIndex+RoundsToWaitToBeEligible < sv.rounder.Index() {
				sv.eligibleList[k] = &validatorData{RoundIndex: v.RoundIndex, Stake: v.Stake}
				delete(sv.waitList, k)
			}
		}
	}

	sv.mut.Unlock()
}

// GetEligibleList returns a list containing nodes from eligible list after a refresh action
func (sv *syncValidators) GetEligibleList() map[string]*validatorData {
	sv.refresh()

	eligibleList := make(map[string]*validatorData, 0)

	sv.mut.RLock()

	for k, v := range sv.eligibleList {
		eligibleList[k] = &validatorData{RoundIndex: v.RoundIndex, Stake: v.Stake}
	}

	sv.mut.RUnlock()

	return eligibleList
}
