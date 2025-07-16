package simulate

//
//import (
//	"github.com/multiversx/mx-chain-core-go/core"
//	"github.com/multiversx/mx-chain-core-go/data/transaction"
//	"github.com/multiversx/mx-chain-go/config"
//	"github.com/multiversx/mx-chain-go/node/chainSimulator"
//	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
//	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
//	"github.com/stretchr/testify/require"
//	"math/big"
//	"testing"
//	"time"
//)
//
//var addr1 = &dtos.AddressState{
//	Address: "erd1fmd662htrgt07xxd8me09newa9s0euzvpz3wp0c4pz78f83grt9qm6pn57",
//	Balance: "13825919478590045170353",
//	Nonce:   func() *uint64 { v := uint64(2738); return &v }(),
//	Pairs: map[string]string{
//		"454c524f4e44657364744d4943452d39653030376101c3": "08021202000122e00208c303120e4d696365436974792023363935311a2000000000000000000500a83d6077dce6489378d56cc0d30aa68ad2a04b7eeddf20f4032a2e516d5745775369394168674d50657534434a664c70577131794b53666d45525a5874373662343239705653553952324c68747470733a2f2f697066732e696f2f697066732f516d5745775369394168674d50657534434a664c70577131794b53666d45525a58743736623432397056535539522f363935312e706e67324d68747470733a2f2f697066732e696f2f697066732f516d5745775369394168674d50657534434a664c70577131794b53666d45525a58743736623432397056535539522f363935312e6a736f6e3a59746167733a4d6963652c3830732c4e6f7374616c6769613b6d657461646174613a516d5745775369394168674d50657534434a664c70577131794b53666d45525a58743736623432397056535539522f363935312e6a736f6e",
//	},
//}
//
//const (
//	defaultPathToInitialConfig = "../../../cmd/node/config/"
//
//	minGasPrice                            = 1000000000
//	maxNumOfBlockToGenerateWhenExecutingTx = 7
//)
//
//func TestSimulateEndpointCheck(t *testing.T) {
//	if testing.Short() {
//		t.Skip("this is not a short test")
//	}
//
//	startTime := time.Now().Unix()
//	roundDurationInMillis := uint64(6000)
//	roundsPerEpoch := core.OptionalUint64{
//		HasValue: true,
//		Value:    15,
//	}
//
//	activationEpoch := uint32(4)
//
//	baseIssuingCost := "1000"
//
//	numOfShards := uint32(3)
//	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
//		BypassTxSignatureCheck:   true,
//		TempDir:                  t.TempDir(),
//		PathToInitialConfig:      defaultPathToInitialConfig,
//		NumOfShards:              numOfShards,
//		GenesisTimestamp:         startTime,
//		RoundDurationInMillis:    roundDurationInMillis,
//		RoundsPerEpoch:           roundsPerEpoch,
//		ApiInterface:             api.NewNoApiInterface(),
//		MinNodesPerShard:         3,
//		MetaChainMinNodes:        3,
//		NumNodesWaitingListMeta:  0,
//		NumNodesWaitingListShard: 0,
//		AlterConfigsFunction: func(cfg *config.Configs) {
//			cfg.EpochConfig.EnableEpochs.EGLDInMultiTransferEnableEpoch = activationEpoch
//			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
//			cfg.EpochConfig.EnableEpochs.AndromedaEnableEpoch = cfg.GeneralConfig.EpochStartConfig.GenesisEpoch + 1
//			cfg.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
//		},
//	})
//	require.Nil(t, err)
//	require.NotNil(t, cs)
//
//	err = cs.SetStateMultiple([]*dtos.AddressState{addr1})
//	require.Nil(t, err)
//
//	err = cs.GenerateBlocks(22)
//	require.Nil(t, err)
//
//	// 00000000000000000500fbdb316a74b878e3e02bfadefb5daf9ceb2e9f3d6c12
//	sndBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(addr1.Address)
//	tx := &transaction.Transaction{
//		SndAddr:   sndBytes,
//		RcvAddr:   sndBytes,
//		Nonce:     *addr1.Nonce,
//		Data:      []byte("MultiESDTNFTTransfer@00000000000000000500fbdb316a74b878e3e02bfadefb5daf9ceb2e9f3d6c12@01@4d4943452d396530303761@01c3@01@6c697374696e67@00000009056bc75e2d6310000000000009056bc75e2d6310000000000000000000000000000445474c44000100000000000000000000000b4d4943452d39653030376100000000000001c300000001010000000201f4"),
//		Signature: []byte("dummy"),
//		GasPrice:  minGasPrice,
//		Value:     big.NewInt(0),
//	}
//
//	res, err := cs.GetNodeHandler(2).GetFacadeHandler().ComputeTransactionGasLimit(tx)
//	require.Nil(t, err)
//	require.NotNil(t, res)
//}
//
//// MultiESDTNFTTransfer@00000000000000000500fbdb316a74b878e3e02bfadefb5daf9ceb2e9f3d6c12@01@4d4943452d396530303761@0193@01@6c697374696e67@000000084563918244f40000000000084563918244f4000000000000000000000000000445474c44000100000000000000000000000b4d4943452d396530303761000000000000019300000001010000000201f4
//
//func TestSimulateMint(t *testing.T) {
//	if testing.Short() {
//		t.Skip("this is not a short test")
//	}
//
//	startTime := time.Now().Unix()
//	roundDurationInMillis := uint64(6000)
//	roundsPerEpoch := core.OptionalUint64{
//		HasValue: true,
//		Value:    20,
//	}
//
//	activationEpoch := uint32(4)
//
//	baseIssuingCost := "1000"
//
//	numOfShards := uint32(3)
//	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
//		BypassTxSignatureCheck:   true,
//		TempDir:                  t.TempDir(),
//		PathToInitialConfig:      defaultPathToInitialConfig,
//		NumOfShards:              numOfShards,
//		GenesisTimestamp:         startTime,
//		RoundDurationInMillis:    roundDurationInMillis,
//		RoundsPerEpoch:           roundsPerEpoch,
//		ApiInterface:             api.NewNoApiInterface(),
//		MinNodesPerShard:         3,
//		MetaChainMinNodes:        3,
//		NumNodesWaitingListMeta:  0,
//		NumNodesWaitingListShard: 0,
//		AlterConfigsFunction: func(cfg *config.Configs) {
//			cfg.EpochConfig.EnableEpochs.EGLDInMultiTransferEnableEpoch = activationEpoch
//			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
//			cfg.EpochConfig.EnableEpochs.AndromedaEnableEpoch = cfg.GeneralConfig.EpochStartConfig.GenesisEpoch + 1
//			cfg.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
//		},
//	})
//	require.Nil(t, err)
//	require.NotNil(t, cs)
//
//	err = cs.GenerateBlocks(22)
//	require.Nil(t, err)
//
//	sndAddr := "erd1nq0lcpt27dv3hp4eusf43ucynm777p8q5tczn0jc82m28m440y4sjejpf4"
//	sndBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(sndAddr)
//	mintContractAddr, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqpgq4q7kqa7uueyfx7x4dnqdxz4x3tf2qjm7ah0sf3kcsa")
//	sndNonce, err := cs.ChainFetcher().GetAddressNonce(sndAddr)
//	require.Nil(t, err)
//
//	tx := &transaction.Transaction{
//		SndAddr:   sndBytes,
//		RcvAddr:   mintContractAddr,
//		Nonce:     sndNonce,
//		Data:      []byte("buy@4d69636543697479@5075626c69634d696e74@01"),
//		Signature: []byte("dummy"),
//		GasPrice:  minGasPrice,
//		Value:     big.NewInt(400000000000000000),
//	}
//
//	res, err := cs.GetNodeHandler(1).GetFacadeHandler().ComputeTransactionGasLimit(tx)
//	require.Nil(t, err)
//	require.NotNil(t, res)
//	require.Equal(t, "", res.ReturnMessage)
//}
