## Test: ex07ChatWithGossip

We’re going to be using 3 separate terminals to act as individual peers.

On your first terminal, go run testP2P.go -l 10000 -secio

Terminal 1  
Follow the instructions that say “Now run…”. Open up a 2nd terminal, go to the same directory and go run testP2P.go -l 10001 -d <given address in the instructions> -secio

Terminal 2  
You’ll see the first terminal detected the new connection!

Terminal 1  
Now following the instructions in the 2nd terminal, open up a 3rd terminal, go to the same working directory and go run testP2P.go -l 10002 -d <given address in the instructions> -secio

Terminal 3  
Check out the 2nd terminal, which detected the connection from the 3rd terminal.

Terminal 2  
Now let’s start inputting our BPM data. Type in “70” in our 1st terminal, give it a few seconds and watch what happens in each terminal.


* Terminal 1 added a new block to its blockchain
* It then broadcast it to Terminal 2
* Terminal 2 compared it against its own blockchain, which only contained its genesis block. It saw Terminal 1 had a longer chain so it replaced its own chain with Terminal 1’s chain. Then it broadcast the new chain to Terminal 3.
* Terminal 3 compared the new chain against its own and replaced it.

All 3 terminals updated their blockchains to the latest state with no central authority! This is the power of Peer-to-Peer.
