package genesis

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/files"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
)

func TestNewStateExporter(t *testing.T) {
	tests := []struct {
		name    string
		args    ArgsNewStateExporter
		exError error
	}{
		{name: "NilCoordinator", args: ArgsNewStateExporter{Marshalizer: &mock.MarshalizerMock{}, ShardCoordinator: nil,
			StateSyncer: &mock.SyncStateStub{}, Writer: &mock.MultiFileWriterStub{}}, exError: data.ErrNilShardCoordinator},
		{name: "NilStateSyncer", args: ArgsNewStateExporter{Marshalizer: &mock.MarshalizerMock{}, ShardCoordinator: mock.NewOneShardCoordinatorMock(),
			StateSyncer: nil, Writer: &mock.MultiFileWriterStub{}}, exError: update.ErrNilStateSyncer},
		{name: "NilMarshalizer", args: ArgsNewStateExporter{Marshalizer: nil, ShardCoordinator: mock.NewOneShardCoordinatorMock(),
			StateSyncer: &mock.SyncStateStub{}, Writer: &mock.MultiFileWriterStub{}}, exError: data.ErrNilMarshalizer},
		{name: "NilWriter", args: ArgsNewStateExporter{Marshalizer: &mock.MarshalizerMock{}, ShardCoordinator: mock.NewOneShardCoordinatorMock(),
			StateSyncer: &mock.SyncStateStub{}, Writer: nil}, exError: epochStart.ErrNilStorage},
		{name: "Ok", args: ArgsNewStateExporter{Marshalizer: &mock.MarshalizerMock{}, ShardCoordinator: mock.NewOneShardCoordinatorMock(),
			StateSyncer: &mock.SyncStateStub{}, Writer: &mock.MultiFileWriterStub{}}, exError: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewStateExporter(tt.args)
			require.Equal(t, tt.exError, err)
		})
	}
}

func TestExportMeta(t *testing.T) {
	t.Parallel()

	testPath := "./testFiles"
	storer := mock.NewStorerMock()
	argsWriter := files.ArgsNewMultiFileWriter{ExportFolder: testPath, ExportStore: storer}
	writer, _ := files.NewMultiFileWriter(argsWriter)

	args := ArgsNewStateExporter{
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		Marshalizer:      &mock.MarshalizerMock{},
		StateSyncer:      &mock.SyncStateStub{},
		Writer:           writer,
	}

	stateExporter, _ := NewStateExporter(args)
	err := stateExporter.exportMeta()
	require.NotNil(t, err)
}
