package rx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceSource(t *testing.T) {
	slice := []string{"A", "B", "C"}

	pipe := Slice(slice...).Create()
	for i := 0; i <= len(slice); i++ {
		pipe.Pull()
		evt := <-pipe.Events()
		if i == len(slice) {
			assert.True(t, evt.Complete)
			continue
		}
		assert.Equal(t, slice[i], evt.Data)
	}
}

func TestSequenceSource(t *testing.T) {
	pipe := Sequence(uint8(0)).Create()
	for i := uint8(0); i <= 10; i++ {
		pipe.Pull()
		evt := <-pipe.Events()
		assert.Equal(t, i, evt.Data)
	}
}