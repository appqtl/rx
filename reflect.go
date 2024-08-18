package rx

import "reflect"

var (
	errorType = reflect.TypeOf((*error)(nil)).Elem()
)

func isErrorType(gtype reflect.Type) bool {
	return gtype.Implements(errorType)
}

func checkImplements(atype reflect.Type, btype reflect.Type) bool {
	if btype.Kind() == reflect.Interface {
		return atype.Implements(btype)
	} else if atype.Kind() == reflect.Interface {
		return btype.Implements(atype)
	}
	return false
}

func compareType(atype reflect.Type, btype reflect.Type) bool {
	return atype == btype || checkImplements(atype, btype)
}

type rxReflectType uint8

const (
	RX_TYPE_UNKNOWN              = iota
	RX_TYPE_SOURCE_FUNC          // func() (T, error)
	RX_TYPE_SOURCE_FUNC_NO_ERROR // func() T
	RX_TYPE_FLOW_FUNC            // func(T) (K, error)
	RX_TYPE_FLOW_FUNC_NO_ERROR   // func(T) K
	RX_TYPE_SINK_FUNC            // func(K) error
	RX_TYPE_SINK_FUNC_NO_ERROR   // func(K)
)

func rxType(t reflect.Type) rxReflectType {
	if t.Kind() == reflect.Func {
		// source-func
		if t.NumIn() == 0 {
			if t.NumOut() == 2 && isErrorType(t.Out(1)) {
				return RX_TYPE_SOURCE_FUNC
			}
			if t.NumOut() == 1 {
				return RX_TYPE_SOURCE_FUNC_NO_ERROR
			}
			return RX_TYPE_UNKNOWN
		}
		if t.NumIn() == 1 {
			if t.NumOut() == 0 {
				return RX_TYPE_SINK_FUNC_NO_ERROR
			}
			if t.NumOut() == 1 && isErrorType(t.Out(0)) {
				return RX_TYPE_SINK_FUNC
			}
			if t.NumOut() == 1 {
				return RX_TYPE_FLOW_FUNC_NO_ERROR
			}
			if t.NumOut() == 2 && isErrorType(t.Out(1)) {
				return RX_TYPE_FLOW_FUNC
			}

		}
	}
	return RX_TYPE_UNKNOWN
}
