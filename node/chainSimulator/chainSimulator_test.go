package chainSimulator

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../cmd/node/config/"
)

func TestNewChainSimulator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         core.OptionalUint64{},
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       1,
		MetaChainMinNodes:      1,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	time.Sleep(time.Second)

	err = chainSimulator.Close()
	assert.Nil(t, err)
}

func TestChainSimulator_GenerateBlocksShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         core.OptionalUint64{},
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       1,
		MetaChainMinNodes:      1,
		InitialRound:           200000000,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	time.Sleep(time.Second)

	err = chainSimulator.GenerateBlocks(30)
	require.Nil(t, err)

	err = chainSimulator.Close()
	assert.Nil(t, err)
}

func TestChainSimulator_GenerateBlocksAndEpochChangeShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       100,
		MetaChainMinNodes:      100,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	facade, err := NewChainSimulatorFacade(chainSimulator)
	require.Nil(t, err)

	genesisAddressWithStake := chainSimulator.initialWalletKeys.InitialWalletWithStake.Address
	initialAccount, err := facade.GetExistingAccountFromBech32AddressString(genesisAddressWithStake)
	require.Nil(t, err)

	time.Sleep(time.Second)

	err = chainSimulator.GenerateBlocks(80)
	require.Nil(t, err)

	accountAfterRewards, err := facade.GetExistingAccountFromBech32AddressString(genesisAddressWithStake)
	require.Nil(t, err)

	assert.True(t, accountAfterRewards.GetBalance().Cmp(initialAccount.GetBalance()) > 0,
		fmt.Sprintf("initial balance %s, balance after rewards %s", initialAccount.GetBalance().String(), accountAfterRewards.GetBalance().String()))

	fmt.Println(chainSimulator.GetRestAPIInterfaces())

	err = chainSimulator.Close()
	assert.Nil(t, err)
}

func TestChainSimulator_SetState(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       1,
		MetaChainMinNodes:      1,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	keyValueMap := map[string]string{
		"01": "01",
		"02": "02",
	}

	address := "erd1qtc600lryvytxuy4h7vn7xmsy5tw6vuw3tskr75cwnmv4mnyjgsq6e5zgj"
	err = chainSimulator.SetKeyValueForAddress(address, keyValueMap)
	require.Nil(t, err)

	err = chainSimulator.GenerateBlocks(1)
	require.Nil(t, err)

	nodeHandler := chainSimulator.GetNodeHandler(0)
	keyValuePairs, _, err := nodeHandler.GetFacadeHandler().GetKeyValuePairs(address, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, keyValueMap, keyValuePairs)
}

func TestChainSimulator_SetEntireState(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       1,
		MetaChainMinNodes:      1,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	balance := "431271308732096033771131"
	contractAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	accountState := &dtos.AddressState{
		Address:          contractAddress,
		Nonce:            0,
		Balance:          balance,
		Code:             "0061736d010000000129086000006000017f60027f7f017f60027f7f0060017f0060037f7f7f017f60037f7f7f0060017f017f0290020b03656e7619626967496e74476574556e7369676e6564417267756d656e74000303656e760f6765744e756d417267756d656e7473000103656e760b7369676e616c4572726f72000303656e76126d42756666657253746f726167654c6f6164000203656e76176d427566666572546f426967496e74556e7369676e6564000203656e76196d42756666657246726f6d426967496e74556e7369676e6564000203656e76136d42756666657253746f7261676553746f7265000203656e760f6d4275666665725365744279746573000503656e760e636865636b4e6f5061796d656e74000003656e7614626967496e7446696e697368556e7369676e6564000403656e7609626967496e744164640006030b0a010104070301000000000503010003060f027f0041a080080b7f0041a080080b074607066d656d6f7279020004696e697400110667657453756d00120361646400130863616c6c4261636b00140a5f5f646174615f656e6403000b5f5f686561705f6261736503010aca010a0e01017f4100100c2200100020000b1901017f419c8008419c800828020041016b220036020020000b1400100120004604400f0b4180800841191002000b16002000100c220010031a2000100c220010041a20000b1401017f100c2202200110051a2000200210061a0b1301017f100c220041998008410310071a20000b1401017f10084101100d100b210010102000100f0b0e0010084100100d1010100e10090b2201037f10084101100d100b210110102202100e220020002001100a20022000100f0b0300010b0b2f0200418080080b1c77726f6e67206e756d626572206f6620617267756d656e747373756d00419c80080b049cffffff",
		CodeHash:         "n9EviPlHS6EV+3Xp0YqP28T0IUfeAFRFBIRC1Jw6pyU=",
		RootHash:         "76cr5Jhn6HmBcDUMIzikEpqFgZxIrOzgNkTHNatXzC4=",
		CodeMetadata:     "BQY=",
		Owner:            "erd1ss6u80ruas2phpmr82r42xnkd6rxy40g9jl69frppl4qez9w2jpsqj8x97",
		DeveloperRewards: "5401004999998",
		Keys: map[string]string{
			"73756d": "0a",
		},
	}

	err = chainSimulator.SetStateMultiple([]*dtos.AddressState{accountState})
	require.Nil(t, err)

	err = chainSimulator.GenerateBlocks(30)
	require.Nil(t, err)

	nodeHandler := chainSimulator.GetNodeHandler(1)
	scAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(contractAddress)
	res, _, err := nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  scAddress,
		FuncName:   "getSum",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)

	counterValue := big.NewInt(0).SetBytes(res.ReturnData[0]).Int64()
	require.Equal(t, 10, int(counterValue))

	time.Sleep(time.Second)

	account, _, err := nodeHandler.GetFacadeHandler().GetAccount(contractAddress, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, accountState.Balance, account.Balance)
	require.Equal(t, accountState.DeveloperRewards, account.DeveloperReward)
	require.Equal(t, accountState.Code, account.Code)
	require.Equal(t, accountState.CodeHash, base64.StdEncoding.EncodeToString(account.CodeHash))
	require.Equal(t, accountState.CodeMetadata, base64.StdEncoding.EncodeToString(account.CodeMetadata))
	require.Equal(t, accountState.Owner, account.OwnerAddress)
	require.Equal(t, accountState.RootHash, base64.StdEncoding.EncodeToString(account.RootHash))
}
