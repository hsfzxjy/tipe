package tipe

import (
	"fmt"
	"reflect"
	"unsafe"
)

type kind struct{ int }

var (
	kPtrLike  = unsafe.Pointer(&kind{0})
	kSimple   = unsafe.Pointer(&kind{1})
	kString   = unsafe.Pointer(&kind{2})
	kAlloc    = unsafe.Pointer(&kind{3})
	kIface    = unsafe.Pointer(&kind{4})
	kDetached = unsafe.Pointer(&kind{5})
)

type iface struct {
	typ  uintptr
	data unsafe.Pointer
}

// Result[T] is a type that can be used to return a value or an error.
// It is similar to Rust's Result<T, E> type.
// Result[T] occupies a fixed size of 3 words.
// For most common types, Result[T] needs no heap allocation.
type Result[T any] struct {
	// This phantom field is used to make sure that Result[T] is not comparable,
	// and also to make sure that Result[T] is not assignable to any other type.
	_      [0][]T
	data   unsafe.Pointer
	field1 uintptr
	field2 uintptr
}

const sliceFlag = ^(^uintptr(0) >> 1)
const x = unsafe.Sizeof(Result[int]{})

// IsOk returns true if the Result[T] contains a value.
func (r Result[T]) IsOk() bool {
	if r.data == kSimple || r.field1&sliceFlag != 0 {
		return true
	}
	switch r.field2 {
	case uintptr(kPtrLike), uintptr(kString), uintptr(kAlloc), uintptr(kIface), uintptr(kDetached):
		return true
	}
	if r.data == nil && r.field1 == 0 && r.field2 == 0 {
		return true
	}
	return false
}

// IsErr returns true if the Result[T] contains an error.
func (r Result[T]) IsErr() bool {
	return !r.IsOk()
}

// Unwrap returns the value contained in the Result[T].
func (r Result[T]) Unwrap() T {
	if r.data == kSimple {
		return *(*T)(unsafe.Pointer(&r.field1))
	}
	if r.field1&sliceFlag != 0 {
		len := r.field1 & ^sliceFlag
		cap := r.field2
		return *(*T)(unsafe.Pointer(&struct {
			ptr unsafe.Pointer
			len int
			cap int
		}{r.data, int(len), int(cap)}))
	}
	switch r.field2 {
	case uintptr(kPtrLike):
		return *(*T)(unsafe.Pointer(&r.data))

	case uintptr(kIface):
		return *(*T)(unsafe.Pointer(&iface{
			typ:  r.field1,
			data: r.data,
		}))

	case uintptr(kString):
		len := r.field1
		str := unsafe.String((*byte)(r.data), len)
		return *(*T)(unsafe.Pointer(&str))

	case uintptr(kAlloc):
		return *(*T)(r.data)

	case uintptr(kDetached):
		return *(*T)(unsafe.Pointer(&r.data))

	case 0:
		if r.data == nil && r.field1 == 0 {
			var zero T
			return zero
		}
	}
	panic("tipe: Result[T] is not ok")
}

// UnwrapErr returns the error contained in the Result[T].
func (r Result[T]) UnwrapErr() error {
	if !r.IsErr() {
		panic("tipe: Result[T] is not err")
	}
	return *(*error)(unsafe.Pointer(&iface{
		typ:  r.field2,
		data: r.data,
	}))
}

// UnwrapOr returns the value contained in the Result[T].
// If the Result[T] is an error, it returns the given value or.
func (r Result[T]) UnwrapOr(or T) T {
	if r.IsOk() {
		return r.Unwrap()
	}
	return or
}

// Tuple returns the value and error contained in the Result[T].
func (r Result[T]) Tuple() (T, error) {
	if r.IsOk() {
		return r.Unwrap(), nil
	}
	var zero T
	return zero, r.UnwrapErr()
}

// TupleBool returns the value and a boolean indicating whether the Result[T] is ok.
func (r Result[T]) TupleBool() (T, bool) {
	if r.IsOk() {
		return r.Unwrap(), true
	}
	var zero T
	return zero, false
}

// String returns a string representation of the Result[T].
func (r Result[T]) String() string {
	if r.IsOk() {
		return fmt.Sprintf("Ok(%v)", r.Unwrap())
	} else {
		return fmt.Sprintf("Err(%v)", r.UnwrapErr())
	}
}

// Fill returns a new Result[T] that contains the given value.
func (r Result[T]) Fill(v T) Result[T] {
	return Ok(v)
}

// FillErr returns a new Result[T] that contains the given error.
func (r Result[T]) FillErr(err error) Result[T] {
	return Err[T](err)
}

// FillTuple returns a new Result[T] that contains the given value or error.
func (r Result[T]) FillTuple(v T, err error) Result[T] {
	return MakeResult(v, err)
}

// Zero returns a new Result[T] that contains the zero value of T.
func (r Result[T]) Zero() Result[T] {
	var zero T
	return Ok(zero)
}

//go:nosplit
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

const ptrSize = unsafe.Sizeof(uintptr(0))

// Ok returns a Result[T] that contains the given value.
func Ok[T any](value T) Result[T] {
	var kind reflect.Kind
	var zero T
	if any(zero) == nil {
		kind = reflect.Interface
	} else {
		kind = reflect.TypeOf(zero).Kind()
	}
	switch kind {
	case reflect.Ptr, reflect.UnsafePointer, reflect.Chan, reflect.Func, reflect.Map:
		return Result[T]{
			data:   *(*unsafe.Pointer)(noescape(unsafe.Pointer(&value))),
			field2: uintptr(kPtrLike),
		}
	case reflect.Interface:
		iface := *(*iface)(noescape(unsafe.Pointer(&value)))
		return Result[T]{
			data:   iface.data,
			field1: iface.typ,
			field2: uintptr(kIface),
		}
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		goto SIMPLE
	case reflect.String:
		str := *(*string)(noescape(unsafe.Pointer(&value)))
		return Result[T]{
			data:   unsafe.Pointer(unsafe.StringData(str)),
			field1: uintptr(len(str)),
			field2: uintptr(kString),
		}
	case reflect.Slice:
		slice := *(*[]byte)(noescape(unsafe.Pointer(&value)))
		return Result[T]{
			data:   unsafe.Pointer(unsafe.SliceData(slice)),
			field1: uintptr(len(slice)) | sliceFlag,
			field2: uintptr(cap(slice)),
		}
	default:
		var TSize = unsafe.Sizeof(zero)
		if TSize < ptrSize {
			goto SIMPLE
		} else if TSize <= 2*ptrSize {
			pBytes := ptrBytes[T]()
			switch pBytes {
			case 0:
				goto SIMPLE
			case ptrSize:
				r := Result[T]{
					field2: uintptr(kDetached),
				}
				*(*T)(unsafe.Pointer(&r.data)) = value
				return r
			}
		}

		escaped := new(T)
		*escaped = value
		return Result[T]{
			data:   unsafe.Pointer(escaped),
			field2: uintptr(kAlloc),
		}
	}
SIMPLE:
	r := Result[T]{
		data: kSimple,
	}
	*(*T)(unsafe.Pointer(&r.field1)) = value
	return r
}

// Err returns a Result[T] that contains the given error.
func Err[T any](err error) Result[T] {
	if err == nil {
		var zero T
		return Ok[T](zero)
	}
	var internal = *(*iface)(noescape(unsafe.Pointer(&err)))
	return Result[T]{
		data:   internal.data,
		field2: internal.typ,
	}
}

// MakeResult returns a Result[T] that contains the given value or error.
func MakeResult[T any](value T, err error) Result[T] {
	if err == nil {
		return Ok[T](value)
	}
	return Err[T](err)
}
