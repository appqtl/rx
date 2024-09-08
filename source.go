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

type inlineSourceFactory func() Pipe

func (isf inlineSourceFactory) Create() Pipe {
	return isf()
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
	Builder[SourceBuilder]
	// Via(any) SourceBuilder
	To(any)
	RunWith(any) Runnable
}

type SourceConnector func() SourceFactory

func (sc SourceConnector) Create() Pipe {
	return sc().Create()
}

func (sc SourceConnector) Via(a any) SourceBuilder {
	return SourceBuilderFunc(func() SourceFactory {
		return inlineSourceFactory(func() Pipe {
			return FlowOf(a).Create(sc.Create())
		})
	})
}

type sourceBuilder struct {
	SourceFactory
	Builder[SourceBuilder]
}

func SourceBuilderFunc(f func() SourceFactory) SourceBuilder {
	sourceConnector := SourceConnector(f)
	return sourceBuilder{
		sourceConnector,
		builder[SourceBuilder]{sourceConnector},
	}
}

func (sb sourceBuilder) To(a any) {
	sb.RunWith(a).Run()
}

func (sb sourceBuilder) RunWith(a any) Runnable {
	return SinkOf(a).Create(sb.Create())
}
