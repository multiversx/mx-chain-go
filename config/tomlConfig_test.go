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

	receiptsStorageSize := 171
	receiptsStorageType := "type3"
	receiptsStorageFile := "path1/file2"
	receiptsStorageTypeDB := "type4"

	logsPath := "pathLogger"
	logsStackDepth := 1010

	accountsStorageSize := 172
	accountsStorageType := "type5"
	accountsStorageFile := "path2/file3"
	accountsStorageTypeDB := "type6"
	accountsStorageBlomSize := 173
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
		ReceiptsStorage: StorageConfig{
			Cache: CacheConfig{
				Capacity: uint32(receiptsStorageSize),
				Type:     receiptsStorageType,
			},
			DB: DBConfig{
				FilePath: receiptsStorageFile,
				Type:     receiptsStorageTypeDB,
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
				Size:     173,
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
        Capacity = ` + strconv.Itoa(txBlockBodyStorageSize) + `
        Type = "` + txBlockBodyStorageType + `"
		Shards = ` + strconv.Itoa(txBlockBodyStorageShards) + `
    [MiniBlocksStorage.DB]
        FilePath = "` + txBlockBodyStorageFile + `"
        Type = "` + txBlockBodyStorageTypeDB + `"

[ReceiptsStorage]
    [ReceiptsStorage.Cache]
        Capacity = ` + strconv.Itoa(receiptsStorageSize) + `
        Type = "` + receiptsStorageType + `"
    [ReceiptsStorage.DB]
        FilePath = "` + receiptsStorageFile + `"
        Type = "` + receiptsStorageTypeDB + `"

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
	protocolSustainabilityPercentage := 0.1
	leaderPercentage := 0.1
	developerPercentage := 0.3
	maxGasLimitPerBlock := "18446744073709551615"
	minGasPrice := "18446744073709551615"
	minGasLimit := "18446744073709551615"
	protocolSustainabilityAddress := "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp"
	denomination := 18

	cfgEconomicsExpected := EconomicsConfig{
		GlobalSettings: GlobalSettings{
			Denomination: denomination,
		},
		RewardsSettings: RewardsSettings{
			LeaderPercentage:                 leaderPercentage,
			ProtocolSustainabilityPercentage: protocolSustainabilityPercentage,
			ProtocolSustainabilityAddress:    protocolSustainabilityAddress,
			DeveloperPercentage:              developerPercentage,
		},
		FeeSettings: FeeSettings{
			MaxGasLimitPerBlock: maxGasLimitPerBlock,
			MinGasPrice:         minGasPrice,
			MinGasLimit:         minGasLimit,
		},
	}

	testString := `
[GlobalSettings]
    Denomination = ` + fmt.Sprintf("%d", denomination) + `
[RewardsSettings]
    ProtocolSustainabilityPercentage = ` + fmt.Sprintf("%.6f", protocolSustainabilityPercentage) + `
	ProtocolSustainabilityAddress = "` + protocolSustainabilityAddress + `"
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
	identity := "test-identity"
	redundancyLevel := int64(0)

	cfgPreferencesExpected := Preferences{
		Preferences: PreferencesConfig{
			NodeDisplayName:            nodeDisplayName,
			DestinationShardAsObserver: destinationShardAsObs,
			Identity:                   identity,
			RedundancyLevel:            redundancyLevel,
		},
	}

	testString := `
[Preferences]
	NodeDisplayName = "` + nodeDisplayName + `"
	DestinationShardAsObserver = "` + destinationShardAsObs + `"
	Identity = "` + identity + `"
	RedundancyLevel = ` + fmt.Sprintf("%d", redundancyLevel) + `
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

	loggingThreshold := 10

	expectedCfg := ApiRoutesConfig{
		Logging: ApiLoggingConfig{
			LoggingEnabled:          true,
			ThresholdInMicroSeconds: loggingThreshold,
		},
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
[Logging]
    LoggingEnabled = true
    ThresholdInMicroSeconds = 10

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
