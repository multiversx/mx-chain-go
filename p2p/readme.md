# P2P protocol description

The `Messenger` interface with its implementation are 
used to define the way to communicate between Elrond nodes. 

There are 2 ways to send data to the other peers:
1. Broadcasting messages on a `pubsub` using topics;
1. Direct sending messages to the connected peers.

The first type is used to send messages that has to reach every node 
(from corresponding shard, metachain, consensus group, etc.) and the second type is
used to resolve requests comming from directly connected peers. 
