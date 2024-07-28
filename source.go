package rx

import (
	"errors"
	"io"
)

type SourceHandler interface {
	OnPull(Outlet)
	OnCancel(Outlet)
}

type Source struct {
	HandlePull   func(outlet Outlet)
	HandleCancel func(outlet Outlet)
}

func (src Source) OnPull(outlet Outlet) {
	if f := src.HandlePull; f != nil {
		f(outlet)
		return
	}
	outlet.Complete()
}

func (src Source) OnCancel(outlet Outlet) {
	if f := src.HandleCancel; f != nil {
		f(outlet)
		return
	}
	outlet.Complete()
}

type SourceFunc[T any] func() (T, error)

func (sf SourceFunc[T]) OnPull(outlet Outlet) {
	t, err := sf()
	if err != nil {
		if errors.Is(err, io.EOF) {
			outlet.Complete()
			return
		}
		outlet.Error(err)
		return
	}
	outlet.Push(t)
}

func (sf SourceFunc[T]) OnCancel(outlet Outlet) {
	outlet.Complete()
}

type SourceFactory interface {
	Create() Pipe
}

type SourceFactoryFunc func() SourceHandler

func (sff SourceFactoryFunc) Create() Pipe {
	result := newPipe()

	go func(handler SourceHandler, in <-chan Command, outlet Outlet) {
		for cmd := range in {
			switch cmd {
			case PULL:
				handler.OnPull(outlet)
			case CANCEL:
				handler.OnCancel(outlet)
			}
		}
	}(sff(), result.commands, result)

	return result
}

type SourceBuilder interface {
	SourceFactory
	To(SinkFactory)
	RunWith(SinkFactory) Runnable
}

type SourceBuilderFunc func() SourceFactory

func (sbf SourceBuilderFunc) Create() Pipe {
	return sbf().Create()
}

func (sbf SourceBuilderFunc) To(builder SinkFactory) {
	sbf.RunWith(builder).Run()
}

func (sbf SourceBuilderFunc) RunWith(builder SinkFactory) Runnable {
	return builder.Create(sbf.Create())
}
