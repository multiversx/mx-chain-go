package hooks

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	errDRWANilMigrationManifest       = errors.New("nil DRWA migration manifest")
	errDRWAEmptyMigrationTokenID      = errors.New("empty DRWA migration token id")
	errDRWAEmptyMigrationPolicyBody   = errors.New("empty DRWA migration policy body")
	errDRWAEmptyMigrationHolder       = errors.New("empty DRWA migration holder")
	errDRWADuplicateMigrationHolder   = errors.New("duplicate DRWA migration holder")
	errDRWAEmptyMigrationHolderBody   = errors.New("empty DRWA migration holder body")
	errDRWAInvalidMigrationVersion    = errors.New("invalid DRWA migration version")
	errDRWAMigrationTooManyOperations = errors.New("DRWA migration exceeds sync operation limit")
	errDRWAInvalidAuthorizedCaller    = errors.New("invalid DRWA authorized caller")
	errDRWANilMigrationCallerSink     = errors.New("nil DRWA migration caller sink")
)

type drwaMigrationHolder struct {
	Address string `json:"address"`
	Version uint64 `json:"version"`
	Body    []byte `json:"body"`
}

type drwaMigrationManifest struct {
	TokenID           string                `json:"token_id"`
	PolicyVersion     uint64                `json:"policy_version"`
	PolicyBody        []byte                `json:"policy_body"`
	Holders           []drwaMigrationHolder `json:"holders"`
	AuthorizedCallers map[string]string     `json:"authorized_callers,omitempty"`
}

type drwaMigrationSnapshot struct {
	TokenID              string
	TokenPolicy          *drwaSyncStoredValue
	Holders              map[string]*drwaSyncStoredValue
	RequestedHolderState map[string]drwaMigrationHolder
}

type drwaMigrationStateReader interface {
	drwaSyncStateAdapter
	GetTokenPolicyStored(tokenID string) (*drwaSyncStoredValue, error)
	GetHolderMirrorStored(tokenID, holder string) (*drwaSyncStoredValue, error)
}

type drwaMigrationAuthorizedCallerSink interface {
	PutAuthorizedCallerAddress(domain string, address []byte) error
}

func validateDRWAMigrationManifest(manifest *drwaMigrationManifest) error {
	if manifest == nil {
		return errDRWANilMigrationManifest
	}
	if strings.TrimSpace(manifest.TokenID) == "" {
		return errDRWAEmptyMigrationTokenID
	}
	if manifest.PolicyVersion == 0 {
		return errDRWAInvalidMigrationVersion
	}
	if len(manifest.PolicyBody) == 0 {
		return errDRWAEmptyMigrationPolicyBody
	}

	seen := make(map[string]struct{}, len(manifest.Holders))
	for _, holder := range manifest.Holders {
		addr := strings.TrimSpace(holder.Address)
		if addr == "" {
			return errDRWAEmptyMigrationHolder
		}
		if holder.Version == 0 {
			return errDRWAInvalidMigrationVersion
		}
		if len(holder.Body) == 0 {
			return errDRWAEmptyMigrationHolderBody
		}
		if _, exists := seen[addr]; exists {
			return fmt.Errorf("%w: %s", errDRWADuplicateMigrationHolder, addr)
		}
		seen[addr] = struct{}{}
	}

	if 1+len(manifest.Holders) > drwaSyncMaxOperations {
		return errDRWAMigrationTooManyOperations
	}

	if len(manifest.AuthorizedCallers) > 0 {
		requiredDomains := []string{
			drwaSyncCallerPolicyRegistry,
			drwaSyncCallerAssetManager,
			drwaSyncCallerRecoveryAdmin,
		}
		for _, domain := range requiredDomains {
			address, normalizeErr := normalizeDRWAAuthorizedCallerAddress(manifest.AuthorizedCallers[domain])
			if normalizeErr != nil || len(address) == 0 {
				return fmt.Errorf("%w: %s", errDRWAInvalidAuthorizedCaller, domain)
			}
		}
	}

	return nil
}

func buildDRWAMigrationEnvelope(manifest *drwaMigrationManifest) (*drwaSyncEnvelope, error) {
	err := validateDRWAMigrationManifest(manifest)
	if err != nil {
		return nil, err
	}

	sortedHolders := append([]drwaMigrationHolder(nil), manifest.Holders...)
	sort.Slice(sortedHolders, func(i, j int) bool {
		return sortedHolders[i].Address < sortedHolders[j].Address
	})

	operations := make([]drwaSyncOperation, 0, 1+len(sortedHolders))
	operations = append(operations, drwaSyncOperation{
		OperationType: drwaSyncOpTokenPolicy,
		TokenID:       manifest.TokenID,
		Version:       manifest.PolicyVersion,
		Body:          manifest.PolicyBody,
	})

	for _, holder := range sortedHolders {
		operations = append(operations, drwaSyncOperation{
			OperationType: drwaSyncOpHolderMirror,
			TokenID:       manifest.TokenID,
			Holder:        holder.Address,
			Version:       holder.Version,
			Body:          holder.Body,
		})
	}

	hash, err := computeDRWASyncHash(drwaSyncCallerRecoveryAdmin, operations)
	if err != nil {
		return nil, err
	}

	return &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerRecoveryAdmin,
		PayloadHash:  hash,
		Operations:   operations,
	}, nil
}

func captureDRWAMigrationSnapshot(adapter drwaMigrationStateReader, manifest *drwaMigrationManifest) (*drwaMigrationSnapshot, error) {
	err := validateDRWAMigrationManifest(manifest)
	if err != nil {
		return nil, err
	}

	tokenPolicy, err := adapter.GetTokenPolicyStored(manifest.TokenID)
	if err != nil {
		return nil, err
	}

	holders := make(map[string]*drwaSyncStoredValue, len(manifest.Holders))
	for _, holder := range manifest.Holders {
		stored, readErr := adapter.GetHolderMirrorStored(manifest.TokenID, holder.Address)
		if readErr != nil {
			return nil, readErr
		}
		holders[holder.Address] = stored
	}

	return &drwaMigrationSnapshot{
		TokenID:     manifest.TokenID,
		TokenPolicy: tokenPolicy,
		Holders:     holders,
		RequestedHolderState: func() map[string]drwaMigrationHolder {
			requested := make(map[string]drwaMigrationHolder, len(manifest.Holders))
			for _, holder := range manifest.Holders {
				requested[holder.Address] = holder
			}
			return requested
		}(),
	}, nil
}

func buildDRWARollbackEnvelope(snapshot *drwaMigrationSnapshot) (*drwaSyncEnvelope, error) {
	if snapshot == nil {
		return nil, errors.New("nil DRWA migration snapshot")
	}

	operations := make([]drwaSyncOperation, 0, 1+len(snapshot.Holders))
	if snapshot.TokenPolicy != nil {
		operations = append(operations, drwaSyncOperation{
			OperationType: drwaSyncOpTokenPolicy,
			TokenID:       snapshot.TokenID,
			Version:       snapshot.TokenPolicy.Version,
			Body:          snapshot.TokenPolicy.Body,
		})
	}

	holderAddresses := make([]string, 0, len(snapshot.Holders))
	for holder := range snapshot.Holders {
		holderAddresses = append(holderAddresses, holder)
	}
	sort.Strings(holderAddresses)

	for _, holder := range holderAddresses {
		stored := snapshot.Holders[holder]
		if stored == nil {
			requested, ok := snapshot.RequestedHolderState[holder]
			if !ok {
				continue
			}
			operations = append(operations, drwaSyncOperation{
				OperationType: drwaSyncOpHolderMirrorDelete,
				TokenID:       snapshot.TokenID,
				Holder:        holder,
				Version:       requested.Version + 1,
			})
			continue
		}
		operations = append(operations, drwaSyncOperation{
			OperationType: drwaSyncOpHolderMirror,
			TokenID:       snapshot.TokenID,
			Holder:        holder,
			Version:       stored.Version,
			Body:          stored.Body,
		})
	}

	hash, err := computeDRWASyncHash(drwaSyncCallerRecoveryAdmin, operations)
	if err != nil {
		return nil, err
	}

	return &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerRecoveryAdmin,
		PayloadHash:  hash,
		Operations:   operations,
	}, nil
}

func persistDRWAMigrationAuthorizedCallers(
	sink drwaMigrationAuthorizedCallerSink,
	manifest *drwaMigrationManifest,
) error {
	if sink == nil {
		return errDRWANilMigrationCallerSink
	}

	err := validateDRWAMigrationManifest(manifest)
	if err != nil {
		return err
	}

	if len(manifest.AuthorizedCallers) == 0 {
		return nil
	}

	requiredDomains := []string{
		drwaSyncCallerAssetManager,
		drwaSyncCallerPolicyRegistry,
		drwaSyncCallerRecoveryAdmin,
	}
	for _, domain := range requiredDomains {
		address, normalizeErr := normalizeDRWAAuthorizedCallerAddress(manifest.AuthorizedCallers[domain])
		if normalizeErr != nil {
			return fmt.Errorf("%w: %s", errDRWAInvalidAuthorizedCaller, domain)
		}
		err = sink.PutAuthorizedCallerAddress(domain, address)
		if err != nil {
			return err
		}
	}

	return nil
}

func normalizeDRWAAuthorizedCallerAddress(value string) ([]byte, error) {
	address := strings.TrimSpace(value)
	if address == "" {
		return nil, errDRWAInvalidAuthorizedCaller
	}

	if len(address) == 32 {
		raw := []byte(address)
		if isAllZeroAddress(raw) {
			return nil, errDRWAInvalidAuthorizedCaller
		}
		return raw, nil
	}

	trimmedHex := strings.TrimPrefix(address, "0x")
	trimmedHex = strings.TrimPrefix(trimmedHex, "0X")
	decoded, err := hex.DecodeString(trimmedHex)
	if err == nil && len(decoded) == 32 {
		if isAllZeroAddress(decoded) {
			return nil, errDRWAInvalidAuthorizedCaller
		}
		return decoded, nil
	}

	return nil, errDRWAInvalidAuthorizedCaller
}

func isAllZeroAddress(addr []byte) bool {
	for _, b := range addr {
		if b != 0 {
			return false
		}
	}
	return true
}
