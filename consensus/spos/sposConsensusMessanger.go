package spos

func (sposWorker *SPOSConsensusWorker) initMessagesReceived() {
	sposWorker.mutMessagesReceived.Lock()

	sposWorker.MessagesReceived = make(map[MessageType][]*ConsensusData)

	sposWorker.MessagesReceived[MtBlockBody] = make([]*ConsensusData, 0)
	sposWorker.MessagesReceived[MtBlockHeader] = make([]*ConsensusData, 0)
	sposWorker.MessagesReceived[MtCommitmentHash] = make([]*ConsensusData, 0)
	sposWorker.MessagesReceived[MtBitmap] = make([]*ConsensusData, 0)
	sposWorker.MessagesReceived[MtCommitment] = make([]*ConsensusData, 0)
	sposWorker.MessagesReceived[MtSignature] = make([]*ConsensusData, 0)

	sposWorker.mutMessagesReceived.Unlock()
}

func (sposWorker *SPOSConsensusWorker) executeReceivedMessages(cnsDta *ConsensusData) {
	sposWorker.mutMessagesReceived.Lock()

	cnsDataList := sposWorker.MessagesReceived[cnsDta.MsgType]
	cnsDataList = append(cnsDataList, cnsDta)
	sposWorker.MessagesReceived[cnsDta.MsgType] = cnsDataList

	for i := MtBlockBody; i <= MtSignature; i++ {
		cnsDataList = sposWorker.MessagesReceived[i]

		if len(cnsDataList) == 0 {
			continue
		}

		sposWorker.executeMessage(cnsDataList)
		cleanedCnsDtaList := sposWorker.getCleanedList(cnsDataList)
		sposWorker.MessagesReceived[i] = cleanedCnsDtaList
	}

	sposWorker.mutMessagesReceived.Unlock()
}

func (sposWorker *SPOSConsensusWorker) executeMessage(cnsDtaList []*ConsensusData) {
	for i, cnsDta := range cnsDtaList {
		if cnsDta == nil {
			continue
		}

		switch cnsDta.MsgType {
		case MtBlockBody:
			if sposWorker.ShouldSync() {
				continue
			}
		case MtBlockHeader:
			if sposWorker.ShouldSync() {
				continue
			}
		case MtCommitmentHash:
			if sposWorker.Cns.Status(SrBlock) != SsFinished {
				continue
			}
		case MtBitmap:
			if sposWorker.Cns.Status(SrBlock) != SsFinished {
				continue
			}
		case MtCommitment:
			if sposWorker.Cns.Status(SrBitmap) != SsFinished {
				continue
			}
		case MtSignature:
			if sposWorker.Cns.Status(SrBitmap) != SsFinished {
				continue
			}
		}

		cnsDtaList[i] = nil

		if ch, ok := sposWorker.MessageChannels[cnsDta.MsgType]; ok {
			ch <- cnsDta
		}
	}
}

func (sposWorker *SPOSConsensusWorker) getCleanedList(cnsDtaList []*ConsensusData) []*ConsensusData {
	cleanedCnsDtaList := make([]*ConsensusData, 0)

	for i := 0; i < len(cnsDtaList); i++ {
		if cnsDtaList[i] == nil {
			continue
		}

		cleanedCnsDtaList = append(cleanedCnsDtaList, cnsDtaList[i])
	}

	return cleanedCnsDtaList
}

func (sposWorker *SPOSConsensusWorker) checkReceivedMessageChannel() {
	for {
		select {
		case cnsDta := <-sposWorker.ReceivedMessageChannel:
			sposWorker.executeReceivedMessages(cnsDta)
		}
	}
}
