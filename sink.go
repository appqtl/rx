package rx

import (
	"errors"
	"fmt"
	"io"
)

var (
	//lint:ignore ST1012 ingnore this warning
	Next = errors.New("next")
)

type SinkHandler interface {
	OnPush(any, EmittableInlet)
	OnError(error, EmittableInlet)
	OnComplete(EmittableInlet)
}

type Sink[T any] struct {
	HandlePush     func(t T, inlet EmittableInlet)
	HandleError    func(err error, inlet EmittableInlet)
	HandleComplete func(inlet EmittableInlet)
}

func (sink Sink[T]) OnPush(v any, inlet EmittableInlet) {
	if f := sink.HandlePush; f != nil {
		if t, ok := v.(T); ok {
			f(t, inlet)
			return
		}
		var t0 T
		sink.OnError(fmt.Errorf("sink: expect type %T got %T", t0, v), inlet)
		return
	}
	inlet.Pull()
}

func (sink Sink[T]) OnError(err error, inlet EmittableInlet) {
	if f := sink.HandleError; f != nil {
		f(err, inlet)
		return
	}
	inlet.Emit(err)
	inlet.Cancel()
}

func (sink Sink[T]) OnComplete(inlet EmittableInlet) {
	if f := sink.HandleComplete; f != nil {
		f(inlet)
		return
	}
	inlet.Close()
}

type SinkFunc[T any] func(t T) error

func (sf SinkFunc[T]) OnPush(v any, inlet EmittableInlet) {
	if t, ok := v.(T); ok {
		if err := sf(t); err != nil {
			if errors.Is(err, io.EOF) {
				inlet.Cancel()
				return
			}
			if errors.Is(err, Next) {
				inlet.Pull()
				return
			}
			sf.OnError(err, inlet)
			return
		}
		inlet.Pull()
		return
	}
	var t0 T
	sf.OnError(fmt.Errorf("sink: expect type %T got %T", t0, v), inlet)
}

func (sf SinkFunc[T]) OnError(err error, inlet EmittableInlet) {
	inlet.Emit(err)
	inlet.Cancel()
}

func (sf SinkFunc[T]) OnComplete(inlet EmittableInlet) {
	inlet.Close()
}

type SinkFactory interface {
	Create(Pipe) Runnable
}

type SinkFactoryFunc func() SinkHandler

func (sff SinkFactoryFunc) Create(pipe Pipe) Runnable {
	emitter := newEmitter(pipe)

	go func(handler SinkHandler, pipe Pipe, inlet EmittableInlet) {
		defer pipe.Close()
		for evt := range pipe.Events() {
			switch evt.Type() {
			case PUSH:
				handler.OnPush(evt.Data, inlet)
			case ERROR:
				handler.OnError(evt.Error, inlet)
			case COMPLETE:
				handler.OnComplete(inlet)
			}
		}
	}(sff(), pipe, emitter)

	return emitter
}
