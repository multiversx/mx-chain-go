package transferandexecutefailable

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
	"github.com/stretchr/testify/require"
)

// forw_raw_transf_exec_fallible_egld_reject.scen
func TestForwRawTransfExecFallibleEgldReject(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	net, owner, forwarderAddr, vaultAddr := DeployAndSetupNet(t, 100000000000000)
	defer net.Close()

	txData := txDataBuilder.NewBuilder()
	txData.Func("transfer_execute_fallible")
	txData.Bytes(vaultAddr)
	txData.Str("reject_funds")
	txData.Str("EGLD-000000").Int64(0).BigInt(big.NewInt(100))

	tx := net.CreateTxUint64(owner, forwarderAddr, 1000, txData.ToBytes())
	tx.GasLimit = net.MaxGasLimit
	net.SignAndSendTx(owner, tx)
	net.Steps(5)

	CheckAddressHasBalance(t, net, forwarderAddr, big.NewInt(1000))
	CheckAddressHasBalance(t, net, vaultAddr, big.NewInt(0))
}

func DeployAndSetupNet(t *testing.T, initialVal uint64) (*integrationTests.TestNetwork, *integrationTests.TestWalletAccount, []byte, []byte) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	net := integrationTests.NewTestNetworkSized(t, 2, 1, 1)
	net.Start()

	net.MintNodeAccountsUint64(initialVal)
	net.Step()

	net.CreateWallets(1)
	net.MintWalletsUint64(initialVal)

	owner := net.Wallets[0]

	forwarderAddr := net.DeployPayableSC(owner, "../testdata/forwarder-barnard.wasm")
	vaultAddr := net.DeployPayableSC(owner, "../testdata/vault-barnard.wasm")

	net.Step()

	return net, owner, forwarderAddr, vaultAddr
}

func CheckAddressHasBalance(t *testing.T, net *integrationTests.TestNetwork, address []byte, balance *big.Int) {
	acc := net.GetAccountHandler(address)
	accountBalance := acc.GetBalance().Uint64()
	require.Equal(t, accountBalance, balance.Uint64())
}
