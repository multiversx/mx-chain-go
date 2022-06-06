package config

import (
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
)

func TestRoundConfigLoad(t *testing.T) {
	t.Parallel()

	testString := `
[[RoundActivations]]
	Name = "test1"
    Round = 0
    Options = [
		"option1",
		"option2",
	]

[[RoundActivations]]
    Name = "test2"
    Round = 1
    Options = [
		"option3",
		"option4",
	]
`

	expectedConfig := &RoundConfig{
		RoundActivations: []ActivationRoundByName{
			{
				Name:    "test1",
				Round:   0,
				Options: []string{"option1", "option2"},
			},
			{
				Name:    "test2",
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
