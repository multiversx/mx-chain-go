package smoke

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	apiData "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	chainTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process"
)

const numShards = uint32(2)

type fixture struct {
	t  *testing.T
	cs chainTests.ChainSimulator
}

func newFixture(t *testing.T, configure ...func(*config.Configs)) *fixture {
	t.Helper()
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         false,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            "../../../cmd/node/config/",
		NumOfShards:                    numShards,
		RoundDurationInMillis:          6000,
		SupernovaRoundDurationInMillis: 600,
		RoundsPerEpoch:                 core.OptionalUint64{HasValue: true, Value: 20},
		SupernovaRoundsPerEpoch:        core.OptionalUint64{HasValue: true, Value: 20},
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               1,
		MetaChainMinNodes:              1,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 2
			cfg.RoundConfig.RoundActivations["SupernovaEnableRound"] = config.ActivationRoundByName{Round: "80"}
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = "1000"
			for _, alter := range configure {
				alter(cfg)
			}
		},
	})
	require.NoError(t, err)
	t.Cleanup(cs.Close)
	chainTests.RequireSupernova(t, cs, numShards, 200)
	return &fixture{t: t, cs: cs}
}

func (f *fixture) wallet(shard uint32) *integrationTests.TestWalletAccount {
	f.t.Helper()
	wallet := integrationTests.CreateTestWalletAccount(f.cs.GetNodeHandler(0).GetShardCoordinator(), shard)
	address, err := f.cs.GetNodeHandler(shard).GetCoreComponents().AddressPubKeyConverter().Encode(wallet.Address)
	require.NoError(f.t, err)
	require.NoError(f.t, f.cs.SetStateMultiple([]*dtos.AddressState{{Address: address, Balance: new(big.Int).Mul(chainTests.OneEGLD, big.NewInt(100)).String()}}))
	require.NoError(f.t, f.cs.GenerateBlocks(1))
	return wallet
}

func (f *fixture) address(address []byte) dtos.WalletAddress {
	f.t.Helper()
	encoded, err := f.cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(address)
	require.NoError(f.t, err)
	return dtos.WalletAddress{Bytes: address, Bech32: encoded}
}

func (f *fixture) account(address []byte) apiData.AccountResponse {
	f.t.Helper()
	account, err := f.cs.GetAccount(f.address(address))
	require.NoError(f.t, err)
	return account
}

func (f *fixture) tx(sender *integrationTests.TestWalletAccount, receiver []byte, value int64, data string, gas uint64) *transaction.Transaction {
	tx := chainTests.GenerateTransaction(sender.Address, f.account(sender.Address).Nonce, receiver, big.NewInt(value), data, gas)
	tx.Version = 2
	f.sign(tx, sender, nil, nil)
	return tx
}

func (f *fixture) sign(tx *transaction.Transaction, sender, guardian, relayer *integrationTests.TestWalletAccount) {
	f.t.Helper()
	message, err := tx.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer, integrationTests.TestTxSignHasher)
	require.NoError(f.t, err)
	tx.Signature, err = sender.SingleSigner.Sign(sender.SkTxSign, message)
	require.NoError(f.t, err)
	if guardian != nil {
		tx.GuardianSignature, err = guardian.SingleSigner.Sign(guardian.SkTxSign, message)
		require.NoError(f.t, err)
	}
	if relayer != nil {
		tx.RelayerSignature, err = relayer.SingleSigner.Sign(relayer.SkTxSign, message)
		require.NoError(f.t, err)
	}
}

func (f *fixture) execute(tx *transaction.Transaction) *transaction.ApiTransactionResult {
	f.t.Helper()
	result, err := f.cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 40)
	if err != nil {
		// A source-shard rejection has no destination transaction. The simulator's
		// convenience method polls only the destination; require a committed source
		// failure before accepting that specific outcome.
		shard := f.cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(tx.SndAddr)
		node := f.cs.GetNodeHandler(shard)
		hash, hashErr := core.CalculateHash(node.GetCoreComponents().InternalMarshalizer(), node.GetCoreComponents().Hasher(), tx)
		require.NoError(f.t, hashErr)
		source, sourceErr := node.GetFacadeHandler().GetTransaction(hexArg(hash), true)
		require.NoError(f.t, sourceErr, "execution error: %v", err)
		require.NotNil(f.t, source)
		require.Contains(f.t, []transaction.TxStatus{transaction.TxStatusInvalid, transaction.TxStatusFail}, source.Status, "execution error: %v", err)
		require.NotEmpty(f.t, source.BlockHash)
		result = source
	}
	require.NotNil(f.t, result)
	// Destination execution can precede source-side SCR refunds. Drain those before
	// taking account snapshots for the next operation's conservation checks.
	require.NoError(f.t, f.cs.GenerateBlocks(8))
	return result
}

func (f *fixture) success(tx *transaction.Transaction) *transaction.ApiTransactionResult {
	f.t.Helper()
	result := f.execute(tx)
	requireExecutionSuccess(f.t, result)
	return result
}

func requireExecutionSuccess(t *testing.T, result *transaction.ApiTransactionResult) {
	t.Helper()
	require.Equal(t, transaction.TxStatusSuccess, result.Status, "%s", executionMessages(result))
	if result.Logs != nil {
		for _, event := range result.Logs.Events {
			require.NotEqual(t, core.SignalErrorOperation, event.Identifier, "processed status alone does not prove SC execution succeeded: %s", event.Topics)
		}
	}
}

func requireExecutionFailure(t *testing.T, result *transaction.ApiTransactionResult) {
	t.Helper()
	if result.Status == transaction.TxStatusFail || result.Status == transaction.TxStatusInvalid {
		return
	}
	require.NotNil(t, result.Logs)
	for _, event := range result.Logs.Events {
		if event.Identifier == core.SignalErrorOperation {
			return
		}
	}
	t.Fatal("expected execution failure status or signalError event")
}

// Capture committed tries as well as account fields. No blocks are generated inside
// a simulation/admission comparison: epoch rewards would legitimately change roots.
func (f *fixture) roots() map[uint32][]byte {
	f.t.Helper()
	roots := make(map[uint32][]byte)
	for _, shard := range []uint32{0, 1, core.MetachainShardId} {
		root, err := f.cs.GetNodeHandler(shard).GetStateComponents().AccountsAdapter().RootHash()
		require.NoError(f.t, err)
		roots[shard] = bytes.Clone(root)
	}
	return roots
}

func (f *fixture) rejected(tx *transaction.Transaction, addresses ...[]byte) {
	f.t.Helper()
	roots := f.roots()
	before := make([]apiData.AccountResponse, len(addresses))
	for i, address := range addresses {
		before[i] = f.account(address)
	}
	_, err := f.cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 2)
	require.Error(f.t, err, "invalid transaction must be rejected at admission")
	require.Equal(f.t, roots, f.roots())
	// Check again after block production, so delayed admission cannot pass unnoticed.
	require.NoError(f.t, f.cs.GenerateBlocks(5))
	for i, address := range addresses {
		after := f.account(address)
		require.Equal(f.t, before[i].Balance, after.Balance)
		require.Equal(f.t, before[i].Nonce, after.Nonce)
		require.Equal(f.t, before[i].RootHash, after.RootHash)
	}
}

func (f *fixture) query(contract []byte, method string) *big.Int {
	f.t.Helper()
	shard := f.cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(contract)
	result, _, err := f.cs.GetNodeHandler(shard).GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{ScAddress: contract, FuncName: method})
	require.NoError(f.t, err)
	require.Equal(f.t, "ok", result.ReturnCode)
	require.Len(f.t, result.ReturnData, 1)
	return new(big.Int).SetBytes(result.ReturnData[0])
}

func (f *fixture) queryValues(contract []byte, method string, args ...[]byte) [][]byte {
	f.t.Helper()
	shard := f.cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(contract)
	result, _, err := f.cs.GetNodeHandler(shard).GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{ScAddress: contract, FuncName: method, Arguments: args})
	require.NoError(f.t, err)
	require.Equal(f.t, "ok", result.ReturnCode, "%s: %s", method, result.ReturnMessage)
	return result.ReturnData
}

func (f *fixture) queryAmount(contract []byte, method string, args ...[]byte) *big.Int {
	f.t.Helper()
	values := f.queryValues(contract, method, args...)
	require.Len(f.t, values, 1, method)
	return new(big.Int).SetBytes(values[0])
}

func (f *fixture) call(sender *integrationTests.TestWalletAccount, receiver []byte, value *big.Int, data string) *transaction.Transaction {
	tx := f.tx(sender, receiver, 0, data, 100_000_000)
	tx.Value = new(big.Int).Set(value)
	f.sign(tx, sender, nil, nil)
	return tx
}

func (f *fixture) fund(wallet *integrationTests.TestWalletAccount, value *big.Int) {
	f.t.Helper()
	require.NoError(f.t, f.cs.SetStateMultiple([]*dtos.AddressState{{Address: f.address(wallet.Address).Bech32, Balance: value.String()}}))
	require.NoError(f.t, f.cs.GenerateBlocks(1))
}

func (f *fixture) epoch() uint32 {
	return f.cs.GetNodeHandler(core.MetachainShardId).GetChainHandler().GetCurrentBlockHeader().GetEpoch()
}

func (f *fixture) until(maxBlocks int, condition func() bool) {
	f.t.Helper()
	for i := 0; i < maxBlocks && !condition(); i++ {
		require.NoError(f.t, f.cs.GenerateBlocks(1))
	}
	require.True(f.t, condition(), "condition not reached within %d blocks", maxBlocks)
}

func amount(t *testing.T, text string) *big.Int {
	t.Helper()
	value, ok := new(big.Int).SetString(text, 10)
	require.True(t, ok)
	return value
}

func hexArg(data []byte) string { return hex.EncodeToString(data) }
