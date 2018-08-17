package broadcast

import (
	"BroadcastBenchmark/peer"
	"fmt"
	"github.com/satori/go.uuid"
)

var peers []peer.Peer

func Broadcast(nodes int, latencty int, peersPerNode int, bandwidth int, blocksize int) (int, int) {

	for i := 0; i < nodes; i++ {
		peers = append(peers, peer.New())

	}

	for i := range peers {
		computeDistance(peers[i], peersPerNode)
	}

	/*for i := 0; i < nodes; i++ {

		fmt.Printf("\n peer %v \n", peers[i].Nr)
		for hash := range peers[i].PeerMap {
			fmt.Printf("%v ", FindByHash(hash))
		}
	}*/

	SendMessage(peers[0], nodes)

	return 10, 20
}

/*func FindByHash(hash uuid.UUID) int {
	for _, user := range peers {
		if user.Id == hash {
			return user.Nr
		}
	}
	return -1
}*/

func SendMessage(user peer.Peer, nodes int) {

	visited := make([]int, nodes)
	var queue []peer.Peer

	visited[user.Nr] = 1

	queue = append(queue, user)

	for len(queue) != 0 {

		var neighbor peer.Peer

		node := queue[0]
		queue = append(queue[:0], queue[1:]...)

		//fmt.Printf("User %v \n", node.Nr)

		for key := range node.PeerMap {
			for _, user := range peers {
				if user.Id == key {
					neighbor = user
				}
			}

			if visited[neighbor.Nr] == 0 {
				visited[neighbor.Nr] = 1

				queue = append(queue, neighbor)
			}

		}

	}

	/*if user.Message == "" {

		for _, node := range msgPath {
			user.MessagePath = append(user.MessagePath, node)
		}
		user.MessagePath = append(user.MessagePath, user.Nr)

		fmt.Printf("Node %v got the message %v and has path %v !\n", user.Id, user.Message, user.MessagePath)

		user.Message = message
		for k := range user.PeerMap {
			id := GetUserById(k)
			if id >= 0 {

				SendMessage(&peers[id], message, user.MessagePath)
			}
		}
	}*/
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
