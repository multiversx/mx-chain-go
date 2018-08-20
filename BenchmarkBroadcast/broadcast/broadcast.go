package broadcast

import (
	"BenchmarkBroadcast/peer"
	"fmt"
	"github.com/satori/go.uuid"
)

var peers []peer.Peer
var tempQueue []peer.Peer

func Broadcast(nodes int, latency int, peersPerNode int, bandwidth int, blocksize int) (int, int) {

	blocksSentInParallel := bandwidth / blocksize

	for i := 0; i < nodes; i++ {
		peers = append(peers, peer.New())
	}

	for i := range peers {
		computeDistance(peers[i], peersPerNode)
	}

	SendMessage(peers[0], nodes, blocksSentInParallel)

	ComputePathLatency()

	systemLatency, nrOfHops := ComputeSystemLatencyAndHopNumber()
	/*for i := range peers {
		fmt.Printf("Peer %v latency %v path %v \n", peers[i].Nr, peers[i].Latency, peers[i].Path)
	}*/

	return systemLatency * latency, nrOfHops
}

func ComputePathLatency() {
	for index := range peers {
		pathLatency := 0
		for _, val := range peers[index].Path {
			pathLatency += peers[val].Latency
		}
		peers[index].PathLatency = pathLatency
	}
}

func ComputeSystemLatencyAndHopNumber() (int, int) {
	systemLatency, nrOfHops := 0, 0
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

func SendMessage(user peer.Peer, nodes int, blocksSentInParallel int) {

	var totalNodeLatency int
	var nodesToSend int
	visited := make([]int, nodes)
	var queue []peer.Peer

	visited[user.Nr] = 1

	queue = append(queue, user)
	tempQueue = append(tempQueue, user)

	for len(queue) != 0 {

		nodesToSend = 0
		var neighbor peer.Peer

		node := queue[0]
		queue = append(queue[:0], queue[1:]...)

		for key := range node.PeerMap {
			for _, user := range peers {
				if user.Id == key {
					neighbor = user
				}
			}

			if visited[neighbor.Nr] == 0 {
				visited[neighbor.Nr] = 1
				nodesToSend++
				neighbor.ParentNode = node.Nr
				(&neighbor).SetPeerPath(node.Path, node.Nr)
				queue = append(queue, neighbor)
				tempQueue = append(tempQueue, neighbor)
			}

		}

		if nodesToSend%blocksSentInParallel == 0 {
			totalNodeLatency = nodesToSend / blocksSentInParallel
		} else {
			totalNodeLatency = nodesToSend/blocksSentInParallel + 1
		}
		node.Latency = totalNodeLatency

		for i := range peers {
			if peers[i].Nr == node.Nr {
				peers[i].Latency = node.Latency
				peers[i].ParentNode = node.ParentNode
				for _, k := range node.Path {
					peers[i].Path = append(peers[i].Path, k)
				}
			}
		}

	}

}

func computeDistance(node peer.Peer, peersPerNode int) {
	distanceToPeers := make([]int, len(peers))
	distanceMap := make(map[uuid.UUID]int)
	var nodeID []byte = (node.Id).Bytes()
	var peerID []byte

	for i, user := range peers {
		peerID = (user.Id).Bytes()

		XOR, err := XORBytes(nodeID, peerID)

		if err != nil {
			fmt.Println(err)
		}

		distanceToPeers[i] = CountSetBits(XOR)

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
		distanceMap[peers[indexForMin].Id] = min
		distanceToPeers[indexForMin] = 0
	}

	node.SetPeersToPeerMap(distanceMap)

}

func XORBytes(a, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("length of byte slices is not equivalent: %d != %d", len(a), len(b))
	}

	buf := make([]byte, len(a))

	for i := range a {
		buf[i] = a[i] ^ b[i]
	}

	return buf, nil
}

func CountSetBits(nr []byte) int {
	count := 0
	for _, byteNo := range nr {
		for i := 0; i < 8; i++ {
			if byteNo&1 == 1 {
				count++
			}
			byteNo >>= 1
		}
	}
	return count
}
