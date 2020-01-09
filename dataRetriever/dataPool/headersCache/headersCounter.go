package headersCache

type numHeadersByShard map[uint32]uint64

func (nhs numHeadersByShard) increment(shardId uint32) {
	if _, ok := nhs[shardId]; !ok {
		nhs[shardId] = 0
	}

	nhs[shardId]++
}

func (nhs numHeadersByShard) decrement(shardId uint32, value int) {
	if _, ok := nhs[shardId]; !ok {
		return
	}

	nhs[shardId] -= uint64(value)
}

func (nhs numHeadersByShard) getCount(shardId uint32) int64 {
	numShardHeaders, ok := nhs[shardId]
	if !ok {
		return 0
	}

	return int64(numShardHeaders)
}

func (nhs numHeadersByShard) totalHeaders() int {
	total := 0

	for _, value := range nhs {
		total += int(value)
	}

	return total
}
