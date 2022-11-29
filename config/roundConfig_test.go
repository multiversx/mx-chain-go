package config

import (
	"encoding/json"
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundConfigLoad(t *testing.T) {
	t.Parallel()

	testString := `
[RoundActivations]
    [RoundActivations.test1]
        Options = ["option1", "option2"]
        Round = "0"
    [RoundActivations.test2]
        Options = ["option3", "option4"]
        Round = "1"
`

	expectedConfig := &RoundConfig{
		RoundActivations: map[string]ActivationRoundByName{
			"test1": {
				Round:   "0",
				Options: []string{"option1", "option2"},
			},
			"test2": {
				Round:   "1",
				Options: []string{"option3", "option4"},
			},
		},
	}

	loadedConfig := &RoundConfig{}
	err := toml.Unmarshal([]byte(testString), loadedConfig)

	assert.Nil(t, err)
	assert.Equal(t, expectedConfig, loadedConfig)
}

func TestNodesSetup(t *testing.T) {
	var nodesCfg *NodesConfig
	payload := `
{
  "startTime": 0,
  "roundDuration": 3000,
  "consensusGroupSize": 1,
  "hysteresis": 0,
  "adaptivity": false,
  "shardConsensusConfiguration": [
    {
      "enableEpoch": 0,
      "minNodes": 2,
      "consensusGroupSize": 2
    }
  ],
  "metachainConsensusConfiguration": [
    {
      "enableEpoch": 0,
      "minNodes": 2,
      "consensusGroupSize": 2
    }
  ],
  "initialNodes": [
    {
      "pubkey": "b9fa402bf620e971bcdcf3ad40d9241a80b72f7df58be97846bfdbc5e35bb6dcd2e6751246380174238c9afa791e8b08ede3ee2128056cc019348b52d610e13c982940aaf426938c6a5dbd92b14a54a9b65771c547f97f7b48697cafe5890701",
      "address": "erd1045r79f5rwvfje65ngwnczxujqerhfxkkt9a73rvy6v95jpfusrq32zyqy",
      "initialRating": 5000001
    },
    {
      "pubkey": "03b4207aec8c312d764dc2efeb4d904a1272c4524f88239a3abef9f5d73c8f77c4e74bf2141170d83f7d36c0c1b2ec0638d3261333bab3f3430f8a181a41172d542680d9dcf00c96c01ec1251eb14d6124dbeda689e23109faef86f9e4b04085",
      "address": "erd1qunp09v79fvdtwcdwhynvmzt5n42rj9u5pn6w0gupazdfs26jndsadhw6s",
      "initialRating": 5000001
    },
    {
      "pubkey": "e14d6f9c4f32787fae58999c732759329a0b7932bfaf33cd65829efe7bd5169ef70e7ea896985d2654f31d138531db16e1d8990880471e2af0deb8362f79ff594701254ea7d78dd0417ec95a358a27b2bf7ef20f906c4384193a2ddfd7f40d06",
      "address": "erd1d5dm8zwsnyg3juznx3e9t7fd4ez6y5xsy98k82s7l99cfn8eefsqrkvfqn",
      "initialRating": 5000001
    },
    {
      "pubkey": "e4c760f5892567626c49eb850adb0412bd33f36e9bd55420a920f1fff7a95e3053b9315f72c08d5798e73a874c046414205263130b43c5116211d96172a85c4ddf665d0086b7e40ce5f3b1126d3ab08734c8cca4cc53f4dd13b13a6357c44a89",
      "address": "erd12hgyk37yuxlaj5vux0wxg907x99t3zjqhyetn6e47068ew0g45lqll39x8",
      "initialRating": 5000001
    },
    {
      "pubkey": "8dd89c1625ae648374612d6c534a21c658db0651cc49c4fa6a377577e60c0d7441f05750bf95de297165e3fbb5a61e034b3616b6d189bf79e560c0bba19238e0030dc7a510aa229b12f7074f358b9929772c6e576ac1c2c89570ac2caa3bff05",
      "address": "erd18l3m5fpsqe3c3dqjyc3ynvvq00vrwqzcq5dqak7tyj7dsj6xutyqvl5eax",
      "initialRating": 5000001
    },
    {
      "pubkey": "5e130cdf7f7e188b9db98f495e4429c3496bb1718432568a4d047b647e449379e7d91105b24097bc4a4e6366913316035737dc6a1a089be2d330f92d2fe4517c3dea69705b14dfdc3a41065e4122da3bb898e4c180a185a6182fb86b4ec04f94",
      "address": "erd1rlvlgv0tedqjsawfl599ghvrnvs78gn0k6dnu95a8fx0ngt37rkqcpurf4",
      "initialRating": 5000001
    }
  ]
}
`

	err := json.Unmarshal([]byte(payload), nodesCfg)
	require.NoError(t, err)
}
