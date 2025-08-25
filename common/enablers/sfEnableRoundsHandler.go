package enablers

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

type sfEnableRoundsHandler struct {
	enableRoundsHandler *enableRoundsHandler
}

func NewSFEnableRoundsHandler(roundsConfig config.RoundConfig, roundNotifier process.RoundNotifier) (*sfEnableRoundsHandler, error) {
	handler := &enableRoundsHandler{}
	err := handler.createAllFlagsMap(roundsConfig)
	if err != nil {
		return nil, err
	}

	roundNotifier.RegisterNotifyHandler(handler)

	return &sfEnableRoundsHandler{handler}, nil
}

func (sf *sfEnableRoundsHandler) SetSuperNovaActivationRound(round uint64) {
	log.Info("SetSuperNovaActivationRound - entered", "round", round)
	superNovaRound, ok := sf.enableRoundsHandler.allFlagsDefined[common.SupernovaRoundFlag]
	if ok {
		log.Info("SetSuperNovaActivationRound", "round", round)
		superNovaRound.activationRound = round
	}
	log.Info("SetSuperNovaActivationRound - exited", "round", round)
}
