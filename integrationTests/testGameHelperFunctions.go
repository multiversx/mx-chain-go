package integrationTests

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
)

// ScCallTxWithParams creates and sends a SC tx call or deploy with all major parameters provided
func ScCallTxWithParams(
	senderNode *TestProcessorNode,
	sk crypto.PrivateKey,
	nonce uint64,
	data string,
	value *big.Int,
	gasLimit uint64,
	gasPrice uint64,
) {

	fmt.Println("Deploying SC...")
	pkBuff, _ := sk.GeneratePublic().ToByteArray()
	txArgsObject := &txArgs{
		nonce:    nonce,
		value:    value,
		rcvAddr:  make([]byte, 32),
		sndAddr:  pkBuff,
		data:     data,
		gasLimit: gasLimit,
		gasPrice: gasPrice,
	}

	txDeploy := generateTx(
		sk,
		senderNode.OwnAccount.SingleSigner,
		txArgsObject,
	)

	_, _ = senderNode.SendTransaction(txDeploy)
	fmt.Println("Delaying for disseminating the deploy tx...")
	time.Sleep(StepDelay)
}

// DeployScTx creates and sends a SC tx
func DeployScTx(nodes []*TestProcessorNode, senderIdx int, scCode string, vmType []byte, initArguments string) {
	fmt.Println("Deploying SC...")
	scCodeMetadataString := "0000"
	data := scCode + "@" + hex.EncodeToString(vmType) + "@" + scCodeMetadataString
	if len(initArguments) > 0 {
		data += "@" + initArguments
	}

	txDeploy := generateTx(
		nodes[senderIdx].OwnAccount.SkTxSign,
		nodes[senderIdx].OwnAccount.SingleSigner,
		&txArgs{
			nonce:    nodes[senderIdx].OwnAccount.Nonce,
			value:    big.NewInt(0),
			rcvAddr:  make([]byte, 32),
			sndAddr:  nodes[senderIdx].OwnAccount.PkTxSignBytes,
			data:     data,
			gasLimit: MaxGasLimitPerBlock - 1,
			gasPrice: MinTxGasPrice,
		})
	nodes[senderIdx].OwnAccount.Nonce++
	_, _ = nodes[senderIdx].SendTransaction(txDeploy)
	fmt.Println("Delaying for disseminating the deploy tx...")
	time.Sleep(StepDelay)

	fmt.Println(MakeDisplayTable(nodes))
}

// PlayerSendsTransaction creates and sends a transaction to the SC
func PlayerSendsTransaction(
	nodes []*TestProcessorNode,
	player *TestWalletAccount,
	scAddress []byte,
	value *big.Int,
	txData string,
	gasLimit uint64,
) {
	txDispatcherNode := getNodeWithinSameShardAsPlayer(nodes, player.Address)
	txScCall := generateTx(
		player.SkTxSign,
		player.SingleSigner,
		&txArgs{
			nonce:    player.Nonce,
			value:    value,
			rcvAddr:  scAddress,
			sndAddr:  player.Address,
			data:     txData,
			gasLimit: gasLimit,
			gasPrice: MinTxGasPrice,
		})
	player.Nonce++
	newBalance := big.NewInt(0)
	newBalance = newBalance.Sub(player.Balance, value)
	player.Balance = player.Balance.Set(newBalance)

	_, _ = txDispatcherNode.SendTransaction(txScCall)
}

func getNodeWithinSameShardAsPlayer(
	nodes []*TestProcessorNode,
	player []byte,
) *TestProcessorNode {
	nodeWithCaller := nodes[0]
	playerShId := nodeWithCaller.ShardCoordinator.ComputeId(player)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() == playerShId {
			nodeWithCaller = node
			break
		}
	}

	return nodeWithCaller
}
