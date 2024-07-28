package rx

import "context"

type Emittable interface {
	Emit(any)
	Close()
}

type EmittableInlet interface {
	Inlet
	Emittable
}

type Emitter interface {
	EmittableInlet
	Runnable
}

type emitter struct {
	emits chan any
	inlet Inlet
}

func newEmitter(inlet Inlet) Emitter {
	return &emitter{
		emits: make(chan any, 1),
		inlet: inlet,
	}
}

func (e *emitter) Pull() {
	e.inlet.Pull()
}

func (e *emitter) Cancel() {
	e.inlet.Cancel()
}

func (e *emitter) Emit(v any) {
	e.emits <- v
}

func (e *emitter) Close() {
	close(e.emits)
}

func (e *emitter) Await() error {
	return e.AwaitWithContext(context.Background())
}

func (e *emitter) AwaitWithContext(ctx context.Context) error {
	_, err := e.ExecuteWithContext(ctx)
	return err
}

func (e *emitter) Run() <-chan any {
	return e.RunWithContext(context.Background())
}

func (e *emitter) RunWithContext(ctx context.Context) <-chan any {
	e.inlet.Pull()
	return runWithContext(ctx, func() (<-chan any, context.CancelFunc) {
		return e.emits, e.Cancel
	})
}

func (e *emitter) Execute() (any, error) {
	return e.ExecuteWithContext(context.Background())
}

func (e *emitter) ExecuteWithContext(ctx context.Context) (any, error) {
	return execute(e.RunWithContext(ctx))
}
