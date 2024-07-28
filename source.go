package rx

import (
	"errors"
	"io"
)

type SourceHandler interface {
	OnPull(Outlet)
	OnCancel(Outlet)
}

type Source[T any] struct {
	Pull   func() (T, error)
	Cancel func()
}

func (src Source[T]) OnPull(outlet Outlet) {
	if f := src.Pull; f != nil {
		t, err := f()
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
	outlet.Complete()
}

func (src Source[T]) OnCancel(outlet Outlet) {
	if f := src.Cancel; f != nil {
		f()
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
}

type SourceBuilderFunc func() SourceFactory

func (sbf SourceBuilderFunc) Create() Pipe {
	return sbf().Create()
}
