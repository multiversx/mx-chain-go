package epochStart

import (
	"encoding/json"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

// UnmarshalMetaTrigger unmarshalls the meta trigger with json, for backwards compatibility
func UnmarshalMetaTrigger(marshaller marshal.Marshalizer, data []byte) (data.MetaTriggerRegistryHandler, error) {
	trig, err := unmarshallMetaTriggerV3(marshaller, data)
	if err == nil {
		return trig, nil
	}

	trig, err = unmarshallMetaTriggerV1(marshaller, data)
	if err == nil {
		return trig, nil
	}

	// for backwards compatibility
	return unmarshallMetaTriggerJson(data)
}

func unmarshallMetaTriggerV3(marshaller marshal.Marshalizer, data []byte) (data.MetaTriggerRegistryHandler, error) {
	state := &block.MetaTriggerRegistryV3{
		EpochStartMeta: &block.MetaBlockV3{},
	}

	err := marshaller.Unmarshal(state, data)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func unmarshallMetaTriggerV1(marshaller marshal.Marshalizer, data []byte) (data.MetaTriggerRegistryHandler, error) {
	state := &block.MetaTriggerRegistry{
		EpochStartMeta: &block.MetaBlock{},
	}

	err := marshaller.Unmarshal(state, data)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func unmarshallMetaTriggerJson(data []byte) (data.MetaTriggerRegistryHandler, error) {
	state := &block.MetaTriggerRegistry{
		EpochStartMeta: &block.MetaBlock{},
	}

	err := json.Unmarshal(data, state)
	if err != nil {
		return nil, err
	}

	return state, nil
}

// CreateMetaRegistryHandler creates the appropriate meta trigger registry handler based on the header type
func CreateMetaRegistryHandler(epochStartMeta data.HeaderHandler) data.MetaTriggerRegistryHandler {
	if epochStartMeta.IsHeaderV3() {
		return &block.MetaTriggerRegistryV3{}
	}

	return &block.MetaTriggerRegistry{}
}

// UnmarshalShardTrigger unmarshalls the shard trigger
func UnmarshalShardTrigger(marshaller marshal.Marshalizer, data []byte) (data.TriggerRegistryHandler, error) {
	trig, err := unmarshalShardTriggerV3(marshaller, data)
	if err == nil {
		return trig, nil
	}

	trig, err = unmarshalShardTriggerV2(marshaller, data)
	if err == nil {
		return trig, nil
	}

	trig, err = unmarshalShardTriggerV1(marshaller, data)
	if err == nil {
		return trig, nil
	}

	// for backwards compatibility
	return unmarshalShardTriggerJson(data)
}

// unmarshalShardTriggerJson unmarshalls the trigger with json, for backwards compatibility
func unmarshalShardTriggerJson(data []byte) (data.TriggerRegistryHandler, error) {
	trig := &block.ShardTriggerRegistry{EpochStartShardHeader: &block.Header{}}
	err := json.Unmarshal(data, trig)
	if err != nil {
		return nil, err
	}

	return trig, nil
}

// unmarshalShardTriggerV3 tries to unmarshal the data into a v3 trigger
func unmarshalShardTriggerV3(marshaller marshal.Marshalizer, data []byte) (data.TriggerRegistryHandler, error) {
	triggerV3 := &block.ShardTriggerRegistryV3{
		EpochStartShardHeader: &block.HeaderV3{},
	}

	err := marshaller.Unmarshal(triggerV3, data)
	if err != nil {
		return nil, err
	}

	if check.IfNil(triggerV3.EpochStartShardHeader) {
		return nil, fmt.Errorf("%w while checking inner epoch start shard header", ErrNilHeaderHandler)
	}
	return triggerV3, nil
}

// unmarshalShardTriggerV2 tries to unmarshal the data into a v2 trigger
func unmarshalShardTriggerV2(marshaller marshal.Marshalizer, data []byte) (data.TriggerRegistryHandler, error) {
	triggerV2 := &block.ShardTriggerRegistryV2{
		EpochStartShardHeader: &block.HeaderV2{},
	}

	err := marshaller.Unmarshal(triggerV2, data)
	if err != nil {
		return nil, err
	}

	if check.IfNil(triggerV2.EpochStartShardHeader) || check.IfNil(triggerV2.EpochStartShardHeader.Header) {
		return nil, fmt.Errorf("%w while checking inner epoch start shard header", ErrNilHeaderHandler)
	}
	return triggerV2, nil
}

// unmarshalShardTriggerV1 tries to unmarshal the data into a v1 trigger
func unmarshalShardTriggerV1(marshaller marshal.Marshalizer, data []byte) (data.TriggerRegistryHandler, error) {
	triggerV1 := &block.ShardTriggerRegistry{
		EpochStartShardHeader: &block.Header{},
	}

	err := marshaller.Unmarshal(triggerV1, data)
	if err != nil {
		return nil, err
	}

	if check.IfNil(triggerV1.EpochStartShardHeader) {
		return nil, fmt.Errorf("%w while checking inner epoch start shard header", ErrNilHeaderHandler)
	}
	return triggerV1, nil
}

// CreateShardRegistryHandler creates the appropriate shard trigger registry handler based on the header type
func CreateShardRegistryHandler(epochStartShardHeader data.HeaderHandler) data.TriggerRegistryHandler {
	if epochStartShardHeader.IsHeaderV3() {
		return &block.ShardTriggerRegistryV3{
			EpochStartShardHeader: &block.HeaderV3{},
		}
	}

	_, ok := epochStartShardHeader.(*block.HeaderV2)
	if ok {
		return &block.ShardTriggerRegistryV2{
			EpochStartShardHeader: &block.HeaderV2{},
		}
	}

	return &block.ShardTriggerRegistry{
		EpochStartShardHeader: &block.Header{},
	}
}
