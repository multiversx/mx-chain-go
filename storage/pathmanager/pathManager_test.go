package pathmanager_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/pathmanager"
	"github.com/stretchr/testify/assert"
)

func TestNewPathManager_EmptyPruningPathTemplateShouldErr(t *testing.T) {
	t.Parallel()

	pm, err := pathmanager.NewPathManager("", "shard_[S]/[I]")
	assert.Nil(t, pm)
	assert.Equal(t, storage.ErrEmptyPruningPathTemplate, err)
}

func TestNewPathManager_EmptyStaticPathTemplateShouldErr(t *testing.T) {
	t.Parallel()

	pm, err := pathmanager.NewPathManager("epoch_[E]/shard_[S]/[I]", "")
	assert.Nil(t, pm)
	assert.Equal(t, storage.ErrEmptyStaticPathTemplate, err)
}

func TestNewPathManager_InvalidPruningPathTemplate_NoShardPlaceholder_ShouldErr(t *testing.T) {
	t.Parallel()

	pm, err := pathmanager.NewPathManager("epoch_[E]/shard/[I]", "shard_[S]/[I]")
	assert.Nil(t, pm)
	assert.Equal(t, storage.ErrInvalidPruningPathTemplate, err)
}

func TestNewPathManager_InvalidPruningPathTemplate_NoEpochPlaceholder_ShouldErr(t *testing.T) {
	t.Parallel()

	pm, err := pathmanager.NewPathManager("epoch/shard_[S]/[I]", "shard_[S]/[I]")
	assert.Nil(t, pm)
	assert.Equal(t, storage.ErrInvalidPruningPathTemplate, err)
}

func TestNewPathManager_InvalidPathPruningTemplate_NoIdentifierPlaceholder_ShouldErr(t *testing.T) {
	t.Parallel()

	pm, err := pathmanager.NewPathManager("epoch_[E]/shard_[S]", "shard_[S]/[I]")
	assert.Nil(t, pm)
	assert.Equal(t, storage.ErrInvalidPruningPathTemplate, err)
}

func TestNewPathManager_InvalidStaticPathTemplate_NoShardPlaceholder_ShouldErr(t *testing.T) {
	t.Parallel()

	pm, err := pathmanager.NewPathManager("epoch_[E]/shard_[S]/[I]", "shard/[I]")
	assert.Nil(t, pm)
	assert.Equal(t, storage.ErrInvalidStaticPathTemplate, err)
}

func TestNewPathManager_InvalidStaticPathTemplate_NoIdentifierPlaceholder_ShouldErr(t *testing.T) {
	t.Parallel()

	pm, err := pathmanager.NewPathManager("epoch_[E]/shard_[S]/[I]", "shard_[S]")
	assert.Nil(t, pm)
	assert.Equal(t, storage.ErrInvalidStaticPathTemplate, err)
}

func TestNewPathManager_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	pm, err := pathmanager.NewPathManager("epoch_[E]/shard_[S]/[I]", "shard_[S]/[I]")
	assert.NotNil(t, pm)
	assert.Nil(t, err)
}

func TestPathManager_PathForEpoch(t *testing.T) {
	t.Parallel()

	type args struct {
		shardId    string
		epoch      uint32
		identifier string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{shardId: "0", epoch: 2, identifier: "table"},
			want: "Epoch_2/Shard_0/table",
		},
		{
			args: args{shardId: "metachain", epoch: 2654, identifier: "table23"},
			want: "Epoch_2654/Shard_metachain/table23",
		},
		{
			args: args{shardId: "0", epoch: 0, identifier: ""},
			want: "Epoch_0/Shard_0/",
		},
		{
			args: args{shardId: "53", epoch: 25839, identifier: "table1"},
			want: "Epoch_25839/Shard_53/table1",
		},
	}
	pruningPathTemplate := "Epoch_[E]/Shard_[S]/[I]"
	staticPathTemplate := "Shard_[S]/[I]"
	pm, _ := pathmanager.NewPathManager(pruningPathTemplate, staticPathTemplate)
	for _, tt := range tests {
		ttCopy := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := pm.PathForEpoch(ttCopy.args.shardId, ttCopy.args.epoch, ttCopy.args.identifier); got != ttCopy.want {
				t.Errorf("PathForEpoch() = %v, want %v", got, ttCopy.want)
			}
		})
	}
}

func TestPathManager_PathForStatic(t *testing.T) {
	t.Parallel()

	type args struct {
		shardId    string
		identifier string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{shardId: "0", identifier: "table"},
			want: "Static/Shard_0/table",
		},
		{
			args: args{shardId: "metachain", identifier: "table23"},
			want: "Static/Shard_metachain/table23",
		},
		{
			args: args{shardId: "0", identifier: ""},
			want: "Static/Shard_0/",
		},
	}
	pruningPathTemplate := "Epoch_[E]/Shard_[S]/[I]"
	staticPathTemplate := "Static/Shard_[S]/[I]"
	pm, _ := pathmanager.NewPathManager(pruningPathTemplate, staticPathTemplate)
	for _, tt := range tests {
		ttCopy := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := pm.PathForStatic(ttCopy.args.shardId, ttCopy.args.identifier); got != ttCopy.want {
				t.Errorf("PathForEpoch() = %v, want %v", got, ttCopy.want)
			}
		})
	}
}
