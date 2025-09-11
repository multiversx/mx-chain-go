package factory

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/state"
)

// StateComponentsMock -
type StateComponentsMock struct {
	PeersAcc                      state.AccountsAdapter
	Accounts                      state.AccountsAdapter
	AccountsAPI                   state.AccountsAdapter
	AccountsProposal              state.AccountsAdapter
	AccountsAdapterAPICalled      func() state.AccountsAdapter
	AccountsAdapterProposalCalled func() state.AccountsAdapter
	AccountsRepo                  state.AccountsRepository
	Tries                         common.TriesHolder
	StorageManagers               map[string]common.StorageManager
	MissingNodesNotifier          common.MissingTrieNodesNotifier
	LeavesRetriever               common.TrieLeavesRetriever
}

// NewStateComponentsMockFromRealComponent -
func NewStateComponentsMockFromRealComponent(stateComponents factory.StateComponentsHolder) *StateComponentsMock {
	return &StateComponentsMock{
		PeersAcc:             stateComponents.PeerAccounts(),
		Accounts:             stateComponents.AccountsAdapter(),
		AccountsAPI:          stateComponents.AccountsAdapterAPI(),
		AccountsProposal:     stateComponents.AccountsAdapterProposal(),
		AccountsRepo:         stateComponents.AccountsRepository(),
		Tries:                stateComponents.TriesContainer(),
		StorageManagers:      stateComponents.TrieStorageManagers(),
		MissingNodesNotifier: stateComponents.MissingTrieNodesNotifier(),
		LeavesRetriever:      stateComponents.TrieLeavesRetriever(),
	}
}

// Create -
func (scm *StateComponentsMock) Create() error {
	return nil
}

// Close -
func (scm *StateComponentsMock) Close() error {
	return nil
}

// CheckSubcomponents -
func (scm *StateComponentsMock) CheckSubcomponents() error {
	return nil
}

// PeerAccounts -
func (scm *StateComponentsMock) PeerAccounts() state.AccountsAdapter {
	return scm.PeersAcc
}

// AccountsAdapter -
func (scm *StateComponentsMock) AccountsAdapter() state.AccountsAdapter {
	return scm.Accounts
}

// AccountsAdapterAPI -
func (scm *StateComponentsMock) AccountsAdapterAPI() state.AccountsAdapter {
	if scm.AccountsAdapterAPICalled != nil {
		return scm.AccountsAdapterAPICalled()
	}
	return scm.AccountsAPI
}

// AccountsAdapterProposal -
func (scm *StateComponentsMock) AccountsAdapterProposal() state.AccountsAdapter {
	if scm.AccountsAdapterProposalCalled != nil {
		return scm.AccountsAdapterProposalCalled()
	}
	return scm.AccountsProposal
}

// AccountsRepository -
func (scm *StateComponentsMock) AccountsRepository() state.AccountsRepository {
	return scm.AccountsRepo
}

// TriesContainer -
func (scm *StateComponentsMock) TriesContainer() common.TriesHolder {
	return scm.Tries
}

// TrieStorageManagers -
func (scm *StateComponentsMock) TrieStorageManagers() map[string]common.StorageManager {
	return scm.StorageManagers
}

// String -
func (scm *StateComponentsMock) String() string {
	return "StateComponentsMock"
}

// MissingTrieNodesNotifier -
func (scm *StateComponentsMock) MissingTrieNodesNotifier() common.MissingTrieNodesNotifier {
	return scm.MissingNodesNotifier
}

// TrieLeavesRetriever -
func (scm *StateComponentsMock) TrieLeavesRetriever() common.TrieLeavesRetriever {
	return scm.LeavesRetriever
}

// IsInterfaceNil -
func (scm *StateComponentsMock) IsInterfaceNil() bool {
	return scm == nil
}
