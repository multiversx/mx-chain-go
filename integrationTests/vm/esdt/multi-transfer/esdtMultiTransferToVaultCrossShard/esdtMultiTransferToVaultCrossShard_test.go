package esdtMultiTransferToVaultCrossShard

import (
	"testing"

	multitransfer "github.com/ElrondNetwork/elrond-go/integrationTests/vm/esdt/multi-transfer"
)

func TestESDTMultiTransferToVaultCrossShard(t *testing.T) {
	multitransfer.EsdtMultiTransferToVault(t, true, "../../testdata/vaultV2.wasm")
}
