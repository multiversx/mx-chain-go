package spos

import (
	"github.com/davecgh/go-spew/spew"
)

func (sposWorker *SPOSConsensusWorker) initReceivedMessages() {
	sposWorker.mutMessagesReceived.Lock()

	sposWorker.ReceivedMessages = make(map[MessageType][]*ConsensusData)

	sposWorker.ReceivedMessages[MtBlockBody] = make([]*ConsensusData, 0)
	sposWorker.ReceivedMessages[MtBlockHeader] = make([]*ConsensusData, 0)
	sposWorker.ReceivedMessages[MtCommitmentHash] = make([]*ConsensusData, 0)
	sposWorker.ReceivedMessages[MtBitmap] = make([]*ConsensusData, 0)
	sposWorker.ReceivedMessages[MtCommitment] = make([]*ConsensusData, 0)
	sposWorker.ReceivedMessages[MtSignature] = make([]*ConsensusData, 0)

	sposWorker.mutMessagesReceived.Unlock()
}

func (sposWorker *SPOSConsensusWorker) displayReceivedMessages() {
	sposWorker.mutMessagesReceived.RLock()

	for i := MtBlockBody; i <= MtSignature; i++ {
		cnsDataList := sposWorker.ReceivedMessages[i]

		if len(cnsDataList) == 0 {
			continue
		}

		for j := 0; j < len(cnsDataList); j++ {
			spew.Dump(cnsDataList[j])
		}
	}

	sposWorker.mutMessagesReceived.RUnlock()
}

func (sposWorker *SPOSConsensusWorker) cleanReceivedMessages() {
	sposWorker.mutMessagesReceived.Lock()

	for i := MtBlockBody; i <= MtSignature; i++ {
		cnsDataList := sposWorker.ReceivedMessages[i]

		if len(cnsDataList) == 0 {
			continue
		}

		cleanedCnsDtaList := sposWorker.getCleanedList(cnsDataList)
		sposWorker.ReceivedMessages[i] = cleanedCnsDtaList
	}

	sposWorker.mutMessagesReceived.Unlock()
}

func (sposWorker *SPOSConsensusWorker) executeReceivedMessages(cnsDta *ConsensusData) {
	sposWorker.mutMessagesReceived.Lock()

	cnsDataList := sposWorker.ReceivedMessages[cnsDta.MsgType]
	cnsDataList = append(cnsDataList, cnsDta)
	sposWorker.ReceivedMessages[cnsDta.MsgType] = cnsDataList

	for i := MtBlockBody; i <= MtSignature; i++ {
		cnsDataList = sposWorker.ReceivedMessages[i]

		if len(cnsDataList) == 0 {
			continue
		}

		sposWorker.executeMessage(cnsDataList)
		cleanedCnsDtaList := sposWorker.getCleanedList(cnsDataList)
		sposWorker.ReceivedMessages[i] = cleanedCnsDtaList
	}

	sposWorker.mutMessagesReceived.Unlock()
}

func (sposWorker *SPOSConsensusWorker) executeMessage(cnsDtaList []*ConsensusData) {
	for i, cnsDta := range cnsDtaList {
		if cnsDta == nil {
			continue
		}

		if sposWorker.dropConsensusMessage(cnsDta) {
			continue
		}

		switch cnsDta.MsgType {
		case MtBlockBody:
			if sposWorker.Cns.Status(SrStartRound) != SsFinished ||
				sposWorker.ShouldSync() {
				continue
			}
		case MtBlockHeader:
			if sposWorker.Cns.Status(SrStartRound) != SsFinished ||
				sposWorker.ShouldSync() {
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

func (sposWorker *SPOSConsensusWorker) getCleanedList(cnsDataList []*ConsensusData) []*ConsensusData {
	cleanedCnsDataList := make([]*ConsensusData, 0)

	for i := 0; i < len(cnsDataList); i++ {
		if cnsDataList[i] == nil {
			continue
		}

		if sposWorker.dropConsensusMessage(cnsDataList[i]) {
			continue
		}

		cleanedCnsDataList = append(cleanedCnsDataList, cnsDataList[i])
	}

	return cleanedCnsDataList
}

func (sposWorker *SPOSConsensusWorker) checkReceivedMessageChannel() {
	for {
		select {
		case cnsDta := <-sposWorker.ReceivedMessageChannel:
			sposWorker.executeReceivedMessages(cnsDta)
		}
	}
}
