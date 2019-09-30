package config

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
)

func TestTomlParser(t *testing.T) {
	txBlockBodyStorageSize := 170
	txBlockBodyStorageType := "type1"
	txBlockBodyStorageShards := 5
	txBlockBodyStorageFile := "path1/file1"
	txBlockBodyStorageTypeDB := "type2"

	logsPath := "pathLogger"
	logsStackDepth := 1010

	accountsStorageSize := 171
	accountsStorageType := "type3"
	accountsStorageFile := "path2/file2"
	accountsStorageTypeDB := "type4"
	accountsStorageBlomSize := 172
	accountsStorageBlomHash1 := "hashFunc1"
	accountsStorageBlomHash2 := "hashFunc2"
	accountsStorageBlomHash3 := "hashFunc3"

	hasherType := "hashFunc4"
	multiSigHasherType := "hashFunc5"

	consensusType := "bn"

	cfgExpected := Config{
		MiniBlocksStorage: StorageConfig{
			Cache: CacheConfig{
				Size:   uint32(txBlockBodyStorageSize),
				Type:   txBlockBodyStorageType,
				Shards: uint32(txBlockBodyStorageShards),
			},
			DB: DBConfig{
				FilePath: txBlockBodyStorageFile,
				Type:     txBlockBodyStorageTypeDB,
			},
		},
		Logger: LoggerConfig{
			Path:            logsPath,
			StackTraceDepth: logsStackDepth,
		},
		AccountsTrieStorage: StorageConfig{
			Cache: CacheConfig{
				Size: uint32(accountsStorageSize),
				Type: accountsStorageType,
			},
			DB: DBConfig{
				FilePath: accountsStorageFile,
				Type:     accountsStorageTypeDB,
			},
			Bloom: BloomFilterConfig{
				Size:     172,
				HashFunc: []string{accountsStorageBlomHash1, accountsStorageBlomHash2, accountsStorageBlomHash3},
			},
		},
		Hasher: TypeConfig{
			Type: hasherType,
		},
		MultisigHasher: TypeConfig{
			Type: multiSigHasherType,
		},
		Consensus: TypeConfig{
			Type: consensusType,
		},
	}

	testString := `
[MiniBlocksStorage]
    [MiniBlocksStorage.Cache]
        Size = ` + strconv.Itoa(txBlockBodyStorageSize) + `
        Type = "` + txBlockBodyStorageType + `"
		Shards = ` + strconv.Itoa(txBlockBodyStorageShards) + `
    [MiniBlocksStorage.DB]
        FilePath = "` + txBlockBodyStorageFile + `"
        Type = "` + txBlockBodyStorageTypeDB + `"

[Logger]
    Path = "` + logsPath + `"
    StackTraceDepth = ` + strconv.Itoa(logsStackDepth) + `

[AccountsTrieStorage]
    [AccountsTrieStorage.Cache]
        Size = ` + strconv.Itoa(accountsStorageSize) + `
        Type = "` + accountsStorageType + `"
    [AccountsTrieStorage.DB]
        FilePath = "` + accountsStorageFile + `"
        Type = "` + accountsStorageTypeDB + `"
    [AccountsTrieStorage.Bloom]
        Size = ` + strconv.Itoa(accountsStorageBlomSize) + `
		HashFunc = ["` + accountsStorageBlomHash1 + `", "` + accountsStorageBlomHash2 + `", "` +
		accountsStorageBlomHash3 + `"]

[Hasher]
	Type = "` + hasherType + `"

[MultisigHasher]
	Type = "` + multiSigHasherType + `"

[Consensus]
	Type = "` + consensusType + `"

`
	cfg := Config{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	assert.Nil(t, err)
	assert.Equal(t, cfgExpected, cfg)
}

func TestTomlEconomicsParser(t *testing.T) {
	communityAddress := "commAddr"
	burnAddress := "burnAddr"
	rewardsValue := uint64(500)
	communityPercentage := 0.1
	leaderPercentage := 0.1
	burnPercentage := 0.8
	minGasPrice := uint64(1)
	minGasLimitForTx := uint64(2)
	minTxFee := uint64(3)

	cfgEconomicsExpected := ConfigEconomics{
		EconomicsAddresses: EconomicsAddresses{
			CommunityAddress: communityAddress,
			BurnAddress:      burnAddress,
		},
		RewardsSettings: RewardsSettings{
			RewardsValue:        rewardsValue,
			CommunityPercentage: communityPercentage,
			LeaderPercentage:    leaderPercentage,
			BurnPercentage:      burnPercentage,
		},
		FeeSettings: FeeSettings{
			MinGasPrice:      minGasPrice,
			MinGasLimitForTx: minGasLimitForTx,
			MinTxFee:         minTxFee,
		},
	}

	testString := `
[EconomicsAddresses]
	CommunityAddress = "` + communityAddress + `"
	BurnAddress = "` + burnAddress + `"
[RewardsSettings]
    RewardsValue = ` + strconv.FormatUint(rewardsValue, 10) + `
    CommunityPercentage = ` + fmt.Sprintf("%.6f", communityPercentage) + `
    LeaderPercentage = ` + fmt.Sprintf("%.6f", leaderPercentage) + `
    BurnPercentage = 	` + fmt.Sprintf("%.6f", burnPercentage) + `
[FeeSettings]
    MinGasPrice = ` + strconv.FormatUint(minGasPrice, 10) + `
    MinGasLimitForTx = ` + strconv.FormatUint(minGasLimitForTx, 10) + `
    MinTxFee = ` + strconv.FormatUint(minTxFee, 10) + `
`

	cfg := ConfigEconomics{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	assert.Nil(t, err)
	assert.Equal(t, cfgEconomicsExpected, cfg)
}
