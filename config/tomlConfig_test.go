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

	consensusType := "bls"

	cfgExpected := Config{
		MiniBlocksStorage: StorageConfig{
			Cache: CacheConfig{
				Capacity: uint32(txBlockBodyStorageSize),
				Type:     txBlockBodyStorageType,
				Shards:   uint32(txBlockBodyStorageShards),
			},
			DB: DBConfig{
				FilePath: txBlockBodyStorageFile,
				Type:     txBlockBodyStorageTypeDB,
			},
		},
		AccountsTrieStorage: StorageConfig{
			Cache: CacheConfig{
				Capacity: uint32(accountsStorageSize),
				Type:     accountsStorageType,
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
		Consensus: ConsensusConfig{
			Type: consensusType,
		},
	}

	testString := `
[MiniBlocksStorage]
    [MiniBlocksStorage.Cache]
        Capacity = ` + strconv.Itoa(txBlockBodyStorageSize) + `
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
        Capacity = ` + strconv.Itoa(accountsStorageSize) + `
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
	communityPercentage := 0.1
	leaderPercentage := 0.1
	developerPercentage := 0.3
	maxGasLimitPerBlock := "18446744073709551615"
	minGasPrice := "18446744073709551615"
	minGasLimit := "18446744073709551615"
	communityAddress := "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp"
	denominatinoCoefficient := "0.000000000000000001"

	cfgEconomicsExpected := EconomicsConfig{
		RewardsSettings: RewardsSettings{
			LeaderPercentage:               leaderPercentage,
			CommunityPercentage:            communityPercentage,
			CommunityAddress:               communityAddress,
			DenominationCoefficientForView: denominatinoCoefficient,
			DeveloperPercentage:            developerPercentage,
		},
		FeeSettings: FeeSettings{
			MaxGasLimitPerBlock: maxGasLimitPerBlock,
			MinGasPrice:         minGasPrice,
			MinGasLimit:         minGasLimit,
		},
	}

	testString := `
[RewardsSettings]
    CommunityPercentage = ` + fmt.Sprintf("%.6f", communityPercentage) + `
	CommunityAddress = "` + communityAddress + `"
	DenominationCoefficientForView = "` + denominatinoCoefficient + `"
    LeaderPercentage = ` + fmt.Sprintf("%.6f", leaderPercentage) + `
	DeveloperPercentage = ` + fmt.Sprintf("%.6f", developerPercentage) + `
[FeeSettings]
	MaxGasLimitPerBlock = "` + maxGasLimitPerBlock + `"
    MinGasPrice = "` + minGasPrice + `"
    MinGasLimit = "` + minGasLimit + `"
`

	cfg := EconomicsConfig{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	assert.Nil(t, err)
	assert.Equal(t, cfgEconomicsExpected, cfg)
}

func TestTomlPreferencesParser(t *testing.T) {
	nodeDisplayName := "test-name"
	destinationShardAsObs := "3"

	cfgPreferencesExpected := Preferences{
		Preferences: PreferencesConfig{
			NodeDisplayName:            nodeDisplayName,
			DestinationShardAsObserver: destinationShardAsObs,
		},
	}

	testString := `
[Preferences]
	NodeDisplayName = "` + nodeDisplayName + `"
	DestinationShardAsObserver = "` + destinationShardAsObs + `"
`

	cfg := Preferences{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	assert.Nil(t, err)
	assert.Equal(t, cfgPreferencesExpected, cfg)
}

func TestTomlExternalParser(t *testing.T) {
	indexerURL := "url"
	elasticUsername := "user"
	elasticPassword := "pass"

	cfgExternalExpected := ExternalConfig{
		ElasticSearchConnector: ElasticSearchConfig{
			Enabled:  true,
			URL:      indexerURL,
			Username: elasticUsername,
			Password: elasticPassword,
		},
	}

	testString := `
[ElasticSearchConnector]
    Enabled = true
    URL = "` + indexerURL + `"
    Username = "` + elasticUsername + `"
    Password = "` + elasticPassword + `"`

	cfg := ExternalConfig{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	assert.Nil(t, err)
	assert.Equal(t, cfgExternalExpected, cfg)
}

func TestAPIRoutesToml(t *testing.T) {
	package0 := "testPackage0"
	route0 := "testRoute0"
	route1 := "testRoute1"

	package1 := "testPackage1"
	route2 := "testRoute2"

	expectedCfg := ApiRoutesConfig{
		APIPackages: map[string]APIPackageConfig{
			package0: {
				Routes: []RouteConfig{
					{Name: route0, Open: true},
					{Name: route1, Open: true},
				},
			},
			package1: {
				Routes: []RouteConfig{
					{Name: route2, Open: false},
				},
			},
		},
	}

	testString := `
     # API routes configuration
[APIPackages]

[APIPackages.` + package0 + `]
	Routes = [
        # test comment
        { Name = "` + route0 + `", Open = true },

        # test comment
        { Name = "` + route1 + `", Open = true },
	]

[APIPackages.` + package1 + `]
	Routes = [
         # test comment
        { Name = "` + route2 + `", Open = false }
    ]
 `

	cfg := ApiRoutesConfig{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	assert.Nil(t, err)
	assert.Equal(t, expectedCfg, cfg)
}
