package rx

import "context"

type Runnable interface {
	Await() error
	AwaitWithContext(context.Context) error
	Run() <-chan any
	RunWithContext(context.Context) <-chan any
	Execute() (any, error)
	ExecuteWithContext(context.Context) (any, error)
}

type StartFunc func() (<-chan any, context.CancelFunc)

func runWithContext(ctx context.Context, start StartFunc) <-chan any {
	out := make(chan any, 1)
	emits, cancel := start()

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				defer cancel()
				out <- ctx.Err()
				return
			case e, open := <-emits:
				if !open {
					return
				}
				out <- e
			}
		}
	}()

	return out
}

func execute(emits <-chan any) (any, error) {
	results := make(chan any, 1)
	go func() {
		defer close(results)
		result := make([]interface{}, 0)
		for e := range emits {
			if err, ok := e.(error); ok {
				results <- err
				return
			}
			result = append(result, e)
		}
		results <- result
	}()

	switch res := (<-results).(type) {
	case error:
		return nil, res
	case []interface{}:
		if len(res) == 0 {
			return struct{}{}, nil
		}
		if len(res) == 1 {
			return res[0], nil
		}
		return res, nil
	default:
		return res, nil
	}
}
