package trie

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceHashesAtPosition(t *testing.T) {
	t.Parallel()

	newData := []string{"aaa", "bbb"}
	initialContainingOne := []string{"one"}
	initialContainingTwo := []string{"one", "two"}
	initialContainingThree := []string{"one", "two", "three"}
	empty := make([]string, 0)

	t.Run("empty initial", func(t *testing.T) {
		result := replaceHashesAtPosition(0, make([]string, 0), make([]string, 0))
		assert.Empty(t, result)
	})
	t.Run("empty initial with index out of bound", func(t *testing.T) {
		result := replaceHashesAtPosition(1, make([]string, 0), make([]string, 0))
		assert.Empty(t, result)
	})
	t.Run("empty initial provided data", func(t *testing.T) {
		result := replaceHashesAtPosition(0, make([]string, 0), newData)
		assert.Empty(t, result)
	})

	t.Run("1 existing and replace on first position", func(t *testing.T) {
		result := replaceHashesAtPosition(0, initialContainingOne, newData)
		expected := []string{"aaa", "bbb"}
		assert.Equal(t, expected, result)
	})
	t.Run("1 existing and replace on an out of bound position", func(t *testing.T) {
		result := replaceHashesAtPosition(1, initialContainingOne, newData)
		expected := []string{"one"}
		assert.Equal(t, expected, result)
	})
	t.Run("1 existing and replace with empty", func(t *testing.T) {
		result := replaceHashesAtPosition(0, initialContainingOne, empty)
		assert.Empty(t, result)
	})
	t.Run("2 existing and replace on first position", func(t *testing.T) {
		result := replaceHashesAtPosition(0, initialContainingTwo, newData)
		expected := []string{"aaa", "bbb", "two"}
		assert.Equal(t, expected, result)
	})
	t.Run("2 existing and replace on second position", func(t *testing.T) {
		result := replaceHashesAtPosition(1, initialContainingTwo, newData)
		expected := []string{"one", "aaa", "bbb"}
		assert.Equal(t, expected, result)
	})
	t.Run("2 existing and replace on an out of bound position", func(t *testing.T) {
		result := replaceHashesAtPosition(2, initialContainingTwo, newData)
		expected := []string{"one", "two"}
		assert.Equal(t, expected, result)
	})
	t.Run("2 existing and replace with empty", func(t *testing.T) {
		result := replaceHashesAtPosition(0, initialContainingTwo, empty)
		expected := []string{"two"}
		assert.Equal(t, expected, result)
	})
	t.Run("3 existing and replace on first position", func(t *testing.T) {
		result := replaceHashesAtPosition(0, initialContainingThree, newData)
		expected := []string{"aaa", "bbb", "two", "three"}
		assert.Equal(t, expected, result)
	})
	t.Run("3 existing and replace on second position", func(t *testing.T) {
		result := replaceHashesAtPosition(1, initialContainingThree, newData)
		expected := []string{"one", "aaa", "bbb", "three"}
		assert.Equal(t, expected, result)
	})
	t.Run("3 existing and replace on third position", func(t *testing.T) {
		result := replaceHashesAtPosition(2, initialContainingThree, newData)
		expected := []string{"one", "two", "aaa", "bbb"}
		assert.Equal(t, expected, result)
	})
	t.Run("3 existing and replace on an out of bound position", func(t *testing.T) {
		result := replaceHashesAtPosition(3, initialContainingThree, newData)
		expected := []string{"one", "two", "three"}
		assert.Equal(t, expected, result)
	})
	t.Run("3 existing and replace with empty", func(t *testing.T) {
		result := replaceHashesAtPosition(0, initialContainingThree, empty)
		expected := []string{"two", "three"}
		assert.Equal(t, expected, result)
	})
	t.Run("3 existing and replace with empty on second position", func(t *testing.T) {
		result := replaceHashesAtPosition(1, initialContainingThree, empty)
		expected := []string{"one", "three"}
		assert.Equal(t, expected, result)
	})
	t.Run("3 existing and replace with empty on third position", func(t *testing.T) {
		result := replaceHashesAtPosition(2, initialContainingThree, empty)
		expected := []string{"one", "two"}
		assert.Equal(t, expected, result)
	})
}
