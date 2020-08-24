package transaction

// func TestNode_ComputeTransactionStatus(t *testing.T) {
// 	t.Parallel()

// 	storer := &mock.ChainStorerMock{
// 		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
// 			return getStorerStub(false)
// 		},
// 	}

// 	shardZeroAddr := []byte("addrShard0")
// 	shardOneAddr := []byte("addrShard1")
// 	shardCoordinator := &mock.ShardCoordinatorMock{
// 		ComputeIdCalled: func(addr []byte) uint32 {
// 			if bytes.Equal(shardZeroAddr, addr) {
// 				return 0
// 			}
// 			return 1
// 		},
// 	}

// 	n, _ := node.NewNode(
// 		node.WithDataStore(storer),
// 		node.WithShardCoordinator(shardCoordinator),
// 	)

// 	rwdTxCrossShard := &rewardTx.RewardTx{RcvAddr: shardZeroAddr}
// 	normalTxIntraShard := &transaction.Transaction{RcvAddr: shardZeroAddr, SndAddr: shardZeroAddr}
// 	normalTxCrossShard := &transaction.Transaction{RcvAddr: shardOneAddr, SndAddr: shardZeroAddr}
// 	unsignedTxIntraShard := &smartContractResult.SmartContractResult{RcvAddr: shardZeroAddr, SndAddr: shardZeroAddr}
// 	unsignedTxCrossShard := &smartContractResult.SmartContractResult{RcvAddr: shardOneAddr, SndAddr: shardZeroAddr}

// 	// cross shard reward tx in storage source shard
// 	shardCoordinator.SelfShardId = core.MetachainShardId
// 	txStatus := n.ComputeTransactionStatus(rwdTxCrossShard, false)
// 	assert.Equal(t, transaction.TxStatusPartiallyExecuted, txStatus)

// 	// cross shard reward tx in pool source shard
// 	shardCoordinator.SelfShardId = core.MetachainShardId
// 	txStatus = n.ComputeTransactionStatus(rwdTxCrossShard, true)
// 	assert.Equal(t, transaction.TxStatusReceived, txStatus)

// 	// intra shard transaction in storage
// 	shardCoordinator.SelfShardId = 0
// 	txStatus = n.ComputeTransactionStatus(normalTxIntraShard, false)
// 	assert.Equal(t, transaction.TxStatusExecuted, txStatus)

// 	// intra shard transaction in pool
// 	shardCoordinator.SelfShardId = 0
// 	txStatus = n.ComputeTransactionStatus(normalTxIntraShard, true)
// 	assert.Equal(t, transaction.TxStatusReceived, txStatus)

// 	// cross shard transaction in storage source shard
// 	shardCoordinator.SelfShardId = 0
// 	txStatus = n.ComputeTransactionStatus(normalTxCrossShard, false)
// 	assert.Equal(t, transaction.TxStatusPartiallyExecuted, txStatus)

// 	// cross shard transaction in pool source shard
// 	shardCoordinator.SelfShardId = 0
// 	txStatus = n.ComputeTransactionStatus(normalTxCrossShard, true)
// 	assert.Equal(t, transaction.TxStatusReceived, txStatus)

// 	// cross shard transaction in storage destination shard
// 	shardCoordinator.SelfShardId = 1
// 	txStatus = n.ComputeTransactionStatus(normalTxCrossShard, false)
// 	assert.Equal(t, transaction.TxStatusExecuted, txStatus)

// 	// cross shard transaction in pool destination shard
// 	shardCoordinator.SelfShardId = 1
// 	txStatus = n.ComputeTransactionStatus(normalTxCrossShard, true)
// 	assert.Equal(t, transaction.TxStatusPartiallyExecuted, txStatus)

// 	// intra shard scr in storage source shard
// 	shardCoordinator.SelfShardId = 0
// 	txStatus = n.ComputeTransactionStatus(unsignedTxIntraShard, false)
// 	assert.Equal(t, transaction.TxStatusExecuted, txStatus)

// 	// intra shard scr in pool source shard
// 	shardCoordinator.SelfShardId = 0
// 	txStatus = n.ComputeTransactionStatus(unsignedTxIntraShard, true)
// 	assert.Equal(t, transaction.TxStatusReceived, txStatus)

// 	// cross shard scr in storage source shard
// 	shardCoordinator.SelfShardId = 0
// 	txStatus = n.ComputeTransactionStatus(unsignedTxCrossShard, false)
// 	assert.Equal(t, transaction.TxStatusPartiallyExecuted, txStatus)

// 	// cross shard scr in pool source shard
// 	shardCoordinator.SelfShardId = 0
// 	txStatus = n.ComputeTransactionStatus(unsignedTxCrossShard, true)
// 	assert.Equal(t, transaction.TxStatusReceived, txStatus)
// }
