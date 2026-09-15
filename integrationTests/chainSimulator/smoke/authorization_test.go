package smoke

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	apiData "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/integrationTests"
	tokenTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/vm"
)

func TestSupernovaSmokeSignedEGLD(t *testing.T) {
	f := newFixture(t)
	sender := f.wallet(0)
	for _, shard := range []uint32{0, 1} {
		t.Run(fmt.Sprintf("receiver-shard-%d", shard), func(t *testing.T) {
			f := &fixture{t: t, cs: f.cs}
			receiver := f.wallet(shard)
			for _, invalid := range []string{"missing-signature", "wrong-signature", "wrong-chain"} {
				tx := f.tx(sender, receiver.Address, 100, "", 50_000)
				switch invalid {
				case "missing-signature":
					tx.Signature = nil
				case "wrong-signature":
					tx.Signature[0] ^= 1
				case "wrong-chain":
					tx.ChainID = []byte("another-chain")
					f.sign(tx, sender, nil, nil)
				}
				f.rejected(tx, sender.Address, receiver.Address)
			}
			beforeSender, beforeReceiver := f.account(sender.Address), f.account(receiver.Address)
			tx := f.tx(sender, receiver.Address, 100, "", 50_000)
			result := f.success(tx)
			require.Equal(t, new(big.Int).Sub(amount(t, beforeSender.Balance), new(big.Int).Add(big.NewInt(100), amount(t, result.Fee))).String(), f.account(sender.Address).Balance)
			require.Equal(t, new(big.Int).Add(amount(t, beforeReceiver.Balance), big.NewInt(100)).String(), f.account(receiver.Address).Balance)
			require.Equal(t, beforeSender.Nonce+1, f.account(sender.Address).Nonce)
			// An already executed signed transaction must not be charged or credited again.
			f.rejected(tx, sender.Address, receiver.Address)
		})
	}
}

func TestSupernovaSmokeSignedRelayed(t *testing.T) {
	f := newFixture(t)
	sender, relayer, receiver := f.wallet(0), f.wallet(0), f.wallet(1)
	for _, invalid := range []string{"sender", "relayer"} {
		tx := f.tx(sender, receiver.Address, 100, "", 100_000)
		tx.RelayerAddr = relayer.Address
		f.sign(tx, sender, nil, relayer)
		if invalid == "sender" {
			tx.Signature[0] ^= 1
		} else {
			tx.RelayerSignature[0] ^= 1
		}
		f.rejected(tx, sender.Address, relayer.Address, receiver.Address)
	}
	beforeSender, beforeRelayer, beforeReceiver := f.account(sender.Address), f.account(relayer.Address), f.account(receiver.Address)
	tx := f.tx(sender, receiver.Address, 100, "", 100_000)
	tx.RelayerAddr = relayer.Address
	f.sign(tx, sender, nil, relayer)
	result := f.success(tx)
	require.Equal(t, new(big.Int).Sub(amount(t, beforeSender.Balance), big.NewInt(100)).String(), f.account(sender.Address).Balance)
	require.Equal(t, new(big.Int).Sub(amount(t, beforeRelayer.Balance), amount(t, result.Fee)).String(), f.account(relayer.Address).Balance)
	require.Equal(t, new(big.Int).Add(amount(t, beforeReceiver.Balance), big.NewInt(100)).String(), f.account(receiver.Address).Balance)
	require.Equal(t, beforeSender.Nonce+1, f.account(sender.Address).Nonce)
	require.Equal(t, beforeRelayer.Nonce, f.account(relayer.Address).Nonce)
}

func TestSupernovaSmokeSignedGuardian(t *testing.T) {
	f := newFixture(t)
	sender, guardian, receiver := f.wallet(0), f.wallet(0), f.wallet(1)
	f.success(f.tx(sender, sender.Address, 0, "SetGuardian@"+hexArg(guardian.Address)+"@"+hexArg([]byte("smoke")), 5_000_000))
	guardianData := func() apiData.GuardianData {
		data, _, err := f.cs.GetNodeHandler(0).GetFacadeHandler().GetGuardianData(f.address(sender.Address).Bech32, apiData.AccountQueryOptions{})
		require.NoError(t, err)
		return data
	}
	pending := guardianData()
	require.Nil(t, pending.ActiveGuardian)
	require.NotNil(t, pending.PendingGuardian)
	require.False(t, pending.Guarded)
	requireExecutionFailure(t, f.execute(f.tx(sender, sender.Address, 0, "GuardAccount", 5_000_000)))
	require.False(t, guardianData().Guarded)
	// Activate through real epoch progression and inspect the account API.
	f.until(100, func() bool { return guardianData().ActiveGuardian != nil })
	require.Equal(t, f.address(guardian.Address).Bech32, guardianData().ActiveGuardian.Address)
	f.success(f.tx(sender, sender.Address, 0, "GuardAccount", 5_000_000))
	for _, invalid := range []string{"missing-guardian", "wrong-guardian"} {
		tx := f.tx(sender, receiver.Address, 100, "", 100_000)
		if invalid == "wrong-guardian" {
			tx.Options = 2
			tx.GuardianAddr = guardian.Address
			f.sign(tx, sender, guardian, nil)
			tx.GuardianSignature[0] ^= 1
		}
		f.rejected(tx, sender.Address, guardian.Address, receiver.Address)
	}
	tx := f.tx(sender, receiver.Address, 100, "", 100_000)
	tx.Options = 2
	tx.GuardianAddr = guardian.Address
	f.sign(tx, sender, guardian, nil)
	beforeSender, beforeGuardian, beforeReceiver := f.account(sender.Address), f.account(guardian.Address), f.account(receiver.Address)
	result := f.success(tx)
	require.Equal(t, new(big.Int).Sub(amount(t, beforeSender.Balance), new(big.Int).Add(big.NewInt(100), amount(t, result.Fee))).String(), f.account(sender.Address).Balance)
	require.Equal(t, new(big.Int).Add(amount(t, beforeReceiver.Balance), big.NewInt(100)).String(), f.account(receiver.Address).Balance)
	require.Equal(t, beforeSender.Nonce+1, f.account(sender.Address).Nonce)
	require.Equal(t, beforeGuardian.Balance, f.account(guardian.Address).Balance)
	require.Equal(t, beforeGuardian.Nonce, f.account(guardian.Address).Nonce)
	replacement := f.wallet(0)
	guardedTx := func(data string, signer *integrationTests.TestWalletAccount) *transaction.Transaction {
		tx := f.tx(sender, sender.Address, 0, data, 5_000_000)
		tx.Options = 2
		tx.GuardianAddr = signer.Address
		f.sign(tx, sender, signer, nil)
		return tx
	}
	f.success(guardedTx("SetGuardian@"+hexArg(replacement.Address)+"@"+hexArg([]byte("replacement")), guardian))
	require.Equal(t, f.address(replacement.Address).Bech32, guardianData().ActiveGuardian.Address)
	require.True(t, guardianData().Guarded)
	require.Nil(t, guardianData().PendingGuardian)
	// A valid signature from the previous guardian no longer authorizes spending.
	f.rejected(guardedTx("UnGuardAccount", guardian), sender.Address, receiver.Address)
	// Exercise the combined sender/guardian/relayer signature boundary with the same guarded account.
	relayer := f.wallet(0)
	relayed := f.tx(sender, receiver.Address, 100, "", 500_000)
	relayed.Options = 2
	relayed.GuardianAddr = replacement.Address
	relayed.RelayerAddr = relayer.Address
	f.sign(relayed, sender, replacement, relayer)
	for _, signature := range []*[]byte{&relayed.Signature, &relayed.GuardianSignature, &relayed.RelayerSignature} {
		(*signature)[0] ^= 1
		f.rejected(relayed, sender.Address, receiver.Address, relayer.Address)
		(*signature)[0] ^= 1
	}
	guardedBefore, relayerBefore, recipientBefore := f.account(sender.Address), f.account(relayer.Address), f.account(receiver.Address)
	relayedResult := f.success(relayed)
	require.Equal(t, new(big.Int).Sub(amount(t, guardedBefore.Balance), big.NewInt(100)).String(), f.account(sender.Address).Balance)
	require.Equal(t, new(big.Int).Add(amount(t, recipientBefore.Balance), big.NewInt(100)).String(), f.account(receiver.Address).Balance)
	require.Equal(t, new(big.Int).Sub(amount(t, relayerBefore.Balance), amount(t, relayedResult.Fee)).String(), f.account(relayer.Address).Balance)
	require.Equal(t, guardedBefore.Nonce+1, f.account(sender.Address).Nonce)
	require.Equal(t, relayerBefore.Nonce, f.account(relayer.Address).Nonce)
	f.success(guardedTx("UnGuardAccount", replacement))
	require.False(t, guardianData().Guarded)
	beforeReceiver = f.account(receiver.Address)
	f.success(f.tx(sender, receiver.Address, 100, "", 50_000))
	require.Equal(t, new(big.Int).Add(amount(t, beforeReceiver.Balance), big.NewInt(100)).String(), f.account(receiver.Address).Balance)

}

func TestSupernovaSmokeSignedESDT(t *testing.T) {
	f := newFixture(t)
	sender, receiver := f.wallet(0), f.wallet(1)
	// The reserved EGLD ticker must fail without retaining the issuance payment
	// or leaving token registry/role storage behind. The next issue is a retry.
	beforeSenderAccount, beforeRegistry := f.account(sender.Address), f.account(core.ESDTSCAddress)
	invalidIssue := tokenTests.IssueNonFungibleTx(beforeSenderAccount.Nonce, sender.Address, []byte("EGLD"), "1000")
	f.sign(invalidIssue, sender, nil, nil)
	failedIssue := f.execute(invalidIssue)
	requireExecutionFailure(t, failedIssue)
	require.Contains(t, executionMessages(failedIssue), vm.ErrCouldNotCreateNewTokenIdentifier.Error())
	require.Equal(t, new(big.Int).Sub(amount(t, beforeSenderAccount.Balance), amount(t, failedIssue.Fee)).String(), f.account(sender.Address).Balance)
	require.Equal(t, beforeSenderAccount.Nonce+1, f.account(sender.Address).Nonce)
	require.Equal(t, beforeSenderAccount.RootHash, f.account(sender.Address).RootHash)
	require.Equal(t, beforeRegistry.Balance, f.account(core.ESDTSCAddress).Balance)
	require.Equal(t, beforeRegistry.RootHash, f.account(core.ESDTSCAddress).RootHash)
	// Reuse the existing token issuance builder; sign the actual transaction.
	issue := tokenTests.IssueTx(f.account(sender.Address).Nonce, sender.Address, []byte("SMOKE"), "1000")
	for _, property := range []string{"canFreeze", "canWipe", "canPause", "canAddSpecialRoles"} {
		issue.Data = append(issue.Data, []byte("@"+hexArg([]byte(property))+"@74727565")...)
	}
	f.sign(issue, sender, nil, nil)
	issued := f.success(issue)
	require.NotNil(t, issued.Logs)
	require.NotEmpty(t, issued.Logs.Events)
	require.NotEmpty(t, issued.Logs.Events[0].Topics)
	token := string(issued.Logs.Events[0].Topics[0])
	balance := func(wallet *integrationTests.TestWalletAccount) *big.Int {
		shard := f.cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(wallet.Address)
		data, _, err := f.cs.GetNodeHandler(shard).GetFacadeHandler().GetESDTData(f.address(wallet.Address).Bech32, token, 0, apiData.AccountQueryOptions{})
		require.NoError(t, err)
		return new(big.Int).Set(data.Value)
	}
	beforeSender, beforeReceiver := balance(sender), balance(receiver)
	data := "ESDTTransfer@" + hexArg([]byte(token)) + "@01"
	invalid := f.tx(sender, receiver.Address, 0, data, 1_000_000)
	invalid.Signature[0] ^= 1
	f.rejected(invalid, sender.Address, receiver.Address)
	require.Equal(t, beforeSender, balance(sender))
	require.Equal(t, beforeReceiver, balance(receiver))
	f.success(f.tx(sender, receiver.Address, 0, data, 1_000_000))
	require.Equal(t, new(big.Int).Sub(beforeSender, big.NewInt(1)), balance(sender))
	require.Equal(t, new(big.Int).Add(beforeReceiver, big.NewInt(1)), balance(receiver))
	f.tokenAdministration(sender, receiver, token)
	f.tokenContractPayments(sender, receiver, []tokenPayment{{token: token}}, false)
}
