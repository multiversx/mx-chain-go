package api

import "fmt"

type fixedPortAPIConfigurator struct {
	restAPIInterface string
	mapShardPort     map[uint32]int
}

// NewFixedPortAPIConfigurator will create a new instance of fixedPortAPIConfigurator
func NewFixedPortAPIConfigurator(restAPIInterface string, mapShardPort map[uint32]int) *fixedPortAPIConfigurator {
	return &fixedPortAPIConfigurator{
		restAPIInterface: restAPIInterface,
		mapShardPort:     mapShardPort,
	}
}

// RestApiInterface will return the api interface for the provided shard
func (f *fixedPortAPIConfigurator) RestApiInterface(shardID uint32) string {
	return fmt.Sprintf("%s:%d", f.restAPIInterface, f.mapShardPort[shardID])
}
