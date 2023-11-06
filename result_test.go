package tipe_test

import (
	"crypto/rand"
	"errors"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/hsfzxjy/tipe"

	"github.com/stretchr/testify/assert"
)

func testType[T any](alloc bool, t *testing.T, values ...T) {
	var zero T
	assert.Equalf(t, unsafe.Sizeof(tipe.Ok(zero)), 3*unsafe.Sizeof(uintptr(0)), "%T", tipe.Ok(zero))
	assert.Equal(t, alloc, tipe.Ok(zero).Alloc())
	for _, value := range append(values, zero) {
		ok := tipe.Ok(value)
		assert.True(t, ok.IsOk())
		assert.False(t, ok.IsErr())
		assert.Equal(t, value, ok.Unwrap())
		switch x := any(value).(type) {
		case []int:
			assert.Equal(t, cap(x), cap(any(ok.Unwrap()).([]int)))
		}
	}
	ok := tipe.Err[T](nil)
	assert.True(t, ok.IsOk())
	assert.False(t, ok.IsErr())
	assert.Equal(t, zero, ok.Unwrap())

	ok = tipe.Result[T]{}
	assert.True(t, ok.IsOk())
	assert.False(t, ok.IsErr())
	assert.Equal(t, zero, ok.Unwrap())

	e := errors.New("foo error")
	err := tipe.Err[T](e)
	assert.False(t, err.IsOk())
	assert.True(t, err.IsErr())
	assert.Equal(t, e, err.UnwrapErr())
}

func testTypeAndPtr[T any](alloc bool, t *testing.T, values ...T) {
	testType[T](alloc, t, values...)
	var ptrValues []*T
	for _, v := range values {
		ptrValues = append(ptrValues, &v)
	}
	testType[*T](false, t, ptrValues...)
}

type F interface{ f() }

type X struct{ int }

func (X) f() {}

func testByteArray[T any](t *testing.T) {
	var data = make([]byte, 32)
	rand.Read(data)
	var zero T
	testTypeAndPtr[T](unsafe.Sizeof(zero) > 2*unsafe.Sizeof(uintptr(0)), t, *(*T)(unsafe.Pointer(&data[0])))
}

func TestResult(t *testing.T) {
	type MyInt int
	type BigType struct{ data [3]uintptr }
	testTypeAndPtr[bool](false, t, true, false)
	testTypeAndPtr[int](false, t, 42, 0, -1, math.MaxInt, math.MinInt)
	testTypeAndPtr[MyInt](false, t, 42, 0, -1, math.MaxInt, math.MinInt)
	testTypeAndPtr[int8](false, t, 42, 0, -1, math.MaxInt8, math.MinInt8)
	testTypeAndPtr[int16](false, t, 42, 0, -1, math.MaxInt16, math.MinInt16)
	testTypeAndPtr[int32](false, t, 42, 0, -1, math.MaxInt32, math.MinInt32)
	testTypeAndPtr[int64](false, t, 42, 0, -1, math.MaxInt64, math.MinInt64)
	testTypeAndPtr[uint](false, t, 42, 0, 1, math.MaxUint, math.MaxUint)
	testTypeAndPtr[uint8](false, t, 42, 0, 1, math.MaxUint8, math.MaxUint8)
	testTypeAndPtr[uint16](false, t, 42, 0, 1, math.MaxUint16, math.MaxUint16)
	testTypeAndPtr[uint32](false, t, 42, 0, 1, math.MaxUint32, math.MaxUint32)
	testTypeAndPtr[uint64](false, t, 42, 0, 1, math.MaxUint64, math.MaxUint64)
	testTypeAndPtr[float32](false, t, 42, 0, 1, math.MaxFloat32, math.SmallestNonzeroFloat32)
	testTypeAndPtr[float64](false, t, 42, 0, 1, math.MaxFloat64, math.SmallestNonzeroFloat64)
	testTypeAndPtr[complex64](false, t, 42, 0, 1, complex(math.MaxFloat32, math.MaxFloat32), complex(math.SmallestNonzeroFloat32, math.SmallestNonzeroFloat32))
	testTypeAndPtr[complex128](false, t, 42, 0, 1, complex(math.MaxFloat64, math.MaxFloat64), complex(math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64))
	testTypeAndPtr[uintptr](false, t, 42, 0, 1, ^uintptr(0))
	testTypeAndPtr[string](false, t, "", "foo")
	testTypeAndPtr[[]int](false, t, nil, []int{1}, make([]int, 0, 45))
	testTypeAndPtr[any](false, t, nil, 1, "foo", false, errors.New("bar"))
	testTypeAndPtr[F](false, t, nil, X{})
	testTypeAndPtr[BigType](true, t, BigType{data: [3]uintptr{43, 42, 41}})
	testTypeAndPtr(false, t, struct {
		int
		byte
		bool
	}{0, 1, false})
	testTypeAndPtr(false, t, struct {
		ptr *struct{ int }
		byte
		bool
	}{&struct{ int }{45}, 1, false})
	testTypeAndPtr(true, t, struct {
		byte
		ptr *struct{ int }
		bool
	}{1, &struct{ int }{45}, false})
	testTypeAndPtr(true, t, struct {
		ptr, ptr2 *struct{ int }
	}{&struct{ int }{45}, nil})
	testByteArray[[0]byte](t)
	testByteArray[[1]byte](t)
	testByteArray[[2]byte](t)
	testByteArray[[3]byte](t)
	testByteArray[[4]byte](t)
	testByteArray[[5]byte](t)
	testByteArray[[6]byte](t)
	testByteArray[[7]byte](t)
	testByteArray[[8]byte](t)
	testByteArray[[9]byte](t)
	testByteArray[[10]byte](t)
	testByteArray[[11]byte](t)
	testByteArray[[12]byte](t)
	testByteArray[[13]byte](t)
	testByteArray[[14]byte](t)
	testByteArray[[15]byte](t)
	testByteArray[[16]byte](t)
}

func TestProperty(t *testing.T) {
	x := tipe.Ok[int](8)
	assert.False(t, reflect.TypeOf(x).AssignableTo(reflect.TypeOf(tipe.Ok[uint](8))))
	assert.False(t, reflect.TypeOf(x).Comparable())
}
