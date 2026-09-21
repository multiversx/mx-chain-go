package bootstrap

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
)

func TestBootstrapHistoricalProofRequestNetworks(t *testing.T) {
	t.Parallel()

	for _, fullArchive := range []bool{false, true} {
		for _, offset := range []uint32{0, 2} {
			t.Run(fmt.Sprintf("full archive %t offset %d", fullArchive, offset), func(t *testing.T) {
				coreComp, cryptoComp := createComponentsForEpochStart()
				args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
				args.PrefsConfig.FullArchive = fullArchive
				args.FlagsConfig.StartInEpochOffset = offset
				args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
					IsFlagEnabledInEpochCalled: func(core.EnableEpochFlag, uint32) bool { return true },
				}
				var mainRequests, archiveRequests int
				messenger := func(peer core.PeerID, count *int) *p2pmocks.MessengerStub {
					return &p2pmocks.MessengerStub{
						ConnectedPeersCalled:        func() []core.PeerID { return []core.PeerID{peer} },
						ConnectedPeersOnTopicCalled: func(string) []core.PeerID { return []core.PeerID{peer} },
						SendToConnectedPeerCalled: func(topic string, buff []byte, gotPeer core.PeerID) error {
							require.Equal(t, peer, gotPeer)
							require.Contains(t, topic, "equivalentProofs")
							request := &dataRetriever.RequestData{}
							require.NoError(t, coreComp.InternalMarshalizer().Unmarshal(request, buff))
							require.Equal(t, uint32(10), request.Epoch)
							*count++
							return nil
						},
					}
				}
				args.MainMessenger = messenger("main", &mainRequests)
				args.FullArchiveMessenger = messenger("archive", &archiveRequests)
				provider, err := NewEpochStartBootstrap(args)
				require.NoError(t, err)
				provider.shardCoordinator, err = sharding.NewMultiShardCoordinator(2, 0)
				require.NoError(t, err)
				provider.whiteListHandler = &testscommon.WhiteListHandlerStub{}
				require.NoError(t, provider.createRequestHandler())

				provider.requestHandler.RequestEquivalentProofByHashForEpoch(0, []byte("historical header"), 10)

				require.Positive(t, mainRequests)
				if fullArchive && offset > 0 {
					require.Positive(t, archiveRequests)
				} else {
					require.Zero(t, archiveRequests)
				}
			})
		}
	}
}
