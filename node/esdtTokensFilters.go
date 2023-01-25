package node

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
)

type getRegisteredNftsFilter struct {
	addressBytes []byte
}

func (f *getRegisteredNftsFilter) filter(_ string, esdtData *systemSmartContracts.ESDTDataV2) bool {
	return !bytes.Equal(esdtData.TokenType, []byte(core.FungibleESDT)) && bytes.Equal(esdtData.OwnerAddress, f.addressBytes)
}

type getTokensWithRoleFilter struct {
	addressBytes []byte
	role         string
}

func (f *getTokensWithRoleFilter) filter(_ string, esdtData *systemSmartContracts.ESDTDataV2) bool {
	for _, esdtRoles := range esdtData.SpecialRoles {
		if !bytes.Equal(esdtRoles.Address, f.addressBytes) {
			continue
		}

		for _, specialRole := range esdtRoles.Roles {
			if bytes.Equal(specialRole, []byte(f.role)) {
				return true
			}
		}
	}

	return false
}

type getAllTokensRolesFilter struct {
	addressBytes []byte
	outputRoles  map[string][]string
}

func (f *getAllTokensRolesFilter) filter(tokenIdentifier string, esdtData *systemSmartContracts.ESDTDataV2) bool {
	for _, esdtRoles := range esdtData.SpecialRoles {
		if !bytes.Equal(esdtRoles.Address, f.addressBytes) {
			continue
		}

		rolesStr := make([]string, 0, len(esdtRoles.Roles))
		for _, roleBytes := range esdtRoles.Roles {
			rolesStr = append(rolesStr, string(roleBytes))
		}

		f.outputRoles[tokenIdentifier] = rolesStr
		return true
	}
	return false
}
