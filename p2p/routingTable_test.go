package p2p_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/libp2p/go-libp2p-peer"
	tu "github.com/libp2p/go-testutil"
	"github.com/stretchr/testify/assert"
)

func TestCalculateDistanceDifferentLengths(t *testing.T) {
	buff1 := []byte{0, 0}
	buff2 := []byte{255}

	assert.Equal(t, uint32(8), p2p.ComputeDistanceAD(peer.ID(string(buff1)), peer.ID(string(buff2))))

	buff1 = []byte{0}
	buff2 = []byte{1, 0}

	assert.Equal(t, uint32(1), p2p.ComputeDistanceAD(peer.ID(string(buff1)), peer.ID(string(buff2))))
}

func TestCalculateDistanceLarge(t *testing.T) {
	buff1 := []byte{0}
	buff2 := []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}

	assert.Equal(t, uint32(8*14), p2p.ComputeDistanceAD(peer.ID(string(buff1)), peer.ID(string(buff2))))
}

func TestCalculateDistance(t *testing.T) {
	pid1 := getPID([]byte{0, 0, 0, 1})
	pid2 := getPID([]byte{1, 0, 0, 0})

	assert.Equal(t, uint32(2), p2p.ComputeDistanceAD(pid1, pid2))
}

func TestRoutingTable(t *testing.T) {
	pid1 := getPID([]byte{0, 0, 0, 1})
	pid2 := getPID([]byte{1, 0, 0, 0})
	pid3 := getPID([]byte{0, 0, 0, 3})

	rt := p2p.NewRoutingTable(pid1)
	assert.Equal(t, 1, rt.Len())
	rt.Update(pid2)
	assert.Equal(t, 2, rt.Len())
	dist, err := rt.GetDistance(pid2)
	assert.Nil(t, err)
	assert.Equal(t, uint32(2), dist)

	rt.Update(pid3)
	assert.Equal(t, 3, rt.Len())
	dist, err = rt.GetDistance(pid3)
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), dist)

	peers, dists := rt.Peers()

	assert.Equal(t, peers[0], pid3)
	assert.Equal(t, dists[0], uint32(1))

	assert.Equal(t, peers[1], pid2)
	assert.Equal(t, dists[1], uint32(2))

	assert.Equal(t, peers[2], pid1)
	assert.Equal(t, dists[2], uint32(0))

}

func TestRoutingTableNotFound(t *testing.T) {
	pid1 := getPID([]byte{0, 0, 0, 1})
	pid2 := getPID([]byte{1, 0, 0, 0})

	rt := p2p.NewRoutingTable(pid1)
	assert.True(t, rt.Has(pid1))
	assert.False(t, rt.Has(pid2))

	_, err := rt.GetDistance(pid2)

	assert.NotNil(t, err)
}

func TestRoutingTableMultiple(t *testing.T) {
	pid1 := getPID([]byte{0, 0, 0, 1})
	pid2 := getPID([]byte{1, 0, 0, 0})

	rt := p2p.NewRoutingTable(pid1)
	assert.Equal(t, 1, rt.Len())
	rt.Update(pid2)
	assert.Equal(t, 2, rt.Len())
	rt.Update(pid2)
	assert.Equal(t, 2, rt.Len())
	dist, err := rt.GetDistance(pid2)
	assert.Nil(t, err)
	assert.Equal(t, uint32(2), dist)
}

func getPID(buff []byte) peer.ID {
	return peer.ID(string(buff))
}

// Looks for race conditions in table operations. For a more 'certain'
// test, increase the loop counter from 1000 to a much higher number
// and set GOMAXPROCS above 1
func TestTableMultithreaded(t *testing.T) {
	local := peer.ID("localPeer")
	tab := p2p.NewRoutingTable(local)
	var peers []peer.ID
	for i := 0; i < 500; i++ {
		peers = append(peers, tu.RandPeerIDFatal(t))
	}

	done := make(chan struct{})
	go func() {
		for i := 0; i < 10000; i++ {
			n := rand.Intn(len(peers))
			tab.Update(peers[n])
		}
		done <- struct{}{}
	}()

	go func() {
		for i := 0; i < 10000; i++ {
			n := rand.Intn(len(peers))
			tab.Update(peers[n])
		}
		done <- struct{}{}
	}()

	<-done
	<-done

	for i := 0; i < len(peers); i++ {
		if !tab.Has(peers[i]) {
			assert.Fail(t, fmt.Sprintf("Not found %v", peers[i].Pretty()))
		}
	}
}

func TestClosestPeers(t *testing.T) {
	pid1 := getPID([]byte{0, 0, 0, 1})
	pid2 := getPID([]byte{1, 0, 0, 0})
	pid3 := getPID([]byte{0, 0, 0, 3})
	pid4 := getPID([]byte{0, 1, 0, 0})
	pid5 := getPID([]byte{0, 0, 1, 0})
	pid6 := getPID([]byte{255, 0, 0, 0})

	rt := p2p.NewRoutingTable(pid1)
	rt.Update(pid2)
	rt.Update(pid3)
	rt.Update(pid6)
	rt.Update(pid4)
	rt.Update(pid5)

	peers := rt.NearestPeers(100)
	assert.Equal(t, 5, len(peers))

	peers = rt.NearestPeersAll()
	assert.Equal(t, 5, len(peers))

	assert.Equal(t, pid3, peers[0])
	assert.Equal(t, pid2, peers[1])
	assert.Equal(t, pid4, peers[2])
	assert.Equal(t, pid5, peers[3])
	assert.Equal(t, pid6, peers[4])

	peers = rt.NearestPeers(5)
	assert.Equal(t, 5, len(peers))

	assert.Equal(t, pid3, peers[0])
	assert.Equal(t, pid2, peers[1])
	assert.Equal(t, pid4, peers[2])
	assert.Equal(t, pid5, peers[3])
	assert.Equal(t, pid6, peers[4])

	peers = rt.NearestPeers(3)
	assert.Equal(t, 3, len(peers))

	assert.Equal(t, pid3, peers[0])
	assert.Equal(t, pid2, peers[1])
	assert.Equal(t, pid4, peers[2])

	rt.Print()
}

func TestLargeSetOfPeers(t *testing.T) {
	id := "Node0"

	rt := p2p.NewRoutingTable(peer.ID(id))

	for i := 1; i <= 200; i++ {
		id = "Node" + strconv.Itoa(i)

		rt.Update(peer.ID(id))

		found := rt.Has(peer.ID(id))

		if !found {
			fmt.Printf("Peer %s not found!\n", id)
		}
	}

	rt.Print()

	fmt.Println()
	fmt.Println("Nearest peers:")

	peers := rt.NearestPeers(13)

	for i := 0; i < len(peers); i++ {
		fmt.Println("-", peers[i].Pretty())
	}
}

func TestLonelyPeers(t *testing.T) {
	peers := make([]string, 0)
	rts := make([]p2p.RoutingTable, 0)
	nearest := map[peer.ID][]peer.ID{}

	//use a simple hash method
	h := sha256.New()

	for i := 0; i < 200; i++ {
		h.Write([]byte("Node" + strconv.Itoa(i)))
		hashed := h.Sum(nil)
		p := hex.EncodeToString(hashed)
		peers = append(peers, p)

		rt := p2p.NewRoutingTable(peer.ID(p))

		rts = append(rts, *rt)
		//nearest = append(nearest, []kbucket.ID{})
	}

	for i := 0; i < len(peers); i++ {
		for j := 0; j < len(peers); j++ {
			if i == j {
				continue
			}

			rts[i].Update(peer.ID(peers[j]))
		}
	}

	nearest = make(map[peer.ID][]peer.ID)

	for i := 0; i < len(peers); i++ {
		nearestPeers := rts[i].NearestPeers(10)

		nearest[peer.ID(peers[i])] = nearestPeers
	}

	for i := 0; i < len(peers); i++ {
		assert.Equal(t, 0, testLonelyPeer(peers[i], peers, nearest))
	}
}

func testLonelyPeer(start string, peers []string, conn map[peer.ID][]peer.ID) int {
	reached := make(map[peer.ID]bool)

	for i := 0; i < len(peers); i++ {
		reached[peer.ID(peers[i])] = peers[i] == start
	}

	job := make(map[peer.ID]bool)
	job[peer.ID(start)] = false

	//traverseRec(start, conn, reached)
	traverseRec2(conn, reached, job)

	notReached := 0

	fmt.Printf("%s has not reached: ", start)
	for i := 0; i < len(peers); i++ {
		if !reached[peer.ID(peers[i])] {
			fmt.Printf("%s, ", peers[i])
			notReached++
		}
	}
	fmt.Println()

	return notReached
}

func traverseRec2(conn map[peer.ID][]peer.ID, reached map[peer.ID]bool, job map[peer.ID]bool) {
	var foundPeerID peer.ID = ""

	//find first element that was not processed
	for k, v := range job {
		if !v {
			foundPeerID = k
			break
		}
	}

	if foundPeerID == "" {
		//done, no more peers were found to process
		return
	}

	job[foundPeerID] = true
	reached[foundPeerID] = true

	peersToCheck := conn[foundPeerID]

	//append sub-peers to list
	for _, pid := range peersToCheck {
		_, found := job[pid]

		if !found {
			job[pid] = false
		}
	}

	traverseRec2(conn, reached, job)
}

func BenchmarkAdd(t *testing.B) {
	local := peer.ID("localPeer")
	tab := p2p.NewRoutingTable(local)

	for i := 0; i < t.N; i++ {
		tab.Update(tu.RandPeerIDFatal(t))
	}
}
