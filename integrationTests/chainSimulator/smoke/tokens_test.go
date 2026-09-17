package smoke

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	apiData "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/integrationTests"
	tokenTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/vm"
)

func (f *fixture) tokenBalance(wallet *integrationTests.TestWalletAccount, token string, nonce uint64) *big.Int {
	f.t.Helper()
	shard := f.cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(wallet.Address)
	data, _, err := f.cs.GetNodeHandler(shard).GetFacadeHandler().GetESDTData(f.address(wallet.Address).Bech32, token, nonce, apiData.AccountQueryOptions{})
	require.NoError(f.t, err)
	require.NotNil(f.t, data)
	require.NotNil(f.t, data.Value)
	return new(big.Int).Set(data.Value)
}

func (f *fixture) tokenRole(owner *integrationTests.TestWalletAccount, token string, endpoint string, roles ...string) {
	f.t.Helper()
	data := endpoint + "@" + hexArg([]byte(token)) + "@" + hexArg(owner.Address)
	for _, role := range roles {
		data += "@" + hexArg([]byte(role))
	}
	f.success(f.tx(owner, core.ESDTSCAddress, 0, data, 100_000_000))
}

func (f *fixture) tokenAdministration(owner, receiver *integrationTests.TestWalletAccount, token string) {
	tokenHex := hexArg([]byte(token))
	transfer := "ESDTTransfer@" + tokenHex + "@01"
	supply := func() *big.Int {
		return new(big.Int).Add(f.tokenBalance(owner, token, 0), f.tokenBalance(receiver, token, 0))
	}
	initialSupply := supply()
	f.tokenRole(owner, token, "setSpecialRole", core.ESDTRoleLocalMint, core.ESDTRoleLocalBurn)
	f.success(f.tx(owner, owner.Address, 0, "ESDTLocalMint@"+tokenHex+"@05", 10_000_000))
	require.Equal(f.t, new(big.Int).Add(initialSupply, big.NewInt(5)), supply())
	f.success(f.tx(owner, owner.Address, 0, "ESDTLocalBurn@"+tokenHex+"@02", 10_000_000))
	require.Equal(f.t, new(big.Int).Add(initialSupply, big.NewInt(3)), supply())
	f.success(f.tx(owner, core.ESDTSCAddress, 0, "unsetBurnRoleGlobally@"+tokenHex, 100_000_000))
	f.tokenRole(owner, token, "unSetSpecialRole", core.ESDTRoleLocalMint, core.ESDTRoleLocalBurn)
	for _, method := range []string{"ESDTLocalMint", "ESDTLocalBurn"} {
		f.t.Logf("revoked %s", method)
		before := supply()
		requireExecutionFailure(f.t, f.execute(f.tx(owner, owner.Address, 0, method+"@"+tokenHex+"@01", 10_000_000)))
		require.Equal(f.t, before, supply())
	}
	for _, mode := range []struct{ disable, enable, extra string }{
		{"pause", "unPause", ""},
		{"freeze", "unFreeze", "@" + hexArg(receiver.Address)},
	} {
		f.t.Logf("administration %s", mode.disable)
		f.success(f.tx(owner, core.ESDTSCAddress, 0, mode.disable+"@"+tokenHex+mode.extra, 100_000_000))
		beforeSender, beforeReceiver := f.tokenBalance(owner, token, 0), f.tokenBalance(receiver, token, 0)
		requireExecutionFailure(f.t, f.execute(f.tx(owner, receiver.Address, 0, transfer, 10_000_000)))
		require.Equal(f.t, beforeSender, f.tokenBalance(owner, token, 0))
		require.Equal(f.t, beforeReceiver, f.tokenBalance(receiver, token, 0))
		f.success(f.tx(owner, core.ESDTSCAddress, 0, mode.enable+"@"+tokenHex+mode.extra, 100_000_000))
		f.success(f.tx(owner, receiver.Address, 0, transfer, 10_000_000))
		require.Equal(f.t, new(big.Int).Sub(beforeSender, big.NewInt(1)), f.tokenBalance(owner, token, 0))
		require.Equal(f.t, new(big.Int).Add(beforeReceiver, big.NewInt(1)), f.tokenBalance(receiver, token, 0))
	}
	beforeSupply, wiped := supply(), f.tokenBalance(receiver, token, 0)
	require.Positive(f.t, wiped.Sign())
	f.success(f.tx(owner, core.ESDTSCAddress, 0, "freeze@"+tokenHex+"@"+hexArg(receiver.Address), 100_000_000))
	f.success(f.tx(owner, core.ESDTSCAddress, 0, "wipe@"+tokenHex+"@"+hexArg(receiver.Address), 100_000_000))
	require.Zero(f.t, f.tokenBalance(receiver, token, 0).Sign())
	require.Equal(f.t, new(big.Int).Sub(beforeSupply, wiped), supply())
}

func TestSupernovaSmokeNFTAndMixedTransfers(t *testing.T) {
	f := newFixture(t)
	owner, receiver := f.wallet(0), f.wallet(1)
	var tokens []string
	builders := []tokenTests.IssueTxFunc{tokenTests.IssueNonFungibleTx, tokenTests.IssueSemiFungibleTx, tokenTests.IssueMetaESDTTx}
	for index, issue := range builders {
		tx := issue(f.account(owner.Address).Nonce, owner.Address, []byte(fmt.Sprintf("SMK%c", 'A'+index)), "1000")
		for _, property := range []string{"canFreeze", "canWipe"} {
			tx.Data = append(tx.Data, []byte("@"+hexArg([]byte(property))+"@74727565")...)
		}
		f.sign(tx, owner, nil, nil)
		result := f.success(tx)
		require.NotNil(t, result.Logs)
		require.NotEmpty(t, result.Logs.Events)
		token := string(result.Logs.Events[0].Topics[0])
		tokens = append(tokens, token)
		roles := []string{core.ESDTRoleNFTCreate}
		if index > 0 {
			roles = append(roles, core.ESDTRoleNFTAddQuantity)
		}
		f.tokenRole(owner, token, "setSpecialRole", roles...)
		quantity := "0a"
		if index == 0 {
			quantity = "01"
		}
		create := "ESDTNFTCreate@" + hexArg([]byte(token)) + "@" + quantity + "@" + hexArg([]byte("smoke")) + "@00@@" + hexArg([]byte("attributes")) + "@" + hexArg([]byte("uri"))
		f.success(f.tx(owner, owner.Address, 0, create, 100_000_000))
		expectedQuantity := int64(10)
		if index == 0 {
			expectedQuantity = 1
		}
		require.Equal(t, big.NewInt(expectedQuantity), f.tokenBalance(owner, token, 1))
		if index > 0 {
			f.tokenRole(owner, token, "setSpecialRole", core.ESDTRoleNFTBurn)
			tokenHex := hexArg([]byte(token))
			f.success(f.tx(owner, owner.Address, 0, "ESDTNFTAddQuantity@"+tokenHex+"@01@03", 20_000_000))
			require.Equal(t, big.NewInt(13), f.tokenBalance(owner, token, 1))
			f.success(f.tx(owner, owner.Address, 0, "ESDTNFTBurn@"+tokenHex+"@01@03", 20_000_000))
			require.Equal(t, big.NewInt(10), f.tokenBalance(owner, token, 1))
			f.success(f.tx(owner, core.ESDTSCAddress, 0, "unsetBurnRoleGlobally@"+tokenHex, 100_000_000))
			f.tokenRole(owner, token, "unSetSpecialRole", core.ESDTRoleNFTAddQuantity, core.ESDTRoleNFTBurn)
			for _, method := range []string{"ESDTNFTAddQuantity", "ESDTNFTBurn"} {
				requireExecutionFailure(t, f.execute(f.tx(owner, owner.Address, 0, method+"@"+tokenHex+"@01@01", 20_000_000)))
				require.Equal(t, big.NewInt(10), f.tokenBalance(owner, token, 1))
			}
		}
	}
	// A second NFT nonce gives a genuine multi-NFT transfer, alongside SFT and MetaESDT.
	nftHex := hexArg([]byte(tokens[0]))
	f.success(f.tx(owner, owner.Address, 0, "ESDTNFTCreate@"+nftHex+"@01@736d6f6b65@00@@@757269", 100_000_000))
	items := []string{hexArg([]byte("EGLD-000000")), "00", "64", nftHex, "01", "01", nftHex, "02", "01", hexArg([]byte(tokens[1])), "01", "02", hexArg([]byte(tokens[2])), "01", "03"}
	data := "MultiESDTNFTTransfer@" + hexArg(receiver.Address) + "@05@" + strings.Join(items, "@")
	badItems := append([]string(nil), items...)
	badItems[len(badItems)-1] = "ff" // valid encoding, insufficient MetaESDT quantity
	bad := "MultiESDTNFTTransfer@" + hexArg(receiver.Address) + "@05@" + strings.Join(badItems, "@")
	before := []int64{1, 10, 10}
	beforeOwner, beforeReceiver := f.account(owner.Address), f.account(receiver.Address)
	failed := f.execute(f.tx(owner, owner.Address, 0, bad, 30_000_000))
	requireExecutionFailure(t, failed)
	require.Equal(t, new(big.Int).Sub(amount(t, beforeOwner.Balance), amount(t, failed.Fee)).String(), f.account(owner.Address).Balance)
	require.Equal(t, beforeOwner.Nonce+1, f.account(owner.Address).Nonce)
	require.Equal(t, beforeReceiver.Balance, f.account(receiver.Address).Balance)
	require.Equal(t, beforeOwner.RootHash, f.account(owner.Address).RootHash)
	require.Equal(t, beforeReceiver.RootHash, f.account(receiver.Address).RootHash)
	for i, token := range tokens {
		require.Equal(t, big.NewInt(before[i]), f.tokenBalance(owner, token, 1))
		require.Zero(t, f.tokenBalance(receiver, token, 1).Sign())
	}
	require.Equal(t, big.NewInt(1), f.tokenBalance(owner, tokens[0], 2))
	require.Zero(t, f.tokenBalance(receiver, tokens[0], 2).Sign())
	beforeOwner = f.account(owner.Address)
	paid := f.success(f.tx(owner, owner.Address, 0, data, 30_000_000))
	expected := new(big.Int).Sub(amount(t, beforeOwner.Balance), amount(t, paid.Fee))
	expected.Sub(expected, big.NewInt(100))
	require.Equal(t, expected.String(), f.account(owner.Address).Balance)
	require.Equal(t, beforeOwner.Nonce+1, f.account(owner.Address).Nonce)
	require.Equal(t, new(big.Int).Add(amount(t, beforeReceiver.Balance), big.NewInt(100)).String(), f.account(receiver.Address).Balance)
	for i, qty := range []int64{1, 2, 3} {
		require.Equal(t, big.NewInt(before[i]-qty), f.tokenBalance(owner, tokens[i], 1))
		require.Equal(t, big.NewInt(qty), f.tokenBalance(receiver, tokens[i], 1))
	}
	require.Zero(t, f.tokenBalance(owner, tokens[0], 2).Sign())
	require.Equal(t, big.NewInt(1), f.tokenBalance(receiver, tokens[0], 2))
	// A nonce-specific freeze must leave the second NFT and other token types usable.
	f.success(f.tx(owner, core.ESDTSCAddress, 0, "freezeSingleNFT@"+nftHex+"@01@"+hexArg(receiver.Address), 100_000_000))
	blocked := "ESDTNFTTransfer@" + nftHex + "@01@01@" + hexArg(owner.Address)
	requireExecutionFailure(t, f.execute(f.tx(receiver, receiver.Address, 0, blocked, 20_000_000)))
	require.Equal(t, big.NewInt(1), f.tokenBalance(receiver, tokens[0], 1))
	require.Equal(t, big.NewInt(1), f.tokenBalance(receiver, tokens[0], 2))
	f.success(f.tx(owner, core.ESDTSCAddress, 0, "unFreezeSingleNFT@"+nftHex+"@01@"+hexArg(receiver.Address), 100_000_000))
	// Both NFT nonces plus SFT and MetaESDT go through cross-shard contract execution.
	contract := f.tokenContractPayments(receiver, owner, []tokenPayment{{tokens[0], 1}, {tokens[0], 2}, {tokens[1], 1}, {tokens[2], 1}}, true)
	// Wipe one NFT nonce and prove the sibling nonce survives in the same account.
	contractAccount := &integrationTests.TestWalletAccount{Address: contract}
	f.success(f.tx(owner, core.ESDTSCAddress, 0, "freezeSingleNFT@"+nftHex+"@01@"+hexArg(contract), 100_000_000))
	f.success(f.tx(owner, core.ESDTSCAddress, 0, "wipeSingleNFT@"+nftHex+"@01@"+hexArg(contract), 100_000_000))
	require.Zero(t, f.tokenBalance(contractAccount, tokens[0], 1).Sign())
	require.Equal(t, big.NewInt(1), f.tokenBalance(contractAccount, tokens[0], 2))
	require.Equal(t, big.NewInt(1), f.tokenBalance(contractAccount, tokens[1], 1))
	require.Equal(t, big.NewInt(1), f.tokenBalance(contractAccount, tokens[2], 1))
}
