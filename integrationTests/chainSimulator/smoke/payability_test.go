package smoke

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestSupernovaSmokeContractPayability(t *testing.T) {
	f := newFixture(t)
	owner, user, remoteUser := f.wallet(0), f.wallet(0), f.wallet(1)
	forwarder, _ := f.deployArtifact(owner, "testdata/promise-probe.wasm", "0506", "")
	for _, mode := range []struct {
		name                 string
		payable, payableBySC bool
	}{
		{"non-payable", false, false},
		{"user-payable", true, false},
		{"SC-only-payable", false, true},
	} {
		t.Run(mode.name, func(t *testing.T) {
			f := &fixture{t: t, cs: f.cs}
			metadata := &vmcommon.CodeMetadata{Upgradeable: true, Readable: true, Payable: mode.payable, PayableBySC: mode.payableBySC}
			contract, _ := f.deployArtifact(owner, "../relayedTx/testData/adder.wasm", hexArg(metadata.ToBytes()), "@00")
			for i, sender := range []*integrationTests.TestWalletAccount{user, remoteUser, user} {
				t.Run(fmt.Sprintf("origin-%s", []string{"user-same-shard", "user-cross-shard", "contract-same-shard"}[i]), func(t *testing.T) {
					f := &fixture{t: t, cs: f.cs}
					receiver, data := contract, ""
					fromContract := i == 2
					if fromContract {
						receiver, data = forwarder, "send@"+hexArg(contract)
					}
					beforeSender, beforeContract, beforeForwarder := f.account(sender.Address), f.account(contract), f.account(forwarder)
					result := f.execute(f.tx(sender, receiver, 100, data, 10_000_000))
					moved := big.NewInt(0)
					if mode.payable || (fromContract && mode.payableBySC) {
						requireExecutionSuccess(t, result)
						moved.SetInt64(100)
					} else {
						requireExecutionFailure(t, result)
					}
					expectedSender := new(big.Int).Sub(amount(t, beforeSender.Balance), amount(t, result.Fee))
					expectedSender.Sub(expectedSender, moved)
					require.Equal(t, expectedSender.String(), f.account(sender.Address).Balance)
					require.Equal(t, beforeSender.Nonce+1, f.account(sender.Address).Nonce)
					require.Equal(t, new(big.Int).Add(amount(t, beforeContract.Balance), moved).String(), f.account(contract).Balance)
					// Pure payment must not execute or mutate the recipient's storage.
					require.Equal(t, beforeContract.RootHash, f.account(contract).RootHash)
					require.Equal(t, beforeForwarder.Balance, f.account(forwarder).Balance)
					require.Equal(t, beforeForwarder.RootHash, f.account(forwarder).RootHash)
				})
			}
		})
	}
}
