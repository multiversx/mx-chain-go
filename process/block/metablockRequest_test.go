package block_test

// TODO[Sorin-nextPR]: move and adapt these tests for headersForBlock
// func TestMetaProcessor_computeExistingAndRequestMissingShardHeaders(t *testing.T) {
// 	t.Parallel()
//
// 	noOfShards := uint32(2)
// 	td := createTestData()
//
// 	t.Run("all referenced shard headers missing", func(t *testing.T) {
// 		t.Parallel()
// 		referencedHeaders := []*shardHeaderData{td[0].referencedHeaderData, td[1].referencedHeaderData}
// 		shardInfo := createShardInfo(referencedHeaders)
// 		metaBlock := &block.MetaBlock{
// 			ShardInfo: shardInfo,
// 		}
//
// 		numCallsMissingAttestation := atomic.Uint32{}
// 		numCallsMissingHeaders := atomic.Uint32{}
// 		arguments := createMetaProcessorArguments(t, noOfShards)
// 		updateRequestsHandlerForCountingRequests(t, arguments, td, metaBlock, &numCallsMissingHeaders, &numCallsMissingAttestation)
//
// 		mp, err := blockProcess.NewMetaProcessor(*arguments)
// 		require.Nil(t, err)
// 		require.NotNil(t, mp)
//
// 		headersForBlock := mp.GetHdrForBlock()
// 		numMissing, missingProofs, numAttestationMissing := mp.ComputeExistingAndRequestMissingShardHeaders(metaBlock)
// 		time.Sleep(100 * time.Millisecond)
// 		require.Equal(t, uint32(2), numMissing)
// 		missingHdrs, _, missingAttesting := headersForBlock.GetMissingData()
// 		require.Equal(t, uint32(2), missingHdrs)
// 		// before receiving all missing headers referenced in metaBlock, the number of missing attestations is not updated
// 		require.Equal(t, uint32(0), numAttestationMissing)
// 		require.Equal(t, uint32(0), missingProofs)
// 		require.Equal(t, uint32(0), missingAttesting)
// 		require.Len(t, headersForBlock.GetHeadersInfoMap(), 2)
// 		require.Equal(t, uint32(0), numCallsMissingAttestation.Load())
// 		require.Equal(t, uint32(2), numCallsMissingHeaders.Load())
// 	})
// 	t.Run("one referenced shard header present and one missing", func(t *testing.T) {
// 		t.Parallel()
// 		referencedHeaders := []*shardHeaderData{td[0].referencedHeaderData, td[1].referencedHeaderData}
// 		shardInfo := createShardInfo(referencedHeaders)
// 		metaBlock := &block.MetaBlock{
// 			ShardInfo: shardInfo,
// 		}
//
// 		numCallsMissingAttestation := atomic.Uint32{}
// 		numCallsMissingHeaders := atomic.Uint32{}
// 		arguments := createMetaProcessorArguments(t, noOfShards)
// 		poolsHolder, ok := arguments.DataComponents.Datapool().(*dataRetrieverMock.PoolsHolderMock)
// 		require.True(t, ok)
//
// 		headersPoolStub := createPoolsHolderForHeaderRequests()
// 		poolsHolder.SetHeadersPool(headersPoolStub)
// 		updateRequestsHandlerForCountingRequests(t, arguments, td, metaBlock, &numCallsMissingHeaders, &numCallsMissingAttestation)
//
// 		mp, err := blockProcess.NewMetaProcessor(*arguments)
// 		require.Nil(t, err)
// 		require.NotNil(t, mp)
//
// 		headersPool := mp.GetDataPool().Headers()
// 		// adding the existing header
// 		headersPool.AddHeader(td[0].referencedHeaderData.headerHash, td[0].referencedHeaderData.header)
// 		numMissing, missingProofs, numAttestationMissing := mp.ComputeExistingAndRequestMissingShardHeaders(metaBlock)
// 		time.Sleep(100 * time.Millisecond)
// 		headersForBlock := mp.GetHdrForBlock()
// 		require.Equal(t, uint32(1), numMissing)
//
// 		missingHdrs, _, missingAttesting := headersForBlock.GetMissingData()
// 		require.Equal(t, uint32(1), missingHdrs)
// 		// before receiving all missing headers referenced in metaBlock, the number of missing attestations is not updated
// 		require.Equal(t, uint32(0), numAttestationMissing)
// 		require.Equal(t, uint32(0), missingProofs)
// 		require.Equal(t, uint32(0), missingAttesting)
// 		require.Len(t, headersForBlock.GetHeadersInfoMap(), 2)
// 		require.Equal(t, uint32(0), numCallsMissingAttestation.Load())
// 		require.Equal(t, uint32(1), numCallsMissingHeaders.Load())
// 	})
// 	t.Run("all referenced shard headers present, all attestation headers missing", func(t *testing.T) {
// 		t.Parallel()
// 		referencedHeaders := []*shardHeaderData{td[0].referencedHeaderData, td[1].referencedHeaderData}
// 		shardInfo := createShardInfo(referencedHeaders)
// 		metaBlock := &block.MetaBlock{
// 			ShardInfo: shardInfo,
// 		}
//
// 		numCallsMissingAttestation := atomic.Uint32{}
// 		numCallsMissingHeaders := atomic.Uint32{}
// 		arguments := createMetaProcessorArguments(t, noOfShards)
// 		poolsHolder, ok := arguments.DataComponents.Datapool().(*dataRetrieverMock.PoolsHolderMock)
// 		require.True(t, ok)
//
// 		headersPoolStub := createPoolsHolderForHeaderRequests()
// 		poolsHolder.SetHeadersPool(headersPoolStub)
// 		updateRequestsHandlerForCountingRequests(t, arguments, td, metaBlock, &numCallsMissingHeaders, &numCallsMissingAttestation)
//
// 		mp, err := blockProcess.NewMetaProcessor(*arguments)
// 		require.Nil(t, err)
// 		require.NotNil(t, mp)
//
// 		headersPool := mp.GetDataPool().Headers()
// 		// adding the existing headers
// 		headersPool.AddHeader(td[0].referencedHeaderData.headerHash, td[0].referencedHeaderData.header)
// 		headersPool.AddHeader(td[1].referencedHeaderData.headerHash, td[1].referencedHeaderData.header)
// 		numMissing, missingProofs, numAttestationMissing := mp.ComputeExistingAndRequestMissingShardHeaders(metaBlock)
// 		time.Sleep(100 * time.Millisecond)
// 		headersForBlock := mp.GetHdrForBlock()
// 		require.Equal(t, uint32(0), numMissing)
// 		missingHdrs, _, missingAttesting := headersForBlock.GetMissingData()
// 		require.Equal(t, uint32(0), missingHdrs)
// 		require.Equal(t, uint32(2), numAttestationMissing)
// 		require.Equal(t, uint32(0), missingProofs)
// 		require.Equal(t, uint32(2), missingAttesting)
// 		require.Len(t, headersForBlock.GetHeadersInfoMap(), 2)
// 		require.Equal(t, uint32(2), numCallsMissingAttestation.Load())
// 		require.Equal(t, uint32(0), numCallsMissingHeaders.Load())
// 	})
// 	t.Run("all referenced shard headers present, one attestation header missing", func(t *testing.T) {
// 		t.Parallel()
// 		referencedHeaders := []*shardHeaderData{td[0].referencedHeaderData, td[1].referencedHeaderData}
// 		shardInfo := createShardInfo(referencedHeaders)
// 		metaBlock := &block.MetaBlock{
// 			ShardInfo: shardInfo,
// 		}
//
// 		numCallsMissingAttestation := atomic.Uint32{}
// 		numCallsMissingHeaders := atomic.Uint32{}
// 		arguments := createMetaProcessorArguments(t, noOfShards)
// 		poolsHolder, ok := arguments.DataComponents.Datapool().(*dataRetrieverMock.PoolsHolderMock)
// 		require.True(t, ok)
//
// 		headersPoolStub := createPoolsHolderForHeaderRequests()
// 		poolsHolder.SetHeadersPool(headersPoolStub)
// 		updateRequestsHandlerForCountingRequests(t, arguments, td, metaBlock, &numCallsMissingHeaders, &numCallsMissingAttestation)
//
// 		mp, err := blockProcess.NewMetaProcessor(*arguments)
// 		require.Nil(t, err)
// 		require.NotNil(t, mp)
//
// 		headersPool := mp.GetDataPool().Headers()
// 		// adding the existing headers
// 		headersPool.AddHeader(td[0].referencedHeaderData.headerHash, td[0].referencedHeaderData.header)
// 		headersPool.AddHeader(td[1].referencedHeaderData.headerHash, td[1].referencedHeaderData.header)
// 		headersPool.AddHeader(td[0].attestationHeaderData.headerHash, td[0].attestationHeaderData.header)
// 		numMissing, missingProofs, numAttestationMissing := mp.ComputeExistingAndRequestMissingShardHeaders(metaBlock)
// 		time.Sleep(100 * time.Millisecond)
// 		headersForBlock := mp.GetHdrForBlock()
// 		require.Equal(t, uint32(0), numMissing)
// 		missingHdrs, _, missingAttesting := headersForBlock.GetMissingData()
// 		require.Equal(t, uint32(0), missingHdrs)
// 		require.Equal(t, uint32(1), numAttestationMissing)
// 		require.Equal(t, uint32(0), missingProofs)
// 		require.Equal(t, uint32(1), missingAttesting)
// 		require.Len(t, headersForBlock.GetHeadersInfoMap(), 3)
// 		require.Equal(t, uint32(1), numCallsMissingAttestation.Load())
// 		require.Equal(t, uint32(0), numCallsMissingHeaders.Load())
// 	})
// 	t.Run("all referenced shard headers present, all attestation headers present", func(t *testing.T) {
// 		t.Parallel()
// 		referencedHeaders := []*shardHeaderData{td[0].referencedHeaderData, td[1].referencedHeaderData}
// 		shardInfo := createShardInfo(referencedHeaders)
// 		metaBlock := &block.MetaBlock{
// 			ShardInfo: shardInfo,
// 		}
//
// 		numCallsMissingAttestation := atomic.Uint32{}
// 		numCallsMissingHeaders := atomic.Uint32{}
// 		arguments := createMetaProcessorArguments(t, noOfShards)
// 		poolsHolder, ok := arguments.DataComponents.Datapool().(*dataRetrieverMock.PoolsHolderMock)
// 		require.True(t, ok)
//
// 		headersPoolStub := createPoolsHolderForHeaderRequests()
// 		poolsHolder.SetHeadersPool(headersPoolStub)
// 		updateRequestsHandlerForCountingRequests(t, arguments, td, metaBlock, &numCallsMissingHeaders, &numCallsMissingAttestation)
//
// 		mp, err := blockProcess.NewMetaProcessor(*arguments)
// 		require.Nil(t, err)
// 		require.NotNil(t, mp)
//
// 		headersPool := mp.GetDataPool().Headers()
// 		// adding the existing headers
// 		headersPool.AddHeader(td[0].referencedHeaderData.headerHash, td[0].referencedHeaderData.header)
// 		headersPool.AddHeader(td[1].referencedHeaderData.headerHash, td[1].referencedHeaderData.header)
// 		headersPool.AddHeader(td[0].attestationHeaderData.headerHash, td[0].attestationHeaderData.header)
// 		headersPool.AddHeader(td[1].attestationHeaderData.headerHash, td[1].attestationHeaderData.header)
// 		numMissing, missingProofs, numAttestationMissing := mp.ComputeExistingAndRequestMissingShardHeaders(metaBlock)
// 		time.Sleep(100 * time.Millisecond)
// 		headersForBlock := mp.GetHdrForBlock()
// 		require.Equal(t, uint32(0), numMissing)
// 		missingHdrs, _, missingAttesting := headersForBlock.GetMissingData()
// 		require.Equal(t, uint32(0), missingHdrs)
// 		require.Equal(t, uint32(0), numAttestationMissing)
// 		require.Equal(t, uint32(0), missingProofs)
// 		require.Equal(t, uint32(0), missingAttesting)
// 		require.Len(t, headersForBlock.GetHeadersInfoMap(), 4)
// 		require.Equal(t, uint32(0), numCallsMissingAttestation.Load())
// 		require.Equal(t, uint32(0), numCallsMissingHeaders.Load())
// 	})
// }
//
// func TestMetaProcessor_receivedShardHeader(t *testing.T) {
// 	t.Parallel()
// 	noOfShards := uint32(2)
// 	td := createTestData()
//
// 	t.Run("receiving the last used in block shard header", func(t *testing.T) {
// 		t.Parallel()
//
// 		numCalls := atomic.Uint32{}
// 		arguments := createMetaProcessorArguments(t, noOfShards)
// 		requestHandler, ok := arguments.ArgBaseProcessor.RequestHandler.(*testscommon.RequestHandlerStub)
// 		require.True(t, ok)
//
// 		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
// 			attestationNonce := td[shardID].attestationHeaderData.header.GetNonce()
// 			if nonce != attestationNonce {
// 				require.Fail(t, fmt.Sprintf("nonce should have been %d", attestationNonce))
// 			}
// 			numCalls.Add(1)
// 		}
//
// 		mp, err := blockProcess.NewMetaProcessor(*arguments)
// 		require.Nil(t, err)
// 		require.NotNil(t, mp)
//
// 		hdrsForBlock := mp.GetHdrForBlock()
// 		hdrsForBlock.IncreaseMissingHeaders()
// 		hdrsForBlock.SetHighestHeaderNonceForShard(0, td[0].referencedHeaderData.header.GetNonce()-1)
// 		hdrsForBlock.AddHeaderInfo(string(td[0].referencedHeaderData.headerHash), headerForBlock.NewHeaderInfo(nil, true, false, false))
//
// 		mp.ReceivedShardHeader(td[0].referencedHeaderData.header, td[0].referencedHeaderData.headerHash)
//
// 		time.Sleep(100 * time.Millisecond)
// 		require.Nil(t, err)
// 		require.NotNil(t, mp)
// 		require.Equal(t, uint32(1), numCalls.Load())
// 		_, _, missingAttesting := hdrsForBlock.GetMissingData()
// 		require.Equal(t, uint32(1), missingAttesting)
// 	})
//
// 	t.Run("shard header used in block received, not latest", func(t *testing.T) {
// 		t.Parallel()
//
// 		numCalls := atomic.Uint32{}
// 		arguments := createMetaProcessorArguments(t, noOfShards)
// 		requestHandler, ok := arguments.ArgBaseProcessor.RequestHandler.(*testscommon.RequestHandlerStub)
// 		require.True(t, ok)
//
// 		// for requesting attestation header
// 		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
// 			attestationNonce := td[shardID].attestationHeaderData.header.GetNonce()
// 			require.Equal(t, nonce, attestationNonce, fmt.Sprintf("nonce should have been %d", attestationNonce))
// 			numCalls.Add(1)
// 		}
//
// 		mp, err := blockProcess.NewMetaProcessor(*arguments)
// 		require.Nil(t, err)
// 		require.NotNil(t, mp)
//
// 		hdrsForBlock := mp.GetHdrForBlock()
// 		hdrsForBlock.IncreaseMissingHeaders()
// 		hdrsForBlock.IncreaseMissingHeaders()
// 		referencedHeaderData := td[1].referencedHeaderData
// 		hdrsForBlock.SetHighestHeaderNonceForShard(0, referencedHeaderData.header.GetNonce()-1)
// 		hdrsForBlock.AddHeaderInfo(string(referencedHeaderData.headerHash), headerForBlock.NewHeaderInfo(nil, true, false, false))
//
// 		mp.ReceivedShardHeader(referencedHeaderData.header, referencedHeaderData.headerHash)
//
// 		time.Sleep(100 * time.Millisecond)
// 		require.Nil(t, err)
// 		require.NotNil(t, mp)
// 		// not yet requested attestation blocks as still missing one header
// 		require.Equal(t, uint32(0), numCalls.Load())
// 		// not yet computed
// 		_, _, missingAttesting := hdrsForBlock.GetMissingData()
// 		require.Equal(t, uint32(0), missingAttesting)
// 	})
// 	t.Run("all needed shard attestation headers received", func(t *testing.T) {
// 		t.Parallel()
//
// 		numCalls := atomic.Uint32{}
// 		arguments := createMetaProcessorArguments(t, noOfShards)
//
// 		poolsHolder, ok := arguments.DataComponents.Datapool().(*dataRetrieverMock.PoolsHolderMock)
// 		require.True(t, ok)
//
// 		headersPoolStub := createPoolsHolderForHeaderRequests()
// 		poolsHolder.SetHeadersPool(headersPoolStub)
// 		requestHandler, ok := arguments.ArgBaseProcessor.RequestHandler.(*testscommon.RequestHandlerStub)
// 		require.True(t, ok)
//
// 		// for requesting attestation header
// 		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
// 			attestationNonce := td[shardID].attestationHeaderData.header.GetNonce()
// 			if nonce != attestationNonce {
// 				require.Fail(t, "nonce should have been %d", attestationNonce)
// 			}
// 			numCalls.Add(1)
// 		}
//
// 		mp, err := blockProcess.NewMetaProcessor(*arguments)
// 		require.Nil(t, err)
// 		require.NotNil(t, mp)
//
// 		hdrsForBlock := mp.GetHdrForBlock()
// 		hdrsForBlock.IncreaseMissingHeaders()
// 		referencedHeaderData := td[0].referencedHeaderData
// 		hdrsForBlock.SetHighestHeaderNonceForShard(0, referencedHeaderData.header.GetNonce()-1)
// 		hdrsForBlock.AddHeaderInfo(string(referencedHeaderData.headerHash), headerForBlock.NewHeaderInfo(nil, true, false, false))
//
// 		// receive the missing header
// 		headersPool := mp.GetDataPool().Headers()
// 		headersPool.AddHeader(referencedHeaderData.headerHash, referencedHeaderData.header)
// 		mp.ReceivedShardHeader(td[0].referencedHeaderData.header, referencedHeaderData.headerHash)
//
// 		time.Sleep(100 * time.Millisecond)
// 		require.Nil(t, err)
// 		require.NotNil(t, mp)
// 		require.Equal(t, uint32(1), numCalls.Load())
// 		_, _, missingAttesting := hdrsForBlock.GetMissingData()
// 		require.Equal(t, uint32(1), missingAttesting)
//
// 		// needs to be done before receiving the last header otherwise it will
// 		// be blocked waiting on writing to the channel
// 		wg := startWaitingForAllHeadersReceivedSignal(t, mp)
//
// 		// receive also the attestation header
// 		attestationHeaderData := td[0].attestationHeaderData
// 		headersPool.AddHeader(attestationHeaderData.headerHash, attestationHeaderData.header)
// 		mp.ReceivedShardHeader(attestationHeaderData.header, attestationHeaderData.headerHash)
// 		wg.Wait()
//
// 		require.Equal(t, uint32(1), numCalls.Load())
// 		_, _, missingAttesting = hdrsForBlock.GetMissingData()
// 		require.Equal(t, uint32(0), missingAttesting)
// 	})
// 	t.Run("all needed shard attestation headers received, when multiple shards headers missing", func(t *testing.T) {
// 		t.Parallel()
//
// 		numCalls := atomic.Uint32{}
// 		arguments := createMetaProcessorArguments(t, noOfShards)
//
// 		poolsHolder, ok := arguments.DataComponents.Datapool().(*dataRetrieverMock.PoolsHolderMock)
// 		require.True(t, ok)
//
// 		headersPoolStub := createPoolsHolderForHeaderRequests()
// 		poolsHolder.SetHeadersPool(headersPoolStub)
// 		requestHandler, ok := arguments.ArgBaseProcessor.RequestHandler.(*testscommon.RequestHandlerStub)
// 		require.True(t, ok)
//
// 		// for requesting attestation header
// 		requestHandler.RequestShardHeaderByNonceCalled = func(shardID uint32, nonce uint64) {
// 			attestationNonce := td[shardID].attestationHeaderData.header.GetNonce()
// 			if nonce != td[shardID].attestationHeaderData.header.GetNonce() {
// 				require.Fail(t, fmt.Sprintf("requested nonce for shard %d should have been %d", shardID, attestationNonce))
// 			}
// 			numCalls.Add(1)
// 		}
//
// 		mp, err := blockProcess.NewMetaProcessor(*arguments)
// 		require.Nil(t, err)
// 		require.NotNil(t, mp)
//
// 		hdrsForBlock := mp.GetHdrForBlock()
// 		hdrsForBlock.IncreaseMissingHeaders()
// 		hdrsForBlock.IncreaseMissingHeaders()
// 		hdrsForBlock.SetHighestHeaderNonceForShard(0, 99)
// 		hdrsForBlock.SetHighestHeaderNonceForShard(1, 97)
// 		hdrsForBlock.AddHeaderInfo(string(td[0].referencedHeaderData.headerHash), headerForBlock.NewHeaderInfo(nil, true, false, false))
// 		hdrsForBlock.AddHeaderInfo(string(td[1].referencedHeaderData.headerHash), headerForBlock.NewHeaderInfo(nil, true, false, false))
//
// 		// receive the missing header for shard 0
// 		headersPool := mp.GetDataPool().Headers()
// 		headersPool.AddHeader(td[0].referencedHeaderData.headerHash, td[0].referencedHeaderData.header)
// 		mp.ReceivedShardHeader(td[0].referencedHeaderData.header, td[0].referencedHeaderData.headerHash)
//
// 		time.Sleep(100 * time.Millisecond)
// 		require.Nil(t, err)
// 		require.NotNil(t, mp)
// 		// the attestation header for shard 0 is not requested as the attestation header for shard 1 is missing
// 		// TODO: refactor request logic to request missing attestation headers as soon as possible
// 		require.Equal(t, uint32(0), numCalls.Load())
// 		_, _, missingAttesting := hdrsForBlock.GetMissingData()
// 		require.Equal(t, uint32(0), missingAttesting)
//
// 		// receive the missing header for shard 1
// 		headersPool.AddHeader(td[1].referencedHeaderData.headerHash, td[1].referencedHeaderData.header)
// 		mp.ReceivedShardHeader(td[1].referencedHeaderData.header, td[1].referencedHeaderData.headerHash)
//
// 		time.Sleep(100 * time.Millisecond)
// 		require.Nil(t, err)
// 		require.NotNil(t, mp)
// 		require.Equal(t, uint32(2), numCalls.Load())
// 		_, _, missingAttesting = hdrsForBlock.GetMissingData()
// 		require.Equal(t, uint32(2), missingAttesting)
//
// 		// needs to be done before receiving the last header otherwise it will
// 		// be blocked writing to a channel no one is reading from
// 		wg := startWaitingForAllHeadersReceivedSignal(t, mp)
//
// 		// receive also the attestation header
// 		headersPool.AddHeader(td[0].attestationHeaderData.headerHash, td[0].attestationHeaderData.header)
// 		mp.ReceivedShardHeader(td[0].attestationHeaderData.header, td[0].attestationHeaderData.headerHash)
//
// 		headersPool.AddHeader(td[1].attestationHeaderData.headerHash, td[1].attestationHeaderData.header)
// 		mp.ReceivedShardHeader(td[1].attestationHeaderData.header, td[1].attestationHeaderData.headerHash)
// 		wg.Wait()
//
// 		time.Sleep(100 * time.Millisecond)
// 		// the receive of an attestation header, if not the last one, will trigger a new request of missing attestation headers
// 		// TODO: refactor request logic to not request recently already requested headers
// 		require.Equal(t, uint32(3), numCalls.Load())
// 		_, _, missingAttesting = hdrsForBlock.GetMissingData()
// 		require.Equal(t, uint32(0), missingAttesting)
// 	})
// }
