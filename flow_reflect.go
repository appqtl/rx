package rx

import (
	"fmt"
	"reflect"
)

func isFlowFunc(a any) (FlowFactory, bool) {
	valA := reflect.ValueOf(a)
	switch rxType(valA.Type()) {
	case RX_TYPE_FLOW_FUNC:
		return FlowFactoryFunc(func() FlowHandler {
			typeA := valA.Type()
			typeParam := typeA.In(0)
			return FlowFunc[any, any](func(x any) (any, error) {
				typeX := reflect.TypeOf(x)
				if compareType(typeParam, typeX) {
					values := valA.Call([]reflect.Value{reflect.ValueOf(x)})
					if !values[1].IsNil() {
						return nil, values[1].Interface().(error)
					}
					return values[0].Interface(), nil
				}
				return nil, fmt.Errorf("flow: expect type %v got %v", typeParam, typeX)
			})
		}), true
	case RX_TYPE_FLOW_FUNC_NO_ERROR:
		return FlowFactoryFunc(func() FlowHandler {
			typeA := valA.Type()
			typeParam := typeA.In(0)
			return FlowFunc[any, any](func(x any) (any, error) {
				typeX := reflect.TypeOf(x)
				if compareType(typeParam, typeX) {
					values := valA.Call([]reflect.Value{reflect.ValueOf(x)})
					return values[0].Interface(), nil
				}
				return nil, fmt.Errorf("flow: expect type %v got %v", typeParam, typeX)
			})
		}), true
	}
	return nil, false
}
