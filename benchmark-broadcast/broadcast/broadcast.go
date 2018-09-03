package broadcast

import (
	"fmt"

	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/benchmark-broadcast/bits"
	"github.com/ElrondNetwork/elrond-go-sandbox/benchmark-broadcast/copy"
	"github.com/ElrondNetwork/elrond-go-sandbox/benchmark-broadcast/types"
	"github.com/satori/go.uuid"
)

var peers []types.Peer

func Broadcast(nodes int, latency float64, peersPerNode int, bandwidth int, blocksize int) (totalLatency float64, nrOfHops, lonelyNodes int) {

	var wg sync.WaitGroup
	blocksSentInParallel := bandwidth / blocksize

	createPeers(nodes)

	for i := range peers {
		wg.Add(1)
		go computeDistance(i, peersPerNode, &wg)

	}
	wg.Wait()

	lonelyNodes = sendMessage(peers[0], nodes, blocksSentInParallel)

	computePathLatency(latency)

	systemLatency, nrOfHops := computeSystemLatencyAndHopNumber()

	peers = peers[:0]
	return systemLatency, nrOfHops, lonelyNodes
}

func createPeers(nodes int) {

	for i := 0; i < nodes; i++ {
		var path []int
		var peerMap []uuid.UUID
		peer := types.Peer{
			Nr:          i,
			Id:          uuid.Must(uuid.NewV4()),
			PeerMap:     peerMap,
			Path:        path,
			PathLatency: 0.0,
			Latency:     -1.0,
		}
		peers = append(peers, peer)
	}
}

func computeDistance(index int, peersPerNode int, wg *sync.WaitGroup) {

	distanceToPeers := make([]int, len(peers))
	var distanceMap []uuid.UUID
	var nodeID = (peers[index].Id).Bytes()

	for i, user := range peers {
		distanceToEveryone(distanceToPeers, nodeID, user, i)
	}

	for j := 0; j < peersPerNode; j++ {
		min := 200
		indexForMin := 0
		for k := 0; k < len(distanceToPeers); k++ {
			if distanceToPeers[k] < min && distanceToPeers[k] != 0 {
				min = distanceToPeers[k]
				indexForMin = k
			}
		}
		distanceMap = append(distanceMap, peers[indexForMin].Id)
		distanceToPeers[indexForMin] = 0
	}

	copy.PeersToPeerMap(&peers[index].PeerMap, distanceMap)

	wg.Done()
}

func distanceToEveryone(distanceToPeers []int, nodeID []byte, user types.Peer, i int) {
	var peerID []byte
	peerID = (user.Id).Bytes()

	XOR, err := bits.XOR(nodeID, peerID)

	if err != nil {
		fmt.Println(err)
	}

	distanceToPeers[i] = bits.CountSet(XOR)
}

func sendMessage(user types.Peer, nodes int, blocksSentInParallel int) int {
	lonelyNode := 0
	var nodesToSend int
	visited := make([]int, nodes)
	var queue []types.Peer

	visited[user.Nr] = 1

	queue = append(queue, user)

	for len(queue) != 0 {

		nodesToSend = 0
		var neighbor types.Peer

		node := queue[0]
		queue = append(queue[:0], queue[1:]...)

		for _, val := range node.PeerMap {
			for _, user := range peers {
				if user.Id == val {
					neighbor = user
				}
			}

			if visited[neighbor.Nr] == 0 {
				visited[neighbor.Nr] = 1
				nodesToSend++
				copy.PeerPath(&neighbor.Path, node.Path, node.Nr)
				queue = append(queue, neighbor)

			}

		}

		node.Latency = (float64)(nodesToSend) / (float64)(blocksSentInParallel)

		for i := range peers {
			if peers[i].Nr == node.Nr {
				peers[i].Latency = node.Latency
				for _, k := range node.Path {
					peers[i].Path = append(peers[i].Path, k)
				}
			}
		}

	}

	for _, val := range visited {
		if val == 0 {
			lonelyNode++
		}
	}

	return lonelyNode

}

func computePathLatency(latency float64) {

	for index := range peers {
		pathLatency := 0.0
		for _, val := range peers[index].Path {
			pathLatency += peers[val].Latency
		}
		pathLatency += (float64)(len(peers[index].Path)) * (float64)(latency)
		peers[index].PathLatency = pathLatency
	}
}

func computeSystemLatencyAndHopNumber() (float64, int) {

	systemLatency, nrOfHops := 0.0, 0
	for i := range peers {
		if len(peers[i].Path) > nrOfHops {
			nrOfHops = len(peers[i].Path)
		}

		if peers[i].PathLatency > systemLatency {
			systemLatency = peers[i].PathLatency
		}
	}

	return systemLatency, nrOfHops
}
