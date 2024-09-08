package rx

type Connector[T any] interface {
	Via(any) T
}

type Builder[T any] interface {
	Connector[T]
	Map(any) T
	Filter(any) T
	Take(uint64) T
	TakeWhile(any) T
	Drop(uint64) T
	DropWhile(any) T
	Fold(any, any) T
	Reduce(any) T
}

type builder[T any] struct {
	Connector[T]
}

func (b builder[T]) create(a any, f func(any) (FlowFactory, error)) T {
	ff, err := f(a)
	if err != nil {
		panic(err)
	}
	return b.Via(ff)
}

func (b builder[T]) Map(a any) T {
	return b.create(a, buildMap)
}

func (b builder[T]) Filter(a any) T {
	return b.create(a, buildFilter)
}

func (b builder[T]) Take(n uint64) T {
	return b.Via(Take(n))
}

func (b builder[T]) TakeWhile(a any) T {
	return b.create(a, buildTakeWhile)
}

func (b builder[T]) Drop(n uint64) T {
	return b.Via(Drop(n))
}

func (b builder[T]) DropWhile(a any) T {
	return b.create(a, buildDropWhile)
}

func (b builder[T]) Fold(v, a any) T {
	return b.create(a, buildFold(v))
}

func (b builder[T]) Reduce(a any) T {
	return b.create(a, buildReduce)
}