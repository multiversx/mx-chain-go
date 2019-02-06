package bn

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

func (wrk *Worker) initMessageChannels() {
	wrk.mutMessageChannels.Lock()

	wrk.MessageChannels = make(map[MessageType]chan *spos.ConsensusData)

	wrk.MessageChannels[MtBlockBody] = make(chan *spos.ConsensusData)
	wrk.MessageChannels[MtBlockHeader] = make(chan *spos.ConsensusData)
	wrk.MessageChannels[MtCommitmentHash] = make(chan *spos.ConsensusData)
	wrk.MessageChannels[MtBitmap] = make(chan *spos.ConsensusData)
	wrk.MessageChannels[MtCommitment] = make(chan *spos.ConsensusData)
	wrk.MessageChannels[MtSignature] = make(chan *spos.ConsensusData)

	wrk.mutMessageChannels.Unlock()
}

func (wrk *Worker) initReceivedMessages() {
	wrk.mutReceivedMessages.Lock()

	wrk.ReceivedMessages = make(map[MessageType][]*spos.ConsensusData)

	wrk.ReceivedMessages[MtBlockBody] = make([]*spos.ConsensusData, 0)
	wrk.ReceivedMessages[MtBlockHeader] = make([]*spos.ConsensusData, 0)
	wrk.ReceivedMessages[MtCommitmentHash] = make([]*spos.ConsensusData, 0)
	wrk.ReceivedMessages[MtBitmap] = make([]*spos.ConsensusData, 0)
	wrk.ReceivedMessages[MtCommitment] = make([]*spos.ConsensusData, 0)
	wrk.ReceivedMessages[MtSignature] = make([]*spos.ConsensusData, 0)

	wrk.mutReceivedMessages.Unlock()
}

func (wrk *Worker) cleanReceivedMessages() {
	wrk.mutReceivedMessages.Lock()

	for i := MtBlockBody; i <= MtSignature; i++ {
		cnsDataList := wrk.ReceivedMessages[i]

		if len(cnsDataList) == 0 {
			continue
		}

		cleanedCnsDtaList := wrk.getCleanedList(cnsDataList)
		wrk.ReceivedMessages[i] = cleanedCnsDtaList
	}

	wrk.mutReceivedMessages.Unlock()
}

func (wrk *Worker) getCleanedList(cnsDataList []*spos.ConsensusData) []*spos.ConsensusData {
	cleanedCnsDataList := make([]*spos.ConsensusData, 0)

	for i := 0; i < len(cnsDataList); i++ {
		if cnsDataList[i] == nil {
			continue
		}

		if wrk.shouldDropConsensusMessage(cnsDataList[i]) {
			continue
		}

		cleanedCnsDataList = append(cleanedCnsDataList, cnsDataList[i])
	}

	return cleanedCnsDataList
}

func (wrk *Worker) shouldDropConsensusMessage(cnsDta *spos.ConsensusData) bool {
	if cnsDta.RoundIndex < wrk.SPoS.Chr.Round().Index() {
		return true
	}

	if cnsDta.RoundIndex == wrk.SPoS.Chr.Round().Index() &&
		wrk.SPoS.Chr.GetSubround() > chronology.SubroundId(SrEndRound) {
		return true
	}

	return false
}

// ReceivedMessage method redirects the received message to the channel which should handle it
func (wrk *Worker) ReceivedMessage(name string, data interface{}, msgInfo *p2p.MessageInfo) {
	if wrk.SPoS.Chr.IsCancelled() {
		return
	}

	cnsDta, ok := data.(*spos.ConsensusData)

	if !ok {
		return
	}

	senderOK := wrk.SPoS.IsNodeInConsensusGroup(string(cnsDta.PubKey))

	if !senderOK {
		return
	}

	if wrk.shouldDropConsensusMessage(cnsDta) {
		return
	}

	if wrk.SPoS.SelfPubKey() == string(cnsDta.PubKey) {
		return
	}

	sigVerifErr := wrk.checkSignature(cnsDta)
	if sigVerifErr != nil {
		return
	}

	wrk.ReceivedMessageChannel <- cnsDta
}

func (wrk *Worker) checkSignature(cnsDta *spos.ConsensusData) error {
	if cnsDta == nil {
		return spos.ErrNilConsensusData
	}

	if cnsDta.PubKey == nil {
		return spos.ErrNilPublicKey
	}

	if cnsDta.Signature == nil {
		return spos.ErrNilSignature
	}

	pubKey, err := wrk.keyGen.PublicKeyFromByteArray(cnsDta.PubKey)

	if err != nil {
		return err
	}

	dataNoSig := *cnsDta
	signature := cnsDta.Signature

	dataNoSig.Signature = nil
	dataNoSigString, err := wrk.marshalizer.Marshal(dataNoSig)

	if err != nil {
		return err
	}

	err = pubKey.Verify(dataNoSigString, signature)

	return err
}

func (wrk *Worker) checkReceivedMessageChannel() {
	for {
		select {
		case cnsDta := <-wrk.ReceivedMessageChannel:
			wrk.executeReceivedMessages(cnsDta)
		}
	}
}

func (wrk *Worker) executeReceivedMessages(cnsDta *spos.ConsensusData) {
	wrk.mutReceivedMessages.Lock()

	msgType := MessageType(cnsDta.MsgType)

	cnsDataList := wrk.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.ReceivedMessages[msgType] = cnsDataList

	for i := MtBlockBody; i <= MtSignature; i++ {
		cnsDataList = wrk.ReceivedMessages[i]

		if len(cnsDataList) == 0 {
			continue
		}

		wrk.executeMessage(cnsDataList)
		cleanedCnsDtaList := wrk.getCleanedList(cnsDataList)
		wrk.ReceivedMessages[i] = cleanedCnsDtaList
	}

	wrk.mutReceivedMessages.Unlock()
}

func (wrk *Worker) executeMessage(cnsDtaList []*spos.ConsensusData) {
	for i, cnsDta := range cnsDtaList {
		if cnsDta == nil {
			continue
		}

		if wrk.boot.ShouldSync() {
			continue
		}

		if wrk.shouldDropConsensusMessage(cnsDta) {
			continue
		}

		msgType := MessageType(cnsDta.MsgType)

		switch msgType {
		case MtBlockBody:
			if wrk.SPoS.Status(SrStartRound) != spos.SsFinished {
				continue
			}
		case MtBlockHeader:
			if wrk.SPoS.Status(SrStartRound) != spos.SsFinished {
				continue
			}
		case MtCommitmentHash:
			if wrk.SPoS.Status(SrBlock) != spos.SsFinished {
				continue
			}
		case MtBitmap:
			if wrk.SPoS.Status(SrBlock) != spos.SsFinished {
				continue
			}
		case MtCommitment:
			if wrk.SPoS.Status(SrBitmap) != spos.SsFinished {
				continue
			}
		case MtSignature:
			if wrk.SPoS.Status(SrBitmap) != spos.SsFinished {
				continue
			}
		}

		cnsDtaList[i] = nil

		if ch, ok := wrk.MessageChannels[msgType]; ok {
			ch <- cnsDta
		}
	}
}

// checkChannels method is used to listen to the channels through which node receives and consumes,
// during the round, different messages from the nodes which are in the validators group
func (wrk *Worker) checkChannels() {
	for {
		select {
		case rcvDta := <-wrk.MessageChannels[MtBlockBody]:
			if wrk.receivedBlockBody(rcvDta) {
				wrk.checkBlockConsensus()
			}
		case rcvDta := <-wrk.MessageChannels[MtBlockHeader]:
			if wrk.receivedBlockHeader(rcvDta) {
				wrk.checkBlockConsensus()
			}
		case rcvDta := <-wrk.MessageChannels[MtCommitmentHash]:
			if wrk.receivedCommitmentHash(rcvDta) {
				wrk.checkCommitmentHashConsensus()
			}
		case rcvDta := <-wrk.MessageChannels[MtBitmap]:
			if wrk.receivedBitmap(rcvDta) {
				wrk.checkBitmapConsensus()
			}
		case rcvDta := <-wrk.MessageChannels[MtCommitment]:
			if wrk.receivedCommitment(rcvDta) {
				wrk.checkCommitmentConsensus()
			}
		case rcvDta := <-wrk.MessageChannels[MtSignature]:
			if wrk.receivedSignature(rcvDta) {
				wrk.checkSignatureConsensus()
			}
		}
	}
}

// sendConsensusMessage sends the consensus message
func (wrk *Worker) sendConsensusMessage(cnsDta *spos.ConsensusData) bool {
	signature, err := wrk.genConsensusDataSignature(cnsDta)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	signedCnsData := *cnsDta
	signedCnsData.Signature = signature

	if wrk.SendMessage == nil {
		log.Error("SendMessage call back function is not set\n")
		return false
	}

	go wrk.SendMessage(&signedCnsData)

	return true
}

func (wrk *Worker) genConsensusDataSignature(cnsDta *spos.ConsensusData) ([]byte, error) {

	cnsDtaStr, err := wrk.marshalizer.Marshal(cnsDta)

	if err != nil {
		return nil, err
	}

	signature, err := wrk.privKey.Sign(cnsDtaStr)

	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (wrk *Worker) broadcastTxBlockBody(blockBody *block.TxBlockBody) error {
	if blockBody == nil {
		return spos.ErrNilTxBlockBody
	}

	message, err := wrk.marshalizer.Marshal(blockBody)

	if err != nil {
		return err
	}

	// send message
	if wrk.BroadcastBlockBody == nil {
		return spos.ErrNilOnBroadcastTxBlockBody
	}

	go wrk.BroadcastBlockBody(message)

	return nil
}

func (wrk *Worker) broadcastHeader(header *block.Header) error {
	if header == nil {
		return spos.ErrNilBlockHeader
	}

	message, err := wrk.marshalizer.Marshal(header)

	if err != nil {
		return err
	}

	// send message
	if wrk.BroadcastHeader == nil {
		return spos.ErrNilOnBroadcastHeader
	}

	go wrk.BroadcastHeader(message)

	return nil
}
