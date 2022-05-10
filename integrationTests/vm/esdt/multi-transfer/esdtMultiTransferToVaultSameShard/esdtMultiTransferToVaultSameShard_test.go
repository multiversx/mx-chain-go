package esdtMultiTransferToVaultSameShard

import (
	"testing"

	multitransfer "github.com/ElrondNetwork/elrond-go/integrationTests/vm/esdt/multi-transfer"
)

func TestESDTMultiTransferToVaultSameShard(t *testing.T) {
	multitransfer.EsdtMultiTransferToVault(t, false, "../../testdata/vaultV2.wasm")
}
