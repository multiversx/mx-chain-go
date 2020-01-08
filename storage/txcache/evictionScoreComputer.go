package txcache

type evictionScoreComputer struct {
	senders         []*txListForSender
	scores          []float64
	quantizedScores []int64
	maxGas          int64
	minGas          int64
	maxSize         int64
	minSize         int64
	maxOrderNumber  int64
	minOrderNumber  int64
	maxTxCount      int64
	minTxCount      int64

	gasRange         int64
	sizeRange        int64
	orderNumberRange int64
	txCountRange     int64
}

func newEvictionScoreComputer(senders []*txListForSender) *evictionScoreComputer {
	computer := &evictionScoreComputer{
		senders: senders,
	}

	computer.computeBounds()
	computer.computeRanges()
	computer.computeScores()
	computer.quantizeScores()
	return computer
}

// computeBounds finds the min and max values for the score parameters (gas, size and so on)
func (computer *evictionScoreComputer) computeBounds() {
	for _, sender := range computer.senders {
		gas := sender.totalGas.Get()
		size := sender.totalBytes.Get()
		txCount := sender.countTx()
		orderNumber := sender.orderNumber

		if gas > computer.maxGas {
			computer.maxGas = gas
		} else if gas < computer.minGas {
			computer.minGas = gas
		}

		if size > computer.maxSize {
			computer.maxSize = gas
		} else if size < computer.minSize {
			computer.minSize = size
		}

		if txCount > computer.maxTxCount {
			computer.maxTxCount = txCount
		} else if txCount < computer.minTxCount {
			computer.minTxCount = txCount
		}

		if orderNumber > computer.maxOrderNumber {
			computer.maxOrderNumber = orderNumber
		} else if orderNumber < computer.minOrderNumber {
			computer.minOrderNumber = orderNumber
		}
	}
}

func (computer *evictionScoreComputer) computeRanges() {
	computer.gasRange = strictlyPositive(computer.maxGas - computer.minGas)
	computer.sizeRange = strictlyPositive(computer.maxSize - computer.minSize)
	computer.orderNumberRange = strictlyPositive(computer.maxOrderNumber - computer.minOrderNumber)
	computer.txCountRange = strictlyPositive(computer.maxTxCount - computer.minTxCount)
}

func strictlyPositive(value int64) int64 {
	if value > 0 {
		return value
	}
	return 1
}

func (computer *evictionScoreComputer) computeScores() {
	for i, sender := range computer.senders {
		computer.scores[i] = computer.computeScore(sender)
	}
}

// A low score means that the sender will be evicted
// A high score means that the sender will not be evicted, most probably
// Score is:
// - inversely proportional to sender's tx count
// - inversely proportional to sender's tx total size
// - directly proportional to sender's order number
// - directly proportional to sender's tx total gas
func (computer *evictionScoreComputer) computeScore(txList *txListForSender) float64 {
	// Normalize score parameters, interval [0..1]
	orderNumber := float64(txList.orderNumber-computer.minOrderNumber) / float64(computer.orderNumberRange)
	gas := float64(txList.totalGas.Get()-computer.minGas) / float64(computer.gasRange)
	txCount := float64(txList.countTx()-computer.minTxCount) / float64(computer.txCountRange)
	size := float64(txList.totalBytes.Get()-computer.minSize) / float64(computer.sizeRange)

	txCount = notTooSmall(txCount)
	size = notTooSmall(size)

	return orderNumber * gas / txCount / size
}

func notTooSmall(value float64) float64 {
	if value > 0.01 {
		return value
	}
	return 0.01
}

func (computer *evictionScoreComputer) quantizeScores() {
	maxScore := float64(0)
	minScore := float64(0)

	for _, score := range computer.scores {
		if score > maxScore {
			maxScore = score
		} else if score < minScore {
			minScore = score
		}
	}

	scoreRange := notTooSmall(maxScore - minScore)

	for i, score := range computer.scores {
		computer.quantizedScores[i] = int64((score - minScore) / scoreRange)
	}
}
