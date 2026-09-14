package smoke

import (
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/stretchr/testify/require"
)

type tokenPayment struct {
	token string
	nonce uint64
}

// Extend the existing issuance/transfer scenarios; no second token fixture.
func (f *fixture) tokenContractPayments(sender, remoteOwner *integrationTests.TestWalletAccount, payments []tokenPayment, multi bool) []byte {
	f.t.Helper()
	contract, _ := f.deployArtifact(remoteOwner, "../../vm/txsFee/testdata/forwarderQueue/vault-promises.wasm", "0506", "")
	contractAccount := &integrationTests.TestWalletAccount{Address: contract}
	relayer := f.wallet(f.cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(sender.Address))
	for _, scenario := range []struct {
		name             string
		success, relayed bool
	}{
		{"failed endpoint refunds tokens", false, false},
		{"failed relayed endpoint refunds tokens", false, true},
		{"successful payment", true, !multi},
	} {
		f.t.Run(scenario.name, func(t *testing.T) {
			f := &fixture{t: t, cs: f.cs}
			endpoint := "missing_endpoint"
			if scenario.success {
				endpoint = "accept_funds"
			}
			receiver := contract
			data := "ESDTTransfer@" + hexArg([]byte(payments[0].token)) + "@01@" + hexArg([]byte(endpoint))
			if multi {
				receiver = sender.Address
				parts := []string{"MultiESDTNFTTransfer", hexArg(contract), hexArg(big.NewInt(int64(len(payments))).Bytes())}
				for _, payment := range payments {
					parts = append(parts, hexArg([]byte(payment.token)), hexArg(new(big.Int).SetUint64(payment.nonce).Bytes()), "01")
				}
				parts = append(parts, hexArg([]byte(endpoint)))
				data = strings.Join(parts, "@")
			}
			tx := f.tx(sender, receiver, 0, data, 100_000_000)
			if scenario.relayed {
				tx.RelayerAddr = relayer.Address
				f.sign(tx, sender, nil, relayer)
				validSignature := append([]byte(nil), tx.RelayerSignature...)
				tx.RelayerSignature[0] ^= 1
				f.rejected(tx, sender.Address, relayer.Address, contract)
				tx.RelayerSignature = validSignature
			}
			beforeSender, beforeRelayer := f.account(sender.Address), f.account(relayer.Address)
			senderTokens, contractTokens := make([]*big.Int, len(payments)), make([]*big.Int, len(payments))
			for i, payment := range payments {
				senderTokens[i] = f.tokenBalance(sender, payment.token, payment.nonce)
				contractTokens[i] = f.tokenBalance(contractAccount, payment.token, payment.nonce)
			}
			result := f.execute(tx)
			execution := result
			if multi {
				// Multi-transfer is addressed to the sender: its successful source
				// transaction is not the cross-shard contract execution outcome.
				var err error
				result, err = f.cs.GetNodeHandler(f.cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(sender.Address)).GetFacadeHandler().GetTransaction(result.Hash, true)
				require.NoError(t, err)
				var remote *transaction.ApiTransactionResult
				for _, scr := range result.SmartContractResults {
					if scr.RcvAddr != f.address(contract).Bech32 {
						continue
					}
					require.Nil(t, remote, "exactly one contract execution")
					remote, err = f.cs.GetNodeHandler(f.cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(contract)).GetFacadeHandler().GetTransaction(scr.Hash, true)
					require.NoError(t, err)
				}
				require.NotNil(t, remote, "cross-shard contract SCR must execute")
				execution = remote
			}
			if scenario.success {
				requireExecutionSuccess(t, execution)
			} else {
				requireExecutionFailure(t, execution)
			}
			for i, payment := range payments {
				moved := int64(0)
				if scenario.success {
					moved = 1
				}
				require.Zero(t, new(big.Int).Sub(senderTokens[i], big.NewInt(moved)).Cmp(f.tokenBalance(sender, payment.token, payment.nonce)), payment.token+"/"+strconv.FormatUint(payment.nonce, 10))
				require.Zero(t, new(big.Int).Add(contractTokens[i], big.NewInt(moved)).Cmp(f.tokenBalance(contractAccount, payment.token, payment.nonce)))
			}
			payerBefore, payer := beforeSender, sender
			if scenario.relayed {
				payerBefore, payer = beforeRelayer, relayer
				require.Equal(t, beforeSender.Balance, f.account(sender.Address).Balance)
			}
			require.Equal(t, new(big.Int).Sub(amount(t, payerBefore.Balance), amount(t, result.Fee)).String(), f.account(payer.Address).Balance)
			require.Equal(t, beforeSender.Nonce+1, f.account(sender.Address).Nonce)
			require.Equal(t, beforeRelayer.Nonce, f.account(relayer.Address).Nonce)
		})
	}
	return contract
}
