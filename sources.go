package rx

import (
	"io"
	"sync"
	"sync/atomic"
)

func NewSource(f func() SourceHandler) SourceBuilder {
	return SourceBuilderFunc(func() SourceFactory {
		return SourceFactoryFunc(f)
	})
}

func Slice[T any](slice ...T) SourceBuilder {
	return NewSource(func() SourceHandler {
		var index int64 = -1
		return SourceFunc[T](func() (T, error) {
			idx := int(atomic.AddInt64(&index, 1))
			if idx < len(slice) {
				return slice[idx], nil
			}
			var t0 T
			return t0, io.EOF
		})
	})
}

type Number interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

func Sequence[T Number](startWith T) SourceBuilder {
	return NewSource(func() SourceHandler {
		var mutex sync.Mutex
		seq := startWith
		return SourceFunc[T](func() (T, error) {
			mutex.Lock()
			defer mutex.Unlock()
			defer func() { seq++ }()
			return seq, nil
		})
	})
}

func ChanSource[T any](channel <-chan T) SourceBuilder {
	return NewSource(func() SourceHandler {
		return SourceFunc[T](func() (T, error) {
			t, open := <-channel
			if !open {
				var t0 T
				return t0, io.EOF
			}
			return t, nil
		})
	})
}
