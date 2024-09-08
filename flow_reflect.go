package rx

import (
	"fmt"
	"io"
	"reflect"
	"sync"
)

func isFlowFunc(a any) (FlowFactory, error) {
	valA := reflect.ValueOf(a)
	typeA := valA.Type()
	switch rxType(valA.Type()) {
	case RX_TYPE_FLOW_FUNC:
		return FlowFactoryFunc(func() FlowHandler {
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
		}), nil
	case RX_TYPE_FLOW_FUNC_NO_ERROR:
		return FlowFactoryFunc(func() FlowHandler {
			typeParam := typeA.In(0)
			return FlowFunc[any, any](func(x any) (any, error) {
				typeX := reflect.TypeOf(x)
				if compareType(typeParam, typeX) {
					values := valA.Call([]reflect.Value{reflect.ValueOf(x)})
					return values[0].Interface(), nil
				}
				return nil, fmt.Errorf("flow: expect type %v got %v", typeParam, typeX)
			})
		}), nil
	}
	return nil, fmt.Errorf("%v is not a valid flow-func of type func[T,K any](T) (K,error) or func[T,K any](T) K", typeA)
}

func buildMap(a any) (FlowFactory, error) {
	valA := reflect.ValueOf(a)
	typeA := valA.Type()
	if typeA.Kind() == reflect.Func && typeA.NumIn() == 1 && typeA.NumOut() == 1 {
		return FlowFactoryFunc(func() FlowHandler {
			typeParam := typeA.In(0)
			return FlowFunc[any, any](func(x any) (any, error) {
				typeX := reflect.TypeOf(x)
				if compareType(typeParam, typeX) {
					values := valA.Call([]reflect.Value{reflect.ValueOf(x)})
					return values[0].Interface(), nil
				}
				return reflect.Zero(typeParam).Interface(), fmt.Errorf("map: expect type %v got %v", typeParam, typeX)
			})
		}), nil
	}
	return nil, fmt.Errorf("%v is not a valid map-func of type func[T,K any](T) K", typeA)
}

func buildFilter(a any) (FlowFactory, error) {
	valA := reflect.ValueOf(a)
	typeA := valA.Type()
	if typeA.Kind() == reflect.Func && typeA.NumIn() == 1 && typeA.NumOut() == 1 && typeA.Out(0).Kind() == reflect.Bool {
		return FlowFactoryFunc(func() FlowHandler {
			typeParam := typeA.In(0)
			return FlowFunc[any, any](func(x any) (any, error) {
				typeX := reflect.TypeOf(x)
				if compareType(typeParam, typeX) {
					values := valA.Call([]reflect.Value{reflect.ValueOf(x)})
					if values[0].Bool() {
						return x, nil
					}
					return reflect.Zero(typeParam).Interface(), Next
				}
				return nil, fmt.Errorf("filter: expect type %v got %v", typeParam, typeX)
			})
		}), nil
	}
	return nil, fmt.Errorf("%v is not a valid filter-func of type func[T any](T) bool", typeA)
}

func buildTakeWhile(a any) (FlowFactory, error) {
	valA := reflect.ValueOf(a)
	typeA := valA.Type()
	if typeA.Kind() == reflect.Func && typeA.NumIn() == 1 && typeA.NumOut() == 1 && typeA.Out(0).Kind() == reflect.Bool {
		return FlowFactoryFunc(func() FlowHandler {
			typeParam := typeA.In(0)
			return FlowFunc[any, any](func(x any) (any, error) {
				typeX := reflect.TypeOf(x)
				if compareType(typeParam, typeX) {
					values := valA.Call([]reflect.Value{reflect.ValueOf(x)})
					if values[0].Bool() {
						return x, nil
					}
					return reflect.Zero(typeParam).Interface(), io.EOF
				}
				return nil, fmt.Errorf("takeWhile: expect type %v got %v", typeParam, typeX)
			})
		}), nil
	}
	return nil, fmt.Errorf("%v is not a valid takeWhile-func of type func[T any](T) bool", typeA)
}

func buildDropWhile(a any) (FlowFactory, error) {
	valA := reflect.ValueOf(a)
	typeA := valA.Type()
	if typeA.Kind() == reflect.Func && typeA.NumIn() == 1 && typeA.NumOut() == 1 && typeA.Out(0).Kind() == reflect.Bool {
		return FlowFactoryFunc(func() FlowHandler {
			typeParam := typeA.In(0)
			var mutex sync.Mutex
			var dropped bool
			return FlowFunc[any, any](func(x any) (any, error) {
				mutex.Lock()
				defer mutex.Unlock()
				typeX := reflect.TypeOf(x)
				if compareType(typeParam, typeX) {
					values := valA.Call([]reflect.Value{reflect.ValueOf(x)})
					if values[0].Bool() && !dropped {
						return reflect.Zero(typeParam).Interface(), Next
					}
					dropped = true
					return x, nil
				}
				return nil, fmt.Errorf("dropWhile: expect type %v got %v", typeParam, typeX)
			})
		}), nil
	}
	return nil, fmt.Errorf("%v is not a valid dropWhile-func of type func[T any](T) bool", typeA)
}

func buildFold(v any) func(any) (FlowFactory, error) {
	return func(a any) (FlowFactory, error) {
		valA := reflect.ValueOf(a)
		typeA := valA.Type()
		typeV := reflect.TypeOf(v)
		if typeA.Kind() == reflect.Func && typeA.NumIn() == 2 && typeA.NumOut() == 1 && compareType(typeV, typeA.In(0)) && compareType(typeV, typeA.Out(0)) {
			return FlowFactoryFunc(func() FlowHandler {
				var mutex sync.Mutex
				value := v
				var complete bool
				return Flow[any]{
					HandlePush: func(x any, io IOlet){
						mutex.Lock()
						defer mutex.Unlock()
						valX := reflect.ValueOf(x)
						typeX := valX.Type()
						if !compareType(typeA.In(1), typeX) {
							io.Error(fmt.Errorf("fold: expect type %v got %v", typeA.In(1), typeX))
							return
						}
						values := []reflect.Value{reflect.ValueOf(value), valX}
						results := valA.Call(values)
						value = results[0].Interface()
						io.Pull()
					},
					HandleComplete: func(io IOlet){
						mutex.Lock()
						defer mutex.Unlock()
						io.Push(value)
						complete = true
					},
					HandlePull: func(io IOlet){
						mutex.Lock()
						defer mutex.Unlock()
						if complete {
							io.Complete()
							return
						}
						io.Pull()
					},
				}
			}), nil
		}
		return nil, fmt.Errorf("%v is not a valid fold-func of type func[T,K any](K, T) K", typeA)
	}
}

func buildReduce(a any) (FlowFactory, error) {
	typeA := reflect.TypeOf(a)
	if typeA.Kind() == reflect.Func && typeA.NumIn() == 2 && typeA.NumOut() == 1 && compareType(typeA.In(0), typeA.In(1)) && compareType(typeA.In(0), typeA.Out(0)) {
		valNew := reflect.New(typeA.In(0))
		return buildFold(valNew.Elem().Interface())(a)
	}
	return nil, fmt.Errorf("%v is not a valid reduce-func of type func[T any](T, T) T", typeA)
}
