package integrationTests

// ShardTopic is the topic string generator for sharded topics
// Will generate topics in the following pattern: shard_0, shard_0_1, shard_0_META, shard_1 and so on
const ShardTopic = "shard"

// GlobalTopic is a global testing that all nodes will bind an interceptor
const GlobalTopic = "global"

// AddressPrefix is a global testing const
const AddressHrp = "erd"
