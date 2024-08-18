package rx

import (
	"fmt"
	"reflect"
)

func isSinkFunc(a any) (SinkFactory, bool) {
	valA := reflect.ValueOf(a)
	switch rxType(valA.Type()) {
	case RX_TYPE_SINK_FUNC:
		return SinkFactoryFunc(func() SinkHandler {
			typeA := valA.Type()
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
		}), true
	case RX_TYPE_SINK_FUNC_NO_ERROR:
		return SinkFactoryFunc(func() SinkHandler {
			typeA := valA.Type()
			typeParam := typeA.In(0)
			return SinkFunc[any](func(x any) error {
				typeX := reflect.TypeOf(x)
				if compareType(typeParam, typeX) {
					valA.Call([]reflect.Value{reflect.ValueOf(x)})
					return nil
				}
				return fmt.Errorf("flow: expect type %v got %v", typeParam, typeX)
			})
		}), true
	}
	return nil, false
}
