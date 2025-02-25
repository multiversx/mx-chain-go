package chaos

import "time"

// Controller -
var Controller ChaosController = newNilChaosController()

// ChaosController -
type ChaosController interface {
	Setup() error
	HandleNodeConfig(config interface{})
	HandleNode(node interface{})
	HandleTransaction(transactionHash []byte, transaction interface{}, senderShard uint32, receiverShard uint32)
	HandlePoint(input PointInput) PointOutput
}

// PointInput -
type PointInput struct {
	Name                 string
	ConsensusState       NilInterfaceChecker
	NodePublicKey        string
	Header               NilInterfaceChecker
	Transaction          NilInterfaceChecker
	CorruptibleVariables []interface{}
}

// PointOutput -
type PointOutput struct {
	HasValue  bool
	Error     error
	Boolean   bool
	Duration  time.Duration
	NumberInt int
}

// NilInterfaceChecker -
type NilInterfaceChecker interface {
	IsInterfaceNil() bool
}

type nilChaosController struct {
}

func newNilChaosController() *nilChaosController {
	return &nilChaosController{}
}

// Setup -
func (controller *nilChaosController) Setup() error {
	return nil
}

// HandleNodeConfig -
func (controller *nilChaosController) HandleNodeConfig(config interface{}) {
}

// HandleNode -
func (controller *nilChaosController) HandleNode(node interface{}) {
}

// HandleTransaction -
func (controller *nilChaosController) HandleTransaction(transactionHash []byte, transaction interface{}, senderShard uint32, receiverShard uint32) {
}

// HandlePoint -
func (controller *nilChaosController) HandlePoint(input PointInput) PointOutput {
	return PointOutput{}
}
