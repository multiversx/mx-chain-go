package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemDB(t *testing.T) {
	t.Parallel()

	instance := NewMemDB()
	assert.NotNil(t, instance)
}

func TestNewlruDB(t *testing.T) {
	t.Parallel()

	t.Run("invalid argument should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewlruDB(0)
		assert.Nil(t, instance)
		assert.NotNil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		instance, err := NewlruDB(1)
		assert.NotNil(t, instance)
		assert.Nil(t, err)
	})
}

func TestNewLevelDB(t *testing.T) {
	t.Parallel()

	t.Run("invalid argument should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewLevelDB(t.TempDir(), 0, 0, 0)
		assert.Nil(t, instance)
		assert.NotNil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		instance, err := NewLevelDB(t.TempDir(), 1, 1, 1)
		assert.NotNil(t, instance)
		assert.Nil(t, err)
		_ = instance.Close()
	})
}

func TestNewSerialDB(t *testing.T) {
	t.Parallel()

	t.Run("invalid argument should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewSerialDB(t.TempDir(), 0, 0, 0)
		assert.Nil(t, instance)
		assert.NotNil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		instance, err := NewSerialDB(t.TempDir(), 1, 1, 1)
		assert.NotNil(t, instance)
		assert.Nil(t, err)
		_ = instance.Close()
	})
}
