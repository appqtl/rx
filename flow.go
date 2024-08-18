package rx

import (
	"errors"
	"fmt"
	"io"
)

type FlowHandler interface {
	OnPull(IOlet)
	OnCancel(IOlet)
	OnPush(any, IOlet)
	OnError(error, IOlet)
	OnComplete(IOlet)
}

type Flow[T any] struct {
	HandlePull     func(IOlet)
	HandleCancel   func(IOlet)
	HandlePush     func(T, IOlet)
	HandleError    func(error, IOlet)
	HandleComplete func(IOlet)
}

func (f Flow[T]) OnPull(io IOlet) {
	if f.HandlePull != nil {
		f.HandlePull(io)
		return
	}
	io.Pull()
}

func (f Flow[T]) OnCancel(io IOlet) {
	if f.HandleCancel != nil {
		f.HandleCancel(io)
		return
	}
	io.Cancel()
}

func (f Flow[T]) OnPush(v any, io IOlet) {
	if f.HandlePush != nil {
		if t, ok := v.(T); ok {
			f.HandlePush(t, io)
			return
		}
		var t0 T
		f.OnError(fmt.Errorf("flow: expect type %T got %T", t0, v), io)
		return
	}
	io.Push(v)
}

func (f Flow[T]) OnError(err error, io IOlet) {
	if f.HandleError != nil {
		f.HandleError(err, io)
		return
	}
	io.Error(err)
}

func (f Flow[T]) OnComplete(io IOlet) {
	if f.HandleComplete != nil {
		f.HandleComplete(io)
		return
	}
	io.Complete()
}

type FlowFunc[T, K any] func(t T) (K, error)

func (ff FlowFunc[T, K]) OnPull(io IOlet) {
	io.Pull()
}

func (ff FlowFunc[T, K]) OnCancel(io IOlet) {
	io.Cancel()
}

func (ff FlowFunc[T, K]) OnPush(v any, iol IOlet) {
	if t, ok := v.(T); ok {
		k, err := ff(t)
		if err != nil {
			if errors.Is(err, io.EOF) {
				iol.Cancel()
				return
			}
			if errors.Is(err, Next) {
				iol.Pull()
				return
			}
			ff.OnError(err, iol)
			return
		}
		iol.Push(k)
		return
	}
	var t0 T
	ff.OnError(fmt.Errorf("flow: expect type %T got %T", t0, v), iol)
}

func (ff FlowFunc[T, K]) OnError(err error, io IOlet) {
	io.Error(err)
}

func (ff FlowFunc[T, K]) OnComplete(io IOlet) {
	io.Complete()
}

type FlowFactory interface {
	Create(Pipe) Pipe
}

type inlineFlowFactory func(Pipe) Pipe

func (iff inlineFlowFactory) Create(pipe Pipe) Pipe {
	return iff(pipe)
}

type FlowFactoryFunc func() FlowHandler

func (fff FlowFactoryFunc) Create(pipe Pipe) Pipe {
	result := newPipe()

	go func(handler FlowHandler, commands <-chan Command, pipe Pipe, iolet IOlet) {
		defer pipe.Close()
		var commandsClosed, eventsClosed bool
		for {
			select {
			case cmd, open := <-commands:
				if !open {
					commandsClosed = true
					if eventsClosed {
						return
					}
					continue
				}
				switch cmd {
				case PULL:
					handler.OnPull(iolet)
				case CANCEL:
					handler.OnCancel(iolet)
				}
			case evt, open := <-pipe.Events():
				if !open {
					eventsClosed = true
					if commandsClosed {
						return
					}
					continue
				}
				switch evt.Type() {
				case PUSH:
					handler.OnPush(evt.Data, iolet)
				case ERROR:
					handler.OnError(evt.Error, iolet)
				case COMPLETE:
					handler.OnComplete(iolet)
				}
			}
		}
	}(fff(), result.commands, pipe, iolet{Inlet: pipe, Outlet: result})

	return result
}

type FlowBuilder interface {
	FlowFactory
	Via(any) FlowBuilder
	To(any) SinkFactory
}

type FlowBuilderFunc func() FlowFactory

func (fbf FlowBuilderFunc) Create(pipe Pipe) Pipe {
	return fbf().Create(pipe)
}

func (fbf FlowBuilderFunc) Via(a any) FlowBuilder {
	return FlowBuilderFunc(func() FlowFactory {
		return inlineFlowFactory(func(pipe Pipe) Pipe {
			return FlowOf(a).Create(fbf().Create(pipe))
		})
	})
}

func (fbf FlowBuilderFunc) To(a any) SinkFactory {
	return inlineSinkFactory(func(pipe Pipe) Runnable {
		return SinkOf(a).Create(fbf().Create(pipe))
	})
}
