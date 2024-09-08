package rx

import (
	"io"
	"sync"
	"sync/atomic"
)

func NewFlow[T, K any](f func(T) (K, error)) FlowBuilder {
	return FlowBuilderFunc(func() FlowFactory {
		return FlowFactoryFunc(func() FlowHandler {
			return FlowFunc[T, K](f)
		})
	})
}

func FlowOf(a any) FlowBuilder {
	switch ft := a.(type) {
	case FlowBuilder:
		return ft
	case FlowFactory:
		return FlowBuilderFunc(func() FlowFactory {
			return ft
		})
	case func() FlowFactory:
		return FlowBuilderFunc(ft)
	default:
		factory, err := isFlowFunc(a)
		if err != nil {
			panic(err)
		}
		return FlowBuilderFunc(func() FlowFactory {
			return factory
		})
	}
}

func Map[T, K any](f func(T) K) FlowBuilder {
	return NewFlow(func(t T) (K, error) {
		return f(t), nil
	})
}

func Filter[T any](f func(T) bool) FlowBuilder {
	return NewFlow(func(t T) (T, error) {
		if f(t) {
			return t, nil
		}
		var t0 T
		return t0, Next
	})
}

func Take(n uint64) FlowBuilder {
	return FlowBuilderFunc(func() FlowFactory {
		return FlowFactoryFunc(func() FlowHandler {
			var count uint64
			return FlowFunc[any, any](func(v any) (any, error) {
				defer atomic.AddUint64(&count, 1)
				if atomic.LoadUint64(&count) < n {
					return v, nil
				}
				return v, io.EOF
			})
		})
	})
}

func TakeWhile[T any](f func(T) bool) FlowBuilder {
	return FlowBuilderFunc(func() FlowFactory {
		return FlowFactoryFunc(func() FlowHandler {
			return FlowFunc[T, T](func(t T) (T, error) {
				if f(t) {
					return t, nil
				}
				var t0 T
				return t0, io.EOF
			})
		})
	})
}

func Drop(n uint64) FlowBuilder {
	return FlowBuilderFunc(func() FlowFactory {
		return FlowFactoryFunc(func() FlowHandler {
			var count uint64
			return FlowFunc[any, any](func(v any) (any, error) {
				defer atomic.AddUint64(&count, 1)
				if atomic.LoadUint64(&count) < n {
					return v, Next
				}
				return v, nil
			})
		})
	})
}

func DropWhile[T any](f func(T) bool) FlowBuilder {
	return FlowBuilderFunc(func() FlowFactory {
		return FlowFactoryFunc(func() FlowHandler {
			var mutex sync.Mutex
			var dropped bool
			return FlowFunc[T, T](func(t T) (T, error) {
				mutex.Lock()
				defer mutex.Unlock()
				if f(t) && !dropped {
					return t, Next
				}
				dropped = true
				return t, nil
			})
		})
	})
}

func Fold[T, K any](k K, f func(K, T) K) FlowBuilder {
	return FlowBuilderFunc(func() FlowFactory {
		return FlowFactoryFunc(func() FlowHandler {
			var m sync.Mutex
			x := k
			var complete bool
			return Flow[T]{
				HandlePush: func(t T, io IOlet) {
					m.Lock()
					defer m.Unlock()
					x = f(x, t)
					io.Pull()
				},
				HandleComplete: func(io IOlet) {
					m.Lock()
					defer m.Unlock()
					io.Push(x)
					complete = true
				},
				HandlePull: func(io IOlet) {
					m.Lock()
					defer m.Unlock()
					if complete {
						io.Complete()
						return
					}
					io.Pull()
				},
			}
		})
	})
}

func Reduce[T any](f func(T, T) T) FlowBuilder {
	var t0 T
	return Fold(t0, f)
}

func Sum[T Number]() FlowBuilder {
	return Reduce(func(a T, b T) T { return a + b })
}
