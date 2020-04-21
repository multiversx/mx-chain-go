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
				Size:   uint32(txBlockBodyStorageSize),
				Type:   txBlockBodyStorageType,
				Shards: uint32(txBlockBodyStorageShards),
			},
			DB: DBConfig{
				FilePath: txBlockBodyStorageFile,
				Type:     txBlockBodyStorageTypeDB,
			},
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
	rewardsValue := "1000000000000000000000000000000000"
	communityPercentage := 0.1
	leaderPercentage := 0.1
	burnPercentage := 0.8
	maxGasLimitPerBlock := "18446744073709551615"
	minGasPrice := "18446744073709551615"
	minGasLimit := "18446744073709551615"

	cfgEconomicsExpected := EconomicsConfig{
		RewardsSettings: RewardsSettings{
			LeaderPercentage: leaderPercentage,
		},
		FeeSettings: FeeSettings{
			MaxGasLimitPerBlock: maxGasLimitPerBlock,
			MinGasPrice:         minGasPrice,
			MinGasLimit:         minGasLimit,
		},
	}

	testString := `
[RewardsSettings]
    RewardsValue = "` + rewardsValue + `"
    CommunityPercentage = ` + fmt.Sprintf("%.6f", communityPercentage) + `
    LeaderPercentage = ` + fmt.Sprintf("%.6f", leaderPercentage) + `
    BurnPercentage =  ` + fmt.Sprintf("%.6f", burnPercentage) + `
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
	testString := `
     # API routes configuration
[APIPackages]

[APIPackages.node]
	Routes = [
        # /node/status will return all metrics stored inside a node
        { Name = "/node/status", Open = true },

        # /node/heartbeatstatus will return all heartbeats messages from the nodes in the network
        { Name = "/node/heartbeatstatus", Open = true },

        # /node/epoch will return data about the current epoch
        { Name = "/node/epoch", Open = true },

        # /node/statistics will return statistics about the chain, such as the peak TPS
        { Name = "/node/statistics", Open = true },

        # /node/p2pstatus will return the metrics related to p2p
        { Name = "/node/p2pstatus", Open = true }
	]

[APIPackages.address]
	Routes = [
         # /address/:address will return data about a given account
        { Name = "/address/:address", Open = true },

         # /address/:address/balance will return the balance of a given account
         { Name = "/address/:address/balance", Open = true }
	]

[APIPackages.hardfork]
	Routes = [
         # /hardfork will receive a trigger request from the client and propagate it for processing
        { Name = "/hardfork", Open = true }
	]

[APIPackages.validator]
	Routes = [
         # /validator/statistics will return a list of validators statistics for all validators
        { Name = "/validator/statistics", Open = true }
	]

[APIPackages.transaction]
	Routes = [
         # /transaction/send will receive a single transaction in JSON format and will propagate it through the network
         # if it's fields are valid. It will return the hash of the transaction
        { Name = "/transaction/send", Open = true },

         # /transaction/send-multiple will receive an array of transactions in JSON format and will propagate through
         # the network those whose field are valid. It will return the number of valid transactions propagated
         { Name = "/transaction/send-multiple", Open = true },

         # /transaction/cost will receive a single transaction in JSON format and will return the estimated cost of it
         { Name = "/transaction/cost", Open = true },

         # /transaction/cost will receive a single transaction in JSON format and will return the estimated cost of it
         { Name = "/transaction/cost", Open = true }
	]
 `

	cfg := ApiRoutesConfig{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	assert.Nil(t, err)
}
