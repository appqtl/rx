package rx

import (
	"fmt"
	"reflect"
)

func isSinkFunc(a any) (SinkFactory, error) {
	valA := reflect.ValueOf(a)
	typeA := valA.Type()
	switch rxType(valA.Type()) {
	case RX_TYPE_SINK_FUNC:
		return SinkFactoryFunc(func() SinkHandler {
			typeParam := typeA.In(0)
			return SinkFunc[any](func(x any) error {
				typeX := reflect.TypeOf(x)
				if compareType(typeParam, typeX) {
					values := valA.Call([]reflect.Value{reflect.ValueOf(x)})
					if !values[0].IsNil() {
						return values[0].Interface().(error)
					}
					return nil
				}
				return fmt.Errorf("flow: expect type %v got %v", typeParam, typeX)
			})
		}), nil
	case RX_TYPE_SINK_FUNC_NO_ERROR:
		return SinkFactoryFunc(func() SinkHandler {
			typeParam := typeA.In(0)
			return SinkFunc[any](func(x any) error {
				typeX := reflect.TypeOf(x)
				if compareType(typeParam, typeX) {
					valA.Call([]reflect.Value{reflect.ValueOf(x)})
					return nil
				}
				return fmt.Errorf("flow: expect type %v got %v", typeParam, typeX)
			})
		}), nil
	}
	return nil, fmt.Errorf("%v is not a valid sink-func of type func[T any](T) error or func[T any](T)", typeA)
}
