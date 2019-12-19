@0xb9f45775755d8a42;
using Go = import "/go.capnp";
$Go.package("capnp");
$Go.import("_");


struct HeaderCapn {
  nonce                  @0:   UInt64;
  prevHash               @1:   Data;
  prevRandSeed           @2:   Data;
  randSeed               @3:   Data;
  pubKeysBitmap          @4:   Data;
  shardId                @5:   UInt32;
  timeStamp              @6:   UInt64;
  round                  @7:   UInt64;
  epoch                  @8:   UInt32;
  blockBodyType          @9:   UInt8;
  signature              @10:  Data;
  leaderSignature        @11:  Data;
  miniBlockHeaders       @12:  List(MiniBlockHeaderCapn);
  peerChanges            @13:  List(PeerChangeCapn);
  rootHash               @14:  Data;
  validatorStatsRootHash @15:  Data;
  metaHdrHashes          @16:  List(Data);
  epochStartMetaHash     @17:  Data;
  txCount                @18:  UInt32;
  chainid                @19:  Data;
}

struct MiniBlockHeaderCapn {
  hash            @0: Data;
  receiverShardID @1: UInt32;
  senderShardID   @2: UInt32;
  txCount         @3: UInt32;
  type            @4: UInt8;
}

struct MiniBlockCapn {
  txHashes        @0:   List(Data);
  receiverShardID @1:   UInt32;
  senderShardID   @2:   UInt32;
  type            @3:   UInt8;
}

struct PeerChangeCapn {
  pubKey       @0:   Data;
  shardIdDest  @1:   UInt32;
}

##compile with:

##
##
##   capnpc  -I$GOPATH/src/github.com/glycerine/go-capnproto -ogo $GOPATH/src/github.com/ElrondNetwork/elrond-go/data/block/capnp/schema.capnp