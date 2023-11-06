package tipe

import (
	"fmt"
	"reflect"
	"unsafe"
)

type kind struct{ int }

var (
	kPtrLike = unsafe.Pointer(&kind{0})
	kSimple  = unsafe.Pointer(&kind{1})
	kString  = unsafe.Pointer(&kind{2})
	kAlloc   = unsafe.Pointer(&kind{3})
	kIface   = unsafe.Pointer(&kind{4})
)

type iface struct {
	typ  uintptr
	data unsafe.Pointer
}

type Result[T any] struct {
	data   unsafe.Pointer
	field1 uintptr
	field2 uintptr
}

const sliceFlag = ^(^uintptr(0) >> 1)

func (r Result[T]) IsOk() bool {
	if r.data == kSimple || r.field2&sliceFlag != 0 {
		return true
	}
	switch r.field1 {
	case uintptr(kPtrLike), uintptr(kString), uintptr(kAlloc), uintptr(kIface):
		return true
	}
	return false
}

func (r Result[T]) IsErr() bool {
	return !r.IsOk()
}

func (r Result[T]) Unwrap() T {
	if r.data == kSimple {
		return *(*T)(unsafe.Pointer(&r.field1))
	}
	if r.field2&sliceFlag != 0 {
		cap := r.field2 & ^sliceFlag
		len := r.field1
		return *(*T)(unsafe.Pointer(&struct {
			ptr unsafe.Pointer
			len int
			cap int
		}{r.data, int(len), int(cap)}))
	}
	switch r.field1 {
	case uintptr(kPtrLike):
		return *(*T)(unsafe.Pointer(&r.data))

	case uintptr(kIface):
		return *(*T)(unsafe.Pointer(&iface{
			typ:  r.field2,
			data: r.data,
		}))

	case uintptr(kString):
		len := r.field2
		str := unsafe.String((*byte)(r.data), len)
		return *(*T)(unsafe.Pointer(&str))

	case uintptr(kAlloc):
		return *(*T)(r.data)
	}
	panic("tipe: Result[T] is not ok")
}

func (r Result[T]) UnwrapErr() error {
	if !r.IsErr() {
		panic("tipe: Result[T] is not err")
	}
	return *(*error)(unsafe.Pointer(&iface{
		typ:  r.field1,
		data: r.data,
	}))
}

func (r Result[T]) UnwrapOr(or T) T {
	if r.IsOk() {
		return r.Unwrap()
	}
	return or
}

func (r Result[T]) Tuple() (T, error) {
	if r.IsOk() {
		return r.Unwrap(), nil
	}
	var zero T
	return zero, r.UnwrapErr()
}

func (r Result[T]) TupleBool() (T, bool) {
	if r.IsOk() {
		return r.Unwrap(), true
	}
	var zero T
	return zero, false
}

func (r Result[T]) String() string {
	if r.IsOk() {
		return fmt.Sprintf("Ok(%v)", r.Unwrap())
	} else {
		return fmt.Sprintf("Err(%v)", r.UnwrapErr())
	}
}

func Ok[T any](value T) Result[T] {
	var zero T
	var anyZero = any(zero)
	var kind reflect.Kind
	if (*iface)(unsafe.Pointer(&anyZero)).typ == 0 {
		kind = reflect.Interface
	} else {
		kind = reflect.TypeOf(zero).Kind()
	}
	switch kind {
	case reflect.Ptr, reflect.UnsafePointer, reflect.Chan, reflect.Func, reflect.Map:
		return Result[T]{
			data:   *(*unsafe.Pointer)(unsafe.Pointer(&value)),
			field1: uintptr(kPtrLike),
		}
	case reflect.Interface:
		iface := *(*iface)(unsafe.Pointer(&value))
		return Result[T]{
			data:   iface.data,
			field1: uintptr(kIface),
			field2: iface.typ,
		}
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		r := Result[T]{
			data: kSimple,
		}
		*(*T)(unsafe.Pointer(&r.field1)) = value
		return r
	case reflect.String:
		str := *(*string)(unsafe.Pointer(&value))
		return Result[T]{
			data:   unsafe.Pointer(unsafe.StringData(str)),
			field1: uintptr(kString),
			field2: uintptr(len(str)),
		}
	case reflect.Slice:
		slice := *(*[]byte)(unsafe.Pointer(&value))
		return Result[T]{
			data:   unsafe.Pointer(unsafe.SliceData(slice)),
			field1: uintptr(len(slice)),
			field2: uintptr(cap(slice)) | sliceFlag,
		}
	default:
		return Result[T]{
			data:   unsafe.Pointer(&value),
			field1: uintptr(kAlloc),
		}
	}
}

func Err[T any](err error) Result[T] {
	if err == nil {
		var zero T
		return Ok[T](zero)
	}
	var internal = *(*iface)(unsafe.Pointer(&err))
	return Result[T]{
		data:   internal.data,
		field1: internal.typ,
	}
}

func MakeResult[T any](value T, err error) Result[T] {
	if err == nil {
		return Ok[T](value)
	}
	return Err[T](err)
}
