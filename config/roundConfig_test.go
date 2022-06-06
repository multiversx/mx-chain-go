package config

import (
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
)

func TestRoundConfigSave(t *testing.T) {
	t.Parallel()

	expectedConfig := &RoundConfig{
		RoundActivations: map[string]ActivationRoundByName{
			"test1": {
				Round:   0,
				Options: []string{"option1", "option2"},
			},
			"test2": {
				Round:   1,
				Options: []string{"option3", "option4"},
			},
		},
	}

	expectedString := `
[RoundActivations]

  [RoundActivations.test1]
    Options = ["option1", "option2"]
    Round = 0

  [RoundActivations.test2]
    Options = ["option3", "option4"]
    Round = 1
`

	bytes, err := toml.Marshal(expectedConfig)

	assert.Nil(t, err)
	assert.Equal(t, expectedString, string(bytes))
}

func TestRoundConfigLoad(t *testing.T) {
	t.Parallel()

	testString := `
[RoundActivations]

    [RoundActivations.test1]
        Options = ["option1", "option2"]
        Round = 0

    [RoundActivations.test2]
        Options = ["option3", "option4"]
        Round = 1
`

	expectedConfig := &RoundConfig{
		RoundActivations: map[string]ActivationRoundByName{
			"test1": {
				Round:   0,
				Options: []string{"option1", "option2"},
			},
			"test2": {
				Round:   1,
				Options: []string{"option3", "option4"},
			},
		},
	}

	loadedConfig := &RoundConfig{}
	err := toml.Unmarshal([]byte(testString), loadedConfig)

	assert.Nil(t, err)
	assert.Equal(t, expectedConfig, loadedConfig)
}
