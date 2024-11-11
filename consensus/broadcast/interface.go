package broadcast

type broadcasterFilterHandler interface {
	shouldSkipShard(shardID uint32) bool
	shouldSkipTopic(topic string) bool
}
