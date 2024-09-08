package rx

import (
	"fmt"
	"sync"
)

func NewSink(f func() SinkHandler) SinkFactory {
	return SinkFactoryFunc(f)
}

func SinkOf(a any) SinkFactory {
	switch st := a.(type) {
	case SinkFactory:
		return st
	case func() SinkHandler:
		return NewSink(st)
	default:
		sf, err := isSinkFunc(a)
		if err != nil {
			panic(err)
		}
		return sf
	}
}

func Empty() SinkFactory {
	return NewSink(func() SinkHandler {
		return Sink[any]{}
	})
}

func ForEach[T any](f func(T)) SinkFactory {
	return NewSink(func() SinkHandler {
		return SinkFunc[T](func(t T) error {
			f(t)
			return nil
		})
	})
}

func Println() SinkFactory {
	return ForEach(func(a any) { fmt.Println(a) })
}

func Printf(format string) SinkFactory {
	return ForEach(func(a any) { fmt.Printf(format, a) })
}

func Collect[T any]() SinkFactory {
	return NewSink(func() SinkHandler {
		var mutex sync.Mutex
		slice := make([]T, 0)
		return Sink[T]{
			HandlePush: func(t T, inlet EmittableInlet) {
				mutex.Lock()
				defer mutex.Unlock()
				slice = append(slice, t)
				inlet.Pull()
			},
			HandleComplete: func(inlet EmittableInlet) {
				mutex.Lock()
				defer mutex.Unlock()
				inlet.Emit(slice)
			},
		}
	})
}

func Replicator() SinkFactory {
	return NewSink(func() SinkHandler {
		return Sink[any]{
			HandlePush: func(t any, inlet EmittableInlet) {
				inlet.Emit(t)
				inlet.Pull()
			},
		}
	})
}
