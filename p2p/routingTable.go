package p2p

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/gogo/protobuf/sortkeys"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
)

type RoutingTable struct {
	list    *list.List
	current peer.ID
	mut     sync.RWMutex
	dists1  map[uint64][]peer.ID
	dists2  map[peer.ID]uint64
	dists3  []uint64

	ComputeDistance func(pid1 peer.ID, pid2 peer.ID) uint64
}

func NewRoutingTable(crt peer.ID) *RoutingTable {
	rt := &RoutingTable{current: crt, list: list.New(),
		dists1: make(map[uint64][]peer.ID), dists2: make(map[peer.ID]uint64),
		ComputeDistance: ComputeDistanceAD}
	rt.Update(crt)

	return rt
}

func (rt *RoutingTable) Peers() ([]peer.ID, []uint64) {
	rt.mut.RLock()
	defer rt.mut.RUnlock()

	ps := make([]peer.ID, 0, rt.list.Len())
	dists := make([]uint64, 0, rt.list.Len())
	for e := rt.list.Front(); e != nil; e = e.Next() {
		id := e.Value.(peer.ID)
		ps = append(ps, id)
		dists = append(dists, rt.dists2[id])
	}

	return ps, dists
}

func (rt *RoutingTable) Has(id peer.ID) bool {
	rt.mut.RLock()
	defer rt.mut.RUnlock()

	for e := rt.list.Front(); e != nil; e = e.Next() {
		if e.Value.(peer.ID) == id {
			return true
		}
	}

	return false
}

func (rt *RoutingTable) Len() int {
	rt.mut.RLock()
	defer rt.mut.RUnlock()

	return rt.list.Len()
}

func (rt *RoutingTable) Update(p peer.ID) {
	rt.mut.Lock()
	defer rt.mut.Unlock()

	//compute distance (current - p)
	dist := uint64(0)
	if rt.ComputeDistance != nil {
		dist = rt.ComputeDistance(rt.current, p)
	}

	//get array from map distance->peers
	rt.dists2[p] = dist
	pids := rt.dists1[dist]
	if pids == nil {
		pids = []peer.ID{}
	}

	if len(pids) == 0 {
		//distance is new, append to list
		rt.dists3 = append(rt.dists3, dist)
		//keep sorted
		sortkeys.Uint64s(rt.dists3[:])
	}

	//search the peer is array of peers
	found := false
	for i := 0; i < len(pids); i++ {
		if pids[i] == p {
			found = true
			break
		}
	}

	if !found {
		//add peer into array, save array
		pids = append(pids, p)
		rt.list.PushFront(p)
	}
	//update distance in map pid->distance
	rt.dists1[dist] = pids
}

func (rt *RoutingTable) GetDistance(p peer.ID) (uint64, error) {
	rt.mut.RLock()
	defer rt.mut.RUnlock()

	if !rt.Has(p) {
		return uint64(0), errors.New(fmt.Sprintf("Peer ID %v was not found!", p.Pretty()))
	}

	return rt.dists2[p], nil
}

func ComputeDistanceAD(p1 peer.ID, p2 peer.ID) uint64 {
	buff1 := []byte(p1)
	buff2 := []byte(p2)

	for len(buff1) < len(buff2) {
		buff1 = append([]byte{0}, buff1...)
	}

	for len(buff2) < len(buff1) {
		buff2 = append([]byte{0}, buff2...)
	}

	var sum uint64 = 0
	for i := 0; i < len(buff1); i++ {
		sum += CountOneBits(buff1[i] ^ buff2[i])
	}

	return sum
}

func CountOneBits(num byte) uint64 {
	var sum uint64 = 0

	operand := byte(128)

	for operand > 0 {
		if (num & operand) > 0 {
			sum++
		}

		operand = operand / 2
	}

	return sum
}

func (rt *RoutingTable) NearestPeers(maxNo int) []peer.ID {
	found := 0
	peers := make([]peer.ID, 0)

	for i := 0; i < len(rt.dists3) && found < maxNo; i++ {
		//get the peers by using the distance as key.
		//started from smallest
		distPeers := rt.dists1[rt.dists3[i]]

		for j := 0; j < len(distPeers) && found < maxNo; j++ {
			if distPeers[j] == rt.current {
				//ignore current peer
				continue
			}

			peers = append(peers, distPeers[j])
			found++
		}
	}

	return peers
}

func (rt *RoutingTable) NearestPeersAll() []peer.ID {
	peers := make([]peer.ID, 0)

	for i := 0; i < len(rt.dists3); i++ {
		//get the peers by using the distance as key.
		//started from smallest
		distPeers := rt.dists1[rt.dists3[i]]

		for j := 0; j < len(distPeers); j++ {
			if distPeers[j] == rt.current {
				//ignore current peer
				continue
			}

			peers = append(peers, distPeers[j])
		}
	}

	return peers
}

func (rt *RoutingTable) Print() {
	for i := 0; i < len(rt.dists3); i++ {
		fmt.Printf("Distance %d:\n", rt.dists3[i])

		pids := rt.dists1[rt.dists3[i]]

		for j := 0; j < len(pids); j++ {
			fmt.Printf("\t %v", pids[j].Pretty())
			if pids[j] == rt.current {
				fmt.Printf("*")
			}
			fmt.Println()
		}
	}
}
