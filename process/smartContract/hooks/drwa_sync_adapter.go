package hooks

import (
	"encoding/json"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type drwaHookStateAdapter struct {
	accounts state.AccountsAdapter
}

func newDRWAHookStateAdapter(accounts state.AccountsAdapter) *drwaHookStateAdapter {
	return &drwaHookStateAdapter{
		accounts: accounts,
	}
}

func buildDRWATokenPolicyKey(tokenIdentifier []byte) []byte {
	return []byte(drwaSyncTokenPolicyPrefix + string(tokenIdentifier) + ":policy")
}

func buildDRWAHolderMirrorKey(tokenIdentifier []byte, address []byte) []byte {
	return []byte(drwaSyncHolderMirrorPrefix + string(tokenIdentifier) + ":" + string(address))
}

func buildDRWAAuthorizedCallerKey(domain string) []byte {
	return []byte("drwa:auth:" + domain)
}

// drwaSyncEvidencePrefix is the storage namespace for all DRWA audit evidence.
// Using a dedicated prefix keeps evidence keys out of the token-policy namespace
// and allows targeted enumeration during forensic inspection.
const drwaSyncEvidencePrefix = "drwa:evidence:"

func buildDRWARecoveryEvidenceKey(tokenIdentifier []byte) []byte {
	return []byte(drwaSyncEvidencePrefix + string(tokenIdentifier) + ":recovery:latest")
}

func buildDRWARecoveryEvidenceHistoryKey(tokenIdentifier []byte, payloadHash []byte) []byte {
	return []byte(fmt.Sprintf("%s%s:recovery:history:%x", drwaSyncEvidencePrefix, string(tokenIdentifier), payloadHash))
}

func buildDRWARolloutEvidenceKey(tokenIdentifier []byte) []byte {
	return []byte(drwaSyncEvidencePrefix + string(tokenIdentifier) + ":rollout:latest")
}

func buildDRWARolloutEvidenceHistoryKey(tokenIdentifier []byte, payloadHash []byte) []byte {
	return []byte(fmt.Sprintf("%s%s:rollout:history:%x", drwaSyncEvidencePrefix, string(tokenIdentifier), payloadHash))
}

func buildDRWARolloutVerificationKey(tokenIdentifier []byte) []byte {
	return []byte(drwaSyncEvidencePrefix + string(tokenIdentifier) + ":rollout:verification:latest")
}

func buildDRWARolloutVerificationHistoryKey(tokenIdentifier []byte, payloadHash []byte) []byte {
	return []byte(fmt.Sprintf("%s%s:rollout:verification:history:%x", drwaSyncEvidencePrefix, string(tokenIdentifier), payloadHash))
}

func buildDRWAHolderDeleteAuditKey(tokenIdentifier []byte, address []byte, version uint64) []byte {
	return []byte(fmt.Sprintf("%s%s:holder-delete:%s:%d", drwaSyncEvidencePrefix, string(tokenIdentifier), string(address), version))
}

type drwaDeleteAuditRecord struct {
	TokenID string `json:"token_id"`
	Holder  string `json:"holder"`
	Version uint64 `json:"version"`
}

func (d *drwaHookStateAdapter) GetTokenPolicyVersion(tokenID string) (uint64, error) {
	systemAccount, err := d.getUserAccount(core.SystemAccountAddress)
	if err != nil {
		return 0, err
	}

	storedValue, err := d.readStoredValue(systemAccount, buildDRWATokenPolicyKey([]byte(tokenID)))
	if err != nil || storedValue == nil {
		return 0, err
	}

	return storedValue.Version, nil
}

func (d *drwaHookStateAdapter) GetHolderMirrorVersion(tokenID, holder string) (uint64, error) {
	holderAccount, err := d.getUserAccount([]byte(holder))
	if err != nil {
		return 0, err
	}

	storedValue, err := d.readStoredValue(holderAccount, buildDRWAHolderMirrorKey([]byte(tokenID), []byte(holder)))
	if err != nil || storedValue == nil {
		return 0, err
	}

	return storedValue.Version, nil
}

func (d *drwaHookStateAdapter) GetAuthorizedCallerAddress(domain string) ([]byte, error) {
	systemAccount, err := d.getUserAccount(core.SystemAccountAddress)
	if err != nil {
		return nil, err
	}

	value, _, err := systemAccount.AccountDataHandler().RetrieveValue(buildDRWAAuthorizedCallerKey(domain))
	if err != nil {
		return nil, err
	}
	if len(value) == 0 {
		return nil, nil
	}

	return append([]byte(nil), value...), nil
}

func (d *drwaHookStateAdapter) PutAuthorizedCallerAddress(domain string, address []byte) error {
	systemAccount, err := d.getUserAccount(core.SystemAccountAddress)
	if err != nil {
		return err
	}

	err = systemAccount.AccountDataHandler().SaveKeyValue(buildDRWAAuthorizedCallerKey(domain), append([]byte(nil), address...))
	if err != nil {
		return err
	}

	return d.accounts.SaveAccount(systemAccount)
}

func (d *drwaHookStateAdapter) PutTokenPolicyBody(tokenID string, version uint64, body []byte) error {
	systemAccount, err := d.getUserAccount(core.SystemAccountAddress)
	if err != nil {
		return err
	}

	return d.writeStoredValue(systemAccount, buildDRWATokenPolicyKey([]byte(tokenID)), version, body)
}

func (d *drwaHookStateAdapter) PutHolderMirrorBody(tokenID, holder string, version uint64, body []byte) error {
	holderAccount, err := d.getUserAccount([]byte(holder))
	if err != nil {
		return err
	}

	return d.writeStoredValue(holderAccount, buildDRWAHolderMirrorKey([]byte(tokenID), []byte(holder)), version, body)
}

func (d *drwaHookStateAdapter) DeleteHolderMirror(tokenID, holder string, version uint64) error {
	holderAccount, err := d.getUserAccount([]byte(holder))
	if err != nil {
		return err
	}

	err = holderAccount.AccountDataHandler().SaveKeyValue(buildDRWAHolderMirrorKey([]byte(tokenID), []byte(holder)), nil)
	if err != nil {
		return err
	}

	err = d.accounts.SaveAccount(holderAccount)
	if err != nil {
		return err
	}

	return d.persistHolderDeleteAudit(tokenID, holder, version)
}

func (d *drwaHookStateAdapter) PersistRecoveryEvidence(tokenID string, payload []byte) error {
	systemAccount, err := d.getUserAccount(core.SystemAccountAddress)
	if err != nil {
		return err
	}

	return d.persistArtifact(systemAccount, buildDRWARecoveryEvidenceKey([]byte(tokenID)), buildDRWARecoveryEvidenceHistoryKey, tokenID, payload)
}

func (d *drwaHookStateAdapter) PersistRolloutEvidence(tokenID string, payload []byte) error {
	systemAccount, err := d.getUserAccount(core.SystemAccountAddress)
	if err != nil {
		return err
	}

	return d.persistArtifact(systemAccount, buildDRWARolloutEvidenceKey([]byte(tokenID)), buildDRWARolloutEvidenceHistoryKey, tokenID, payload)
}

func (d *drwaHookStateAdapter) PersistRolloutVerification(tokenID string, payload []byte) error {
	systemAccount, err := d.getUserAccount(core.SystemAccountAddress)
	if err != nil {
		return err
	}

	return d.persistArtifact(systemAccount, buildDRWARolloutVerificationKey([]byte(tokenID)), buildDRWARolloutVerificationHistoryKey, tokenID, payload)
}

func (d *drwaHookStateAdapter) Snapshot() int {
	return d.accounts.JournalLen()
}

func (d *drwaHookStateAdapter) Rollback(snapshot int) error {
	return d.accounts.RevertToSnapshot(snapshot)
}

func (d *drwaHookStateAdapter) readStoredValue(account vmcommon.UserAccountHandler, key []byte) (*drwaSyncStoredValue, error) {
	value, _, err := account.AccountDataHandler().RetrieveValue(key)
	if err != nil {
		return nil, err
	}
	if len(value) == 0 {
		return nil, nil
	}

	storedValue := &drwaSyncStoredValue{}
	err = json.Unmarshal(value, storedValue)
	if err != nil {
		return nil, err
	}

	return storedValue, nil
}

func (d *drwaHookStateAdapter) writeStoredValue(account vmcommon.UserAccountHandler, key []byte, version uint64, body []byte) error {
	payload, err := json.Marshal(&drwaSyncStoredValue{
		Version: version,
		Body:    body,
	})
	if err != nil {
		return err
	}

	err = account.AccountDataHandler().SaveKeyValue(key, payload)
	if err != nil {
		return err
	}

	return d.accounts.SaveAccount(account)
}

func (d *drwaHookStateAdapter) getUserAccount(address []byte) (vmcommon.UserAccountHandler, error) {
	accountHandler, err := d.accounts.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	userAccount, ok := accountHandler.(vmcommon.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return userAccount, nil
}

func (d *drwaHookStateAdapter) persistArtifact(
	systemAccount vmcommon.UserAccountHandler,
	latestKey []byte,
	historyKeyBuilder func([]byte, []byte) []byte,
	tokenID string,
	payload []byte,
) error {
	payloadCopy := append([]byte(nil), payload...)
	payloadHash := keccak.NewKeccak().Compute(string(payloadCopy))

	err := systemAccount.AccountDataHandler().SaveKeyValue(latestKey, payloadCopy)
	if err != nil {
		return err
	}

	err = systemAccount.AccountDataHandler().SaveKeyValue(historyKeyBuilder([]byte(tokenID), payloadHash), payloadCopy)
	if err != nil {
		return err
	}

	return d.accounts.SaveAccount(systemAccount)
}

func (d *drwaHookStateAdapter) persistHolderDeleteAudit(tokenID, holder string, version uint64) error {
	systemAccount, err := d.getUserAccount(core.SystemAccountAddress)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(&drwaDeleteAuditRecord{
		TokenID: tokenID,
		Holder:  holder,
		Version: version,
	})
	if err != nil {
		return err
	}

	err = systemAccount.AccountDataHandler().SaveKeyValue(buildDRWAHolderDeleteAuditKey([]byte(tokenID), []byte(holder), version), payload)
	if err != nil {
		return err
	}

	return d.accounts.SaveAccount(systemAccount)
}
