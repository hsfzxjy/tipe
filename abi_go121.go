//go:build go1.21
// +build go1.21

package tipe

import "unsafe"

func ptrBytes[T any]() uintptr {
	var zero T
	var a = any(zero)
	return (*struct {
		typ *struct {
			_        uintptr
			PtrBytes uintptr
		}
		_ uintptr
	})(unsafe.Pointer(&a)).typ.PtrBytes
}
