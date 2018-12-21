package sync

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// RoundsToWaitToBeEligible holds the number of rounds after one node could be moved from wait list to eligible list
const RoundsToWaitToBeEligible = 5

// RoundsToWaitToBeRemoved holds the number of rounds after one node could be removed from eligible list
const RoundsToWaitToBeRemoved = 5

type validatorData struct {
	RoundIndex int32
	Stake      big.Int
}

// syncValidators implements a validators sync mechanism
type syncValidators struct {
	eligibleList map[string]*validatorData

	mut sync.RWMutex

	rounder  chronology.Rounder
	accounts state.AccountsAdapter
}

// NewSyncValidators creates a new syncValidators object
func NewSyncValidators(
	rounder chronology.Rounder,
	accounts state.AccountsAdapter,
) (*syncValidators, error) {
	err := checkSyncValidatorsNilParameters(rounder, accounts)
	if err != nil {
		return nil, err
	}

	sv := syncValidators{
		rounder:  rounder,
		accounts: accounts,
	}

	sv.eligibleList = make(map[string]*validatorData, 0)

	return &sv, nil
}

// checkSyncValidatorsNilParameters method will check the input parameters for nil values
func checkSyncValidatorsNilParameters(
	rounder chronology.Rounder,
	accounts state.AccountsAdapter,
) error {
	if rounder == nil {
		return process.ErrNilRound
	}

	if accounts == nil {
		return process.ErrNilAccountsAdapter
	}

	return nil
}

// refresh method executes a refreshing / syncing operation of the validators in the eligible list
func (sv *syncValidators) refresh() error {
	account, err := sv.accounts.GetJournalizedAccount(state.RegistrationAddress)
	if err != nil {
		return err
	}

	regsData := account.BaseAccount().RegistrationData
	if regsData == nil {
		return nil
	}

	sv.mut.Lock()
	sv.eligibleList = make(map[string]*validatorData, 0)
	sv.mut.Unlock()

	for _, regData := range regsData {
		switch regData.Action {
		case state.ArRegister:
			sv.processRegisterRequests(&regData)
		case state.ArUnregister:
			sv.processUnregisterRequests(&regData)
		}
	}

	return nil
}

// processRegisterRequests method adds validators in eligible list after a register request and a certain number
// of rounds
func (sv *syncValidators) processRegisterRequests(regData *state.RegistrationData) {
	sv.mut.Lock()

	// if the validator already exists in the eligible list, its stake will be directly increased with the stake
	// from the new registration request, without needing to wait for some certain rounds, as it should be
	// already synchronized
	k := string(regData.NodePubKey)
	if v, isInEligibleList := sv.eligibleList[k]; isInEligibleList {
		sv.eligibleList[k] = &validatorData{
			RoundIndex: v.RoundIndex,
			Stake:      *big.NewInt(0).Add(&v.Stake, &regData.Stake),
		}
	} else { // if the validator is not in the eligible list it should wait for some certain rounds until it
		// would be moved there
		if regData.RoundIndex+RoundsToWaitToBeEligible < sv.rounder.Index() {
			sv.eligibleList[k] = &validatorData{
				RoundIndex: regData.RoundIndex,
				Stake:      *big.NewInt(0).Set(&regData.Stake),
			}
		}
	}

	sv.mut.Unlock()
}

// processUnregisterRequests method removes validators from eligible list after an unregister request and a certain
// number of rounds
func (sv *syncValidators) processUnregisterRequests(regData *state.RegistrationData) {
	sv.mut.Lock()

	// if the validator is in eligible list it could be removed only after some certain rounds
	k := string(regData.NodePubKey)
	if _, isInEligibleList := sv.eligibleList[k]; isInEligibleList {
		if regData.RoundIndex+RoundsToWaitToBeRemoved < sv.rounder.Index() {
			delete(sv.eligibleList, k)
		}
	}

	sv.mut.Unlock()
}

// GetEligibleList method returns a list containing nodes from eligible list after a refresh action
func (sv *syncValidators) GetEligibleList() map[string]*validatorData {
	log.LogIfError(sv.refresh())

	eligibleList := make(map[string]*validatorData, 0)

	sv.mut.RLock()

	for k, v := range sv.eligibleList {
		eligibleList[k] = &validatorData{RoundIndex: v.RoundIndex, Stake: *big.NewInt(0).Set(&v.Stake)}
	}

	sv.mut.RUnlock()

	return eligibleList
}
