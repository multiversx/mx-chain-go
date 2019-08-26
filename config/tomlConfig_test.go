package config

import (
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

	communityAddress := "community"
	burnAddress := "burnAddress"

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
		RewardConfig: RewardConfig{
			CommunityAddress: communityAddress,
			BurnAddress:      burnAddress,
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

[RewardConfig]
	CommunityAddress = "` + communityAddress + `"
    BurnAddress = "` + burnAddress + `"
`
	cfg := Config{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	assert.Nil(t, err)
	assert.Equal(t, cfgExpected, cfg)
}
