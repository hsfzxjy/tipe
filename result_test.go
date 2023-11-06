package tipe_test

import (
	"errors"
	"math"
	"testing"
	"tipe"

	"github.com/stretchr/testify/assert"
)

func testType[T any](t *testing.T, values ...T) {
	var zero T
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
	e := errors.New("foo error")
	err := tipe.Err[T](e)
	assert.False(t, err.IsOk())
	assert.True(t, err.IsErr())
	assert.Equal(t, e, err.UnwrapErr())
}

type F interface{ f() }

type X struct{ int }

func (X) f() {}

func TestResult(t *testing.T) {
	testType[bool](t, true, false)
	testType[int](t, 42, 0, -1, math.MaxInt, math.MinInt)
	testType[int8](t, 42, 0, -1, math.MaxInt8, math.MinInt8)
	testType[int16](t, 42, 0, -1, math.MaxInt16, math.MinInt16)
	testType[int32](t, 42, 0, -1, math.MaxInt32, math.MinInt32)
	testType[int64](t, 42, 0, -1, math.MaxInt64, math.MinInt64)
	testType[uint](t, 42, 0, 1, math.MaxUint, math.MaxUint)
	testType[uint8](t, 42, 0, 1, math.MaxUint8, math.MaxUint8)
	testType[uint16](t, 42, 0, 1, math.MaxUint16, math.MaxUint16)
	testType[uint32](t, 42, 0, 1, math.MaxUint32, math.MaxUint32)
	testType[uint64](t, 42, 0, 1, math.MaxUint64, math.MaxUint64)
	testType[float32](t, 42, 0, 1, math.MaxFloat32, math.SmallestNonzeroFloat32)
	testType[float64](t, 42, 0, 1, math.MaxFloat64, math.SmallestNonzeroFloat64)
	testType[complex64](t, 42, 0, 1, complex(math.MaxFloat32, math.MaxFloat32), complex(math.SmallestNonzeroFloat32, math.SmallestNonzeroFloat32))
	testType[complex128](t, 42, 0, 1, complex(math.MaxFloat64, math.MaxFloat64), complex(math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64))
	testType[uintptr](t, 42, 0, 1, ^uintptr(0))
	testType[string](t, "", "foo")
	testType[[]int](t, nil, []int{1}, make([]int, 0, 45))
	testType[any](t, nil, 1, "foo", false, errors.New("bar"))
	testType[F](t, nil, X{})
}
